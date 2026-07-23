param(
    [switch]$SkipFormat
)

$ErrorActionPreference = "Stop"
$repository = Split-Path -Parent $PSScriptRoot
Push-Location $repository
try {
    if (-not $SkipFormat) {
        $goFiles = Get-ChildItem -Path cmd, configs, internal -Recurse -Filter *.go |
            ForEach-Object { $_.FullName }
        gofmt -w $goFiles
    }
    go test ./...
    go vet ./...
    go build ./cmd/data-toolkit
    openspec validate establish-data-toolkit-v1-foundation --strict
}
finally {
    Pop-Location
}
