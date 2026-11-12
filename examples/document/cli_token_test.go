package serviceapp_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mejango/blossm-toolbox-go/cli/blossm"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	eventstore "github.com/mejango/blossm-toolbox-go/procedures/event_stores/generic"
	"github.com/mejango/blossm-toolbox-go/procedures/gateways"
	projection "github.com/mejango/blossm-toolbox-go/procedures/projection"
	runtimeauth "github.com/mejango/blossm-toolbox-go/procedures/server/auth"
	serviceapp "github.com/mejango/dia-go/examples/document"
	coremodel "github.com/mejango/dia-go/models/core"
	dia "github.com/mejango/dia-go/services/identity"
)

func internalConfig(t *testing.T, c cli.Config) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "config.json")
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}
func internalSigningKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("JWT_SIGNER", "local")
	t.Setenv("JWT_KMS_KEY", "")
	t.Setenv("JWT_PUBLIC_KEY", "")
	t.Setenv("JWT_PRIVATE_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: raw})))
	return key
}

func internalExample() cli.Config {
	c := cli.Example("gcp", "postgres")
	c.Signing = &cli.Signing{Provider: "local", Issuer: "dia.example.test", Audience: "example.test", Subject: "configured-worker"}
	c.Secrets["JWT_PRIVATE_KEY"] = cli.Secret{Reference: "sample-jwt-key", Version: "1"}
	return c
}

type tokenRoots struct {
	eventlib.Reader
	root eventlib.Root
}

func (r tokenRoots) Roots(ctx context.Context, yield func(eventlib.Root) error) error {
	return yield(r.root)
}
func (r tokenRoots) Count(context.Context, eventlib.Root) (int64, error) { return 1, nil }

type tokenBackend struct {
	eventlib.Backend
	root eventlib.Root
}

func (b *tokenBackend) View(ctx context.Context, fn func(eventlib.Reader) error) error {
	return fn(tokenRoots{root: b.root})
}

type tokenReplay struct{ requests []projection.Request }

func (p *tokenReplay) Enqueue(_ context.Context, r projection.Request) error {
	p.requests = append(p.requests, r)
	return nil
}

func TestInternalTokenDrivesServiceAppReplay(t *testing.T) {
	internalSigningKey(t)
	c := internalExample()
	ctx := context.Background()
	jwt, close, err := runtimeauth.Open[coremodel.SessionContext](ctx, c.Signing.Issuer, c.Signing.Audience)
	if err != nil {
		t.Fatal(err)
	}
	defer close()
	verifier, err := transport.ServiceJWTVerifier(*jwt, c.Signing.Subject)
	if err != nil {
		t.Fatal(err)
	}
	root := eventlib.Root{Network: c.Signing.Audience, Service: "documents", Domain: serviceapp.Domain, ID: "doc"}
	backend := &tokenBackend{root: root}
	identityBackend := &tokenBackend{root: root}
	producer := &tokenReplay{}
	identity := &dia.Service{Network: c.Signing.Audience, JWT: jwt, Store: &eventstore.Store{Backend: identityBackend, Local: true}}
	app, err := serviceapp.New(ctx, serviceapp.Config{Store: &eventstore.Store{Backend: backend, Local: true}, Identity: identity, Service: "documents", Internal: verifier, Replay: producer})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(app.Handler)
	defer server.Close()
	input := filepath.Join(t.TempDir(), "replay.json")
	if err = os.WriteFile(input, []byte(`{"runId":"cli-run","limit":10}`), 0600); err != nil {
		t.Fatal(err)
	}
	config := internalConfig(t, c)
	for _, kind := range []string{"valid", "wrong-subject", "wrong-audience"} {
		t.Run(kind, func(t *testing.T) {
			minted := c
			signing := *c.Signing
			minted.Signing = &signing
			if kind == "wrong-subject" {
				minted.Signing.Subject = "other-worker"
			}
			if kind == "wrong-audience" {
				minted.Signing.Audience = "other.test"
			}
			var token, stderr bytes.Buffer
			if err := cli.Run(ctx, []string{"internal-token", "--config", internalConfig(t, minted), "--ttl", "2m", "--execute"}, &token, &stderr); err != nil {
				t.Fatal(err)
			}
			value := strings.TrimSpace(token.String())
			if kind == "valid" {
				claims, err := jwt.Verify(value, "service")
				if err != nil || claims.Subject != c.Signing.Subject || len(claims.Audience) != 1 || claims.Audience[0] != c.Signing.Audience || claims.Expires-claims.NotBefore != 120 || claims.ID == "" {
					t.Fatal(claims, err)
				}
			}
			before := len(producer.requests)
			var output bytes.Buffer
			e := cli.Execution{HTTP: server.Client(), Getenv: func(string) string { return value }}
			err := cli.RunWith(ctx, []string{"replay", "--config", config, "--url", server.URL + "/projections/documents/replay", "--input", input, "--idempotency-key", "cli-replay", "--execute"}, &output, &stderr, e)
			if kind == "valid" {
				var ack transport.ProjectionReplayOutput
				if err != nil || json.Unmarshal(output.Bytes(), &ack) != nil || ack.RunID != "cli-run" || ack.Enqueued != 1 || !ack.Complete || len(producer.requests) != before+1 || producer.requests[before].RunID != "cli-run" || producer.requests[before].Root != root {
					t.Fatal("minted replay failed", ack, err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), "401") || len(producer.requests) != before {
					t.Fatal("foreign authority accepted", err)
				}
			}
		})
	}
}
