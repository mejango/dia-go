// Package serviceapp is a Dia-owned document example composed from Toolbox
// procedure owners. Dia owns its identity, authorization and domain decisions.
package serviceapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	canonicaljson "github.com/mejango/blossm-toolbox-go/lib/canonical_json"
	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/command"
	commandrpc "github.com/mejango/blossm-toolbox-go/procedures/command/rpc"
	"github.com/mejango/blossm-toolbox-go/procedures/event_stores/generic"
	"github.com/mejango/blossm-toolbox-go/procedures/fact"
	"github.com/mejango/blossm-toolbox-go/procedures/job"
	dia "github.com/mejango/dia-go/services/identity"
)

const Domain = "document"
const ViewName = "documents"

type Document struct {
	Owner eventlib.Root `json:"owner"`
	Text  string        `json:"text"`
}
type SetDocument struct {
	Text     string `json:"text"`
	Expected *int64 `json:"expected,omitempty"`
}
type documentReducer struct{}

func (documentReducer) Normalize(action string, payload json.RawMessage) (json.RawMessage, error) {
	var d Document
	if action != "set" || canonicaljson.Decode(payload, &d) != nil || d.Owner.Network == "" || d.Owner.Service != "core" || d.Owner.Domain != "principal" || d.Owner.ID == "" || len(d.Text) > 4096 {
		return nil, eventlib.ErrInvalid
	}
	return canonicaljson.Encode(d)
}
func (r documentReducer) Apply(state json.RawMessage, e eventlib.Event) (json.RawMessage, error) {
	b, err := r.Normalize(e.Input.Action, e.Input.Payload)
	if err != nil {
		return nil, err
	}
	var next, old Document
	if err = canonicaljson.Decode(b, &next); err != nil {
		return nil, err
	}
	if err = canonicaljson.Decode(state, &old); err != nil {
		return nil, err
	}
	var contextValues command.Context
	if err = canonicaljson.Decode(e.Input.Context, &contextValues); err != nil {
		return nil, err
	}
	actor, err := principal(contextValues, e.Input.Root.Network)
	if err != nil || actor != next.Owner || (old.Owner.ID != "" && old.Owner != next.Owner) {
		return nil, dia.ErrForbidden
	}
	return b, nil
}
func principal(values command.Context, network string) (eventlib.Root, error) {
	var p eventlib.Root
	if canonicaljson.Decode(values["principal"], &p) != nil || p.Network != network || p.Service != "core" || p.Domain != "principal" || p.ID == "" {
		return p, dia.ErrForbidden
	}
	return p, nil
}
func (a *App) root(id string) eventlib.Root {
	return eventlib.Root{Network: a.network, Service: a.service, Domain: Domain, ID: id}
}
func (a *App) read(ctx context.Context, values command.Context, id string) (eventlib.Aggregate, error) {
	owner, err := principal(values, a.network)
	if err != nil {
		return eventlib.Aggregate{}, err
	}
	if strings.TrimSpace(id) == "" {
		return eventlib.Aggregate{}, dia.ErrInvalid
	}
	aggregate, err := a.Store.Aggregate(ctx, a.root(id))
	if err != nil {
		return aggregate, err
	}
	if aggregate.Version < 0 {
		return aggregate, eventlib.ErrNotFound
	}
	var state Document
	if err = canonicaljson.Decode(aggregate.State, &state); err != nil {
		return aggregate, err
	}
	if state.Owner != owner {
		return eventlib.Aggregate{}, eventlib.ErrNotFound
	}
	return aggregate, nil
}
func (a *App) command(queue command.Queue) command.Procedure {
	return command.Procedure{Trace: a.trace("set"), Store: a.Store, Queue: queue, Local: a.Store.Local, RequiresRoot: true, RequiredContexts: []string{"principal"}, Main: func(ctx context.Context, in command.Request) (command.Result, error) {
		if in.Root == nil || in.Root.Network != a.network || in.Root.Service != a.service || in.Root.Domain != Domain || strings.TrimSpace(in.Root.ID) == "" {
			return command.Result{}, command.ErrInvalid
		}
		owner, err := principal(in.Context, a.network)
		if err != nil {
			return command.Result{}, err
		}
		// Reject known ownership violations before queue admission. The reducer
		// repeats this check transactionally because ownership may race this read.
		current, err := a.Store.Aggregate(ctx, *in.Root)
		if err != nil {
			return command.Result{}, err
		}
		if current.Version >= 0 {
			var document Document
			if err := canonicaljson.Decode(current.State, &document); err != nil {
				return command.Result{}, err
			}
			if document.Owner != owner {
				return command.Result{}, dia.ErrForbidden
			}
		}
		var input SetDocument
		if err = canonicaljson.Decode(in.Payload, &input); err != nil || len(input.Text) > 4096 {
			return command.Result{}, command.ErrInvalid
		}
		body, err := canonicaljson.Encode(Document{Owner: owner, Text: input.Text})
		if err != nil {
			return command.Result{}, err
		}
		return command.Result{Events: []eventlib.Input{{Root: *in.Root, Action: "set", Payload: body, Expected: input.Expected}}, Then: func(_ context.Context, receipts []command.Receipt) (command.Contribution, error) {
			// Response construction is repeatable. External effects must separately
			// inspect COMMITTED; this callback is never a durable action protocol.
			if len(receipts) != 1 {
				return command.Contribution{}, errors.New("document receipt count")
			}
			id, err := canonicaljson.Encode(in.Root.ID)
			if err != nil {
				return command.Contribution{}, err
			}
			return command.Contribution{Body: map[string]json.RawMessage{"id": id}, StatusCode: 200}, nil
		}}, nil
	}}
}
func (a *App) fact() fact.Procedure {
	return fact.Procedure{RequiredContexts: []string{"principal"}, ReadFn: func(ctx context.Context, in fact.Request) (json.RawMessage, error) {
		aggregate, err := a.read(ctx, in.Context, in.Root)
		if err != nil {
			return nil, err
		}
		return canonicaljson.Encode(aggregate)
	}, StreamFn: func(ctx context.Context, in fact.Request, yield fact.Yield) error {
		owner, err := principal(in.Context, a.network)
		if err != nil {
			return err
		}
		return a.Store.StreamAggregates(ctx, a.network, a.service, Domain, in.Root, func(v eventlib.Aggregate) error {
			var d Document
			if err := canonicaljson.Decode(v.State, &d); err != nil {
				return err
			}
			if d.Owner != owner {
				return nil
			}
			b, err := canonicaljson.Encode(v)
			if err != nil {
				return err
			}
			return yield(ctx, b)
		})
	}}
}
func (a *App) verifyJob() job.Procedure {
	return job.Procedure{Run: func(ctx context.Context, in job.Input) error {
		var p struct {
			ID string `json:"id"`
		}
		if canonicaljson.Decode(in.Payload, &p) != nil {
			return dia.ErrInvalid
		}
		var values command.Context
		if canonicaljson.Decode(in.Context, &values) != nil {
			return dia.ErrForbidden
		}
		var claims dia.Claims
		owner, ownerErr := principal(values, a.network)
		if canonicaljson.Decode(in.Claims, &claims) != nil || ownerErr != nil || claims.Kind != "access" || claims.Subject != owner.ID || claims.Context.Principal != owner {
			return dia.ErrForbidden
		}
		_, err := a.read(ctx, values, p.ID)
		return err
	}}
}

