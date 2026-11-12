# Standalone publication

## Plan refinement

- **Objective:** Publish the verified Go module and a separately buildable Dia application to the user-selected public repositories.
- **System fit:** Toolbox owns shared behavior and the Dia runtime package; dia-go owns an independent executable/configuration pinned to Toolbox. Existing JavaScript repositories remain references.
- **Reuse and simplicity:** Extract the existing Dia runtime without behavioral changes, use thin entrypoints, mechanically replace the module path, and adapt the existing CI to the repository root.
- **Evidence and unknowns:** Original verification and reviews are retained as historical evidence. Publication-specific checks cover extraction, module relocation and independent consumption; live cloud behavior remains unverified.
- **Verification:** Run package race/vet after extraction, full unit/race/vet after relocation, build containers, smoke the independent Dia entrypoint and verify exact public remote commits.
- **Resource budget:** Reuse installed dependencies and existing native fixtures; avoid domain duplication or a broad package split. Source publication is explicitly authorized, deployment is not.

- [x] Verify the standalone consumer against the published module.
- [ ] Push the consumer and verify its public remote identity.

## Plan refinement

- **Objective:** Make Dia a genuine independent third-party application and organize related Go packages/models into navigable capability folders, with no Dia source or dependency in Toolbox.
- **System fit:** Toolbox owns generic procedures/stores/adapters/security/transport/CLI. Dia owns application domain models, policy, services, routes, composition and its Dia-backed cross-service example. Dependencies flow only from Dia to the public framework. Original JS and verified workspace snapshots remain preserved.
- **Reuse and simplicity:** Move existing implementation/tests with their owners; extract reusable JWT signing and HTTP error facilities once before relocating consumers. Keep application policy out of generic tokens. Group Go packages without separately versioned modules or compatibility aliases. Preserve runtime behavior; no rewrite of storage or protocol.
- **Evidence and unknowns:** Inspection found hidden Dia dependencies in transport, JWT runtime/CLI and the document reference app. Resolve these explicitly; move the Dia-backed reference example with Dia rather than calling its identity integration generic. Reuse all current tests, and document prior evidence as historical. No live provider verification is claimed.
- **Verification:** Enforce Toolbox absence of Dia imports/code via an automated boundary check; compile/tests/race/vet both independent modules, native PostgreSQL/Mongo and container lifecycles, CLI generated configuration/build, dependency/security checks. Verify the application with only public framework exports; temporary development workspace must not hide private API access or enter committed go.mod.
- **Resource budget:** Delegate auth/transport extraction and application moves with disjoint ownership. Parent owns package grouping/build/config/module and composed gates. Local branches preserve published mains. Prepare paired reviewable PRs; no merge or source-split main update without approval under root AGENTS release rules.

- [x] Extract reusable security/HTTP boundaries and remove framework dependency on Dia.
- [x] Move Dia models/services/runtime/routes/tests and its identity-backed example into dia-go.
- [x] Group related Go packages/models and update all imports, CLI/build/configuration/docs.
- [x] Add independent-consumer/dependency-boundary checks and run complete relevant gates.
- [x] Prepare paired reviewable branches/PRs with exact evidence and coordinated dependency pin.

### Application move owner

- [x] Coordinate generic auth extraction before relocating identity files.
- [x] Move identity, wallet, runtime and executable sources with every test/document; retain package names and existing Dia executable test.
- [x] Rewrite moved imports to Dia-owned services/runtime/httpapi; retain generic transport CORS where applicable.
- [x] Verify source inventory, stale import absence and scoped compilation/tests when extraction/module handoffs are ready. Parent owns module/build/root docs and final independent-public-module gates.

Move-owner review: all43 moved files were SHA256-verified before import edits,
including15 tests and7 Markdown files; existing cmd/dia/main_test.go retained.
Toolbox dia/wallet/diaruntime/cmd/dia trees are absent. Package names remain dia,
wallet and diaruntime. Identity auth/signing files handed intact to the extraction
owner. Runtime calls httpapi.API/SessionProjection, generic transport.CORS and
`runtimeauth.Open` with typed session context, issuer dia.NETWORK and audience NETWORK. No module or
build/root documentation edits by this owner. Scoped compile via a temporary
/tmp workspace was attempted offline; it reached in-progress token/runtimeauth
and httpapi extraction errors plus an unavailable pubsub1.51.0 module in the
default cache. This is not passing compile evidence; rerun after extraction and
module-cache handoffs. Final public-module verification remains parent-owned.

