param([string]$DatabaseUrl = $env:DATABASE_URL)
if (-not $DatabaseUrl) { throw 'DATABASE_URL is required' }
$env:STORE_MODE = 'postgres'
$env:SEED_DEMO = 'true'
$env:DATABASE_URL = $DatabaseUrl
go run ./cmd/server
