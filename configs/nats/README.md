# Local NATS TLS/JWT files

Generate local TLS certs and JWT/account credentials outside source control:

```bash
go run ./cli nats-dev-assets
```

Required outputs:

- `configs/nats/certs/ca.pem`
- `configs/nats/certs/server.pem`
- `configs/nats/certs/server-key.pem`
- `configs/nats/jwt/operator.jwt`
- `configs/nats/jwt/account.jwt`
- `configs/nats/jwt/account.seed`
- `configs/nats/creds/sys.creds`
- `configs/nats/creds/core.creds`

Core reads `configs/nats/jwt/account.seed` by default for local development.
Restart the NATS container after regenerating these files.
