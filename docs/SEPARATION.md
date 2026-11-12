# Independent application verification

Verified locally on2026-11-12 with Go1.27.4 and `GOWORK=off`. The application
resolves the public module
`github.com/mejango/blossm-toolbox-go@v0.0.0-20261112142659-0007d3d81d36`
through the normal module cache/checksum verification, with no local replacement.

| Gate | Evidence |
| --- | --- |
| `make check build` | PASS: four deployment profiles, boundary check, formatting, vet, unit/race, module verification/tidy diff, all three entrypoints |
| `make native` | PASS: real PostgreSQL/Mongo application, identity, wallet, HTTP and cross-service contracts |
| Reviewed outbox fix | PASS: native race `./cmd/document ./integration`, both backends and remote owner modes, exact retained identity notice; scoped vet |
| `make container-smoke` | PASS: rebuilt `/dia` and `/document`, real HTTP lifecycles |
| govulncheck v1.9.0 | No vulnerabilities found |

The full independent gates preceded the focused document outbox fix. The changed
command and integration packages and both rebuilt containers were then rerun;
domain/storage implementation did not change. The final container log records
the exact measured duration. [Raw logs and hashes](separation-results.json) and
[Go source manifest](separation-source.json) separate this evidence from the old
thin-wrapper publication record.

Remote mode requires the paired Dia runtime to publish the shared identity
outbox. The document process owns only its document outbox. Inline mode retains
a self-contained demonstration. Live GCP/AWS and blockchain operations remain
unverified and no deployment occurred.

The independent Claude Fable5.2 review and follow-up closed the concrete outbox
finding and reported no blocker within their bounded reviewed scope. Canonical
review reports are in the paired Toolbox review branch under
`docs/reviews/separation-2026-11-12`. Neither review is a full security audit.
