## Context

The Data Toolkit is a new local product intended to replace one-off data-manipulation scripts. The canonical requirements are maintained in the approved Obsidian specification, while this OpenSpec change defines the first implementation-ready contract. Project instructions, OpenSpec planning artifacts, code, tests, and local skill sources are maintained together in the single repository at `C:\Users\nevet\data_toolkit`.

The product must process immutable local and Google-backed sources, preserve deterministic behavior and stable order, support Hebrew and emoji, operate within configurable local resource limits, and remain extensible without adding architectural layers. The initial scale is a roughly 40 MB JSON source and CSV files up to 500,000 rows.

## Goals / Non-Goals

**Goals:**

- Deliver one versioned workflow contract used by terminal commands, the interactive TUI, and Codex.
- Keep logical operations independent of file formats and keep file adapters independent of user interfaces.
- Provide deterministic mapping, planning, execution, validation, and reporting.
- Support bounded-memory execution and safe immutable-source file handling.
- Implement the approved JSON, CSV, Excel, TXT, and Google Sheets/Drive adapters.
- Make configuration, operations, adapters, table styles, interfaces, and workflows independently extensible through stable contracts.
- Prove the architecture through the three approved MVP workflows.

**Non-Goals:**

- Editing source files in place.
- Preserving source spreadsheet formatting.
- Executing or authoring spreadsheet formulas.
- Recalculating Excel workbooks through Excel automation in the MVP.
- Supporting time zones in the canonical model.
- Managing Apps Script through `clasp`.
- Adding a GUI, server, multi-user service, or remote execution layer in the MVP.
- Supporting arbitrary user code inside workflows.

## Decisions

### 1. One modular Go executable

The runtime SHALL be one Go executable with packages for contracts, configuration, canonical data, operations, adapters, orchestration, reporting, an application service/composition root, terminal commands, and TUI concerns. Package dependencies point inward toward contracts and canonical types; operations never import file adapters, adapters never import interface code, and only interfaces import the application boundary.

This avoids cross-language IPC, deployment drift, and duplicate schemas. A Python runtime was considered because of its data ecosystem, but it would weaken the approved Go performance boundary and reintroduce script fragmentation. Separate services were rejected because the product is local and does not need network distribution.

### 2. Versioned YAML/JSON contract with typed normalization

YAML is the human-editable workflow and configuration format. JSON is the equivalent integration format. Both decode into the same typed Go structs and pass the same versioned schema validation before planning. Free-form instructions never reach an engine.

Keeping separate interface-specific request types was rejected because it would allow terminal, TUI, and Codex behavior to diverge.

### 3. Canonical values preserve meaning separately from display

A canonical cell is a tagged value with a distinct null kind. Strings retain Unicode. Integers and JSON numbers preserve their original lexical value until an operation requests numeric interpretation; decimal arithmetic uses an exact decimal representation rather than binary floating point. Date and time values are local values without a time-zone field. Each cell can retain source provenance needed for reports without exposing adapter-specific types to operations.

The configured `"-"`, three-decimal precision, date, time, and phone formats are render defaults, not internal storage formats. Nested JSON is not flattened implicitly: a workflow projects explicit paths, and an unprojected object or array can only be preserved as canonical JSON when the target operation or adapter declares support.

Using strings for every cell was rejected because it makes type validation and deterministic comparisons unreliable. Converting every value eagerly was rejected because it risks losing exact source representation.

### 4. Row streams plus bounded spooling

Adapters expose row streams with table and column metadata. Streaming operations such as projection, filtering, renaming, normalization, trimming, and row deletion compose without materializing the whole dataset. Blocking operations such as global sort, some deduplication modes, splitting, and cross-source comparisons use bounded in-memory buffers and deterministic temporary spools under the configured memory budget.

An always-in-memory table was rejected because it cannot honor the configurable memory contract at the approved scale. A streaming-only model was also rejected because joins, global sorting, and some validations require bounded materialization.

### 5. Explicit Boolean expression trees and complement universes

Conditions are schema-validated expression trees. `AND`, `OR`, and `XOR` nodes are commutative for two-source comparisons: source order cannot alter the match set when projection is unchanged. An expression containing `NOT` carries `not_scope` normalized by the interface to `source_1`, `source_2`, or `both`; omitted user choice becomes `both` before the orchestrator receives the request.

For `not_scope: both`, the result preserves a `source_id` discriminator. Each source has a branch projection; compatible columns can share an output field, and source-only fields become canonical null in the other branch. The planner rejects an ambiguous combined projection rather than discarding data.

Inferring `NOT` from SQL-style left/right defaults was rejected because the user explicitly requires source choice and a symmetric default.

### 6. Mapping is a full preflight phase

Before planning, each adapter emits a mapping report with sheets, physical positions, original headers, global aliases, inferred types, counts, observed formats, header/subheader structure, and cleanliness findings. Mixed-type behavior is resolved through workflow/configuration policy and interface approval metadata. Leading/trailing whitespace creates a targeted `trim` repair only when detected.

