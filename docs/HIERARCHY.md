# Application and framework hierarchy

Dia remains an independent consumer of Toolbox's public module. Its own model,
service, HTTP, runtime, deployment and executable directories retain their
application responsibilities. Restoring familiar framework folder names does
not move Dia policy or implementation back into Toolbox.

Toolbox's lower-to-higher capability layers are lib, generic RPC/vendors/onchain/
middleware, procedures, CLI and cmd. These arrows describe capabilities supplied
to consumers; imports point toward allowed lower layers. Internally, Toolbox lib
packages depend only on other lib packages, never on procedures or adapters.
See the framework [hierarchy](https://github.com/mejango/blossm-toolbox-go/blob/d45c8b4e036cf190b2aaccedc9efdfa78701f640/docs/HIERARCHY.md)
for the full navigation map and owning layer checker. This link names the exact pinned framework revision.

Dia's models stay under `models/{core,system,comms}`, domain operations under
`services/{identity,wallet}`, and application adapters/composition under `httpapi`,
`runtime`, `deployment`, `examples/document` and `cmd`. Public framework imports
have been relocated to the familiar hierarchy; this document introduces no
compatibility wrappers, application copy or local replacement dependency.

The previous source-split logs/manifests/reviews remain historical and unchanged.
Independent module and runtime verification passed; see
[HIERARCHY_VERIFICATION.md](HIERARCHY_VERIFICATION.md). Neither a path map nor a documentation link constitutes a
passing gate or live provider/chain evidence.

## Verification finding

Final native verification exposed a pre-existing Mongo contention schedule:
five labeled known-abort attempts exhausted 50ms of retry waits before the
winning command receipt was visible. The run and its secondary test-lifecycle
race are preserved in verification-logs/hierarchy. The fixture now drains all
workers before fatal assertions. A separate bounded-backoff fix passed verification
at the framework adapter owner; no success is inferred from an absent receipt.
