param(
    [string]$DatabaseUrl = $(if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://postgres:postgres@localhost:5432/cactus?sslmode=disable" }),
    [string]$PsqlPath = $(if ($env:PSQL_PATH) { $env:PSQL_PATH } else { "psql" }),
    [switch]$SkipExperimentLookup
)

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "go is required to seed demo data."
}

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    & go run ./cli seed
}
finally {
    Pop-Location
}

if ($LASTEXITCODE -ne 0) {
    throw "Demo seed command failed."
}

if ($SkipExperimentLookup) {
    return
}

if (-not (Get-Command $PsqlPath -ErrorAction SilentlyContinue)) {
    Write-Warning "psql not found: $PsqlPath. Skipping experiment lookup."
    return
}

$query = "SELECT id FROM workflow_experiment WHERE name = 'load-test-version-split' AND deleted_at IS NULL ORDER BY id DESC LIMIT 1;"
$experimentId = & $PsqlPath -v ON_ERROR_STOP=1 -X -A -t -d $DatabaseUrl -c $query

if ([string]::IsNullOrWhiteSpace($experimentId)) {
    throw "Experiment created by seed not found. Check seed status."
}

Write-Host "load-test-version-split experiment id: $experimentId"
