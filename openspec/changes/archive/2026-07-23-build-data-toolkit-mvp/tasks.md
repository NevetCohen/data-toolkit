## Execution Metadata for Remaining Work

Each unchecked task has one machine-readable execution line:

`Execution: model=<model>; effort=<effort>; wave=<wave>; depends_on=[<task ids>]`

- `gpt-5.6-terra` is the bounded implementation worker; `gpt-5.6-sol` owns architecture, subtle semantics, security, bounded-memory algorithms, and final integration.
- Effort is the minimum recommended starting level, not a promise that a harder task will never require escalation.
- `ultra` is reserved for coordinating an entire parallel wave. It is not assigned as an individual task effort because it is a multi-agent mode rather than a single-agent reasoning level.
- Tasks in one wave may run concurrently only when they have different file/package owners. Arrows in the wave table are serialized subchains.
- The evidence and cost basis for these decisions is documented in `GPT-5.6_PROGRAMMING_CAPABILITY_AND_ROUTING_GUIDE.md`.

### Section Dependency Graph

| Section | Requires | Unlocks |
| --- | --- | --- |
| 7. JSON and CSV Vertical Slice | Sections 1–6 | 15.2 and the JSON/CSV part of 15.3 |
| 8. Mapping and Cleaning | Sections 1–6 | 9.4, 11.7, mapping-conformant adapters, and interface mapping flows |
| 9. Core Logical Operations | Sections 1–6; 9.4 also requires 8.5 and 8.7 | Section 10, MVP 1, MVP 3, and executable interfaces |
| 10. Cross-Source Boolean Operations | 10.1, 9.1, 9.2, and 9.3 | Cross-source TUI behavior and interface equivalence |
| 11. Excel Adapter and MVP 1 | Adapter foundation; MVP 1 also requires 8.7 and 9.3 | Excel baseline in 15.3 |
| 12. Cleanup-Script Replacement and MVP 3 | 12.1–12.3, selected Section 8/9 operations, and TXT support | Cleanup baseline in 15.3 |
| 13. TXT and Google Adapters | Adapter foundation; mapping readers require 8.2 | TXT-dependent MVP 3 and complete adapter surface |
| 14. Terminal, TUI, and Codex Interfaces | Stable mapping, operation, adapter, and orchestration contracts | 15.6 and 15.8 |
| 15. Acceptance, Performance, and Completion Gates | Runs incrementally; final gates require Sections 7–14 | Change completion |

### Parallel Execution Waves

| Wave | Tasks that may run in parallel |
| --- | --- |
| W0 | 8.1; 10.1; 11.6; 12.1; 13.3; 13.7; 15.2 |
| W1 | 8.2; 9.1; 11.1; 12.2 |
| W2 | 8.3; 9.2; 9.3; 11.2; 12.3; 13.1; 13.4 |
| W3 | Mapping owner: 8.4 → 8.5 → 8.6; operation owners: 9.5, 9.6, 9.8, 9.10; 11.3; 13.2; 13.5 |
| W4 | 8.7; operation owners: 9.7, 9.9, 9.11; 10.2; 11.4; 13.6 |
| W5 | 9.4; 10.3; 11.5; 12.4; 14.1; 14.4 |
| W6 | 10.4; 11.7; 12.5; 14.2; 14.5 |
| W7 | 10.5; 12.6; 14.3; 14.6; 15.3 |
| W8 | 10.6; 14.7; 15.4 |
| W9 | 14.8; 15.5 |
| W10 | 14.9; 15.1 |
| W11 | 15.6 |
| W12 | 15.7 |
| W13 | 15.8 |

Conflict controls:

- Freeze the shared adapter capability contract after W1; format owners use separate JSON/CSV, Excel, TXT, and Google implementation files.
- Assign one owner per wave to central operation-registry edits. Independent operation implementations may use separate files.
- Assign one orchestration owner to the mapping/clean-handoff wiring in 8.7.
- Serialize durable documentation work as 13.7 → 14.8 → 15.8.
- Task 12.3 is a decision gate. If the reference script requires a reusable operation not already specified, add that explicit task before 12.4.

