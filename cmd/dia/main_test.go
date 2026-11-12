package main

import (
	"context"
	"testing"

	"github.com/mejango/blossm-toolbox-go/diaruntime"
)

func TestMissingNetworkFailsBeforeOpeningClients(t *testing.T) {
	t.Setenv("NETWORK", "")
	if err := diaruntime.Run(context.Background()); err == nil {
		t.Fatal("unconfigured application reported success")
	}
}
