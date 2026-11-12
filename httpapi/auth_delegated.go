package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"

	canonicaljson "github.com/mejango/blossm-toolbox-go/lib/canonical_json"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/command"
	"github.com/mejango/blossm-toolbox-go/procedures/gateways"
	dia "github.com/mejango/dia-go/services/identity"
)

// DelegatedCommandAuthorizer is installed on a commandrpc.CommandClient used inside an
// authenticated Main. parent must be the transport.Framework-authenticated request, never
// raw client JSON. jwt is the issuing peer's signer with the target service's
// exact audience; peerSubject identifies that peer, not the delegated user.
// Delegation keeps user context and never outlives its parent credential.
func DelegatedCommandAuthorizer(jwt dia.JWT, identity *dia.Service, peerSubject string) (func(context.Context, command.Request) (http.Header, error), error) {
	if identity == nil || jwt.Signer == nil || jwt.VerifyKey == nil || jwt.Issuer == "" || jwt.Audience == "" || peerSubject == "" {
		return nil, errors.New("delegation signer, target audience, peer and Dia state validator required")
	}
	return func(ctx context.Context, parent command.Request) (http.Header, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var c dia.Claims
		if canonicaljson.Decode(parent.Claims, &c) != nil || (c.Kind != "access" && c.Kind != "service") {
			return nil, dia.ErrUnauthorized
		}
		// The procedure's authenticated context and its audit claims must agree.
		// Prevent application code accidentally substituting a body-selected actor.
		claimContext, err := canonicaljson.Encode(c.Context)
		if err != nil {
			return nil, err
		}
		actualContext, err := canonicaljson.Encode(parent.Context)
		if err != nil || !bytes.Equal(claimContext, actualContext) {
			return nil, dia.ErrUnauthorized
		}
		if err = identity.AuthorizeDelegatedContext(ctx, c); err != nil {
			return nil, err
		}
		now := time.Now()
		if jwt.Now != nil {
			now = jwt.Now()
		}
		expires := min(c.Expires, now.Add(time.Minute).Unix())
		if c.NotBefore > now.Unix() || expires <= now.Unix() {
			return nil, dia.ErrUnauthorized
		}
		transaction := parent.Transaction
		if transaction.ID == "" {
			transaction.ID = parent.ID
		}
		if transaction.Issued == "" {
			transaction.Issued = parent.Issued
		}
		binding, err := delegatedTransactionID(transaction)
		if err != nil {
			return nil, err
		}
		delegated := dia.Claims{Issuer: jwt.Issuer, Audience: []string{jwt.Audience}, Subject: peerSubject, ID: binding, Kind: "service", Context: c.Context, NotBefore: now.Unix(), Expires: expires}
		token, err := jwt.Sign(ctx, delegated)
		if err != nil {
			return nil, err
		}
		return http.Header{"Authorization": {"Bearer " + token}}, nil
	}, nil
}

// DelegatedCommandVerifier first authenticates the configured peer, then asks
// the Dia owner to validate the CURRENT delegated user's session/scene state.
// Use DelegatedCommandCheck as the route policy too, or DelegatedCommandGateway,
// to bind the credential to the forwarded parent transaction before execution.
func DelegatedCommandVerifier(jwt dia.JWT, identity *dia.Service, peerSubject string) (transport.TokenVerifier, error) {
	if identity == nil {
		return nil, errors.New("delegated Dia state validator required")
	}
	verify, err := transport.ServiceJWTVerifier(jwt, peerSubject)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, token string) (transport.Identity, error) {
		verified, err := verify(ctx, token)
		if err != nil {
			return transport.Identity{}, err
		}
		var claims dia.Claims
		if canonicaljson.Decode(verified.Claims, &claims) != nil || len(claims.Audience) != 1 || claims.Audience[0] != jwt.Audience {
			return transport.Identity{}, dia.ErrUnauthorized
		}
		// A generic peer token is not sufficient: it must carry the bound delegated
		// transaction marker and a current end-user session owned by this network.
		if len(claims.ID) != len(delegatedIDPrefix)+64 || claims.ID[:len(delegatedIDPrefix)] != delegatedIDPrefix {
			return transport.Identity{}, dia.ErrUnauthorized
		}
		if err = identity.AuthorizeDelegatedContext(ctx, claims); err != nil {
			return transport.Identity{}, err
		}
		return verified, nil
	}, nil
}

// DelegatedCommandCheck is a mandatory companion to the token-only verifier.
// Compare the original parent trace before Command.Execute appends the receiving
// procedure's trace; body claims/context remain ignored by transport.Framework.
func DelegatedCommandCheck(ctx context.Context, identity transport.Identity, request transport.GatewayRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r := request.Request
	if r == nil || r.Method != "POST" || r.Pattern != "POST /commands/{name}" || r.PathValue("name") == "" {
		return dia.ErrForbidden
	}
	var claims dia.Claims
	if canonicaljson.Decode(identity.Claims, &claims) != nil || claims.Kind != "service" {
		return dia.ErrUnauthorized
	}
	var in command.Request
	if err := transport.InspectGatewayJSON(r, &in); err != nil {
		return err
	}
	binding, err := delegatedTransactionID(in.Transaction)
	if err != nil || binding != claims.ID {
		return dia.ErrForbidden
	}
	return nil
}

// DelegatedCommandGateway makes token verification, current user-state checking
// and transaction binding inseparable. The caller still selects exact allowed
// command names and enforces command/resource permissions at its route owner.
// It must never be installed as a delivery-envelope or store-peer authenticator.
func DelegatedCommandGateway(jwt dia.JWT, identity *dia.Service, peerSubject string) (func(*http.Request) (transport.Identity, error), error) {
	verifier, err := DelegatedCommandVerifier(jwt, identity, peerSubject)
	if err != nil {
		return nil, err
	}
	authenticate := transport.Authenticator{Bearer: verifier}
	return func(r *http.Request) (transport.Identity, error) {
		verified, err := authenticate.Authenticate(r)
		if err != nil {
			return transport.Identity{}, err
		}
		if err = DelegatedCommandCheck(r.Context(), verified, transport.GatewayRequest{Request: r, Name: r.PathValue("name")}); err != nil {
			return transport.Identity{}, err
		}
		return verified, nil
	}, nil
}

const delegatedIDPrefix = "command-delegation-v1:"

func delegatedTransactionID(transaction eventlib.Transaction) (string, error) {
	if transaction.ID == "" || len(transaction.ID) > 1024 || transaction.Issued == "" {
		return "", dia.ErrInvalid
	}
	issued, err := time.Parse(time.RFC3339Nano, transaction.Issued)
	if err != nil {
		return "", dia.ErrInvalid
	}
	for _, trace := range transaction.Path {
		if trace.Network == "" || trace.Host == "" || trace.Procedure == "" || trace.Hash == "" {
			return "", dia.ErrInvalid
		}
	}
	// Claims are authenticated separately and are not forwarded by commandrpc.CommandClient.
	// Nil and empty trace slices represent the same empty parent trace.
	path := append([]eventlib.Trace{}, transaction.Path...)
	digest, err := canonicaljson.Digest(struct {
		Domain        string           `json:"domain"`
		TransactionID string           `json:"transactionId"`
		Issued        string           `json:"issued"`
		Path          []eventlib.Trace `json:"path"`
	}{delegatedIDPrefix, transaction.ID, issued.UTC().Format(time.RFC3339Nano), path})
	if err != nil {
		return "", err
	}
	return delegatedIDPrefix + digest, nil
}
