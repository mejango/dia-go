//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	eventlib "github.com/mejango/blossm-toolbox-go/lib/event"
	"github.com/mejango/blossm-toolbox-go/procedures/event_stores/generic"
	mongoStore "github.com/mejango/blossm-toolbox-go/vendors/mongodb"
	"github.com/mejango/blossm-toolbox-go/vendors/postgres"
)

func newStore(t *testing.T, kind string) *eventstore.Store {
	t.Helper()
	ctx := context.Background()
	ns := fmt.Sprintf("test_%d", time.Now().UnixNano())
	var b eventlib.Backend
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
	return &eventstore.Store{Backend: b, Reducers: map[string]eventlib.Reducer{}}
}