## 1. Repository and Architecture Foundation

- [x] 1.1 Inspect `C:\Users\nevet\data_toolkit`, read any local instructions, and record the pre-existing state before adding files.
  - Pre-existing state recorded on 2026-07-19: the directory exists and is empty, with no local instructions, README, Go module, or Git repository.
- [x] 1.2 Initialize or verify the Go module and pin the approved Go toolchain without modifying unrelated existing files.
- [x] 1.3 Create package boundaries for contracts, configuration, canonical data, operations, adapters, orchestration, reporting, CLI, and TUI with dependency-boundary tests or checks.
- [x] 1.4 Add and pin only the required external modules: Excelize, Cobra, Bubble Tea/Bubbles, `golang.org/x/text`, exact-decimal support, YAML/schema validation, and Google API clients.
- [x] 1.5 Add formatting, vet, unit-test, race-test, and build commands that run without live Google credentials.
- [x] 1.6 Create read-only fixture conventions, generated-output directories, and source-immutability helpers for local acceptance tests.
- [x] 1.7 Add architecture documentation that identifies the planning source of truth and the approved implementation repository boundary.

## 2. Workflow and Result Contracts

- [x] 2.1 Define versioned Go structs for sources, sheets, operations, outputs, queries, validations, overrides, interface metadata, and run requests.
- [x] 2.2 Define YAML and JSON decoding into the same normalized request and add equivalence tests.
- [x] 2.3 Implement strict schema validation with path-specific errors for unknown versions, operations, fields, and enum values.
- [x] 2.4 Define recursive Boolean expression nodes for comparisons, `AND`, `OR`, `XOR`, and `NOT` with explicit grouping tests.
- [x] 2.5 Normalize cross-source `NOT` requests to explicit `source_1`, `source_2`, or `both` scope and test the `both` default.
- [x] 2.6 Define output projection, branch projection, per-output override, and query-result schemas.
- [x] 2.7 Define started and final result contracts covering outputs, query answers, resolved settings, validations, warnings, exceptions, and errors.
- [x] 2.8 Add round-trip fixtures for supported YAML/JSON workflows and rejection fixtures for ambiguous or incomplete requests.

## 3. Configuration and Secret Boundaries

- [x] 3.1 Define the versioned configuration schema and precedence from built-ins through user, workflow, and per-output settings.
- [x] 3.2 Implement the approved null, decimal, date, time, phone, UTF-8, dedup-keep, and `not_scope: both` defaults with unit tests.
- [x] 3.3 Implement global alias and named-style configuration with uniqueness and default-style validation.
- [x] 3.4 Implement output directory, filename template, TXT delimiter, collision, mixed-type, exception-detail, and repair settings.
- [x] 3.5 Implement maximum-memory, lock-retry, temporary-workspace, failed-run-retention, and recent-workflow settings without hard-coded user preferences.
- [x] 3.6 Capture the resolved non-secret configuration in run reports and test every precedence layer.
- [x] 3.7 Define an external credential-provider interface and add redaction tests proving secrets never enter workflows, config snapshots, logs, reports, or errors.

## 4. Canonical Data Model

- [x] 4.1 Implement tagged canonical values with distinct null, Unicode string, Boolean, exact integer/decimal, local date, local time, and supported canonical JSON kinds.
- [x] 4.2 Preserve raw numeric lexemes and implement exact numeric comparison and formatting tests without binary-float loss.
- [x] 4.3 Implement sheet, table, column, row, cell, and provenance structures with stable source and row identity.
- [x] 4.4 Implement physical position, original header, global alias, type authority, header, and subheader metadata with validation.
- [x] 4.5 Implement null rendering separately from internal null and test coexistence with the literal string `"-"`.
- [x] 4.6 Define explicit nested-JSON path projection and canonical-JSON preservation rules with no silent flattening.
- [x] 4.7 Implement stable row-stream interfaces and bounded spool primitives for blocking operations.
- [x] 4.8 Add Unicode, Hebrew, emoji, stable-order, and provenance unit fixtures across canonical serialization boundaries.

