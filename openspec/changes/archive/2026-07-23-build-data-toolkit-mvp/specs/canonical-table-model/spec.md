## ADDED Requirements

### Requirement: Canonical sheets and tables
The canonical model SHALL represent explicitly addressable sheets and SHALL keep file-format details outside logical operations.

#### Scenario: Select a non-default sheet
- **WHEN** a workflow selects a named Excel or Google sheet
- **THEN** the adapter emits only the selected sheet as a canonical table with its sheet identity preserved

### Requirement: Complete column identity
Each canonical column SHALL preserve its physical source position, original header, optional global logical alias, inferred or declared data type, and optional header and subheader.

#### Scenario: Table begins after column A
- **WHEN** an Excel table begins at physical column `C`
- **THEN** the first canonical column retains physical position `C` rather than being relabeled `A`

#### Scenario: Global alias is reused
- **WHEN** two source headers are mapped to the same configured global alias
- **THEN** workflows can address that alias consistently while reports retain both original headers and positions

### Requirement: Header structure preservation
The file engine SHALL exclude a merged whole-table title row from data headers and SHALL preserve both header and subheader metadata when both rows exist.

#### Scenario: Header and subheader rows
- **WHEN** a source contains a header row followed by a subheader row
- **THEN** each affected canonical column contains both labels and the first data row is not consumed as metadata

### Requirement: Distinct internal null
The canonical model SHALL represent null independently from strings and SHALL render null as `"-"` by default without conflating it with a source hyphen.

#### Scenario: Null and hyphen coexist
- **WHEN** one source cell is null and another contains the literal string `"-"`
- **THEN** the canonical values remain distinct even if the default output displays both using the configured null rendering

### Requirement: Meaning-preserving canonical values
The canonical model SHALL preserve Unicode strings, booleans, exact integer and decimal source representations, local date/time values without time zones, and explicit JSON projections without silent lossy conversion.

#### Scenario: Exact JSON number
- **WHEN** a JSON number has more precision than a binary floating-point value can preserve
- **THEN** its canonical representation retains the exact source number until an explicit numeric operation is applied

#### Scenario: Nested JSON is not implicitly flattened
- **WHEN** a JSON object contains nested fields and the workflow projects selected paths
- **THEN** only declared paths become columns and no unrequested path is silently flattened or discarded by projection logic

### Requirement: Stable row order and provenance
Canonical row streams SHALL preserve source order and source identity unless an explicit operation changes order.

#### Scenario: Stable filtering
- **WHEN** a filter removes some rows without sorting
- **THEN** surviving rows retain their relative source order and provenance

