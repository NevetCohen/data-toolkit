## ADDED Requirements

### Requirement: Stable file-adapter contract
Each adapter SHALL expose source inspection, basic and extended mapping, canonical row reading, canonical row writing where supported, staged output validation, and declared format capabilities.

#### Scenario: Unsupported adapter feature
- **WHEN** a workflow requests a capability not declared by the selected adapter
- **THEN** preflight rejects the request before reading data rows

### Requirement: Immutable local source snapshots
The file engine SHALL open local sources for shared read, create and validate a per-run immutable snapshot, and process only that snapshot.

#### Scenario: Source changes after snapshot
- **WHEN** the original source changes after snapshot validation
- **THEN** the active run continues deterministically from the validated snapshot and reports its source provenance

#### Scenario: Exclusive source lock
- **WHEN** a source remains unreadable after configured lock retries
- **THEN** the run stops before logical processing with a precise lock error and does not use Volume Shadow Copy

### Requirement: Staged output publication
Adapters SHALL write outputs to staged files, validate them, and publish final paths only after validation under the resolved collision policy.

#### Scenario: Validation failure
- **WHEN** a staged output fails structural validation
- **THEN** no successful final output is published and the report identifies the retained or removed staged artifact according to policy

#### Scenario: Locked output uses alternate name
- **WHEN** a target is locked and the resolved collision action is `alternate_name`
- **THEN** a validated output is published under a deterministic alternate name and reported

### Requirement: JSON and CSV adapters
The JSON and CSV adapters SHALL support streaming or bounded-memory mapping and row conversion, preserve Unicode, exact source values and stable order, and write new outputs without changing sources.

#### Scenario: Large JSON field projection
- **WHEN** a roughly 40 MB JSON source is projected to selected CSV fields
- **THEN** the adapters complete within the resolved memory budget and preserve projected value order and encoding

### Requirement: TXT delimiter behavior
The TXT adapter SHALL detect candidate input delimiters during mapping and SHALL use the configured or per-workflow delimiter when writing.

#### Scenario: Line-delimited source and comma output
- **WHEN** mapping detects line-delimited input and the workflow requests comma-delimited output
- **THEN** the adapter reads the detected records and writes the configured delimiter without modifying the source

### Requirement: Excel displayed values and styles
The Excel adapter SHALL address explicit sheets, read cached computed values rather than formulas, warn when a formula has no cached result, support bounded reading/writing, and apply a named final style instead of preserving source formatting.

#### Scenario: Missing cached formula result
- **WHEN** an Excel formula cell has no cached computed value
- **THEN** mapping reports a warning and the workflow warning/approval policy determines whether execution continues

### Requirement: Google Sheets and Drive integration
The Google adapter SHALL use Sheets API v4 and Drive API with external least-privilege OAuth credentials, address explicit sheets, support displayed-value reads and styled writes, and MUST NOT require `clasp`.

#### Scenario: Read formatted values
- **WHEN** a workflow requests displayed Google Sheets values
- **THEN** the adapter returns formatted values under the canonical model without exposing formulas to the logical engine

### Requirement: Unicode and RTL-safe output
All adapters SHALL preserve Hebrew, emoji, and other Unicode characters; text output SHALL default to UTF-8 and spreadsheet output SHALL use native Unicode with style-level RTL presentation where selected.

#### Scenario: Hebrew and emoji round trip
- **WHEN** a row containing Hebrew and emoji is read and written without a declared transformation
- **THEN** the output preserves the characters exactly