## 5. Adapter and File-Safety Foundation

- [x] 5.1 Define the adapter registry and capability contract for inspection, mapping, reading, writing, validation, and format support.
- [x] 5.2 Implement shared-read source opening, configurable retries, and precise Windows lock errors.
- [x] 5.3 Implement validated per-run immutable snapshots and provenance checks showing that processing uses the snapshot.
- [x] 5.4 Implement staged output files in the target filesystem and publish only after validation.
- [x] 5.5 Implement `overwrite`, deterministic `alternate_name`, and `block` collision behavior, including locked-output tests.
- [x] 5.6 Reject every workflow whose output resolves to a local or Google source identity regardless of collision settings.
- [x] 5.7 Implement temporary-artifact cleanup and retention policy with cancellation and failed-validation tests.

## 6. Orchestration and Reporting Core

- [x] 6.1 Implement contract and configuration validation as the first orchestration stage.
- [x] 6.2 Implement multi-source mapping orchestration and require all mapping reports before final planning.
- [x] 6.3 Implement plan steps that identify engine, input, output, repairs, resource budget, and validations.
- [x] 6.4 Implement started-status emission only after blocking preflight checks succeed.
- [x] 6.5 Implement deterministic plan execution with cancellation, bounded spooling, and maximum-memory enforcement.
- [x] 6.6 Implement validation gates that prevent failed outputs from being reported or published as successful.
- [x] 6.7 Implement ordered exception reports with source, sheet, row, column, observed class, value reference, and reason.
- [x] 6.8 Implement combined file/query final results and tests for success, warnings, partial exceptions, and hard failure.

## 7. JSON and CSV Vertical Slice

> Section dependency: requires Sections 1–6; unlocks scale testing in 15.2 and the JSON/CSV baseline in 15.3.

- [x] 7.1 Implement streaming JSON inspection and row reading with exact numbers, Unicode, explicit paths, and stable order.
- [x] 7.2 Implement JSON mapping for top-level and explicitly projected nested fields, missing values, and type counts.
- [x] 7.3 Implement streaming CSV inspection and row reading with header mapping, Unicode, missing fields, and stable order.
- [x] 7.4 Implement UTF-8 CSV writing with explicit projection order and structural validation.
- [x] 7.5 Connect JSON read, canonical projection, CSV write, reporting, and staged publication through the orchestrator.
- [x] 7.6 Derive a read-only representative fixture for the approved JSON source and define exact expected headers, rows, values, and source checks.
- [x] 7.7 Implement and pass MVP 2 for `שם פרטי`, `שם מלא`, `entityID`, `TZ`, `ישוב`, and `מייל` under the resolved memory budget.
  - Verified on 2026-07-19: focused MVP 2 tests pass against the approved 4,665-row source under a 128 MiB memory budget.

## 8. Mapping and Cleaning

> Section dependency: requires Sections 1–6; unlocks normalization integration, adapter mapping conformance, MVP 1, and interface mapping flows.

