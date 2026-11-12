# Application deployment validation

## Plan refinement

- **Objective:** Reject invalid Dia identity/document application deployment configuration before generic public CLI artifact generation, preserving application secrets and namespace authority.
- **System fit:** Toolbox CLI owns provider resource/env/signing schema validation and artifact rendering. Dia owns selected application requirements. Parent owns command/Makefile composition and execution; this package never executes providers or reads secret values.
- **Reuse and simplicity:** Wrap public cli.Config.Validate, RuntimeEnv and cli.Prepare; validate required secret references rather than duplicate runtime secret decoding. Freeze schema with its owner before implementation. Coordinate reusable pure app checks with runtime owners if runtime constraints need factoring.
- **Evidence and unknowns:** Current runtime requires HMAC/encryption keys and signing; document additionally requires isolated identity/document namespaces and peer/service settings. Schema v2 Application/Env/Signing is in progress. No coherent-source test gate or live deployment evidence is available yet.
- **Verification:** Table-driven negative tests must show missing secrets, absent/mismatched signing, invalid application vars and equal namespaces fail before any plan is returned. Positive identity/document configurations must preserve app environment and secret references in generic generated artifacts. Hold execution until parent readiness.
- **Resource budget:** Own deployment/, parent-authorized cmd/dia-config and four Dia deploy/examples, no module, generic CLI, runtime or Makefile edits. One narrow app validator and preparation wrapper, no custom deployment executor or provider calls.

- [x] Freeze public schema with Boole; parent accepted Validate/Prepare/ReadConfig and exact application.Package selection.
- [x] Implement application validation and public-CLI-only preparation.
- [x] Add failure/positive artifact tests and hand off command/config integration.
- [ ] Run scoped checks after parent readiness.

Parent accepted command contract: --config path is required; optional --out dir
requests preparation and WritePlan. Without --out only validation runs. No runner
or execute option exists. Generic CLI remains the sole artifact writer.

Schema frozen with Boole: optional Database/Publication/Signing pointers,
Application{Package,Binary,Entrypoint}, Env map, optional Peers render NETWORK and
SERVICE. Generic RuntimeEnv owns infra keys; application env owns PUBLISHER or
PUBLICATION_MODE and LOCAL_MODE. Parent authorizes four Dia examples; Boole owns
only Toolbox examples. Runtime secret decoding is deliberately not duplicated.

Source review: gofmt, JSON parsing, plan checker and whitespace checks pass.
All four examples preserve their prior provider/database/publication/secret/delivery
objects exactly; only application profile/schema and explicit application env and
signing trust changed. Go tests are intentionally unrun pending parent readiness.
No module edits, provider execution, secret-value reads or remote operations.
Tests cover both profiles/providers, app-invalid no-output, namespace isolation,
required secrets, signer trust and optional feature secret references.

Verification update after CLI owner's frozen-schema release: offline scoped
`go test ./deployment ./cmd/dia-config` PASS (3.601s and1.282s), and scoped
`go vet ./deployment ./cmd/dia-config` PASS on installed Go1.27.4. These runs used
an uncommitted /tmp workspace joining the two exports, so they prove composed
source behavior, not the later independently published dependency pin. Tests
exercise all four examples through actual public rendering and command output.
No full application/tooling suites or provider calls were run by this owner.

## Plan refinement

- **Objective:** Make paired identity/document deployment examples address the same real identity store while keeping document data and publication consumers isolated.
- **System fit:** Generic Database.Name owns Mongo physical database selection; Dia examples own paired network/trust, identity prefix and separate publication resources. Runtime/module owners remain unchanged.
- **Reuse and simplicity:** Keep public validation/rendering; change four example configurations and add one cross-profile regression using their prepared environment rather than another runtime validator.
- **Evidence and unknowns:** Review found independently valid examples used incompatible identity namespaces and shared publication consumers. Boole is adding required Mongo Name, forbidden on PostgreSQL. No live provisioning is implied.
- **Verification:** Assert paired database references and Mongo physical database agree, identity prefix matches document identity lookup, document prefix differs, signing/network trust agrees, and publication topics/subscriptions/queues differ; run scoped tests/vet after schema handoff.
- **Resource budget:** Deployment/examples/tests only plus correction lesson, no runtime/module edits or remote operations. Reuse existing four-example fixtures and offline workspace.

Correction complete: identity namespace is dia_identity; document namespace is
 documents and its identity lookup is dia_identity. Mongo physical Name is diaapp
for both profiles. Document publication uses document-events/document-projection,
including matching AWS queueURL/ARN; Dia retains dia-events/dia-projection. This
supersedes the earlier exact-preservation note for these erroneous fields.
TestPairedApplicationsShareIdentityAndIsolatePublications verifies actual prepared
environment, shared endpoint/trust/secret refs and isolated consumer targets.
Fresh offline Go1.27.4 `go test ./deployment ./cmd/dia-config -count=1` passed
(1.881s/2.202s); scoped vet passed. Temporary workspace evidence only, no live
provider observations. No runtime or module edits.
