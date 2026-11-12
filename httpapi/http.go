package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	canonicaljson "github.com/mejango/blossm-toolbox-go/lib/canonical_json"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/event_stores/generic"
	"github.com/mejango/blossm-toolbox-go/procedures/gateways"
	coremodel "github.com/mejango/dia-go/models/core"
	dia "github.com/mejango/dia-go/services/identity"
)

type API struct {
	Dia   *dia.Service
	Store *eventstore.Store
}

func (a API) Handler() http.Handler {
	mux := http.NewServeMux()
	a.commands(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_ = transport.Write(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /commands/session/start", transport.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Device coremodel.Device `json:"device"`
		}
		if e := transport.Decode(w, r, &in); e != nil {
			return e
		}
		ip, _, e := net.SplitHostPort(r.RemoteAddr)
		if e == nil {
			in.Device.IP = ip
		}
		out, e := a.Dia.StartSession(r.Context(), r.Header.Get("Idempotency-Key"), in.Device)
		if e != nil {
			return e
		}
		return transport.Write(w, 201, out)
	}))
	mux.HandleFunc("POST /commands/session/terminate", transport.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		if e := a.Dia.TerminateSession(r.Context(), token(r), r.Header.Get("Idempotency-Key")); e != nil {
			return e
		}
		return transport.Write(w, 200, map[string]bool{"completed": true})
	}))
	mux.HandleFunc("GET /facts/public-key", transport.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		key, e := a.Dia.PublicKey()
		if e != nil {
			return e
		}
		return transport.Write(w, 200, map[string]string{"algorithm": "ES256", "pem": string(key)})
	}))
	mux.HandleFunc("GET /views/session", transport.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		c, e := a.Dia.AuthenticateAccess(r.Context(), token(r))
		if e != nil {
			return e
		}
		var p eventlib.Projection
		e = a.Store.Backend.View(r.Context(), func(reader eventlib.Reader) error {
			var e error
			p, e = reader.Projection(r.Context(), c.Context.Session.Key())
			return e
		})
		if e != nil {
			return e
		}
		if p.Deleted {
			return dia.ErrForbidden
		}
		return transport.Write(w, 200, p)
	}))
	return mux
}

// SessionProjection replaces a single session row with the authoritative
// aggregate and revision. Duplicate/out-of-order notices cannot roll it back.
type SessionProjection struct{ Store *eventstore.Store }

func (p SessionProjection) Publish(ctx context.Context, notice eventlib.Publication) error {
	root := notice.Event.Input.Root
	if root.Service != "core" || root.Domain != "session" {
		return nil
	}
	a, e := p.Store.Aggregate(ctx, root)
	if e != nil {
		return e
	}
	var state coremodel.SessionState
	if e = json.Unmarshal(a.State, &state); e != nil {
		return e
	}
	body, e := canonicaljson.Encode(struct {
		Device coremodel.Device `json:"device"`
	}{state.Device})
	if e != nil {
		return e
	}
	_, e = p.Store.Project(ctx, eventlib.Projection{ID: root.Key(), Root: root, Version: a.Version, Body: body, Deleted: state.Terminated != ""})
	return e
}

func token(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}
