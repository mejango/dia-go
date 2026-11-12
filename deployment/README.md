# Dia deployment validation

Use the application validator before preparing or executing any deployment.
Toolbox schema v2 validates generic resource identifiers and rendering; it does
not know Dia's required keys, trust domain or identity namespace policy.

```sh
go run ./cmd/dia-config --config deploy/examples/gcp-postgres.json
go run ./cmd/dia-config --config deploy/examples/aws-mongo.json
go run ./cmd/dia-config --config deploy/examples/gcp-document-postgres.json
go run ./cmd/dia-config --config deploy/examples/aws-document-mongo.json
```

Without `--out`, the command only reads and validates the configuration. With
`--out /path/to/new-plan`, it calls `deployment.Prepare` and public `cli.WritePlan`
to create review artifacts. Existing output is refused by the shared writer.
There is no execute flag or runner. Any later deployment execution is a separate,
explicit action against the reviewed configuration and provider prerequisites.

Public application API:

- `Validate(cli.Config) error` runs generic validation followed by application policy.
- `ReadConfig(path) (cli.Config, error)` uses the public strict decoder and application validation.
- `Prepare(cli.Config) (cli.Plan, error)` validates before delegating all rendering to the public CLI.

Profiles are selected by exact `application.package`, either `./cmd/dia` or
`./cmd/document`. Both require database, publication, signing, NETWORK and secret
references for DIA_HMAC_KEY and DIA_ENCRYPTION_KEY. Issuer is `dia.<NETWORK>` and
audience is NETWORK, matching the native applications' trust construction.
Generic signing validation owns the local/private versus KMS/pinned-public-key
requirements. References are validated; no key or provider credential is read.
Runtime remains responsible for decoding actual secret values and verifying
32-byte key lengths and cryptographic formats.

Dia requires LOCAL_MODE=false and PUBLISHER matching the provider. Document
requires SERVICE, INTERNAL_SUBJECT, PUBLICATION_MODE matching the provider,
delivery settings, and a valid IDENTITY_DATABASE_NAMESPACE distinct from the
document database namespace. The generic namespace validator is reused. If a
static signing subject is configured, it must match INTERNAL_SUBJECT. Peer
routing uses the generic validated peers object; it does not become user authority.
Optional wallet, Twilio and updates settings require their corresponding secret
references. Endpoint contents and actual secret material still undergo native
runtime validation. This package does not replicate provider clients or runtime
parsers, and valid configuration does not establish live reachability or IAM.

The four examples preserve explicit provider resources, AWS raw SNS delivery and
queue size, application secrets and separate document delivery resources. They
are review templates, not evidence of provisioned resources.

Paired examples use one identity store: both Mongo profiles name physical database
`diaapp` and the same endpoint secret, while identity data uses namespace
`dia_identity` and document data uses namespace `documents`. Document's
IDENTITY_DATABASE_NAMESPACE is `dia_identity`. The PostgreSQL pair shares its
endpoint secret with the same namespace separation. Network, JWT trust and
identity cryptographic secret references agree within each pair.

Dia publishes to `dia-events` with consumer `dia-projection`; document publishes
to `document-events` with consumer `document-projection`. GCP subscriptions and
AWS queue URLs/ARNs are separate, so the two consumers do not compete for unrelated
events. Document delivery also uses its own `document-delivery` queue. The paired
configuration test checks these relationships in prepared artifacts; it does not
prove that externally provisioned resources match the configured identifiers.
