# Test Wholesale Endpoints
# RESPONSIBILITY: Test batch MPQ and wholesale reset endpoints (Windows PowerShell)

param(
    [string]$ApiUrl = "http://localhost:3000",
    [string]$TenantId = "yumna_bertigamart",
    [string]$JwtToken
)

if ([string]::IsNullOrEmpty($JwtToken)) {
    Write-Host "ERROR: JwtToken parameter is required" -ForegroundColor Red
    Write-Host "Usage: .\test_wholesale_endpoints.ps1 -JwtToken 'your_token_here'"
    exit 1
}

Write-Host "`nTesting Wholesale Endpoints" -ForegroundColor Yellow
Write-Host "API URL: $ApiUrl"
Write-Host "Tenant: $TenantId`n"

$headers = @{
    "Authorization" = "Bearer $JwtToken"
    "x-tenant-id" = $TenantId
    "Content-Type" = "application/json"
}

# Test 1: GET /api/wholesale/settings
Write-Host "Test 1: GET /api/wholesale/settings" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/api/wholesale/settings" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: POST /api/wholesale/shopee/batch-mpq
Write-Host "Test 2: POST /api/wholesale/shopee/batch-mpq" -ForegroundColor Yellow
try {
    $body = @{
        items = @(
            @{sku = "TEST-SKU-001"; price = 50000}
        )
        mpq = 5
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$ApiUrl/api/wholesale/shopee/batch-mpq" -Method Post -Headers $headers -Body $body
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: POST /api/wholesale/shopee/batch-delete-skus
Write-Host "Test 3: POST /api/wholesale/shopee/batch-delete-skus" -ForegroundColor Yellow
try {
    $body = @{
        skus = @("TEST-SKU-001")
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$ApiUrl/api/wholesale/shopee/batch-delete-skus" -Method Post -Headers $headers -Body $body
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: POST /api/wholesale/shopee/batch-wholesale-reset
Write-Host "Test 4: POST /api/wholesale/shopee/batch-wholesale-reset" -ForegroundColor Yellow
try {
    $body = @{
        items = @(
            @{sku = "TEST-SKU-001"; price = 50000}
        )
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$ApiUrl/api/wholesale/shopee/batch-wholesale-reset" -Method Post -Headers $headers -Body $body
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

Write-Host "Testing completed!" -ForegroundColor Green
