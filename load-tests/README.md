# Load tests

This directory contains local API load-test scripts.

## Send messages at 100 RPS

Run:

```powershell
.\load-tests\messages-send.hey.ps1
```

Defaults:

- `BASE_URL=http://localhost`
- `RATE=100`
- `DURATION=1m`
- `CONCURRENCY=50`
- `PUBLIC_TOKEN=demo-public-token`
- `PRIVATE_TOKEN=demo-private-token`
- `PAYLOAD_FILE=load-tests/messages-send.payload.json`

Override values with parameters:

```powershell
.\load-tests\messages-send.hey.ps1 `
  -Duration 5m `
  -Rate 100 `
  -Concurrency 50 `
  -BaseUrl http://localhost `
  -PublicToken demo-public-token `
  -PrivateToken demo-private-token `
  -PayloadFile .\load-tests\messages-send.payload.json
```

`hey` applies `-q` per worker, not globally. The script converts total `-Rate`
to per-worker `-q` with `ceil(Rate / Concurrency)`. With the defaults this runs:

```powershell
hey -z 1m -c 50 -q 2 -m POST ...
```

That is 50 workers * 2 QPS per worker = 100 requests per second. Increase
`-Concurrency` if responses are slow and `hey` cannot keep the requested rate.

The request body is passed via `hey -D` from `messages-send.payload.json`.
This avoids Windows PowerShell breaking multiline JSON into separate native
command arguments.
