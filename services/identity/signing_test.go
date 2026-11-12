package dia

import (
	"context"
	"testing"
	"time"

	coremodel "github.com/mejango/dia-go/models/core"
)

type signingFunc func(context.Context, []byte) ([]byte, error)

func (f signingFunc) SignES256(ctx context.Context, message []byte) ([]byte, error) {
	return f(ctx, message)
}

func TestServicePropagatesSigningContext(t *testing.T) {
	service, _, _ := coreService(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	local := service.JWT.Signer
	calls := 0
	service.JWT.Signer = signingFunc(func(got context.Context, message []byte) ([]byte, error) {
		calls++
		if got != ctx {
			t.Fatal("service dropped request context")
		}
		return local.SignES256(got, message)
	})
	secret, err := NewSessionSecret()
	if err != nil {
		t.Fatal(err)
	}
	start, err := service.StartSession(ctx, secret, coremodel.Device{Type: "desktop", Platform: "browser", Agent: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	challenge, err := service.SaveSession(ctx, start.Token, "save", "person@example.com", "")
	if err != nil || challenge.Token == "" || calls != 2 {
		t.Fatal("challenge signing context", err, calls)
	}
}
