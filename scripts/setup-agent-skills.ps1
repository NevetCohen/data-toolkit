[CmdletBinding()]
param(
    [string]$GlobalSkillsRoot = (Join-Path $env:USERPROFILE 'tools-hub\skills-inventory')
)

$ErrorActionPreference = 'Stop'

$repositoryRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$skillsRoot = Join-Path $repositoryRoot '.agents\skills'
$localOpenSpecRoot = Join-Path $repositoryRoot '.codex\skills'
$globalSkillsRoot = [IO.Path]::GetFullPath($GlobalSkillsRoot)

$globalSkills = @(
    'ai-ml-domain',
    'data-toolkit-domain',
    'go-programming',
    'llm-markdown-programming',
    'obsidian',
    'obsidian-artifact-creation',
    'obsidian-collaboration',
    'obsidian-domain',
    'obsidian-search-notes',
    'obsidian-update-recent-md-files-list',
    'programming-project',
    'python-programming',
    'tabular-data-programming'
)

$localSkills = @(
    'openspec-apply-change',
    'openspec-archive-change',
    'openspec-explore',
    'openspec-propose',
    'openspec-sync-specs',
    'data-toolkit-google-sheets'
)

New-Item -ItemType Directory -Path $skillsRoot -Force | Out-Null

function Set-SkillJunction {
    param(
        [Parameter(Mandatory)]
        [string]$Name,
        [Parameter(Mandatory)]
        [string]$Target
    )

    $resolvedTarget = [IO.Path]::GetFullPath($Target)
    if (-not (Test-Path -LiteralPath $resolvedTarget -PathType Container)) {
        throw "Skill target does not exist: $resolvedTarget"
    }

    $junction = Join-Path $skillsRoot $Name
    if (Test-Path -LiteralPath $junction) {
        $item = Get-Item -LiteralPath $junction -Force
        $currentTarget = if ($item.Target) {
            [IO.Path]::GetFullPath([string]$item.Target)
        } else {
            ''
        }
        if (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -and
            $currentTarget.Equals($resolvedTarget, [StringComparison]::OrdinalIgnoreCase)) {
            return
        }
        if (-not ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
            throw "Refusing to replace non-junction skill path: $junction"
        }
        Remove-Item -LiteralPath $junction -Force
    }
    New-Item -ItemType Junction -Path $junction -Target $resolvedTarget | Out-Null
}

foreach ($skill in $globalSkills) {
    Set-SkillJunction -Name $skill -Target (Join-Path $globalSkillsRoot $skill)
}
foreach ($skill in $localSkills) {
    Set-SkillJunction -Name $skill -Target (Join-Path $localOpenSpecRoot $skill)
}

Get-ChildItem -LiteralPath $skillsRoot -Force |
    Sort-Object Name |
    Select-Object Name, Target
