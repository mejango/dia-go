GO ?= go
IMAGE ?= dia-go:local
CONTAINER_NETWORK ?= blossm-go-native_default
CONTAINER_POSTGRES_URL ?= postgres://postgres:postgres@postgres:5432/blossm?sslmode=disable
export GOWORK := off

.PHONY: check build container container-smoke
check:
	@test -z "$$($(GO) fmt ./...)" || { echo 'Formatting changed files'; exit 1; }
	$(GO) vet -mod=readonly ./...
	$(GO) test -mod=readonly -race ./...
	$(GO) mod verify
	$(GO) mod tidy -diff
build:
	$(GO) build -mod=readonly -trimpath -o bin/dia ./cmd/dia
container:
	docker build --platform linux/amd64 -t $(IMAGE) .
	$(MAKE) container-smoke
container-smoke:
	CONTAINER_IMAGE=$(IMAGE) CONTAINER_NETWORK=$(CONTAINER_NETWORK) CONTAINER_POSTGRES_URL='$(CONTAINER_POSTGRES_URL)' $(GO) test -mod=readonly -tags=containerintegration -count=1 -timeout=3m github.com/mejango/blossm-toolbox-go/integration -run '^TestContainerSessionLifecycle$$'
