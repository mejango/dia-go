# Dia architecture

`github.com/mejango/dia-go` is an independent third-party application consuming
[Blossm Toolbox Go](https://github.com/mejango/blossm-toolbox-go) through public
module APIs. Dia owns its models, authorization policy, domain services, HTTP
application, runtime and executable examples. Toolbox must neither import Dia
nor supply its application implementation.

## Owners and entrypoints

| Directory | Responsibility |
| --- | --- |
| [models/core](../models/core) | Identity, principal/account, session/scene and role/group data contracts |
| [models/system](../models/system) | Node, key and connection data contracts |
| [models/comms](../models/comms) | Communications data contracts |
| [services/identity](../services/identity) | Domain reducers, permissions, live revocation, recovery, system and communications behavior; package name remains dia |
| [services/wallet](../services/wallet) | Owner-authorized direct custody, durable signing/nonce state and reconciliation |
| [httpapi](../httpapi) | Dia routes, identity adapters and delegated service credentials over generic transport |
| [runtime](../runtime) | Application configuration, client ownership, workers and graceful shutdown |
| [cmd/dia](../cmd/dia) | Identity/system/comms/custody executable |
| [examples/document](../examples/document), [cmd/document](../cmd/document) | Identity-backed document application demonstrating generic procedures, publication, projections and cross-service calls |
| [integration](../integration) | Application/native/container and cross-service acceptance assertions |

Models have no transport or framework execution policy. Services adapt their
application data to Toolbox's event/store/procedure contracts. HTTP composition
binds authenticated caller context and claims before effects. Toolbox generic
ES256 token types/signers remain separate from Dia's claims, expiry ceilings,
session/scene revocation and permissions policy.

## Data integrity, authority and recovery

Every retained application operation uses the generic event protocol and native
backend integrity. Local mode preserves real transactions while bypassing
publication infrastructure. Identity and document namespaces/outboxes remain
separate. Required group membership with zero groups denies access. Named views
use matching logical names at producer and reader boundaries; application query
writes follow committed source events rather than creating projection revisions.

Request IDs and exact durable receipts support lost-response recovery. Secret
inputs are represented by keyed request digests where required. Derived nested
commands bind stable child identity and immutable source versions; retrying after
source changes cannot silently retarget intent. Committed child effects cannot
be rolled back by a later parent failure. Delivery credentials remain distinct
from ordinary store peers, and deterministic terminal rejection requires durable
recording before acknowledgment.

Custodial signing and the account nonce frontier commit before broadcast. Unknown
submission retains its nonce/hash; retries rebroadcast identical bytes. Only
canonical sufficiently confirmed receipt evidence marks a final chain outcome.
Cloud calls, message sending and chain transactions remain explicit operations.
Relayr payments and JB/shop are deferred scope, not successful stubs.

The selected security model remains one signing trust domain per network. Shared
signers can issue peer subjects; subject/audience checks do not provide isolation
from a compromised signer in that same trust domain. Runtime owns cancellation,
request/worker shutdown and client closure; arbitrary callbacks must cooperate
with their contexts.

## Coverage navigation and verification

The complete original154-package inventory and59-family behavior ledger remain
canonical in Toolbox, including delegated Dia-owned behavior:

- [Capability inventory](https://github.com/mejango/blossm-toolbox-go/blob/main/docs/capabilities.md)
- [Behavior assertions](https://github.com/mejango/blossm-toolbox-go/blob/main/docs/behavior-coverage.md)
- [Machine ledger](https://github.com/mejango/blossm-toolbox-go/blob/main/docs/behavior-coverage.json)

Search `dia-go:` in machine source/test paths for application-owned records.
Families include identity/revocation, accounts/challenges, roles/groups,
system/nodes/keys, communications, custody and the document integration flow.
Generic and application assertions can jointly support a family; this is not a
claim of one-to-one JS test equivalence. No package or reference test is removed
merely because its current owner is an application repository.

See [HIERARCHY.md](HIERARCHY.md) for the framework relocation and unchanged application owners. Links below/above use main provisionally for current navigation; the prior source-split pins and evidence describe their own snapshot. Historical combined-tree logs/manifests/reviews remain unchanged and do
not verify the current boundary. Fresh independent module, native/container and
security gates belong to the parent verification record. Refer to this
repository's [README](../README.md) and [PUBLICATION](PUBLICATION.md) for public
module pins and execution evidence; the source split does not establish live
provider/chain verification or deployment.

Toolbox's behavior checker accepts `DIA_SOURCE=/path/to/dia-go`, defaulting to a
sibling checkout for documentation verification only. Neither application nor
framework builds may rely on that environment variable or a sibling filesystem.
