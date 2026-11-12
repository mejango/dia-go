//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	mongoStore "github.com/mejango/blossm-toolbox-go/storage/mongo"
	"github.com/mejango/blossm-toolbox-go/storage/postgres"
	"github.com/mejango/blossm-toolbox-go/stores/event"
)

// This fixture tests publication ownership, not application reducer policy.
type publicationOwnershipReducer struct{}

func (publicationOwnershipReducer) Normalize(_ string, payload json.RawMessage) (json.RawMessage, error) {
	return eventstore.Canonical(payload)
}
func (publicationOwnershipReducer) Apply(_ json.RawMessage, event eventstore.Event) (json.RawMessage, error) {
	return event.Input.Payload, nil
}

type publicationCapture []eventstore.Publication

func (p *publicationCapture) Publish(_ context.Context, notice eventstore.Publication) error {
	*p = append(*p, notice)
	return nil
}

func TestPublicationOwnershipNative(t *testing.T) {
	for _, driver := range []string{"postgres", "mongo"} {
		for _, mode := range []string{"gcp", "aws"} {
			t.Run(driver+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				namespace := fmt.Sprintf("ownership_%d", time.Now().UnixNano())
				open := func(suffix string) *eventstore.Store {
					t.Helper()
					var backend eventstore.Backend
					var err error
					if driver == "postgres" {
						uri := os.Getenv("POSTGRES_URL")
						if uri == "" {
							t.Fatal("POSTGRES_URL required")
						}
						backend, err = postgres.Open(ctx, uri, namespace+suffix)
					} else {
						uri := os.Getenv("MONGODB_URI")
						if uri == "" {
							t.Fatal("MONGODB_URI required")
						}
						backend, err = mongoStore.Open(ctx, mongoStore.Config{URI: uri, Database: "blossm_test", Namespace: namespace + suffix})
					}
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						closeCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
						defer stop()
						if err := backend.Close(closeCtx); err != nil {
							t.Error(err)
						}
					})
					return &eventstore.Store{Backend: backend, Reducers: map[string]eventstore.Reducer{
						"wallet/wallet": publicationOwnershipReducer{}, "notes/document": publicationOwnershipReducer{},
					}}
				}
				identity, documents := open("_identity"), open("_documents")
				appendEvent := func(store *eventstore.Store, service, domain string) {
					t.Helper()
					receipt, err := store.Append(ctx, []eventstore.Input{{
						Root:   eventstore.Root{Network: "local", Service: service, Domain: domain, ID: "one"},
						Action: "set", Key: "fixture", Created: "2026-11-12T00:00:00Z", Payload: json.RawMessage(`{"value":"retained"}`),
						Transaction: eventstore.Transaction{ID: "fixture", Issued: "2026-11-12T00:00:00Z"},
					}})
					if err != nil {
						t.Fatal(err)
					}
					if len(receipt) != 1 || receipt[0].Outcome != eventstore.Committed {
						t.Fatal("event not committed", receipt)
					}
				}
				pending := func(store *eventstore.Store) []eventstore.Publication {
					t.Helper()
					var notices []eventstore.Publication
					err := store.Backend.View(ctx, func(reader eventstore.Reader) error {
						var err error
						notices, err = reader.Pending(ctx, 10)
						return err
					})
					if err != nil {
						t.Fatal(err)
					}
					return notices
				}
				appendEvent(identity, "wallet", "wallet")
				appendEvent(documents, "notes", "document")
				identityBefore, documentsBefore := pending(identity), pending(documents)
				if len(identityBefore) != 1 || len(documentsBefore) != 1 {
					t.Fatal("both durable outboxes must start populated")
				}
				var remote publicationCapture
				for _, owner := range publicationOwners(mode, documents, identity, &remote) {
					if err := owner.store.Deliver(ctx, owner.publisher, 10); err != nil {
						t.Fatal(err)
					}
				}
				if !reflect.DeepEqual([]eventstore.Publication(remote), documentsBefore) {
					t.Fatal("remote publisher must receive only the exact document notice", remote)
				}
				if len(pending(documents)) != 0 {
					t.Fatal("acknowledged document notice was not deleted")
				}
				if !reflect.DeepEqual(pending(identity), identityBefore) {
					t.Fatal("document runtime mutated or drained the identity wallet notice")
				}
				var identityPublisher publicationCapture
				if err := identity.Deliver(ctx, &identityPublisher, 10); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual([]eventstore.Publication(identityPublisher), identityBefore) {
					t.Fatal("identity owner did not recover the original wallet notice")
				}
				if len(pending(identity)) != 0 {
					t.Fatal("identity owner acknowledgment did not delete its notice")
				}
			})
		}
	}
}
