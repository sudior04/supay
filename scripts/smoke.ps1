param([string]$BaseUrl = "http://localhost:8080")
$ErrorActionPreference = "Stop"
foreach ($path in @("/healthz", "/readyz", "/api/v1/users/status", "/api/v1/products/status", "/api/v1/wallets/status")) {
 $response = Invoke-WebRequest -UseBasicParsing -Uri ($BaseUrl.TrimEnd("/") + $path)
 if ($response.StatusCode -ne 200) { throw "Smoke failed: $path" }
 Write-Host "OK $path"
}
Write-Host "Health/status routes reachable. Run scripts/smoke-catalog-pin.ps1 for PIN and catalog checks."
