## 1. Contract Freeze and Baseline

- [x] 1.1 Run `openspec validate implement-revised-data-toolkit-v1 --strict` and save the successful pre-implementation result. <!-- model=gpt-5.6-luna; effort=low -->
- [x] 1.2 Run `go test ./...`, `go vet ./...`, and `go build ./cmd/data-toolkit`; record the current foundation result. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 1.3 Create a traceability table with one row for every requirement and scenario in all nine capability specs. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 1.4 Map every traceability row to at least one implementation task and one focused test task. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 1.5 Record the exact current references to cell provenance, retired combined-engine names, old CLI request flags, TUI, TXT, native Google format, `boolean.not_scope`, and `interfaces.recent_workflows`. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 1.6 Verify no implementation task contains an unresolved contract verb such as “define schema”, “choose fields”, or “decide behavior”. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 1.7 Add an implementation stop rule: any missing field, enum, operation, API code, or Skill contract requires an OpenSpec edit and strict revalidation first. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 1.8 Read `Invoke-RawMemberCleanup.Streaming.py` without executing it and inventory each externally visible transformation and failure behavior. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 1.9 Add sanitized one-behavior cleanup fixtures without copying private production rows into the repository. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 1.10 Verify that exactly four cleanup fixtures remain mapped as V1 coverage evidence through the fixed ten-operation catalog; record the other fourteen as `outside_v1` for a future OpenSpec change that this change MUST NOT create. <!-- model=gpt-5.6-sol; effort=high -->

## 2. Primitive and Canonical Data Contracts

- [x] 2.1 Implement the exact identity aliases from `Primitive identities and enums` with their validation helpers. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.2 Implement the ten fixed enums, including `RemoteCollisionPolicy`, and reject unknown values before planning. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.3 Add table-driven tests for every valid and invalid identity and enum value. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 2.4 Implement immutable `Value{typeID, encoded}` with copied constructor and accessor bytes. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 2.5 Register canonical `null`, `string`, `integer`, `decimal`, `boolean`, `date`, and `time` descriptors at `v1`. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.6 Implement canonical null separately from empty string and literal `-`. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.7 Add exact integer and decimal tests proving no `float64` conversion. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 2.8 Implement `Cell{ColumnID, Value}` without provenance, lineage, or rollback fields. <!-- model=gpt-5.6-luna; effort=low -->
- [x] 2.9 Implement all eight exact `ColumnDescriptor` fields and their required/optional validation. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.10 Implement `Schema{ID, SourceID, SheetID, Columns}` with ordered unique columns. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.11 Implement `Row{ID, SourceID, SheetID, Ordinal, Cells}` and exact schema matching. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 2.12 Implement `Dataset{Schema, Rows}` and reject nil or invalid streams. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.13 Make `RowStream.Close` idempotent and `Next` cancellation-aware with `io.EOF` termination. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 2.14 Add Hebrew, emoji, NFC, multiline, null, empty, hyphen, boolean-zero, and exact-number fixtures. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 2.15 Add row/schema mismatch tests for source, sheet, count, order, column ID, and type ID. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 2.16 Update external extension tests to compile against only the exact canonical contracts. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 2.17 Search root code for stale mandatory per-cell provenance and remove only the runtime dependency. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 2.18 Run focused `internal/core/...` tests and read back the final public structures. <!-- model=gpt-5.6-sol; effort=medium -->

## 3. Descriptors, Findings, and Terminal Structures

- [x] 3.1 Implement `DataTypeDescriptor` with the exact four fields and seven canonical kinds. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.2 Implement `StyleDescriptor` with exact `name` and `right_to_left` fields. <!-- model=gpt-5.6-luna; effort=low -->
- [x] 3.3 Implement the exact twelve-field `OperationDescriptor` shared by the Reader, Logical, and Writer registries. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.4 Implement `OperationRef` with resolved ID, version, and owner. <!-- model=gpt-5.6-luna; effort=low -->
- [x] 3.5 Implement all six optional/required `SourceLocator` fields without per-cell persistence and exact `OutputLocator{output_path,row_ordinal?,column_id?,physical_position?}` for reopened output failures. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.6 Implement `Finding` with stable code, severity, message, locator, and decision flag. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.7 Implement sorted `ErrorDetail{key,value}` values with redaction. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 3.8 Implement `ValidationResult` with the four fixed validation kinds and optional `OutputLocator` only for reopened output failures. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.9 Implement `ExceptionSummary` with optional first locator and external report path. <!-- model=gpt-5.6-luna; effort=medium -->
- [x] 3.10 Implement `PublishedOutput` with final path, format, hash, and size. <!-- model=gpt-5.6-luna; effort=low -->
- [x] 3.11 Implement every exact `RunReport` field and required/conditional presence rule. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 3.12 Add JSON golden tests for succeeded, failed, blocked, and cancelled reports. <!-- model=gpt-5.6-terra; effort=high -->
- [x] 3.13 Add contract tests proving findings and exceptions remain in reports and do not enter compact receipts. <!-- model=gpt-5.6-terra; effort=high -->

