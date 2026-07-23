## ADDED Requirements

### Requirement: Closed deterministic operation registry
The logical engine SHALL execute only versioned operations registered under the closed workflow schema, and each operation SHALL declare deterministic parameters, accepted inputs, outputs, and validation rules.

#### Scenario: Unknown operation
- **WHEN** a workflow names an unregistered operation
- **THEN** schema validation rejects it before source processing

### Requirement: Field projection and renaming
Output field selection SHALL be expressed as projection in the output definition, while field renaming SHALL be an ordered logical operation.

#### Scenario: Project then write
- **WHEN** an output projects six fields and a prior operation renames one of them
- **THEN** the output contains exactly the resolved six fields under their final names

### Requirement: Boolean filtering
The filter operation SHALL evaluate explicit nested expression trees with `AND`, `OR`, `XOR`, and `NOT` and SHALL preserve stable order for surviving rows.

#### Scenario: Nested positive expression
- **WHEN** a workflow filters using `(A == B OR C == D) AND E == F`
- **THEN** the engine evaluates the declared tree without relying on textual precedence

### Requirement: Symmetric positive cross-source logic
For two-source expressions using only `AND`, `OR`, and `XOR`, swapping the source declarations SHALL NOT change the truth set or result when the same projection is requested.

#### Scenario: OR comparison is source-order independent
- **WHEN** source 1 and source 2 are swapped for `source_1.A == source_2.B OR source_1.C == source_2.D` with equivalent remapped references
- **THEN** the engine produces the same logical matches and projected result

### Requirement: Scoped NOT semantics
A cross-source expression containing `NOT` SHALL evaluate its complement relative to `source_1`, `source_2`, or `both`, as declared by `not_scope`; the normalized default SHALL be `both`.

#### Scenario: NOT over both sources
- **WHEN** `NOT (source_1.A == source_2.B OR source_1.C == source_2.D)` is executed with `not_scope: both`
- **THEN** unmatched rows from both sources are returned with source identity preserved

#### Scenario: Ambiguous combined projection
- **WHEN** `not_scope: both` combines incompatible source schemas without a valid branch projection
- **THEN** planning fails rather than silently discarding source-specific fields

### Requirement: Format and text operations
The engine SHALL support declared format normalization, string concatenation, and regex-based text matching or replacement without applying undeclared transformations.

#### Scenario: Date normalization
- **WHEN** a date value `020226` is processed by a declared normalization operation targeting `02.02.2026`
- **THEN** the result is `02.02.2026` and the operation report records the transformation

### Requirement: Deterministic deduplication
Deduplication SHALL use column `#` as its default key when present, otherwise all columns, and SHALL retain the stable first row by default with explicit `last` and `error` alternatives.

#### Scenario: Duplicate IDs keep first
- **WHEN** two stable-order rows share the same `#` value and retain policy is omitted
- **THEN** the first row is retained and the duplicate is reported

#### Scenario: No ID requires full-row equality
- **WHEN** no `#` column exists and two rows differ in any column
- **THEN** they are not treated as duplicates by the default rule

### Requirement: Safe row deletion
The engine SHALL provide separate operations for deleting rows by explicit condition and deleting wholly empty rows, and SHALL NOT delete a row solely because one cleaned field is empty while another field contains data.

#### Scenario: Partially populated row survives
- **WHEN** the cleaned column is empty but another column contains data
- **THEN** empty-row deletion preserves the row

### Requirement: Sorting and splitting
The engine SHALL support deterministic sorting and splitting by simple or nested Boolean criteria, with stable tie behavior and explicit output partitions.

#### Scenario: Complex split
- **WHEN** a workflow splits rows using a nested criterion
- **THEN** each row is assigned according to the declared partition rules and partition validations report counts

