## Why

Recurring data work is currently implemented as task-specific scripts, which makes behavior inconsistent, hard to validate, and expensive to reuse. The approved Data Toolkit specification defines a deterministic local product that replaces those scripts with one versioned workflow contract and extensible Go-based engines.

## What Changes

- Establish a versioned YAML/JSON workflow contract shared by all interfaces.
- Define a canonical table model with stable row order, global column aliases, true internal nulls, sheet metadata, and configurable display formats.
- Add source mapping and preflight handling for inferred types, format variants, mixed-type columns, whitespace, formula-value warnings, file locks, and output collisions.
- Add deterministic logical operations for filtering, cross-source comparisons, Boolean expression trees, normalization, trimming, deduplication, row deletion, sorting, splitting, and text transformation.
- Define symmetric `AND`/`OR`/`XOR` behavior across two sources and scoped `NOT` behavior with configurable `not_scope: both` as the default.
- Add adapters for JSON, CSV, Excel, TXT, and Google Sheets/Drive, with immutable source snapshots, staged output publication, selectable styles, and Unicode-safe output.
- Add a Go orchestrator that validates, plans, executes, verifies, and reports every workflow under configurable resource and exception policies.
- Add an interactive terminal UI, reusable terminal commands, and Codex-facing documentation and skills over the same contract.
- Deliver and verify the three approved MVP workflows against realistic fixtures and the existing cleanup script as a behavioral source of truth.

## Capabilities

### New Capabilities

- `workflow-contract`: Versioned workflow, interface metadata, expression-tree, output, override, validation, and report schemas.
- `canonical-table-model`: Canonical tables, sheets, column identity, global aliases, types, nulls, formats, Unicode, and stable ordering.
- `configuration-policies`: Schema-validated defaults for formatting, cleaning, Boolean scope, outputs, collisions, locks, memory, and runtime behavior.
- `source-mapping-cleaning`: Preflight mapping, type/format counts, dirty-input classification, approval behavior, and targeted repairs.
- `logical-operations`: Deterministic table operations, cross-source comparisons, Boolean semantics, deduplication, splitting, sorting, and text processing.
- `file-adapters`: JSON, CSV, Excel, TXT, and Google Sheets/Drive adapters, source snapshots, displayed values, output styles, staging, and validation.
- `workflow-orchestration`: Request validation, plan construction, engine ordering, bounded execution, verification, exception reporting, and failure handling.
- `user-interfaces`: Interactive CLI/TUI, reusable terminal commands, recent workflows, pasted contracts, and Codex discovery and invocation.
- `mvp-acceptance`: Fixtures and acceptance behavior for the city filter, JSON field extraction, and cleanup-script replacement workflows.

### Modified Capabilities

- None. The project has no existing OpenSpec capability specifications.

## Impact

- Creates the initial implementation contract in the unified project and implementation repository at `C:\Users\nevet\data_toolkit`.
- Introduces a modular Go executable and packages for contracts, canonical data, operations, adapters, orchestration, configuration, reporting, CLI/TUI, and Codex integration.
- Adds Excelize, Cobra, Bubble Tea/Bubbles, `golang.org/x/text`, and the official Google Sheets/Drive Go API packages where the standard library is insufficient.
- Adds external OAuth credential handling for Google adapters; credentials are not stored in workflows or normal configuration.
- Requires fixture access to the three approved MVP sources and read-only inspection of the existing Python cleanup script and comparison outputs.
