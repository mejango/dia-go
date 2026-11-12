GO ?= go
IMAGE ?= dia-go:local
CONTAINER_NETWORK ?= blossm-go-native_default
CONTAINER_POSTGRES_URL ?= postgres://postgres:postgres@postgres:5432/blossm?sslmode=disable
export GOWORK := off

.PHONY: check build container container-smoke native config-check
check: config-check
	python3 tools/check_boundaries.py
	@test -z "$$($(GO) fmt ./...)" || { echo 'Formatting changed files'; exit 1; }
	$(GO) vet -mod=readonly ./...
	$(GO) test -mod=readonly -count=1 ./...
	$(GO) test -mod=readonly -race ./...
	$(GO) mod verify
	$(GO) mod tidy -diff
build:
	$(GO) build -mod=readonly -trimpath -o bin/dia ./cmd/dia
	$(GO) build -mod=readonly -trimpath -o bin/document ./cmd/document
	$(GO) build -mod=readonly -trimpath -o bin/dia-config ./cmd/dia-config
container:
	docker build --platform linux/amd64 -t $(IMAGE) .
	$(MAKE) container-smoke
container-smoke:
	CONTAINER_IMAGE=$(IMAGE) CONTAINER_NETWORK=$(CONTAINER_NETWORK) CONTAINER_POSTGRES_URL='$(CONTAINER_POSTGRES_URL)' $(GO) test -mod=readonly -tags=containerintegration -count=1 -timeout=3m ./integration

native:
	@test -n "$$POSTGRES_URL" || { echo 'POSTGRES_URL required'; exit 1; }
	@test -n "$$MONGODB_URI" || { echo 'MONGODB_URI required'; exit 1; }
	$(GO) test -mod=readonly -race -p=1 -tags=integration -count=1 -timeout=10m ./...

config-check:
	$(GO) run -mod=readonly ./cmd/dia-config --config deploy/examples/gcp-postgres.json
	$(GO) run -mod=readonly ./cmd/dia-config --config deploy/examples/aws-mongo.json
	$(GO) run -mod=readonly ./cmd/dia-config --config deploy/examples/gcp-document-postgres.json
	$(GO) run -mod=readonly ./cmd/dia-config --config deploy/examples/aws-document-mongo.json