## 4. Exact Revised Configuration

- [ ] 4.1 Replace Go root configuration fields with the twelve exact required root fields. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.2 Implement all seven exact `DisplayConfig` fields and fixed UTF-8/NFC constraints. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.3 Implement exact `AliasConfig` and case-insensitive uniqueness validation. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.4 Implement exact `SemanticTypeConfig` and registered-data-type validation. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.5 Implement all six exact `StyleConfig` fields and `#RRGGBB` validation. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 4.6 Implement all seven exact `OutputConfig` fields and their enums/defaults. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.7 Implement all four exact `CleaningConfig` fields and defaults. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 4.8 Implement exact `ExceptionConfig{action,detail}`. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 4.9 Implement all six exact `RuntimeConfig` fields, durations, and safe numeric checks. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.10 Implement all five exact `ReportConfig` fields and enforce the two required true values. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 4.11 Implement `CLIConfig.non_overridable_config_paths` as sorted unique existing JSON Pointers with the four mandatory protected defaults. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 4.12 Implement `GoogleSheetsConfig` with destination folder, title template, and `block`/`alternate_name` remote collision policy only. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.13 Update `config-v1.schema.json` with every exact field and `additionalProperties:false` at every object. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 4.14 Update embedded defaults to the exact values stated in the Infrastructure spec. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.15 Remove `Boolean`, `Interfaces`, `TXTDelimiter`, and their patch fields from active root code. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.16 Implement immutable defaults -> user -> CLI patches -> workflow -> output precedence using only allowed override fields. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 4.17 Compute stable effective-configuration digests from normalized resolved configuration. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 4.18 Add YAML/JSON equivalence golden tests for the complete document. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 4.19 Add rejection tests for every removed field, unknown field, invalid reference, duplicate name, enum, duration, and range. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 4.20 Add config-patch tests for an allowed scalar field, duplicate pointer, missing pointer, invalid structural change, protected pointer, protected ancestor, and protected descendant. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 4.21 Add precedence tests proving CLI patches lose to workflow/output overrides and affect the effective-config digest. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 4.22 Run focused config tests and compare Go fields, schema properties, defaults, and spec tables programmatically. <!-- model=gpt-5.6-sol; effort=high -->

## 5. API Envelopes and Error Catalog

- [ ] 5.1 Implement generic `RequestEnvelope<T>` with exactly `schema_version`, `request_id`, and `payload`. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 5.2 Implement generic `ResultEnvelope<T>` with exclusive `result`/`error` branches. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.3 Add request-correlation and `additionalProperties:false` golden tests for both envelopes. <!-- model=gpt-5.6-terra; effort=medium -->
- [x] 5.4 Implement `APIError` with all eight exact fields and no unlisted field, including exactly one `SourceLocator` or `OutputLocator` variant when present and conditional `RunFailureContext.receipt_id` only after receipt append. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 5.5 Register the eighteen exact API error codes from the Application API spec, including `resource_limit_exceeded`. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 5.6 Implement error-code validation and deterministic sorted details. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.7 Implement `CapabilitiesRequest` as an empty strict object. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 5.8 Implement split `reader_formats` and `writer_formats` in `CapabilitiesResult` and normalized descriptor sorting. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.9 Implement `DescribeOperationRequest{operation_id,version}` with ambiguity handling. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.10 Implement exact `ConfigValidationRequest` and `ConfigValidationResult` payloads. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.11 Implement exact `InspectionRequest` limits and defaults. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 5.12 Implement exact `WorkflowValidationRequest` and `WorkflowValidationResult` payloads, including resolved non-secret Google Sheets policy. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 5.13 Implement exact `RunRequest` including ordered `ConfigOverride[]` and optional expected workflow digest. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 5.14 Implement the six exact application API methods and reserve Go `error` for process-level failure only. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 5.15 Add fake-API contract tests for success, product error, and process-level error on each method. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 5.16 Add workflow-digest-mismatch tests proving failure before source access. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 5.17 Add API golden tests for Reader/Writer component codes, resource-limit execution exit, pre-workspace error envelopes, pre-finalization `run_context` without `receipt_id`, post-receipt `run_context` with `receipt_id`, and finalized failed-run report envelopes. <!-- model=gpt-5.6-sol; effort=xhigh -->

