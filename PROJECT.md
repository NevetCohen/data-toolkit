# Data Toolkit

Purpose: build a local, deterministic, extensible V1 product that replaces
one-off data-manipulation scripts without modifying source data.

Canonical product brief:

`C:\Users\nevet\האחסון שלי\1 - אישי\Vaults\Programming_Vault\Programming\ארגז כלי דאטה\אפיון מחודש 04-08-2026.md`

The active implementation-ready contract is OpenSpec change
`implement-revised-data-toolkit-v1`. The initial brief and the archived
foundation proposal remain historical inputs and are not changed by this work.

## V1 product architecture

- Infrastructure: identities, immutable `Schema + RowStream + Value`,
  registries, configuration, resources, reports, and receipts.
- Reader Engine: inspection, shared-read lock, immutable snapshot/hash, decode,
  and source schema/dirty-input validation.
- Logical Engine: explicitly registered, format-neutral operations in declared
  request order.
- Writer Engine: staged rendering, optional Excel style, reopen validation,
  assertions, atomic publish, and finalization.
- Orchestrator: validation, configuration resolution, deterministic planning,
  ordering, and engine dispatch only.
- CLI: strict JSON workflow files, typed variables, config patches, JSON API
  envelopes, and exit classes.
- Codex Adapter: runtime discovery, template construction, run/recovery
  procedure, and the Google Sheets XLSX bridge.

The existing Go foundation is intentionally not reclassified as a finished V1
implementation by this documentation change. Future implementation will align
its packages with these contracts.

## Core contracts

- Sources are immutable; Reader validates and hashes a snapshot before Logical
  Engine receives values.
- The public model is only `Schema + RowStream + Value`; batching is an optional
  internal optimization.
- Workflows are deterministic and come from a strict file-backed template.
- CLI variable substitution precedes strict workflow validation and digest.
- Configuration precedence is defaults, user config, CLI patches, workflow
  overrides, then output overrides; protected paths cannot be patched.
- Writer validates the reopened staging artifact and all user assertions before
  atomic publication. After workspace creation, it finalizes every terminal
  path, returns the complete terminal `RunReport` through the Application API,
  and appends exactly one receipt.
- Logical work is bounded by resource reservations/spill. A nonrecoverable
  limit returns `resource_limit_exceeded` and cannot publish output.
- Google Sheets is not a Go format. The sixth Codex Skill uses the verified
  Google Drive connector to export/import local XLSX, resolves remote title
  collisions before import, and never deletes a Drive file.

## V1 scope

Local formats are CSV, JSON, and Excel through separate Reader/Writer
capabilities. The V1 logical catalog contains ten operations including keyed
deduplication and stable multi-key sorting. Google Sheets is an agent-mediated
XLSX bridge only. Joins, queries, TXT, TUI, remote overwrite, runtime plugins,
and runtime credentials are out of scope.

## Quality and documentation

Run `openspec validate implement-revised-data-toolkit-v1 --strict` for the
active contract. Codex agents building workflows must read
`docs/codex/workflow-json-format.md`; Google Sheet requests also use
`.agents/skills/data-toolkit-google-sheets/SKILL.md` once implemented.

The repository at `C:\Users\nevet\data_toolkit` contains code, tests,
OpenSpec changes, fixtures, documentation, and project-local skills. Do not
modify `legacy/mvp` as part of V1 runtime work.
