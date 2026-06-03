param(
    [string]$BaseUrl = $(if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost" }),
    [string]$PublicToken = $(if ($env:PUBLIC_TOKEN) { $env:PUBLIC_TOKEN } else { "demo-public-token" }),
    [string]$PrivateToken = $(if ($env:PRIVATE_TOKEN) { $env:PRIVATE_TOKEN } else { "demo-private-token" }),
    [int]$Rate = $(if ($env:RATE) { [int]$env:RATE } else { 100 }),
    [int]$Concurrency = $(if ($env:CONCURRENCY) { [int]$env:CONCURRENCY } else { 50 }),
    [string]$Duration = $(if ($env:DURATION) { $env:DURATION } else { "1m" }),
    [int]$TimeoutSeconds = $(if ($env:TIMEOUT_SECONDS) { [int]$env:TIMEOUT_SECONDS } else { 20 }),
    [string]$PayloadFile = $(if ($env:PAYLOAD_FILE) { $env:PAYLOAD_FILE } else { Join-Path $PSScriptRoot "messages-send.payload.json" })
)

if ($Concurrency -lt 1) {
    throw "Concurrency must be greater than 0."
}
if (-not (Test-Path -LiteralPath $PayloadFile)) {
    throw "Payload file not found: $PayloadFile"
}

$ratePerWorker = [math]::Ceiling($Rate / $Concurrency)
$effectiveRate = $ratePerWorker * $Concurrency

if ($effectiveRate -ne $Rate) {
    Write-Warning "hey limits QPS per worker. Effective rate is $effectiveRate RPS with Concurrency=$Concurrency and -q $ratePerWorker."
}

$resolvedPayloadFile = (Resolve-Path -LiteralPath $PayloadFile).Path

hey `
    -z $Duration `
    -c $Concurrency `
    -q $ratePerWorker `
    -m POST `
    -t $TimeoutSeconds `
    -H "Content-Type: application/json" `
    -H "X-Public-Token: $PublicToken" `
    -H "X-Private-Token: $PrivateToken" `
    -D $resolvedPayloadFile `
    "$BaseUrl/api/v1/messages/send"