## 6. Compact Run Receipt Store

- [x] 6.1 Implement `RunReceipt` with exactly the thirteen allowed fields and conditional source/output/error fields. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 6.2 Reject every unlisted receipt field and fail serialization tests if one appears. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 6.3 Compute `ReceiptID` deterministically from the normalized terminal record excluding `receipt_id`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 6.4 Implement the local append-only receipt-store interface. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 6.5 Implement atomic append behavior for the configured local receipt store. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 6.6 Implement retention using `ReportConfig.retention` without mutating retained records. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 6.7 Strip `APIError.run_context` before storing a terminal error in a receipt. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 6.8 Add redaction tests for credentials, tokens, prompts, row values, full requests, and full configuration. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 6.9 Add compactness tests excluding findings, exceptions, validations, and lifecycle events. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 6.10 Add receipt tests for pre-snapshot failure, post-snapshot failure, cancellation, and success. <!-- model=gpt-5.6-terra; effort=high -->

## 7. Reader Requests, Descriptors, and Inspection

- [ ] 7.1 Implement all seven exact `SourceRequest` fields and format-specific field validation. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.2 Implement `ReaderFormatDescriptor` and shared `FormatLimits` with only `inspect` and `read` capabilities. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.3 Register `reader.inspect`, `reader.snapshot`, and `reader.read` with fixed owner, exposure, and streaming mode. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.4 Add reader-registration tests rejecting descriptor/executor mismatch, duplicate IDs, and an empty capability set. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.5 Implement every exact `InspectionReport` common field and exclusive format-detail branch. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.6 Implement `InspectedTable`, sample rows/values, and column counts with bounded sizes. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.7 Implement all exact `InspectionDecision` codes, including `select_type`, and request field paths. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.8 Implement exact `CSVInspectionDetails` fields and enums. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.9 Implement exact `JSONInspectionDetails` fields and root-kind restriction. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.10 Implement exact `ExcelInspectionDetails` fields and formula counts. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.11 Add JSON golden tests for one complete report of each format. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.12 Add ambiguity tests for sheet, delimiter, header row, root path, duplicate column, type, and dirty input. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.13 Add bounded sample and finding-limit tests. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 7.14 Implement shared-read locking with exact retry, interval, timeout, and cancellation policy. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.15 Add lock tests for immediate success, retry success, timeout, and cancellation. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.16 Implement `Snapshot` with exact source, path, hash, and size fields. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 7.17 Copy, hash with SHA-256, close, and verify the immutable snapshot before row access. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.18 Add tests proving active source files are opened read-only and never written. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 7.19 Implement direct reader decoding to `Schema + RowStream + Value` without a public batch stream. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.20 Validate schema and dirty values before emitting any row to Logical Engine. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.21 Add source-change, blocked dirty-input, and exact first-locator tests. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 7.22 Reject a Reader lifecycle operation in a user `OperationCall` before source access. <!-- model=gpt-5.6-luna; effort=medium -->

## 8. Writer Requests, Descriptors, and Terminal Lifecycle