- [x] 8.1 Implement basic mapping for source, sheet, physical position, original header, global alias, and resolved type.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W0; depends_on=[5.1, 6.2]`.
- [x] 8.2 Implement extended mapping counts for values, inferred types, observed formats, headers/subheaders, and cleanliness findings.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W1; depends_on=[8.1]`.
- [ ] 8.3 Implement user-declared type authority and validated inference fallback with provenance.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W2; depends_on=[8.1, 8.2]`.
- [ ] 8.4 Implement mixed-type block, approval, exception-report, and ignore policies with stable-order tests.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W3; depends_on=[8.2, 8.3, 3.4]`.
- [ ] 8.5 Implement phone, date, time, and numeric format-variant detection and pre-logical normalization planning.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W3; depends_on=[8.2, 8.3, 3.2]`.
- [ ] 8.6 Implement leading/trailing whitespace detection and targeted `trim` repair while preserving internal whitespace.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W3; depends_on=[8.2]`.
- [ ] 8.7 Add a clean-handoff gate proving unresolved blocking findings never reach logical operations.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W4; depends_on=[8.4, 8.5, 8.6, 6.1, 6.2]`.

## 9. Core Logical Operations

> Section dependency: requires Sections 1–6; normalization integration also requires 8.5 and 8.7. Unlocks cross-source operations, MVP workflows, and executable interfaces.

- [x] 9.1 Implement the versioned operation registry and reject undeclared operations or parameters.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W1; depends_on=[2.3, 6.3]`.
- [x] 9.2 Implement output projection and the separate deterministic rename-fields operation.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W2; depends_on=[9.1, 2.6]`.
  - Verified on 2026-07-23: projection remained adapter-owned, rename moved into `internal/operation`, collision and provenance tests pass, and the orchestrator contains no row-transformation implementation.
- [ ] 9.3 Implement single-source Boolean filtering over explicit expression trees with stable-order tests.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W2; depends_on=[9.1, 2.4, 4.7]`.
- [ ] 9.4 Implement format normalization for approved date, time, phone, decimal, and configured formats with transformation reporting.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W5; depends_on=[8.5, 8.7, 9.1, 3.2]`.
- [ ] 9.5 Implement explicit string concatenation and regex match/replace operations with Unicode tests.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W3; depends_on=[9.1, 4.8]`.
- [ ] 9.6 Implement deduplication by `#` when present and full-row equality otherwise.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W3; depends_on=[9.1, 4.7]`.
- [ ] 9.7 Implement stable `first`, explicit `last`, and `error` dedup retain policies with duplicate reports.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W4; depends_on=[9.6, 3.2]`.
- [ ] 9.8 Implement wholly-empty-row deletion without deleting partially populated rows.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W3; depends_on=[9.1, 4.5]`.
- [ ] 9.9 Implement the separate delete-rows-by-condition operation using the shared expression schema.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W4; depends_on=[9.1, 9.3]`.
- [ ] 9.10 Implement stable sorting with deterministic tie behavior and bounded external spooling.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W3; depends_on=[9.1, 4.7, 6.5]`.
- [ ] 9.11 Implement splitting into declared output partitions by simple or nested Boolean criteria with count validation.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W4; depends_on=[9.1, 9.3, 2.6, 6.6]`.

## 10. Cross-Source Boolean Operations

> Section dependency: requires 10.1, 9.1, 9.2, and 9.3; unlocks cross-source TUI behavior and interface equivalence.

- [x] 10.1 Inspect the read-only `csv_compare` reference and record reusable comparison semantics and fixtures without copying format logic into the logical engine.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W0; depends_on=[]`.
- [ ] 10.2 Implement bounded cross-source comparison indexes or partitions under the configured memory budget.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W4; depends_on=[10.1, 9.1, 9.3, 4.7, 6.5]`.
- [ ] 10.3 Implement and test source-order equivalence for `AND`, `OR`, and `XOR` after reference remapping.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W5; depends_on=[10.2, 9.3, 2.4]`.
- [ ] 10.4 Implement `NOT` complements for `source_1`, `source_2`, and `both` with stable source identity.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W6; depends_on=[10.2, 10.3, 2.5, 4.3]`.
- [ ] 10.5 Implement compatible and branch-specific projections for `not_scope: both`, filling absent branch fields with canonical null.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W7; depends_on=[10.4, 9.2, 4.5]`.
- [ ] 10.6 Reject ambiguous combined projections and add fixtures proving that source-specific data is never silently discarded.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W8; depends_on=[10.5, 2.6]`.

## 11. Excel Adapter and MVP 1

> Section dependency: adapter work requires the Section 5 foundation; MVP 1 additionally requires 8.7 and 9.3. Unlocks the Excel baseline in 15.3.

