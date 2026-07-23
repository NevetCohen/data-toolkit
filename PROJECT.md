# Data Toolkit

Purpose: Build a local, deterministic, fast, and extensible V1 product that
replaces one-off data-manipulation scripts.

Canonical requirements:

`C:\Users\nevet\האחסון שלי\1 - אישי\Vaults\Programming_Vault\Programming\ארגז כלי דאטה\אפיון ראשוני.md`

## V1 Product Architecture

- `internal/core`: canonical identities, data types, cells, columns, tables, and
  row streams.
- `internal/logical`: explicitly registered, format-neutral operations.
- `internal/fileengine`: explicitly registered format capabilities and shared
  file-operation contracts.
- `internal/orchestrator`: workflow validation, deterministic planning, and
  registry dispatch.
- `internal/application`: the shared in-process API for CLI, TUI, and future
  Codex integration.
- `internal/interfaces`: interface clients with no product or registry logic.

The active product foundation is specified by OpenSpec change
`establish-data-toolkit-v1-foundation`. The previous MVP implementation is
preserved as a separate, non-runtime module under `legacy/mvp`.

## V1 Extension Policy

- A template is a versioned descriptor, behavior interface, constructor,
  validation, and explicit instance-owned registry.
- Extensions are compiled and composed explicitly; there are no dynamic
  plugins, `init` registration, or executable configuration.
- JSON, CSV, Excel, TXT, and Google are scaffolds only until separate changes
  specify and implement them.
- Columns are metadata. Values remain in bounded-memory row streams.

## Core Contracts

- Source files are immutable.
- Workflows are deterministic and use a closed set of operations.
- Dirty input stops execution unless the workflow explicitly defines repair or ignore behavior.
- Defaults and user preferences belong in configuration.
- Excel and Google Sheets formulas are consumed as displayed values.
- Output formatting uses selectable named styles with one configurable default.
- Sheets can be addressed explicitly within workbooks and Google spreadsheets.
- Final results include output files or query answers, validation results, exception reports, and error details.

## Target Scale

- JSON files around 40 MB.
- CSV datasets up to 500,000 rows.
- Streaming or bounded-memory execution where it materially improves performance.

## Foundation Acceptance

- `go test ./...`, `go vet ./...`, and `go build ./cmd/data-toolkit` pass in
  the root module.
- External tests add a data type, logical operation, and file format through
  implementation and explicit registration only.
- The embedded complete defaults pass the strict V1 schema and resolve to typed
  configuration.
- CLI and TUI scaffolds use the same application API.
- The root module never imports or builds `legacy/mvp`.

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