- [ ] 8.1 Implement exact `CSVOutputOptions` fields and enums. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 8.2 Implement all six exact `OutputRequest` fields and source/output path inequality. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 8.3 Implement `WriterFormatDescriptor` and shared `FormatLimits` with only `write`, `style`, and `validate` capabilities. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 8.4 Register `writer.write`, `writer.style`, `writer.validate`, `writer.publish`, and `writer.finalize` with fixed owner, exposure, and streaming mode. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 8.5 Add writer-registration tests rejecting descriptor/executor mismatch, duplicate IDs, and invalid capability sets. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 8.6 Implement staged output outside the final target path and exact `StagedArtifact` fields. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 8.7 Implement optional Excel-only named style from resolved `output.default_style` and `styles`. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 8.8 Close the staged artifact and reopen it before any mandatory output validation. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 8.9 Implement exact `OutputValidation{format,schema,assertions}`, preserve assertion order, and retain `OutputLocator` fields for reopened failures. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 8.10 Implement local `block` collision behavior. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 8.11 Implement deterministic local `alternate_name` collision behavior. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 8.12 Implement safe local `overwrite` behavior without a partial final output. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 8.13 Publish atomically only after mandatory format/schema checks and all requested assertions pass. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 8.14 Implement `writer.finalize` exactly once after workspace creation for success, reader failure, logical failure, writer failure, and cancellation. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 8.15 Close registered streams and resources in reverse acquisition order during finalization. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 8.16 Implement success cleanup and failed-run retention as separate policy paths. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 8.17 Assemble the terminal `RunReport` and append exactly one compact receipt in finalization. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 8.18 Add lifecycle fixtures for every terminal outcome, cleanup/retention result, no partial publication, and one receipt. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 8.19 Reject a Writer lifecycle operation in a user `OperationCall` before source access. <!-- model=gpt-5.6-luna; effort=medium -->

## 9. Workflow Contract and Orchestrator

- [ ] 9.1 Implement all eight exact `WorkflowRequest` fields and strict decoding. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.2 Implement all four exact `OperationCall` fields and unique step IDs. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.3 Implement the three allowed `WorkflowOverrides` groups and reject every forbidden override. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.4 Implement exact `ExceptionPolicy` fields, defaults, and maximum-detail range. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 9.5 Implement strict `exact_columns` validation requests. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 9.6 Implement strict `all_rows_match` validation requests using the shared expression contract. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.7 Reject multiple sources, outputs, unknown fields, queries, missing versions, and non-workflow operations before source access. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.8 Implement `WorkflowTemplateDocument` with fixed `$schema`, `schema_version`, `variables`, and `workflow` fields. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 9.9 Implement contiguous positive variable IDs, unique descriptive names, declared types, required/default semantics, and strict defaults. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.10 Implement value-only `{"$var": id}` references and reject interpolation, field-name, and object-key substitution. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.11 Implement deterministic typed substitution before strict `WorkflowRequest` decoding and workflow-digest calculation. <!-- model=gpt-5.6-terra; effort=xhigh -->
- [ ] 9.12 Add template fixtures for contiguous IDs, multiple variables, defaults, invalid type, missing reference, and identical digest. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 9.13 Normalize resolved workflows deterministically and compute the exact workflow digest. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.14 Build the fixed ordered plan: `reader.snapshot`, `reader.read`, request operations, `writer.write`, optional `writer.style`, `writer.validate`, `writer.publish`, and `writer.finalize`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.15 Preserve request logical-operation order without inferring a missing operation. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 9.16 Dispatch logical calls only through Logical Engine and lifecycle steps only through their Reader or Writer registry. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.17 Add cancellation checks at every blocking, stream, write, validation, and publication boundary. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.18 Keep progress transient and exclude lifecycle-event streams from results and receipts. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 9.19 Verify the Orchestrator does not construct reports or append receipts; Writer finalization owns both terminal effects. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 9.20 Add plan golden tests and failure tests for Reader, Logical Engine, Writer, and cancellation terminal paths. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 9.21 Add architecture tests rejecting operation-kind switches and format-specific orchestrator branches. <!-- model=gpt-5.6-sol; effort=high -->

## 10. Expression Contract and First Logical Slice

- [ ] 10.1 Implement exact `TypedLiteral{type_id,value}` conditional validation. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 10.2 Implement strict `LogicalGroup`, `LogicalNot`, and `ConstantComparison` tagged nodes. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 10.3 Validate every fixed comparison operator, type rule, child count, and conditional field. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 10.4 Add expression tests for nested `and`, `or`, `not`, null predicates, text operators, and exact decimals. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 10.5 Register `table.project@v1` with its exact descriptor and strict parameter schema. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 10.6 Implement project schema transformation and row-stream wrapping by ordered stable column IDs. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 10.7 Add project tests for Hebrew headers, missing IDs, duplicate IDs, and output order. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 10.8 Register `row.filter@v1` with its exact descriptor and strict parameter schema. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 10.9 Implement typed comparisons through registered data-type handlers. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 10.10 Implement stable streaming filter evaluation without inferred precedence. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 10.11 Add filter tests for null, empty string, hyphen, case sensitivity, regex, and type mismatch. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 10.12 Add discovery and describe golden tests for both operation descriptors. <!-- model=gpt-5.6-luna; effort=medium -->

