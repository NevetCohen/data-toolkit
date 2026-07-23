param(
    [Parameter(Position = 0)]
    [ValidateSet("format", "vet", "unit", "race", "build", "all")]
    [string]$Task = "all"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repositoryRoot = Split-Path -Parent $PSScriptRoot

function Invoke-Go {
    param(
        [Parameter(Mandatory)]
        [string[]]$GoArguments
    )

    & go @GoArguments
    if ($LASTEXITCODE -ne 0) {
        throw "go $($GoArguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

function Invoke-Task {
    param(
        [Parameter(Mandatory)]
        [string]$SelectedTask
    )

    switch ($SelectedTask) {
        "format" { Invoke-Go -GoArguments @("fmt", "./...") }
        "vet" { Invoke-Go -GoArguments @("vet", "./...") }
        "unit" { Invoke-Go -GoArguments @("test", "./...") }
        "race" {
            $previousCgoEnabled = [Environment]::GetEnvironmentVariable("CGO_ENABLED", "Process")
            $previousCompiler = [Environment]::GetEnvironmentVariable("CC", "Process")
            $compiler = Get-Command gcc -ErrorAction SilentlyContinue
            if ($null -eq $compiler) {
                $scoopCompiler = Join-Path $env:USERPROFILE "scoop\apps\gcc\current\bin\gcc.exe"
                if (-not (Test-Path -LiteralPath $scoopCompiler -PathType Leaf)) {
                    throw "race tests require GCC; install it or make gcc available on PATH"
                }
                $compilerPath = $scoopCompiler
            } else {
                $compilerPath = $compiler.Source
            }

            try {
                [Environment]::SetEnvironmentVariable("CGO_ENABLED", "1", "Process")
                [Environment]::SetEnvironmentVariable("CC", $compilerPath, "Process")
                Invoke-Go -GoArguments @("test", "-race", "./...")
            } finally {
                [Environment]::SetEnvironmentVariable("CGO_ENABLED", $previousCgoEnabled, "Process")
                [Environment]::SetEnvironmentVariable("CC", $previousCompiler, "Process")
            }
        }
        "build" { Invoke-Go -GoArguments @("build", "./...") }
        default { throw "unsupported development task: $SelectedTask" }
    }
}

Push-Location $repositoryRoot
try {
    if ($Task -eq "all") {
        foreach ($selectedTask in @("format", "vet", "unit", "race", "build")) {
            Invoke-Task -SelectedTask $selectedTask
        }
    } else {
        Invoke-Task -SelectedTask $Task
    }
} finally {
    Pop-Location
}
