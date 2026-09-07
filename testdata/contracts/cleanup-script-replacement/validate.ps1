$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$manifestPath = Join-Path $root 'manifest.json'
$manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
$catalog = @(
    'table.project', 'table.rename', 'row.filter', 'text.trim',
    'value.normalize', 'value.replace', 'text.regex_replace',
    'row.delete_empty', 'row.deduplicate', 'row.sort'
)

if ($manifest.schema_version -ne 'cleanup-script-replacement-fixtures-v1') {
    throw 'Unexpected manifest schema_version.'
}

$seen = @{}
$mappedCount = 0
$outsideV1Count = 0
foreach ($fixture in $manifest.fixtures) {
    if ([string]::IsNullOrWhiteSpace($fixture.id) -or $seen.ContainsKey($fixture.id)) {
        throw "Fixture IDs must be non-empty and unique: $($fixture.id)"
    }
    $seen[$fixture.id] = $true

    if ([string]::IsNullOrWhiteSpace($fixture.behavior) -or [string]::IsNullOrWhiteSpace($fixture.expected)) {
        throw "Fixture $($fixture.id) needs one behavior and one expected outcome."
    }
    if ($fixture.PSObject.Properties.Name -contains 'operation_ids') {
        if ($fixture.PSObject.Properties.Name -contains 'outside_v1') {
            throw "Fixture $($fixture.id) cannot have both operation_ids and outside_v1."
        }
        if (@($fixture.operation_ids).Count -eq 0) {
            throw "Fixture $($fixture.id) needs at least one V1 operation ID."
        }
        $mappedCount++
        foreach ($operationId in $fixture.operation_ids) {
            if ($catalog -notcontains $operationId) {
                throw "Fixture $($fixture.id) uses non-catalog operation ID $operationId."
            }
        }
    } elseif ([string]::IsNullOrWhiteSpace($fixture.outside_v1)) {
        throw "Fixture $($fixture.id) needs operation_ids or outside_v1."
    } else {
        $outsideV1Count++
    }
    if ($fixture.PSObject.Properties.Name -contains 'input') {
        $inputPath = Join-Path $root $fixture.input
        if (-not (Test-Path -LiteralPath $inputPath -PathType Leaf)) {
            throw "Fixture $($fixture.id) input is missing: $($fixture.input)"
        }
    }
}

if ($mappedCount -ne 4 -or $outsideV1Count -ne 14) {
    throw "Expected exactly 4 V1-mapped fixtures and 14 outside_v1 fixtures; found $mappedCount mapped and $outsideV1Count outside_v1."
}

Write-Output "Validated $($manifest.fixtures.Count) cleanup fixtures: exactly 4 V1-mapped and 14 outside_v1; $($seen.Count) unique IDs; catalog mapping is closed."