## 11. CSV Vertical Slice

- [ ] 11.1 Add CSV fixtures for comma, tab, semicolon, quotes, multiline, Hebrew, emoji, null, duplicate/empty headers, and exact numbers. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 11.2 Implement closed-set delimiter inspection and exact ambiguity decisions. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.3 Implement header-row inspection with stable column IDs and findings. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.4 Implement bounded type/count/sample inspection. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.5 Implement the UTF-8/NFC CSV reader as a canonical `RowStream`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.6 Decode null, empty string, hyphen, exact numbers, quotes, and multiline fields exactly. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.7 Implement the staged CSV writer with exact delimiter, line ending, and null policy. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.8 Implement deterministic quoting, escaping, and `\N` collision failure. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.9 Reopen CSV output and implement mandatory format/schema validation. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.10 Implement exact-columns and all-rows-match assertions on reopened output. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 11.11 Register the complete CSV descriptor and all implemented provider interfaces. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 11.12 Add adapter contract tests for every fixture and invalid option. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.13 Add the application CSV workflow using filter, project, output validation, report, and receipt. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.14 Assert source read-only behavior and the exact processed snapshot hash. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 11.15 Assert deterministic output bytes, workflow digest, validations, and receipt fields. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.16 Assert cancellation closes streams and publishes no output. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 11.17 Measure bounded memory and pass the CSV gate before JSON or Excel registration. <!-- model=gpt-5.6-sol; effort=high -->

## 12. Remaining Eight Logical Operations

- [ ] 12.1 Register `table.rename@v1` with the exact descriptor and `RenameParameters` schema. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.2 Implement rename without changing stable IDs and test duplicate resulting headers. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.3 Register `text.trim@v1` with the exact descriptor and `TrimParameters` schema. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.4 Implement Unicode edge trim only and test preservation of internal whitespace. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.5 Register `value.normalize@v1` with exact target and invalid-action schemas. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.6 Implement registered-type normalization and construct new immutable values. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.7 Test normalize success, `error`, `keep`, semantic-type mismatch, and first locator. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.8 Register `value.replace@v1` with exact typed match/replacement parameters. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.9 Implement typed scalar replacement and test type mismatch rejection. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.10 Register `text.regex_replace@v1` with exact RE2 parameters and 4096-byte pattern limit. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.11 Implement regex replacement and test invalid patterns, replace-first, and replace-all. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.12 Register `row.delete_empty@v1` with an empty strict parameter object. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 12.13 Implement whole-row null/empty deletion and test preservation of hyphen, zero, false, and partially populated rows. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.14 Register `row.deduplicate@v1` with exact keys and keep policy. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.15 Implement bounded-spool `keep=first` semantics within the run reservation. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.16 Implement bounded-spool `keep=last` semantics within the run reservation. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.17 Implement bounded-spool `keep=error` and the exact first duplicate locator. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.18 Preserve stable source order, exact keys, cancellation, and cleanup for every deduplication mode after spill. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 12.19 Test empty dedup keys resolving to `#` or all columns exactly. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.20 Register `row.sort@v1` with exact `SortParameters` and ordered `SortKey` schema. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 12.21 Implement multi-key comparison through registered comparable data types and explicit ascending/descending direction. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.22 Implement explicit null-first/null-last ordering independent of key direction. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.23 Implement NFC Unicode code-point, case-sensitive string ordering. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.24 Preserve input order for rows equal on every sort key. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.25 Implement bounded external merge/spool sorting and release reservations after merge phases. <!-- model=gpt-5.6-sol; effort=xhigh -->
- [ ] 12.26 Add sort fixtures for multi-key order, every null placement, Unicode, stable ties, many unique keys, spill, cancellation, and cleanup. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.27 Add one descriptor golden and one invalid-parameter test for each of the eight operations. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 12.28 Add an operation-chain test proving no intermediate CSV round trip. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 12.29 Run the structural cleanup-fixture mapping against the complete ten-operation catalog; verify exactly four mapped coverage behaviors and fourteen `outside_v1` entries, without creating a future change. <!-- model=gpt-5.6-sol; effort=high -->

