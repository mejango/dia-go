package httpapi

import (
	"context"
	"errors"

	canonicaljson "github.com/mejango/blossm-toolbox-go/lib/canonical_json"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/gateways"
	dia "github.com/mejango/dia-go/services/identity"
)

// DiaVerifier delegates to persisted session/scene or connection/key revocation
// checks, not merely JWT signature verification. Select one kind per service.
func DiaVerifier(service *dia.Service, kind string) (transport.TokenVerifier, error) {
	if service == nil || (kind != "access" && kind != "connection") {
		return nil, errors.New("Dia service and access/connection kind required")
	}
	return func(ctx context.Context, token string) (transport.Identity, error) {
		var c dia.Claims
		var err error
		if kind == "access" {
			c, err = service.AuthenticateAccess(ctx, token)
		} else {
			c, err = service.AuthenticateConnection(ctx, token)
		}
		if err != nil {
			return transport.Identity{}, err
		}
		return transport.IdentityFromClaims(c)
	}, nil
}

// DiaPermission checks one exact subject selected by trusted application policy.
// Never accept an owner/context from the body as the authenticated identity.
func DiaPermission(service *dia.Service, subject eventlib.Root, permission string) transport.GatewayCheck {
	return func(ctx context.Context, identity transport.Identity, _ transport.GatewayRequest) error {
		if service == nil || subject.ID == "" || permission == "" {
			return dia.ErrForbidden
		}
		var claims dia.Claims
		if canonicaljson.Decode(identity.Claims, &claims) != nil {
			return dia.ErrUnauthorized
		}
		return service.RequirePermission(ctx, claims, subject, permission)
	}
}
