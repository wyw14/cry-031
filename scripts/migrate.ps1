param([string]$DatabaseUrl = $env:DATABASE_URL)
if (-not $DatabaseUrl) { throw 'DATABASE_URL is required' }
$root = Split-Path -Parent $PSScriptRoot
Get-ChildItem -LiteralPath (Join-Path $root 'migrations') -Filter '*.sql' | Sort-Object Name | ForEach-Object {
    & psql $DatabaseUrl -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