## 13. JSON Adapter

- [ ] 13.1 Add JSON fixtures for root arrays, nested roots, missing fields, null, empty strings, Hebrew, emoji, and exact numbers. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 13.2 Implement exact root-path inspection and `select_root_path` decisions. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.3 Implement scalar and nested path reporting without flatten or explode. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 13.4 Implement the bounded JSON reader as a canonical row stream. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.5 Preserve integers and decimals without `float64`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.6 Apply the explicit missing-field policy and exact first locator. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.7 Reject object/array projection with `unknown_capability` or `invalid_workflow` as specified by validation stage. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 13.8 Implement deterministic staged JSON writing for scalar canonical tables. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.9 Reopen JSON output and validate declared schema. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.10 Register the complete JSON descriptor and provider interfaces. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.11 Add contract tests for every JSON fixture and error path. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 13.12 Add the exact six-field JSON-to-CSV workflow and assertions. <!-- model=gpt-5.6-luna; effort=medium -->

## 14. Excel Adapter

- [ ] 14.1 Add the approved Excelize v2 dependency to the root module and record its license. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 14.2 Add XLSX fixtures for multiple sheets, Hebrew headers, cached formulas, missing caches, nulls, and styles. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 14.3 Implement worksheet-list inspection and first-sheet default. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.4 Implement exact missing-sheet validation before row access. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 14.5 Implement header-row inspection and stable column IDs. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.6 Implement cached/displayed formula reading without launching Excel. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.7 Emit exact formula-cache findings and counts. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 14.8 Implement the bounded Excel reader as a canonical row stream. <!-- model=gpt-5.6-terra; effort=xhigh -->
- [ ] 14.9 Implement staged Excel writing with one exact worksheet. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.10 Apply the named style only at final writing, including RTL. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.11 Reopen Excel output and validate format, sheet, schema, and assertions. <!-- model=gpt-5.6-terra; effort=xhigh -->
- [ ] 14.12 Register the complete Excel descriptor and provider interfaces. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 14.13 Add contract tests for every sheet, formula, style, and failure fixture. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 14.14 Add the exact `ישוב = בת ים` Excel workflow and assertions. <!-- model=gpt-5.6-terra; effort=high -->

## 15. CLI JSON Surface

- [ ] 15.1 Refactor the CLI root to use separate stdout and stderr writers. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.2 Implement `capabilities --json` using exact empty request and result payloads. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.3 Implement `describe --kind operation --id <id> [--version <v>] --json`. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.4 Implement `config validate --request <file> --json` with strict envelope decoding. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.5 Implement `inspect --request <file> --json` with strict envelope decoding. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.6 Implement `workflow validate --file <workflow.json> [--config <config>] [--var <id>=<value>] [--config-set <json-pointer>=<json-value>] --json`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.7 Implement `run --file <workflow.json> [--config <config>] [--var <id>=<value>] [--config-set <json-pointer>=<json-value>] --expected-workflow-digest <digest> --json`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.8 Emit exactly one result envelope to stdout on every JSON command path. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 15.9 Implement exit codes `0,2,3,4,5,6,7,8,9,10,130` with the exact error mapping. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 15.10 Add stdout/stderr golden tests for every exit class. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.11 Add request-correlation tests for every command. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.12 Add fake-API tests proving the CLI owns no product validation or registry logic. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.13 Remove active TUI composition and supported-V1 advertising without deleting legacy reference evidence. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.14 Parse repeatable `--var` values against the template declaration before application API validation. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.15 Parse repeatable `--config-set` pointer/value pairs as JSON and preserve their supplied order. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 15.16 Reject duplicate variables, unknown variable IDs, missing required variables, invalid JSON values, and declared-type mismatches with one JSON error envelope. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.17 Reject duplicate config pointers, invalid pointer syntax, nonexistent paths, protected paths, protected ancestors, protected descendants, and invalid structural patches. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 15.18 Add PowerShell golden command fixtures covering single-quoted paths, strings, JSON values, and escaped quotes. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 15.19 Add CLI template fixtures for defaults, deterministic substitution, strict post-substitution workflow rejection, and expected-digest drift. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.20 Add config-patch precedence fixtures showing defaults -> user config -> CLI patch -> workflow override -> output override. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 15.21 Run the built binary against CSV, JSON, and Excel template-file fixtures. <!-- model=gpt-5.6-terra; effort=high -->

