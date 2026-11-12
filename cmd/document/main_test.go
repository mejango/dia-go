package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/event_stores/generic"
	"github.com/mejango/dia-go/httpapi"
	dia "github.com/mejango/dia-go/services/identity"
)

type countingPublisher struct{ calls atomic.Int32 }

func (p *countingPublisher) Publish(context.Context, eventlib.Publication) error {
	p.calls.Add(1)
	return nil
}
func TestPublicationRouting(t *testing.T) {
	documents := &countingPublisher{}
	router := publicationRouter{documents: documents, network: "network", service: "notes"}
	for _, tc := range []struct {
		network, service, domain string
		want                     error
	}{{"network", "notes", "document", nil}, {"network", "core", "session", dia.ErrForbidden}, {"network", "wallet", "wallet", dia.ErrForbidden}, {"network", "system", "node", dia.ErrForbidden}, {"network", "comms", "message", dia.ErrForbidden}, {"network", "notes", "wallet", dia.ErrForbidden}, {"other", "notes", "document", dia.ErrForbidden}, {"network", "foreign", "document", dia.ErrForbidden}} {
		err := router.Publish(context.Background(), eventlib.Publication{Event: eventlib.Event{Input: eventlib.Input{Root: eventlib.Root{Network: tc.network, Service: tc.service, Domain: tc.domain}}}})
		if !errors.Is(err, tc.want) {
			t.Fatal(err)
		}
	}
	if documents.calls.Load() != 1 {
		t.Fatal("wrong publication destination")
	}
}
func TestPublicationOwners(t *testing.T) {
	documents, identity := &eventstore.Store{}, &eventstore.Store{}
	publisher := &countingPublisher{}
	for _, mode := range []string{"local", "inline", "gcp", "aws"} {
		t.Run(mode, func(t *testing.T) {
			owners := publicationOwners(mode, documents, identity, publisher)
			want := 1
			if mode == "local" {
				want = 0
			}
			if mode == "inline" {
				want = 2
			}
			if len(owners) != want {
				t.Fatalf("owners = %d, want %d", len(owners), want)
			}
			if want == 0 {
				return
			}
			if owners[0].store != documents || owners[0].publisher != publisher {
				t.Fatal("document outbox lost its publisher")
			}
			if mode == "inline" {
				projection, ok := owners[1].publisher.(httpapi.SessionProjection)
				if owners[1].store != identity || !ok || projection.Store != identity {
					t.Fatal("inline identity outbox requires direct session projection")
				}
			}
		})
	}
}

func TestHTTPServerCancellationJoins(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	exited := make(chan struct{})
	server := &http.Server{BaseContext: func(net.Listener) context.Context { return ctx }, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done(); close(exited) })}
	done := make(chan error, 1)
	go func() { done <- serve(ctx, cancel, server, listener, make(chan error)) }()
	clientDone := make(chan struct{})
	go func() {
		defer close(clientDone)
		response, e := http.Get("http://" + listener.Addr().String())
		if e == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop")
	}
	select {
	case <-exited:
	default:
		t.Fatal("handler not joined")
	}
	<-clientDone
}
