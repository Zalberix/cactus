param(
    [string]$BaseUrl = $(if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost" }),
    [string]$PublicToken = $(if ($env:PUBLIC_TOKEN) { $env:PUBLIC_TOKEN } else { "demo-public-token" }),
    [string]$PrivateToken = $(if ($env:PRIVATE_TOKEN) { $env:PRIVATE_TOKEN } else { "demo-private-token" }),
    [int]$Rate = $(if ($env:RATE) { [int]$env:RATE } else { 50 }),
    [int]$Concurrency = $(if ($env:CONCURRENCY) { [int]$env:CONCURRENCY } else { 50 }),
    [string]$Duration = $(if ($env:DURATION) { $env:DURATION } else { "1m" }),
    [int]$TimeoutSeconds = $(if ($env:TIMEOUT_SECONDS) { [int]$env:TIMEOUT_SECONDS } else { 20 }),
    [string]$PayloadFile = $(if ($env:PAYLOAD_FILE) { $env:PAYLOAD_FILE } else { Join-Path $PSScriptRoot "messages-send.payload.json" }),
    [string]$Process = $(if ($env:PROCESS) { $env:PROCESS } else { "2.v2" }),
    [int]$Experimental = $(if ($env:EXPERIMENT_ID) { [int]$env:EXPERIMENT_ID } else { 0 }),
    [switch]$ShowResponse = [bool]($env:SHOW_RESPONSE -eq "1" -or $env:SHOW_RESPONSE -eq "true" -or $env:SHOW_RESPONSE -eq "yes")
)

if ($Concurrency -lt 1) {
    throw "Concurrency must be greater than 0."
}
if (-not (Test-Path -LiteralPath $PayloadFile)) {
    throw "Payload file not found: $PayloadFile"
}

$payload = Get-Content -Path $PayloadFile -Raw -Encoding UTF8 | ConvertFrom-Json
$payload | Add-Member -Name process -Value $Process -MemberType NoteProperty -Force
if ($Experimental -gt 0) {
    $payload | Add-Member -Name experimental -Value $Experimental -MemberType NoteProperty -Force
} else {
    $null = $payload.PSObject.Properties.Remove("experimental")
}

$generatedPayloadFile = Join-Path $env:TEMP ("messages-send.payload.{0}.json" -f ([Guid]::NewGuid().ToString("N")))
$payloadJson = $payload | ConvertTo-Json -Depth 20

Write-Host "Payload that will be sent:" -ForegroundColor Cyan
Write-Host $payloadJson

$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($generatedPayloadFile, $payloadJson, $utf8NoBom)

if ($ShowResponse) {
    Write-Host ""
    Write-Host "Request response preview (single call):" -ForegroundColor Cyan
    $requestUrl = "$BaseUrl/api/v1/messages/send"
    try {
        $response = Invoke-WebRequest -Method POST -Uri $requestUrl -UseBasicParsing -Headers @{
            "Content-Type"  = "application/json"
            "X-Public-Token" = $PublicToken
            "X-Private-Token" = $PrivateToken
        } -Body $payloadJson -TimeoutSec $TimeoutSeconds
        Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
        try {
            $json = $response.Content | ConvertFrom-Json
            Write-Host ($json | ConvertTo-Json -Depth 20)
        } catch {
            Write-Host $response.Content
        }
    } catch {
        Write-Host "Status: request error" -ForegroundColor Red
        $responseError = $_.Exception.Response
        if ($responseError -ne $null) {
            Write-Host "HTTP status: $($responseError.StatusCode.value__)" -ForegroundColor Red
            try {
                $reader = New-Object System.IO.StreamReader($responseError.GetResponseStream())
                $errorBody = $reader.ReadToEnd()
                $reader.Close()
                if (-not [string]::IsNullOrWhiteSpace($errorBody)) {
                    try {
                        $errorJson = $errorBody | ConvertFrom-Json
                        Write-Host ($errorJson | ConvertTo-Json -Depth 20)
                    } catch {
                        Write-Host $errorBody
                    }
                } else {
                    Write-Host $_.Exception.Message
                }
            } catch {
                Write-Host $_.Exception.Message
            }
        } else {
            Write-Host $_.Exception.Message
        }
    }
}

$ratePerWorker = [math]::Ceiling($Rate / $Concurrency)
$effectiveRate = $ratePerWorker * $Concurrency

if ($effectiveRate -ne $Rate) {
    Write-Warning "hey limits QPS per worker. Effective rate is $effectiveRate RPS with Concurrency=$Concurrency and -q $ratePerWorker."
}

try {
    hey `
        -z $Duration `
        -c $Concurrency `
        -q $ratePerWorker `
        -m POST `
        -t $TimeoutSeconds `
        -H "Content-Type: application/json" `
        -H "X-Public-Token: $PublicToken" `
        -H "X-Private-Token: $PrivateToken" `
        -D $generatedPayloadFile `
        "$BaseUrl/api/v1/messages/send"
} finally {
    Remove-Item -LiteralPath $generatedPayloadFile -Force -ErrorAction SilentlyContinue
}
