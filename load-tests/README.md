# Load tests

This directory contains local API load-test scripts.

## Send messages at 50 RPS by default

Run:

```powershell
.\load-tests\messages-send.hey.ps1
```

Defaults:

- `BASE_URL=http://localhost`
- `RATE=50`
- `DURATION=1m`
- `CONCURRENCY=50`
- `PUBLIC_TOKEN=demo-public-token`
- `PRIVATE_TOKEN=demo-private-token`
- `PAYLOAD_FILE=load-tests/messages-send.payload.json`
- `PROCESS=2.v2`
- `EXPERIMENT_ID=0` (if set, `experimental` is added to request payload and version split uses experiment)

Seed a split experiment for load test as part of demo seed:

```powershell
go run ./cli seed
# or
.\load-tests\seed-messages-experiment.ps1
```

Then call the load script with the returned experiment id (or with `-Experimental`):

```powershell
# if needed, fetch the experiment id from DB
psql "postgres://postgres:postgres@localhost:5432/cactus?sslmode=disable" -Atc ^
  "SELECT id FROM workflow_experiment WHERE name='load-test-version-split' AND deleted_at IS NULL ORDER BY id DESC LIMIT 1;"

# example
.\load-tests\messages-send.hey.ps1 -Experimental 1
```

Override values with parameters:

```powershell
.\load-tests\messages-send.hey.ps1 `
  -Duration 5m `
  -Rate 100 `
  -Concurrency 50 `
  -BaseUrl http://localhost `
  -PublicToken demo-public-token `
  -PrivateToken demo-private-token `
  -PayloadFile .\load-tests\messages-send.payload.json `
  -Process 2.v2 `
  -Experimental 3
```

`hey` applies `-q` per worker, not globally. The script converts total `-Rate`
to per-worker `-q` with `ceil(Rate / Concurrency)`. With the defaults this runs:

```powershell
hey -z 1m -c 50 -q 2 -m POST ...
```

With `-Rate 100` and `-Concurrency 50`, it runs `50 * 2 = 100` requests per second. Increase
`-Concurrency` if responses are slow and `hey` cannot keep the requested rate.

The request body is passed via `hey -D` from `messages-send.payload.json`.
This avoids Windows PowerShell breaking multiline JSON into separate native
command arguments.