- [x] 11.1 Implement Excel workbook inspection, explicit sheet selection, physical column positions, and stable row streaming with Excelize.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W1; depends_on=[5.1, 5.3, 4.7]`.
- [ ] 11.2 Implement merged whole-table title exclusion and header/subheader preservation.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W2; depends_on=[11.1, 4.4]`.
- [ ] 11.3 Read cached computed formula values, emit warnings for missing cached results, and never launch Excel to recalculate.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W3; depends_on=[11.1, 6.7]`.
- [ ] 11.4 Implement bounded Excel writing with Unicode, declared columns, configurable display formats, and named table styles.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W4; depends_on=[11.1, 3.3, 4.7, 5.4]`.
- [ ] 11.5 Validate staged Excel structure, sheets, headers, row counts, displayed values, and style selection before publication.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W5; depends_on=[11.4, 5.4, 6.6]`.
- [x] 11.6 Derive a read-only fixture and exact expected assertions from the approved `all_final.xlsx` source.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W0; depends_on=[]`.
- [ ] 11.7 Implement and pass MVP 1 filtering `ישוב == בת ים` with stable order, valid styling, and source-immutability checks.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W6; depends_on=[8.7, 9.3, 11.1, 11.2, 11.3, 11.4, 11.5, 11.6]`.

## 12. Cleanup-Script Replacement and MVP 3

> Section dependency: requires 12.1–12.3, the selected Section 8/9 operations, and TXT support from 13.1–13.2. Unlocks the cleanup baseline in 15.3.

- [x] 12.1 Read `Invoke-RawMemberCleanup.Streaming.py` and its comparison outputs and document every externally observable transformation and edge case.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W0; depends_on=[]`.
- [x] 12.2 Derive minimal, representative, immutable input/output fixtures and explicit equivalence assertions from the script.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W1; depends_on=[12.1]`.
- [ ] 12.3 Map each script behavior to an existing reusable operation or add a separately specified reusable operation when required.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W2; depends_on=[12.1, 12.2, 9.1]`.
- [ ] 12.4 Express the full cleanup behavior as a versioned workflow without embedding arbitrary Python or one-off code.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W5; depends_on=[12.3, 8.6, 8.7, 9.2, 9.3, 9.5, 9.6, 9.7]`.
- [ ] 12.5 Run script and toolkit against the same fixtures and pass value, inclusion, normalization, ordering, exception, and source-immutability assertions.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W6; depends_on=[12.4, 13.1, 13.2, 6.6]`.
- [ ] 12.6 Record any intentionally unsupported script behavior as a failing acceptance gap rather than silently changing semantics.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W7; depends_on=[12.5]`.

## 13. TXT and Google Adapters

> Section dependency: requires the adapter foundation; mapping readers require the stabilized 8.2 contract. TXT unlocks MVP 3, and the full section unlocks the final interface and adapter gates.

- [ ] 13.1 Implement TXT delimiter candidate detection and include evidence in mapping reports.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W2; depends_on=[8.2, 5.1]`.
- [ ] 13.2 Implement configured and per-workflow TXT writing with UTF-8, Unicode, and stable-order tests.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W3; depends_on=[13.1, 3.4, 5.4]`.
- [x] 13.3 Implement the external OAuth credential provider and least-privilege Sheets/Drive client construction.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W0; depends_on=[3.7, 5.1]`.
- [ ] 13.4 Implement Google spreadsheet and sheet discovery, displayed-value reading, canonical mapping, and Unicode preservation.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W2; depends_on=[13.3, 8.2]`.
- [ ] 13.5 Implement staged Google output creation or copy, value writing, named formatting, validation, and source-identity protection.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W3; depends_on=[13.3, 13.4, 5.6, 3.3]`.
- [ ] 13.6 Add fake-client unit tests that require no credentials and an explicitly activated live integration smoke test.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W4; depends_on=[13.3, 13.4, 13.5]`.
- [x] 13.7 Document that `clasp` is not required and keep Apps Script management outside the adapter contract.
  - Execution: `model=gpt-5.6-terra; effort=low; wave=W0; depends_on=[]`.

## 14. Terminal, TUI, and Codex Interfaces

> Section dependency: scaffolding may start earlier, but completion requires stable mapping, operation, adapter, cross-source, and orchestration contracts. Unlocks 15.6 and 15.8.

- [ ] 14.1 Implement Cobra commands for config validation, workflow validation, basic mapping, extended mapping, execution, recent workflows, and help.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W5; depends_on=[8.7, 9.1, 6.8]`.
- [ ] 14.2 Support original-header column references in terminal commands without requiring a global alias.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W6; depends_on=[14.1, 8.1]`.
- [ ] 14.3 Implement saved and pasted YAML/JSON workflows with exact validation feedback and configurable recent-workflow retention.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W7; depends_on=[14.1, 14.2, 2.2, 3.5]`.
- [ ] 14.4 Implement the Bubble Tea TUI source-intake, mapping, additional-source, output/query, and operation-editing flow.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W5; depends_on=[8.7, 9.1, 6.8]`.
- [ ] 14.5 Implement explicit Boolean grouping and `not_scope` selection in the TUI with `both` as the resolved default.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W6; depends_on=[14.4, 10.4, 2.5]`.
- [ ] 14.6 Implement approval, started, progress, final report, validation, exception, error, save, reload, and new-workflow TUI states.
  - Execution: `model=gpt-5.6-sol; effort=xhigh; wave=W7; depends_on=[14.4, 14.5, 6.4, 6.8]`.
