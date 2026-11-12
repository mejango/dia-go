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
- [ ] Prepare paired reviewable branches/PRs with exact evidence and coordinated dependency pin.

### Application move owner

- [x] Coordinate generic auth extraction before relocating identity files.
- [x] Move identity, wallet, runtime and executable sources with every test/document; retain package names and existing Dia executable test.
- [x] Rewrite moved imports to Dia-owned services/runtime/httpapi; retain generic transport CORS where applicable.
- [ ] Verify source inventory, stale import absence and scoped compilation/tests when extraction/module handoffs are ready. Parent owns module/build/root docs and final independent-public-module gates.

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
