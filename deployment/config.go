// Package deployment adds Dia application policy to the public Toolbox CLI.
// It validates references only: secret contents and live resources remain runtime
// responsibilities. No function in this package executes deployment commands.
package deployment

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mejango/blossm-toolbox-go/cli/blossm"
)

// Validate checks generic configuration first, then the selected application's
// required settings. Package selection is explicit; image names are not policy.
func Validate(c cli.Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Application.Package != "./cmd/dia" && c.Application.Package != "./cmd/document" {
		return errors.New("Dia deployment requires application.package ./cmd/dia or ./cmd/document")
	}
	if c.Database == nil || c.Publication == nil {
		return errors.New("Dia applications require database and publication configuration")
	}
	env := c.RuntimeEnv()
	network := env["NETWORK"]
	if strings.TrimSpace(network) == "" || strings.ContainsAny(network, " \t\r\n\x00") {
		return errors.New("NETWORK must be an explicit nonempty network identifier")
	}
	if c.Signing == nil {
		return errors.New("Dia applications require signing configuration")
	}
	// Both native applications construct exactly this trust domain. Merely
	// supplying a syntactically valid but different generic signer is insufficient.
	if c.Signing.Issuer != "dia."+network || c.Signing.Audience != network {
		return errors.New("Dia signing issuer must be dia.<NETWORK> and audience must equal NETWORK")
	}
	for _, key := range []string{"DIA_HMAC_KEY", "DIA_ENCRYPTION_KEY"} {
		if err := requireSecret(c, key); err != nil {
			return err
		}
	}
	if c.Application.Package == "./cmd/dia" {
		if env["LOCAL_MODE"] != "false" || env["PUBLISHER"] != c.Provider {
			return errors.New("Dia deployment requires LOCAL_MODE=false and PUBLISHER matching provider")
		}
	} else {
		for _, key := range []string{"SERVICE", "INTERNAL_SUBJECT", "IDENTITY_DATABASE_NAMESPACE"} {
			if strings.TrimSpace(env[key]) == "" {
				return fmt.Errorf("document requires %s", key)
			}
		}
		if env["PUBLICATION_MODE"] != c.Provider {
			return errors.New("document PUBLICATION_MODE must match provider")
		}
		if c.Delivery == nil {
			return errors.New("document deployment requires delivery configuration")
		}
		if env["IDENTITY_DATABASE_NAMESPACE"] == c.Database.Namespace {
			return errors.New("identity and document namespaces must differ")
		}
		// Reuse the framework's namespace validator instead of introducing a second
		// regexp which could diverge from database deployment rules.
		identityConfig := c
		identityDatabase := *c.Database
		identityDatabase.Namespace = env["IDENTITY_DATABASE_NAMESPACE"]
		identityConfig.Database = &identityDatabase
		if err := identityConfig.Validate(); err != nil {
			return fmt.Errorf("identity database namespace: %w", err)
		}
		if c.Signing.Subject != "" && c.Signing.Subject != env["INTERNAL_SUBJECT"] {
			return errors.New("configured signing subject must match document INTERNAL_SUBJECT")
		}
		for _, key := range []string{"OUTBOUND_URL", "OUTBOUND_SERVICE", "DELEGATED_SUBJECT"} {
			if env[key] != "" && c.Peers == nil {
				return errors.New("document peer settings require validated peers configuration")
			}
		}
	}
	// Optional runtime features must not produce artifacts missing their secrets.
	if configured(c, "EVM_RPC_URLS") {
		if err := requireSecret(c, "WALLET_ENCRYPTION_KEY"); err != nil {
			return err
		}
	}
	if configured(c, "TWILIO_ACCOUNT_SID") {
		if err := requireSecret(c, "TWILIO_AUTH_TOKEN"); err != nil {
			return err
		}
		if !configured(c, "TWILIO_FROM") {
			return errors.New("TWILIO_FROM required with TWILIO_ACCOUNT_SID")
		}
	}
	if configured(c, "UPDATES_PUBLISH_URL") {
		if err := requireSecret(c, "UPDATES_PUBLISH_TOKEN"); err != nil {
			return err
		}
	}
	return nil
}

func configured(c cli.Config, key string) bool {
	_, secret := c.Secrets[key]
	return strings.TrimSpace(c.Env[key]) != "" || secret
}

func requireSecret(c cli.Config, key string) error {
	if _, ok := c.Env[key]; ok {
		return fmt.Errorf("%s must be a secret reference, not plaintext environment", key)
	}
	if _, ok := c.Secrets[key]; !ok {
		return fmt.Errorf("required application secret reference %s is missing", key)
	}
	return nil
}

// ReadConfig uses the public strict config decoder, then applies Dia policy.
func ReadConfig(path string) (cli.Config, error) {
	c, err := cli.ReadConfig(path)
	if err != nil {
		return cli.Config{}, err
	}
	if err = Validate(c); err != nil {
		return cli.Config{}, err
	}
	return c, nil
}

// Prepare cannot return artifacts until application validation succeeds. The
// framework remains the sole renderer; the caller may explicitly WritePlan.
func Prepare(c cli.Config) (cli.Plan, error) {
	if err := Validate(c); err != nil {
		return cli.Plan{}, err
	}
	return cli.Prepare(c)
}
