## ADDED Requirements

### Requirement: Mapping precedes planning
The orchestrator SHALL obtain a mapping report for every source before finalizing an execution plan.

#### Scenario: Multi-source workflow mapping
- **WHEN** a workflow declares two source files
- **THEN** both sources are mapped and validated before cross-source operations are planned

### Requirement: Basic and extended mapping
Basic mapping SHALL return source, sheet, and column identities; extended mapping SHALL additionally return value counts, inferred-type counts, observed format counts, header structure, and cleanliness findings.

#### Scenario: Extended type report
- **WHEN** a column contains 50 recognized dates and 2 text values
- **THEN** extended mapping reports both counts and classifies the column as mixed-type

### Requirement: User and inferred types
The mapping contract SHALL accept an explicit user-declared column type or infer it from source values and SHALL record which authority supplied the resolved type.

#### Scenario: User declaration overrides inference
- **WHEN** a workflow declares an allowed type for a column and observed values satisfy that type
- **THEN** mapping records the user declaration as authoritative

### Requirement: Mixed-type policy
A mixed-type column SHALL block execution unless an approval-capable interface or an explicit workflow/configuration policy resolves it as block, exception report, or ignore non-matching values.

#### Scenario: Non-approval interface has no policy
- **WHEN** mixed types are detected and `approval_before_run` is false with no resolving policy
- **THEN** the orchestrator returns an error and does not start logical processing

#### Scenario: Exception policy preserves order
- **WHEN** an exception-report policy excludes incompatible values
- **THEN** accepted rows retain stable order and excluded rows are reported with provenance and reason

### Requirement: Format normalization planning
When one semantic type appears in several recognized formats, mapping SHALL report the variants and SHALL plan normalization to the resolved format before logical processing.

#### Scenario: Multiple phone formats
- **WHEN** a phone column contains `526105412`, `052-6105412`, and `+972526105412`
- **THEN** mapping identifies one semantic type with multiple formats and plans normalization under the selected phone rules

### Requirement: Targeted whitespace repair
Mapping SHALL detect leading and trailing whitespace and SHALL insert a trim repair only for affected columns; it MUST NOT remove internal whitespace by default.

#### Scenario: Name contains internal space
- **WHEN** a value contains leading whitespace and an internal space between name parts
- **THEN** repair removes only the leading whitespace and preserves the internal space

### Requirement: Clean handoff to logical engine
The file engine SHALL NOT pass a column to the logical engine until all blocking findings have been resolved and all planned pre-logical repairs have completed.

#### Scenario: Unresolved dirty column
- **WHEN** a blocking cleanliness finding remains unresolved
- **THEN** the logical engine receives no rows for that workflow and the run returns the blocking report