Silently cleaning all values was rejected because it hides data-quality decisions. Stopping on every format variation was rejected because equivalent phone/date representations are intended to normalize automatically.

### 7. Stable deduplication and deletion semantics

The default duplicate key is the column named `#` when present; otherwise all columns form the key. Stable source order determines the default retained row (`first`), with explicit `last` and `error` alternatives. Conditional row deletion is a separate operation. Empty-row deletion cannot remove a row that contains data outside the cleaned field.

Unordered map-based deduplication was rejected because it would make retained records and output order nondeterministic.

### 8. Immutable snapshots and staged publication

Local sources are opened for shared read, copied to a validated per-run snapshot, and processed only from that snapshot. Exclusive read locks use configurable retries and then stop before data processing. The MVP does not use Volume Shadow Copy.

Outputs are written to staged files in the target filesystem, validated, and renamed or replaced only under the selected collision policy. A failed validation never publishes a partial successful output.

Reading a live changing source throughout a run was rejected because it violates determinism. Direct writes to final paths were rejected because failures could leave corrupt outputs.

### 9. Adapter choices and spreadsheet values

Standard-library streaming is preferred for JSON, CSV, and TXT. Excelize handles Excel input/output and styles. The Excel adapter reads cached computed values and warns when a formula lacks a cached result; it does not launch Excel. Google Sheets uses Sheets API v4 and Drive API with external OAuth credentials and least-privilege scopes. Displayed Sheets values are requested when the workflow asks for displayed values. `clasp` is excluded.

### 10. Configuration precedence and secrets

Precedence is built-in defaults, then user configuration, then workflow overrides, then per-output overrides. The resolved configuration is captured in the execution report. OAuth tokens and client secrets are loaded from an external credential provider and are never serialized into workflows, reports, or ordinary config files.

### 11. Shared interface services

Cobra provides reusable terminal commands. Bubble Tea and Bubbles provide the interactive TUI. Both call the same application service and workflow validator. Codex documentation and skills discover the schema and invoke the executable instead of duplicating operations or validating outputs themselves.

### 12. Vertical-slice delivery

Implementation starts with contract/config foundations and one end-to-end local slice, then broadens adapters and operations. Each approved MVP workflow becomes a committed fixture suite. The cleanup Python script is read-only behavioral source material; fixtures and equivalence assertions are derived before its replacement workflow is implemented.

## Risks / Trade-offs

- **[Inferred types are ambiguous]** → Preserve counts and evidence in mapping, require policy or approval for mixed types, and never silently coerce incompatible values.
- **[Cross-source comparisons exceed memory]** → Use deterministic partitioning or temporary spools under the memory budget and add scale fixtures before optimizing.
- **[Excel cached formula values are missing or stale]** → Warn when missing, report provenance, and document that MVP does not recalculate workbooks.
- **[Windows locks prevent snapshot or publication]** → Retry within configured limits, fail before processing when sources are unreadable, and preserve staged diagnostics without reporting success.
- **[Google OAuth broadens security exposure]** → Use least-privilege scopes, an external token store, adapter fakes for ordinary tests, and explicit integration-test activation.
- **[A broad MVP delays usable value]** → Gate delivery by vertical slices and make the first two acceptance workflows operational before the full interactive TUI.
- **[Global aliases collide or drift]** → Validate alias uniqueness and version alias configuration; reports always retain original headers and physical positions.
- **[Exact decimal handling costs performance]** → Preserve lexical values and parse exact decimals only for operations that require numeric semantics.

## Migration Plan

1. Create the Go module and package boundaries in `C:\Users\nevet\data_toolkit` without modifying source data.
2. Implement versioned contracts, configuration resolution, canonical values, provenance, reports, and fixture infrastructure.
3. Deliver a JSON/CSV vertical slice and MVP 2.
4. Add Excel snapshot, mapping, cached-value, styling, and publication support and deliver MVP 1.
5. Derive cleanup-script fixtures, add required logical operations, and deliver MVP 3.
6. Add TXT and Google Sheets/Drive adapters behind the same contracts.
7. Add terminal commands, interactive TUI, documentation, and Codex skills.
8. Record runtime and memory baselines and convert them into regression thresholds.

There is no existing runtime deployment to migrate. Project and OpenSpec artifacts were consolidated into the implementation repository before its Git baseline. Rollback consists of reverting implementation commits and generated outputs; source datasets remain untouched. Workflow schema versions must remain readable or fail with an explicit unsupported-version error.

## Open Questions

- What exact output filenames and validation assertions should become the committed contract for MVP 1 and MVP 2 after the first fixture runs?
- What safe numeric lock-retry, memory, and temporary-retention defaults does the baseline on the local machine support?
- Which behaviors and edge cases discovered in `Invoke-RawMemberCleanup.Streaming.py` must become independent reusable operations rather than parameters of an existing operation?
