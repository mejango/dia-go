package main

import (
	"context"
	"testing"

	"github.com/mejango/dia-go/runtime"
)

func TestMissingNetworkFailsBeforeOpeningClients(t *testing.T) {
	t.Setenv("NETWORK", "")
	if err := diaruntime.Run(context.Background()); err == nil {
		t.Fatal("unconfigured application reported success")
	}
}
