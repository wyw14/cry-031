$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    gofmt -w ./cmd ./internal ./tests
    go build ./...
    go test ./...
    go vet ./...
    Push-Location ./web
    try {
        npm test
        npm run build
    } finally { Pop-Location }
} finally { Pop-Location }
