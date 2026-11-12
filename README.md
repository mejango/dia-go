# Dia Go

Dia is an independent application built with
[Blossm Toolbox Go](https://github.com/mejango/blossm-toolbox-go). This repository
owns its models, domain behavior, authorization policy, HTTP routes, runtime,
configuration and application tests. Toolbox is a pinned public dependency.
There is no local replacement or requirement for a sibling Toolbox checkout.

## Application groups

| Folder | Responsibility |
| --- | --- |
| `models/core` | Typed identity, accounts, sessions, challenges, roles, groups and scenes |
| `models/system` | Typed nodes and keys |
| `models/comms` | Typed communication delivery state |
| `services/identity` | Core, system and communication operations, authorization and recovery |
| `services/wallet` | Custodial wallet ownership, durable journal and lifecycle |
| `httpapi` | Dia routes and Dia-specific framework authentication adapters |
| `runtime`, `cmd/dia` | Infrastructure composition and application lifecycle |
| `examples/document`, `cmd/document` | Cross-service document example using Dia identity |
| `deployment`, `cmd/dia-config` | Dia-specific deployment validation before generic framework generation |

Domain models are grouped without multiplying modules or copying framework code.
The document example deliberately integrates Dia; it is not a prerequisite for
other applications using Toolbox. Generic signing remains in Toolbox, while Dia
owns its typed context, issuer, token kinds and persisted revocation policy.

## Build and verify

Use Go 1.27.4:

```sh
go mod download
make check build
# Disposable native PostgreSQL and Mongo replica fixtures are required:
make native
make container
```

`make check` enforces application ownership and runs formatting, unit/race, vet
and module checks. Native application suites run in this repository, including
identity, wallets and cross-service recovery. Container tests exercise both
shipped HTTP executables through session/retry/revocation and document workflows.

[Runtime configuration](runtime/README.md), [application architecture](docs/ARCHITECTURE.md),
[identity services](services/identity/README.md), [wallet](services/wallet/README.md)
and [current verification](docs/SEPARATION.md), [publication history](docs/PUBLICATION.md) give the concrete contracts.
Local mode bypasses remote publication infrastructure while retaining real
storage and authorization. One network is one signing trust domain.

Core identity, system, communications and custodial direct-EVM flows are retained.
Relayr payments and JB/shop remain explicitly deferred. No live cloud deployment
or blockchain/provider verification is claimed. Deployment generation is
preparation only; execution requires separate authorization.
