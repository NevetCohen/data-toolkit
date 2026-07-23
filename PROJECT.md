# Data Toolkit

Purpose: Develop a local, deterministic, fast, and extensible toolkit that replaces one-off data-manipulation scripts.

Canonical requirements:

`C:\Users\nevet\האחסון שלי\1 - אישי\Vaults\Programming_Vault\Programming\ארגז כלי דאטה\אפיון ראשוני.md`

## Product Architecture

- Go logical engine: filtering, mappings, comparisons, normalization, splitting, transformations, and data queries.
- File engine: adapters between files and the canonical table model.
- Workflow orchestrator: validates requests, builds an execution plan, runs operations, verifies outputs, and reports errors and exceptions.
- Interfaces: Cobra terminal commands, an interactive Bubble Tea TUI, and Codex integration through documentation and skills.

## Initial Formats

- JSON
- CSV
- Excel
- TXT
- Google Sheets through an explicit Google Drive integration

## Core Contracts

- Source files are immutable.
- Workflows are deterministic and use a closed set of operations.
- Dirty input stops execution unless the workflow explicitly defines repair or ignore behavior.
- Defaults and user preferences belong in configuration.
- Excel and Google Sheets formulas are consumed as displayed values.
- Output formatting uses selectable named styles with one configurable default.
- Sheets can be addressed explicitly within workbooks and Google spreadsheets.
- Final results include output files or query answers, validation results, exception reports, and error details.

## Initial Scale

- JSON files around 40 MB.
- CSV datasets up to 500,000 rows.
- Streaming or bounded-memory execution where it materially improves performance.

## MVP Acceptance Workflows

1. Filter the `ישוב` column in:
   `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\all_final.xlsx`

2. Extract selected fields to CSV from:
   `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\ext_members_from_12072026_to_14072026_no_full_phone_2026-07-14-12-35-21.json`

3. Replace the behavior of:
   `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py`

## Repository

`C:\Users\nevet\data_toolkit` is the single repository and source of truth for
project instructions, OpenSpec planning, implementation code, tests, fixtures,
documentation, and repository-local skills. Machine-local skill junctions under
`.agents\skills` are recreated by `scripts\setup-agent-skills.ps1` and are not
versioned.

## Local Skills

- `ai-ml-domain`
- `data-toolkit-domain`
- `go-programming`
- `llm-markdown-programming`
- `obsidian`
- `obsidian-artifact-creation`
- `obsidian-collaboration`
- `obsidian-domain`
- `obsidian-search-notes`
- `obsidian-update-recent-md-files-list`
- `openspec-apply-change`
- `openspec-archive-change`
- `openspec-explore`
- `openspec-propose`
- `openspec-sync-specs`
- `programming-project`
- `python-programming`
- `tabular-data-programming`