## 16. Six-Skill Codex Adapter

- [ ] 16.1 Create `.agents/skills/data-toolkit/SKILL.md` with the exact root trigger, inputs, outputs, handoffs, and same-path stop condition. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.2 Add root-Skill fixtures for supported routing, unsupported join, same path, and clarification. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.3 Create `.agents/skills/data-toolkit-inspect/SKILL.md` with the exact inspection-only contract. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 16.4 Add inspection-Skill fixtures for known source, sheet, delimiter, header, root, duplicate column, and dirty input. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.5 Create `.agents/skills/data-toolkit-build-workflow/SKILL.md` with runtime describe, construction, validation, and stop rules. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.6 Require the builder Skill to read `docs/codex/workflow-json-format.md` before it writes a template file. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 16.7 Add builder fixtures for valid template construction, variables/defaults, missing capability, unknown parameter, and result-changing ambiguity. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 16.8 Create `.agents/skills/data-toolkit-run/SKILL.md` with the exact template path, user-config path, ordered variables and config patches, both digests, correlation, and report-output rules. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.9 Add run fixtures for success, source failure, validation failure, publication failure, and cancellation. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.10 Create `.agents/skills/data-toolkit-recover/SKILL.md` with exact eligibility and one-correction limit. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 16.11 Add recovery fixtures for removable field, explicit version, syntax normalization, single option, user decision, and second failure. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.12 Create `.agents/skills/data-toolkit-google-sheets/SKILL.md` with source/destination triggers, tool-availability probe, stop rules, and local-XLSX-only boundary. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.13 Probe the exact metadata, export/import, paginated search, spreadsheet metadata/cell readback, and update/move Drive operations before a Google run; report a capability gap with connector/manual-XLSX fallback when one is unavailable. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.14 Implement the source-Sheet procedure: metadata/MIME verification, XLSX export, returned absolute workspace path, local handoff, and finally cleanup. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.15 Add source-Sheet fixtures for missing connector, inaccessible file, non-Sheets MIME, 10-MB export gap, successful export, cancellation, and no Drive deletion. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 16.16 Resolve the destination folder and title, then search every result page for an exact native-Sheets title collision in that parent. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.17 Implement remote `block` and deterministic `alternate_name` suffix selection before any import; prohibit overwrite. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.18 Import the exact validated local XLSX as native Google Sheets and verify conversion, MIME, spreadsheet ID, URL, tabs, and metadata. <!-- model=gpt-5.6-terra; effort=xhigh -->
- [ ] 16.19 Implement destination-folder movement only after parent readback and verify parents after `google_drive_update_file`. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.20 Perform bounded structural/display cell readback, return the verified link with the local receipt, and clean only the exact local XLSX on every terminal path. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.21 Add destination-Sheet fixtures for `block`, paginated `alternate_name`, style/title policy, folder move, upload failure, structural/display readback, local cleanup, and no remote overwrite/deletion. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 16.22 Add a Skill-family inventory test requiring exactly the six specified names and locations. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 16.23 Add a drift test rejecting embedded operation IDs, format IDs, or parameter schemas in Skill text. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.24 Add handoff-schema tests for every adjacent Skill pair, including the Google bridge boundaries. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 16.25 Validate all six Skills with the repository Skill validator. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 16.26 Run one CSV, one JSON, one Excel, and one Google-bridge prompt through the full Skill chain before the evaluation suite. <!-- model=gpt-5.6-sol; effort=high -->

## 17. Practical Acceptance Workflows

