# Dia runtime

`Run(context.Context)` owns the Dia HTTP application lifecycle, database/signing
clients, publication workers and optional wallet RPC clients. Configuration is
supplied through environment variables; the application entrypoint is `cmd/dia`.
Context cancellation shuts down and joins workers before clients close.
`LogStopped` reports only controlled lifecycle metadata and never serializes
provider causes.

This package belongs to Dia, importing its identity/wallet services and HTTP API.
It consumes generic database, publication, signing and HTTP facilities from the
public Toolbox module. There is one application runtime implementation.

## Configuration

Required: `NETWORK`, `DATABASE_DRIVER` (`postgres` or `mongo`),
`DATABASE_NAMESPACE`, the selected `POSTGRES_URL` or `MONGODB_URI` and
`MONGODB_DATABASE`, and explicit `LOCAL_MODE`. Signing uses `JWT_SIGNER`
(`local`, `gcp`, `aws`) with its private-key or KMS/public-key configuration.
Dia chooses issuer `dia.<NETWORK>` and audience `<NETWORK>`. Supply independent
base64 32-byte `DIA_HMAC_KEY` and `DIA_ENCRYPTION_KEY`. `PORT` defaults to8080.

Remote mode requires the provider publication configuration consumed by Toolbox.
Optional communications use SMTP, Twilio or the configured updates publisher.
Wallets require explicit `EVM_RPC_URLS` and a separate `WALLET_ENCRYPTION_KEY`;
private keys remain encrypted and never reach the RPC provider. Missing adapters
fail explicitly. See services/identity and services/wallet for domain contracts.

The `cmd/document` example additionally requires a distinct
`IDENTITY_DATABASE_NAMESPACE`, `SERVICE` and `INTERNAL_SUBJECT`, with explicit outbound
peer/delivery/replay settings where enabled. It is an application example using
Dia identity, not part of the framework's infrastructure runtime.
