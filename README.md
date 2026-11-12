# Dia Go

Dia is the application built on [Blossm Toolbox Go](https://github.com/mejango/blossm-toolbox-go).
This repository owns a separately buildable HTTP-container entrypoint and its
application configuration. `go.mod` pins the shared implementation to one exact
Toolbox commit. The first export keeps Dia domain libraries and `diaruntime` in
that shared module; this entrypoint does not duplicate them or claim a complete
physical library split.

Retained scope includes identity, accounts, sessions/revocation, scenes, roles,
groups, system nodes/keys/connections, durable communications and custodial direct
EVM operations. Relayr payments and JB/shop remain deferred. One network is one
signing trust domain. No live deployment or provider/chain verification is claimed.

## Build and verify

Use Go1.27.4. Dependencies come from the pinned public module; no sibling checkout
or local `replace` directive is required.

```sh
go mod download
make check build
```

The shared [runtime documentation](https://github.com/mejango/blossm-toolbox-go/tree/main/diaruntime)
and [deployment guide](https://github.com/mejango/blossm-toolbox-go/blob/main/deploy/README.md)
list configuration. Required values include NETWORK, DATABASE_NAMESPACE,
DATABASE_DRIVER, the database URL, explicit LOCAL_MODE, JWT_SIGNER with its key
configuration, and base64-encoded 32-byte DIA_HMAC_KEY/DIA_ENCRYPTION_KEY.
Configure wallet/provider endpoints only when those capabilities are needed.
Local mode bypasses publication infrastructure while preserving real storage
and authorization. Use synthetic secrets and disposable databases for tests.

```sh
# Build locally; does not push or deploy the image.
docker build --platform linux/amd64 -t dia-go:local .
# After starting the pinned disposable PostgreSQL fixture:
make container-smoke IMAGE=dia-go:local
```

The smoke test reuses the pinned framework's executable HTTP lifecycle assertion:
it runs this image, creates a session, verifies retry identity, terminates it and
checks revocation. CI starts/cleans only disposable fixtures. Native domain,
wallet and storage suites remain with their owning Toolbox implementation.

`deploy/examples` contains illustrative GCP/PostgreSQL and AWS/Mongo configuration
for the shared Blossm CLI. Review resource identifiers, credentials, image and
provider prerequisites before any explicit execution. This repository does not
provision infrastructure or publish images automatically.

See [publication verification](docs/PUBLICATION.md) for exact local evidence.
