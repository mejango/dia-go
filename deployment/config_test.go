package deployment

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mejango/blossm-toolbox-go/cli/blossm"
)

func example(t *testing.T, name string) cli.Config {
	t.Helper()
	c, err := ReadConfig(filepath.Join("..", "deploy", "examples", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestApplicationExamplesPreserveArtifactSettings(t *testing.T) {
	for _, name := range []string{"gcp-postgres", "aws-mongo", "gcp-document-postgres", "aws-document-mongo"} {
		t.Run(name, func(t *testing.T) {
			c := example(t, name)
			env, secretNames := preparedSettings(t, c)
			for key, value := range c.RuntimeEnv() {
				if env[key] != value {
					t.Fatalf("%s=%q want %q", key, env[key], value)
				}
			}
			for _, key := range []string{"DIA_HMAC_KEY", "DIA_ENCRYPTION_KEY", "JWT_PRIVATE_KEY"} {
				if !secretNames[key] {
					t.Fatal("missing secret reference", key)
				}
				if _, ok := env[key]; ok {
					t.Fatal("secret rendered plaintext", key)
				}
			}
			if c.Application.Package == "./cmd/document" && (env["IDENTITY_DATABASE_NAMESPACE"] == env["DATABASE_NAMESPACE"] || env["PUBLICATION_MODE"] != c.Provider || env["SERVICE"] != "documents" || env["INTERNAL_SUBJECT"] != "document-peer") {
				t.Fatal("document settings lost", env)
			}
			if c.Application.Package == "./cmd/dia" && (env["PUBLISHER"] != c.Provider || env["LOCAL_MODE"] != "false") {
				t.Fatal("Dia settings lost", env)
			}
		})
	}
}

func TestApplicationRejectsBeforePreparingArtifacts(t *testing.T) {
	cases := map[string]func(*cli.Config){
		"hmac":                func(c *cli.Config) { delete(c.Secrets, "DIA_HMAC_KEY") },
		"encryption":          func(c *cli.Config) { delete(c.Secrets, "DIA_ENCRYPTION_KEY") },
		"plaintext secret":    func(c *cli.Config) { delete(c.Secrets, "DIA_HMAC_KEY"); c.Env["DIA_HMAC_KEY"] = "not-a-reference" },
		"signer":              func(c *cli.Config) { c.Signing = nil },
		"issuer":              func(c *cli.Config) { c.Signing.Issuer = "foreign" },
		"audience":            func(c *cli.Config) { c.Signing.Audience = "foreign" },
		"network":             func(c *cli.Config) { delete(c.Env, "NETWORK") },
		"database":            func(c *cli.Config) { c.Database = nil },
		"publication":         func(c *cli.Config) { c.Publication = nil },
		"profile":             func(c *cli.Config) { c.Application.Package = "./cmd/other" },
		"namespace collision": func(c *cli.Config) { c.Env["IDENTITY_DATABASE_NAMESPACE"] = c.Database.Namespace },
		"namespace syntax":    func(c *cli.Config) { c.Env["IDENTITY_DATABASE_NAMESPACE"] = "bad-namespace" },
		"service":             func(c *cli.Config) { delete(c.Env, "SERVICE") },
		"internal peer":       func(c *cli.Config) { delete(c.Env, "INTERNAL_SUBJECT") },
		"subject mismatch":    func(c *cli.Config) { c.Signing.Subject = "other" },
		"delivery":            func(c *cli.Config) { c.Delivery = nil },
		"mode":                func(c *cli.Config) { c.Env["PUBLICATION_MODE"] = "local" },
		"unvalidated peers":   func(c *cli.Config) { c.Env["OUTBOUND_URL"] = "https://peer.invalid" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			c := example(t, "gcp-document-postgres")
			change(&c)
			p, err := Prepare(c)
			if err == nil || len(p.Artifacts) != 0 {
				t.Fatal("invalid configuration produced plan", p, err)
			}
		})
	}
	for _, key := range []string{"LOCAL_MODE", "PUBLISHER"} {
		t.Run(key, func(t *testing.T) {
			c := example(t, "gcp-postgres")
			delete(c.Env, key)
			if _, err := Prepare(c); err == nil {
				t.Fatal("missing app setting accepted", key)
			}
		})
	}
}

func TestOptionalApplicationFeaturesRequireSecretReferences(t *testing.T) {
	for _, feature := range []struct{ env, secret string }{{"EVM_RPC_URLS", "WALLET_ENCRYPTION_KEY"}, {"TWILIO_ACCOUNT_SID", "TWILIO_AUTH_TOKEN"}, {"UPDATES_PUBLISH_URL", "UPDATES_PUBLISH_TOKEN"}} {
		t.Run(feature.env, func(t *testing.T) {
			c := example(t, "gcp-postgres")
			c.Env[feature.env] = "configured"
			c.Env["TWILIO_FROM"] = "configured"
			if err := Validate(c); err == nil {
				t.Fatal("missing optional secret accepted")
			}
			c.Secrets[feature.secret] = cli.Secret{Reference: "feature-secret", Version: "1"}
			if err := Validate(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSecretReferencedFeatureStillRequiresItsKey(t *testing.T) {
	c := example(t, "gcp-postgres")
	c.Secrets["EVM_RPC_URLS"] = cli.Secret{Reference: "private-rpc-endpoints", Version: "1"}
	if err := Validate(c); err == nil {
		t.Fatal("secret-referenced wallet configuration bypassed required key")
	}
}

func preparedSettings(t *testing.T, c cli.Config) (map[string]string, map[string]bool) {
	t.Helper()
	p, err := Prepare(c)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "prepared-only" {
		t.Fatal(p.Status)
	}
	file, entry, envField := "cloud-run.json", "command", "env"
	if c.Provider == "aws" {
		file, entry, envField = "ecs-task.json", "entryPoint", "environment"
	}
	var artifact map[string]any
	if err = json.Unmarshal(p.Artifacts[file], &artifact); err != nil {
		t.Fatal(err)
	}
	var container map[string]any
	if c.Provider == "gcp" {
		container = artifact["spec"].(map[string]any)["template"].(map[string]any)["spec"].(map[string]any)["containers"].([]any)[0].(map[string]any)
	} else {
		container = artifact["containerDefinitions"].([]any)[0].(map[string]any)
	}
	if container[entry].([]any)[0] != c.Application.Entrypoint {
		t.Fatal("wrong entrypoint", container[entry])
	}
	env := map[string]string{}
	secretNames := map[string]bool{}
	for _, v := range container[envField].([]any) {
		field := v.(map[string]any)
		key := field["name"].(string)
		if value, ok := field["value"].(string); ok {
			env[key] = value
		} else {
			secretNames[key] = true
		}
	}
	if c.Provider == "aws" {
		for _, v := range container["secrets"].([]any) {
			secretNames[v.(map[string]any)["name"].(string)] = true
		}
	}

	return env, secretNames
}

func TestPairedApplicationsShareIdentityAndIsolatePublications(t *testing.T) {
	for _, pair := range []struct{ identity, document string }{{"gcp-postgres", "gcp-document-postgres"}, {"aws-mongo", "aws-document-mongo"}} {
		t.Run(pair.identity, func(t *testing.T) {
			identity, document := example(t, pair.identity), example(t, pair.document)
			idEnv, _ := preparedSettings(t, identity)
			docEnv, _ := preparedSettings(t, document)
			if identity.Database.Driver != document.Database.Driver || identity.Database.Secret != document.Database.Secret {
				t.Fatal("applications do not share database endpoint reference")
			}
			if idEnv["DATABASE_NAMESPACE"] != "dia_identity" || docEnv["IDENTITY_DATABASE_NAMESPACE"] != idEnv["DATABASE_NAMESPACE"] {
				t.Fatal("document cannot address the identity prefix", idEnv, docEnv)
			}
			if docEnv["DATABASE_NAMESPACE"] != "documents" || docEnv["DATABASE_NAMESPACE"] == idEnv["DATABASE_NAMESPACE"] {
				t.Fatal("document data is not isolated")
			}
			if identity.Database.Driver == "mongo" {
				if identity.Database.Name != "diaapp" || document.Database.Name != identity.Database.Name || idEnv["MONGODB_DATABASE"] != "diaapp" || docEnv["MONGODB_DATABASE"] != idEnv["MONGODB_DATABASE"] {
					t.Fatal("identity lookup addresses a different physical Mongo database", idEnv, docEnv)
				}
			}
			if idEnv["NETWORK"] != docEnv["NETWORK"] || identity.Signing.Issuer != document.Signing.Issuer || identity.Signing.Audience != document.Signing.Audience {
				t.Fatal("paired signing trust differs")
			}
			for _, key := range []string{"JWT_PRIVATE_KEY", "DIA_HMAC_KEY", "DIA_ENCRYPTION_KEY"} {
				if identity.Secrets[key] != document.Secrets[key] {
					t.Fatal("paired trust/crypto reference differs", key)
				}
			}
			if idEnv["PUBLICATION_TOPIC"] == docEnv["PUBLICATION_TOPIC"] {
				t.Fatal("unrelated event streams share publication topic")
			}
			if identity.Provider == "gcp" {
				if idEnv["PUBLICATION_SUBSCRIPTION"] == docEnv["PUBLICATION_SUBSCRIPTION"] {
					t.Fatal("unrelated consumers compete on one subscription")
				}
			} else {
				if idEnv["PUBLICATION_QUEUE_URL"] == docEnv["PUBLICATION_QUEUE_URL"] || identity.Publication.QueueARN == document.Publication.QueueARN {
					t.Fatal("unrelated consumers compete on one queue")
				}
				if docEnv["DELIVERY_QUEUE_URL"] == docEnv["PUBLICATION_QUEUE_URL"] || docEnv["DELIVERY_QUEUE_URL"] == idEnv["PUBLICATION_QUEUE_URL"] {
					t.Fatal("delivery queue overlaps publication consumers")
				}
			}
		})
	}
}
