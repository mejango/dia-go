# Independent hierarchy verification, 2026-11-12

Dia pins the public Toolbox module `v0.0.0-20261112155157-d45c8b4e036c`. All final gates ran with
GOWORK=off, no replace directive and no committed workspace, using Go 1.27.4.
Module resolution and checksums are recorded in the verified-pin log.

Passed: all four application configuration profiles; ownership boundary;
formatting; vet; full unit/race suites; module verify/tidy-diff; builds of dia,
document and dia-config; native PostgreSQL and MongoDB; both application container
lifecycles; and the pinned vulnerability checker (no vulnerabilities found).
Native identity passed in 80.398s, wallet in 20.620s, integration in 7.597s and
document ownership in 3.255s. Container lifecycles passed in 2.371s. Final image:
sha256:2cb26f92714749bef8feb3c284981a5e81c71260cc9fcd4f7e3676614e2fcfdd.

The initial native run exposed Mongo's short known-abort retry budget and a test
that failed before joining its workers. The fixture now drains every result and
still requires every concurrent request to succeed. The framework's separate
bounded jitter/backoff fix preserves uncertain-commit and single-attempt-read
semantics. Ten repetitions of the real concurrent Mongo journey passed before
this final independent run. The initial failure is archived, not relabeled as a
passing gate. No authentication, domain, custody or wire contract was changed.

Exact logs are in verification-logs/hierarchy; files named verified record this
final public pin. hierarchy-source.json hashes current source/build inputs.
Earlier check/pin logs are intermediate checkpoints. The previous source-split
records remain unchanged. No live deployment/provider/chain verification is claimed.

Reproduce with the pinned disposable fixtures in deploy/compose.integration.json,
initialize Mongo using .github/workflows/verify.yml, set POSTGRES_URL and
MONGODB_URI to localhost fixtures, and run:

```sh
GOWORK=off make check build native
GOWORK=off make container IMAGE=dia-go:hierarchy
GOWORK=off go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

For a differently named Docker fixture project, pass CONTAINER_NETWORK and
CONTAINER_POSTGRES_URL to make container. Dia remains an ordinary third-party
consumer and all application models, policy and runtime remain in this repository.
