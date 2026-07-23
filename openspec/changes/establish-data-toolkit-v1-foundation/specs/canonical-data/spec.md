## ADDED Requirements

### Requirement: Versioned canonical data types
Every canonical data type SHALL provide a stable identity, versioned descriptor,
canonical parsing, value validation, deterministic comparison, and rendering.

#### Scenario: Register a custom data type
- **WHEN** a test supplies a valid new data-type handler to a fresh registry
- **THEN** the registry can parse, validate, compare, render, and describe that type without editing core dispatch code

### Requirement: Immutable canonical values
A canonical value SHALL retain its data-type identity and canonical encoded
content without exposing mutable internal storage.

#### Scenario: Caller mutates input bytes
- **WHEN** a value is constructed and the caller later mutates the source byte slice
- **THEN** the canonical value and its encoded output remain unchanged

### Requirement: Canonical cell provenance
Each cell SHALL associate a canonical value with a column identity and complete
source, sheet, row, and column provenance.

#### Scenario: Cell validation
- **WHEN** a cell is validated with missing provenance or a mismatched column identity
- **THEN** validation fails with the invalid field identified

### Requirement: Column metadata and engine binding
Each column SHALL preserve identity, header, optional subheader, data type, type
authority, optional source range, and an opaque binding between the file engine
and logical engine.

#### Scenario: Logical operation receives a bound column
- **WHEN** logical code receives a column originating from a spreadsheet range
- **THEN** it can preserve the binding and range without importing or interpreting file-engine types

### Requirement: Row-stream data transport
Canonical table data SHALL travel through deterministic row streams, and column
metadata MUST NOT contain all column cell values.

#### Scenario: Large logical input
- **WHEN** a logical operation receives a table with a large number of rows
- **THEN** it consumes rows incrementally through the stream contract rather than requiring column materialization