- [ ] 17.1 Fix the exact output filename and assertions for the Excel city-filter fixture file. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 17.2 Run the sanitized Excel city-filter workflow and compare every output row predicate. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 17.3 Confirm approved local Excel source/output normalized paths differ before execution. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 17.4 Run the approved local Excel workflow and record source snapshot hash, output hash, validations, and receipt ID. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 17.5 Fix the exact output filename and assertions for the JSON six-field fixture file. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 17.6 Run the sanitized JSON workflow and compare exact headers, order, scalar values, Unicode, and nulls. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 17.7 Confirm approved local JSON source/output normalized paths differ before execution. <!-- model=gpt-5.6-luna; effort=low -->
- [ ] 17.8 Run the approved local JSON workflow and record source snapshot hash, output hash, validations, and receipt ID. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 17.9 Build only the four mapped cleanup coverage workflows from the fixed ten-operation catalog; do not attempt full script replacement or equivalence. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 17.10 Run only the four mapped sanitized cleanup fixtures and compare their recorded outputs and operation-defined failure behavior. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 17.11 Verify the legacy script is neither imported nor invoked at runtime; record the fourteen `outside_v1` entries and their future-change boundary without creating that change. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 17.12 Write the reproducible scoped-coverage acceptance report without private row values or a full replacement/equivalence claim. <!-- model=gpt-5.6-sol; effort=high -->

## 18. Reliability, Performance, Documentation, and Final Gates

- [ ] 18.1 Create at least 20 versioned Codex cases covering every required format, value edge, ambiguity, failure, and validation. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.2 Add expected commands, handoffs, request digest, output, validations, safety, and receipt assertions to every case. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.3 Run the suite with GPT-5.6 Luna at medium effort and record first-attempt and one-correction outcomes. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 18.4 Fix only the descriptor, schema, error, or Skill contract responsible for each failure; do not add planner inference. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.5 Re-run until safety is 100 percent, first-attempt success is at least 90 percent, and one-correction success is 100 percent. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.6 Record the exact environment and method for the approximately 40 MB JSON baseline. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 18.7 Record the exact environment and method for the CSV baseline up to 500,000 rows. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 18.8 Create the non-sensitive wide fixture with 141–146 columns, 33–35 percent nulls, text-heavy distribution, and hundreds of MB or equivalent rows. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.9 Record a row-oriented baseline on the wide fixture before considering any internal columnar batch prototype. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 18.10 Record wall time, CPU time, peak RSS, Go heap, total allocations, GC count/pause, and spill bytes for the row-oriented baseline. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.11 Verify baseline row count/order, null distinctions, exact numeric values, normalized output hash, and first-error locator. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.12 Measure allocations and GC before evaluating 4K, 16K, or 64K internal batch sizes; do not expose `BatchStream` publicly. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.13 Reject any batch prototype whose semantic-fidelity or first-error-location results differ from the row-oriented baseline. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 18.14 Atomically implement shared run-budget reservation accounting and Reader, Logical, and Writer hooks for material state, release, and bounded-spool spill transitions. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.15 Add `resource_limit_exceeded` fixtures proving spill-before-limit when possible and no publication when impossible. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.16 Set numeric time and peak-memory thresholds only from the recorded baselines. <!-- model=gpt-5.6-sol; effort=high -->
- [ ] 18.17 Add repeatable threshold-regression checks. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.18 Add external extension tests for one data type, one logical operation, and one Reader or Writer format capability without orchestrator edits. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.19 Run dependency-boundary tests rejecting root imports of `legacy/mvp`, forbidden subsystem directions, and Google Drive credentials/network calls from Go runtime packages. <!-- model=gpt-5.6-sol; effort=medium -->
- [ ] 18.20 Update `PROJECT.md`, technical documentation, template instruction, and Hebrew specification to nine capability specs and seven subsystems. <!-- model=gpt-5.6-terra; effort=medium -->
- [ ] 18.21 Search active code/docs for retired combined-engine terms, TXT, Google as a native format, queries, TUI, MCP, HTTP, remote, scaffold support, old API, old config, and old Skill claims. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 18.22 Compare implemented structs, JSON schemas, Reader/Writer/logical registries, CLI table, workflow-template instruction, and Skill inventory mechanically against the spec catalogs. <!-- model=gpt-5.6-terra; effort=high -->
- [ ] 18.23 Run formatting, focused tests, `go test ./...`, `go vet ./...`, and `go build ./cmd/data-toolkit`. <!-- model=gpt-5.6-luna; effort=medium -->
- [ ] 18.24 Run strict OpenSpec validation and confirm every traceability row has passing implementation evidence. <!-- model=gpt-5.6-sol; effort=high -->