// A stable procedure contract identity keeps request retries independent of builds.
func (a *App) trace(name string) *eventlib.Trace {
	return &eventlib.Trace{Network: a.network, Host: a.service, Procedure: name, Hash: "document-v1"}
}

// copyCommand invokes a fixed configured service from inside authenticated Main.
// It promises no transaction across services: a child can commit before the
// parent response is lost, and retry recovers that same child's stable event key.
func (a *App) copyCommand(client commandrpc.CommandClient) command.Procedure {
	return command.Procedure{Trace: a.trace("copy"), RequiredContexts: []string{"principal"}, Main: func(ctx context.Context, in command.Request) (command.Result, error) {
		var input struct {
			Source        string `json:"source"`
			SourceVersion *int64 `json:"sourceVersion"`
			Destination   string `json:"destination"`
		}
		if canonicaljson.Decode(in.Payload, &input) != nil || strings.TrimSpace(input.Destination) == "" || input.SourceVersion == nil || *input.SourceVersion < 0 {
			return command.Result{}, command.ErrInvalid
		}
		owner, err := principal(in.Context, a.network)
		if err != nil {
			return command.Result{}, err
		}
		aggregate, err := a.Store.History(ctx, a.root(input.Source), eventstore.HistoryOptions{ThroughNumber: input.SourceVersion})
		if err != nil {
			return command.Result{}, err
		}
		var document Document
		if err := canonicaljson.Decode(aggregate.State, &document); err != nil {
			return command.Result{}, err
		}
		if aggregate.LastIncludedNumber != *input.SourceVersion || document.Owner != owner {
			return command.Result{}, eventlib.ErrNotFound
		}
		payload, err := canonicaljson.Encode(SetDocument{Text: document.Text})
		if err != nil {
			return command.Result{}, err
		}
		target := strings.Split(client.Target, "/")
		root := eventlib.Root{Network: target[0], Service: target[1], Domain: Domain, ID: input.Destination}
		child, err := client.Execute(ctx, in, "set", "copy", command.Request{Root: &root, Payload: payload})
		if err != nil {
			return command.Result{}, err
		}
		body, err := canonicaljson.Encode(child)
		if err != nil {
			return command.Result{}, err
		}
		return command.Result{Response: command.Contribution{StatusCode: http.StatusOK, Body: map[string]json.RawMessage{"child": body}}}, nil
	}}
}