- [ ] 14.7 Verify that all UI labels, prompts, errors, schemas, and technical interface docs are English while user data remains unchanged.
  - Execution: `model=gpt-5.6-terra; effort=medium; wave=W8; depends_on=[14.1, 14.4, 14.6]`.
- [ ] 14.8 Create README and Codex skills that discover supported schemas, invoke the executable for supported work, and avoid one-off data scripts.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W9; depends_on=[14.1, 14.4, 14.7, 13.7]`.
- [ ] 14.9 Add interface-contract tests proving terminal, TUI, and Codex-generated workflows normalize and execute equivalently.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W10; depends_on=[14.3, 14.6, 14.8, 10.6]`.

## 15. Acceptance, Performance, and Completion Gates

> Section dependency: individual gates run incrementally; baselines require all three MVPs, and final completion requires Sections 7–14.

- [ ] 15.1 Run source-immutability checks around every unit, integration, and MVP fixture that touches approved source material.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W10; depends_on=[5.3, 5.6, 5.7, 7.7, 11.7, 12.5]`.
- [x] 15.2 Add scale fixtures near 40 MB JSON and 500,000 CSV rows and verify bounded memory, stable order, Unicode, and deterministic output.
  - Execution: `model=gpt-5.6-sol; effort=max; wave=W0; depends_on=[7.7, 4.7, 6.5]`.
- [ ] 15.3 Record runtime and peak-memory baselines for all three MVP workflows on the target machine.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W7; depends_on=[7.7, 11.7, 12.5, 12.6]`.
- [ ] 15.4 Select and document safe numeric defaults for memory, file-lock retries, and temporary-artifact retention from measured baselines.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W8; depends_on=[15.3, 3.5]`.
- [ ] 15.5 Convert accepted baselines into repeatable regression thresholds and document exceptions for slower diagnostic/report modes.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W9; depends_on=[15.4]`.
- [ ] 15.6 Run Go formatting, vet, unit tests, race tests, integration tests, builds, and all MVP acceptance workflows successfully.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W11; depends_on=[8.7, 9.11, 10.6, 11.7, 12.6, 13.6, 14.9, 15.1, 15.2, 15.5]`.
- [ ] 15.7 Run `openspec validate build-data-toolkit-mvp --strict` and resolve every proposal, design, spec, scenario, and task validation error.
  - Execution: `model=gpt-5.6-sol; effort=high; wave=W12; depends_on=[15.6]`.
- [ ] 15.8 Update implementation documentation and Codex project metadata to match the verified repository, commands, schemas, and acceptance results.
  - Execution: `model=gpt-5.6-terra; effort=high; wave=W13; depends_on=[14.8, 14.9, 15.3, 15.6, 15.7]`.
