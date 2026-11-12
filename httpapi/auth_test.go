package httpapi

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	canonicaljson "github.com/mejango/blossm-toolbox-go/lib/canonical_json"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/gateways"
	coremodel "github.com/mejango/dia-go/models/core"
	systemmodel "github.com/mejango/dia-go/models/system"
	dia "github.com/mejango/dia-go/services/identity"
)

type authState struct {
	dia.Store
	rows map[eventlib.Root]eventlib.Aggregate
}

func (s *authState) Aggregate(ctx context.Context, r eventlib.Root) (eventlib.Aggregate, error) {
	if err := ctx.Err(); err != nil {
		return eventlib.Aggregate{}, err
	}
	v, ok := s.rows[r]
	if !ok {
		return v, eventlib.ErrNotFound
	}
	return v, nil
}
func (s *authState) put(r eventlib.Root, v any) {
	body, _ := canonicaljson.Encode(v)
	s.rows[r] = eventlib.Aggregate{Root: r, Version: 1, State: body}
}
func gatewayJWT(t *testing.T) (dia.JWT, dia.Claims) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().Unix()
	return dia.JWT{Signer: dia.LocalSigner{Key: key}, VerifyKey: &key.PublicKey, Issuer: "issuer", Audience: "n"}, dia.Claims{Issuer: "issuer", Audience: []string{"n"}, Subject: "p", ID: "jti", Kind: "access", NotBefore: now - 1, Expires: now + 600, Context: coremodel.SessionContext{Network: "n", Session: eventlib.Root{Network: "n", Service: "core", Domain: "session", ID: "s"}, Principal: eventlib.Root{Network: "n", Service: "core", Domain: "principal", ID: "p"}}}
}
func TestGatewayDiaSessionAndConnectionAuthority(t *testing.T) {
	jwt, claims := gatewayJWT(t)
	state := &authState{rows: map[eventlib.Root]eventlib.Aggregate{}}
	session := coremodel.SessionState{Expires: claims.Expires, Principal: claims.Context.Principal}
	state.put(claims.Context.Session, session)
	subject := eventlib.Root{Network: "n", Service: "core", Domain: "account", ID: "account"}
	state.put(claims.Context.Principal, coremodel.PrincipalState{Roles: []coremodel.RoleGrant{{ID: "AccountAdmin", Subject: subject}}})
	service := &dia.Service{JWT: &jwt, Store: state, Network: "n"}
	verify, err := DiaVerifier(service, "access")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Sign(context.Background(), claims)
	if err != nil {
		t.Fatal(err)
	}
	id, err := verify(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if err = DiaPermission(service, subject, "core:account:update")(context.Background(), id, transport.GatewayRequest{}); err != nil {
		t.Fatal(err)
	}
	foreign := subject
	foreign.ID = "other"
	if err = DiaPermission(service, foreign, "core:account:update")(context.Background(), id, transport.GatewayRequest{}); !errors.Is(err, dia.ErrForbidden) {
		t.Fatal("foreign subject allowed", err)
	}
	session.Terminated = "terminated"
	state.put(claims.Context.Session, session)
	if _, err = verify(context.Background(), token); !errors.Is(err, dia.ErrUnauthorized) {
		t.Fatal("revoked session accepted", err)
	}
	session.Terminated = ""
	state.put(claims.Context.Session, session)
	wrong := jwt
	wrong.Audience = "another-service"
	wrongService := *service
	wrongService.JWT = &wrong
	wrongVerify, _ := DiaVerifier(&wrongService, "access")
	if _, err = wrongVerify(context.Background(), token); !errors.Is(err, dia.ErrUnauthorized) {
		t.Fatal("cross-service token accepted", err)
	}
	key := eventlib.Root{Network: "n", Service: "system", Domain: "key", ID: "key"}
	claims.Kind = "connection"
	claims.Context.Subject = key
	state.put(key, systemmodel.KeyState{Principal: claims.Context.Principal})
	state.put(claims.Context.Principal, coremodel.PrincipalState{Subject: key})
	token, err = jwt.Sign(context.Background(), claims)
	if err != nil {
		t.Fatal(err)
	}
	verify, _ = DiaVerifier(service, "connection")
	if _, err = verify(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	state.put(key, systemmodel.KeyState{Principal: claims.Context.Principal, Revoked: "revoked"})
	if _, err = verify(context.Background(), token); !errors.Is(err, dia.ErrUnauthorized) {
		t.Fatal("revoked key accepted", err)
	}
}