### Final independent verification

The early extraction errors above are historical. Final independent public-module
verification passed with GOWORK=off and the pinned Toolbox review commit: unit,
race, vet, native PostgreSQL/MongoDB, generated configuration, security checks
and both application container lifecycles. The document outbox ownership fix
passed affected native and rebuilt container gates. Exact evidence and limits
are in docs/SEPARATION.md.

Prepared review: https://github.com/mejango/dia-go/pull/1 and
https://github.com/mejango/blossm-toolbox-go/pull/1. Both remain open and unmerged;
hosted CI was still running at handoff. Task-owned disposable database fixtures
have been removed after verification.

## Plan refinement

- **Objective:** Restore familiar JS organizational naming and enforce a clear dependency hierarchy with lib at the bottom, while keeping Dia independently owned and all runtime behavior intact.
- **System fit:** Shared types/serialization/helpers underpin vendors and procedure logic. Adapter names return to vendors; stores return to procedures/event_stores and view_stores; transport/composition use the closest original names. Dia consumes updated public imports, never moves back into Toolbox.
- **Reuse and simplicity:** Relocate existing code and extract current shared primitives once; no wrappers, legacy paths, new runtimes or protocol changes. Preserve Go executable entrypoints and necessary package cohesion.
- **Evidence and unknowns:** JS names are authoritative navigation references, but observed JS lib/RPC and logging/vendor cycles contradict the user's intended hierarchy. Use current Go imports to enforce a clean DAG rather than reproducing historical violations.
- **Verification:** Add automated dependency-layer checks, including lib-only internal imports and forbidden-edge fixtures; preserve tests and canonical representations. Run both modules' formatting/vet/unit/race, native and container gates, generated-config checks and independent public-pin validation before new paired PR completion.
- **Resource budget:** Delegate shared-primitives extraction with exclusive Go write ownership first; parent prepares path mapping and gates. Separate behavior-preserving extraction from mechanical folder moves. Preserve old evidence as historical; no merge of new PRs without final approval.

- [x] Extract bottom-layer event protocol and small helpers without changing behavior.
- [x] Restore familiar folders and update all consumers, docs and generators coherently.
- [x] Enforce folder dependency rules with executable regression checks.
- [ ] Verify both independent modules and prepare reviewable changes.

## Plan refinement

- **Objective:** Resolve the native concurrent-challenge failure discovered during final hierarchy verification before completing the requested correction.
- **System fit:** Mongo's known-abort retry policy protects atomic event/outbox writes; Dia reconciles only an already-visible committed receipt. Preserve uncertain-commit behavior, single-attempt streaming reads and exact operation receipts.
- **Reuse and simplicity:** Keep the existing labeled-error retry owner. Use bounded jittered exponential waits (10ms initial ceiling, 250ms cap) and ten default total attempts; retain explicit1..100 configuration. Join all concurrent native fixture workers before reporting failure.
- **Evidence and unknowns:** Preserved final native log proves five callback attempts, TransientTransactionError true, uncertainty false, and exhausted WriteConflict before the winner's receipt was visible. A premature test fatal left another worker logging after test completion, explaining the secondary race. Earlier passing runs did not expose this schedule.
- **Verification:** Add delay-bound/cancellation tests; retain native retry-budget, unknown-commit and single-attempt View tests. Repeatedly run the real concurrent Dia journey, then rerun affected full native suites and final independent module/container gates. Retain initial failing evidence.
- **Resource budget:** Separate this behavioral retry fix from the completed mechanical hierarchy commits. Mongo owner changes only its adapter/tests/docs; parent fixes the test lifecycle and public pin. No error-text retries, new conflict classification or cloud writes.

- [x] Correct bounded Mongo contention backoff and prove existing uncertainty/read contracts.
- [x] Join native fixture workers on failure and reproduce concurrent Dia completion repeatedly.
- [ ] Reverify final public pin, native adapters/application and containers, preserving failed evidence.
