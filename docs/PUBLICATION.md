# Repository ownership correction

The initial publication below is historical. The independent-application refactor
moves all Dia models, services, runtime, HTTP policy and application tests into
dia-go. Toolbox owns only generic framework facilities and has no Dia dependency.
See README.md and docs/ARCHITECTURE.md at this revision for current ownership.
Fresh refactor verification is recorded separately; initial publication results
are not evidence for later source changes.

# Publication verification

Verified locally on 2026-11-12 with Go 1.27.4.

- Public dependency: github.com/mejango/blossm-toolbox-go at
  v0.0.0-20261112135315-ec5d8542cda1 (commit
  ec5d8542cda1323c96ee3b071ddb52bcba859f80). No local replace.
- `make check build`: formatting, vet, race test, module verification, tidy diff
  and independent executable build passed.
- Linux amd64 Docker build passed, image digest
  sha256:d1db45d731e88d5b710b0461cdff4a4a3e40fd3b5a2f10d6cab490da51730ddf.
- `make container-smoke IMAGE=dia-go:publish` with the disposable local PostgreSQL
  fixture passed (5.651s): real HTTP session creation, retry and revocation.
- `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`: no vulnerabilities found.

Shared framework/native suite evidence belongs to the pinned Toolbox dependency.
This consumer verification does not claim a new run of every shared domain test.
GitHub Actions runs independently after push. No live cloud deployment occurred.
