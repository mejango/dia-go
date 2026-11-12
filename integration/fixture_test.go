//go:build integration

package integration

import (
	"context"
	"fmt"
	mongoStore "github.com/mejango/blossm-toolbox-go/storage/mongo"
	"github.com/mejango/blossm-toolbox-go/storage/postgres"
	"github.com/mejango/blossm-toolbox-go/stores/event"
	"os"
	"testing"
	"time"
)

func newStore(t *testing.T, kind string) *eventstore.Store {
	t.Helper()
	ctx := context.Background()
	ns := fmt.Sprintf("test_%d", time.Now().UnixNano())
	var b eventstore.Backend
	var e error
	if kind == "postgres" {
		uri := os.Getenv("POSTGRES_URL")
		if uri == "" {
			t.Fatal("POSTGRES_URL required")
		}
		b, e = postgres.Open(ctx, uri, ns)
	} else {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			t.Fatal("MONGODB_URI required")
		}
		b, e = mongoStore.Open(ctx, mongoStore.Config{URI: uri, Database: "blossm_test", Namespace: ns, MaxAttempts: 20})
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := b.Close(context.Background()); e != nil {
			t.Error(e)
		}
	})
	return &eventstore.Store{Backend: b, Reducers: map[string]eventstore.Reducer{}}
}
