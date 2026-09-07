## ADDED Requirements

### Requirement: Explicit logical operation registry
Each logical operation SHALL provide a versioned descriptor, parameter schema,
constructor, validation, deterministic executor, input/output shape, streaming
behavior, and explicit instance-owned registration.

#### Scenario: Duplicate logical operation
- **WHEN** two constructors register the same normalized operation identity and version
- **THEN** the registry rejects the duplicate deterministically

### Requirement: Validate before consuming rows
Every logical operation SHALL validate its complete parameter object and input
schema before reading the first row from its input stream.

#### Scenario: Unknown parameter
- **WHEN** an operation call contains a parameter outside its strict schema
- **THEN** validation fails and the input stream remains unconsumed

### Requirement: Projection and rename operations
The catalog SHALL provide projection and rename operations that use stable
column IDs, preserve requested column order, produce a new schema, and reject
missing or duplicate output identities.

#### Scenario: Project Hebrew columns
- **WHEN** projection selects stable IDs for Hebrew-named source columns in a declared order
- **THEN** the output schema and every row contain exactly those columns in that order

### Requirement: Constant-only Boolean filter
The catalog SHALL provide a filter whose parameters are an explicit expression
tree of `and`, `or`, `not`, and typed comparison leaves between one column and a
constant. It SHALL reject column-to-column and cross-source comparisons.

#### Scenario: Nested filter expression
- **WHEN** a valid nested expression combines constant comparisons
- **THEN** the operation evaluates the explicit tree without inferring precedence and preserves relative order of matching rows

#### Scenario: Column comparison rejected
- **WHEN** a comparison leaf identifies another column as its right operand
- **THEN** validation fails before row consumption

### Requirement: Cleaning and normalization catalog
The catalog SHALL provide leading/trailing trim, registered type or format
normalization, scalar or regex replacement, delete-empty-row, keyed
deduplication, and stable sorting with explicit parameters and deterministic findings.

#### Scenario: Empty-field safety
- **WHEN** a selected field is empty but another field in the row contains data
- **THEN** delete-empty-row does not delete the row unless the operation's explicit whole-row condition matches

#### Scenario: Stable deduplication
- **WHEN** duplicate rows are found and `keep=first` is selected
- **THEN** the first row in source order is retained and later duplicates are reported

### Requirement: Immutable streaming execution
Logical operations SHALL never mutate input values, rows, schemas, or source
files and SHALL wrap streams or use bounded deterministic spooling only when the
declared operation requires it.

#### Scenario: Changed value
- **WHEN** normalization changes a value
- **THEN** the output contains a newly constructed valid `Value` and the input value remains unchanged

#### Scenario: Operation chain
- **WHEN** several streamable operations are chained
- **THEN** data is not round-tripped through an intermediate CSV file

### Requirement: Exact operation descriptor
Every logical descriptor SHALL expose the following complete structure through
`capabilities` and `describe`.

| Field | JSON type | Required | Contract |
|---|---:|:---:|---|
| `id` | `OperationID` | yes | Exact catalog ID |
| `version` | string | yes | `v1` |
| `owner` | `OperationOwner` | yes | `logical_engine` |
| `exposure` | `OperationExposure` | yes | `workflow` |
| `name` | string | yes | English name |
| `description` | string | yes | English behavior summary |
| `deterministic` | boolean | yes | Must be `true` |
| `streaming_mode` | `StreamingMode` | yes | Catalog value below |
| `input_shape` | string | yes | `table` |
| `output_shape` | string | yes | `table` |
| `parameter_schema` | object | yes | Draft 2020-12 JSON Schema with `additionalProperties:false` |
| `required_capabilities` | string[] | yes | Sorted; may be empty |

#### Scenario: Incomplete descriptor
- **WHEN** an operation omits a descriptor field or exposes a parameter schema with open additional properties
- **THEN** registration fails before capability discovery

### Requirement: Exact typed literal and expression structures
Filter and replacement contracts SHALL use a typed literal and a strict tagged
expression union.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `TypedLiteral` | `type_id` | `DataTypeID` | yes | Registered type |
|  | `value` | string | conditional | Canonical parse input; omitted only when `type_id=null` |
| `LogicalGroup` | `op` | string | yes | `and` or `or` |
|  | `children` | `ExpressionNode[]` | yes | At least two |
| `LogicalNot` | `op` | string | yes | `not` |
|  | `child` | `ExpressionNode` | yes | Exactly one |
| `ConstantComparison` | `op` | string | yes | `eq`, `ne`, `lt`, `lte`, `gt`, `gte`, `contains`, `starts_with`, `ends_with`, `matches`, `is_null`, or `is_not_null` |
|  | `column_id` | `ColumnID` | yes | Existing input column |
|  | `value` | `TypedLiteral` | conditional | Required except for null predicates |
|  | `case_sensitive` | boolean | no | Default `true`; text operators only |

Each expression node SHALL contain only the fields of its selected `op` family.
Ordering comparisons require matching comparable data types. Text operators
require `string`; `matches` uses Go RE2 syntax and a validated bounded pattern.

#### Scenario: Invalid tagged expression
- **WHEN** an expression mixes group, not, or comparison fields or uses a text operator on a non-string column
- **THEN** operation validation fails before reading the first row

### Requirement: Complete V1 logical operation catalog
The registry SHALL contain exactly these logical operations and parameter
contracts. Every array preserves request order and rejects duplicate column IDs. This catalog is not a replacement or equivalence contract for `Invoke-RawMemberCleanup.Streaming.py`: only the four behaviors mapped in the sanitized cleanup manifest are V1 coverage evidence; the other fourteen are `outside_v1` for a future OpenSpec change that this change MUST NOT create.

| ID | Version | Streaming | Parameter structure | Output schema effect |
|---|---|---|---|---|
| `table.project` | `v1` | `streaming` | `ProjectParameters` | Select/reorder columns |
| `table.rename` | `v1` | `streaming` | `RenameParameters` | Change source header only; stable IDs unchanged |
| `row.filter` | `v1` | `streaming` | `FilterParameters` | None |
| `text.trim` | `v1` | `streaming` | `TrimParameters` | None |
| `value.normalize` | `v1` | `streaming` | `NormalizeParameters` | Target data type/semantic metadata may change |
| `value.replace` | `v1` | `streaming` | `ReplaceParameters` | None; replacement type must match column |
| `text.regex_replace` | `v1` | `streaming` | `RegexReplaceParameters` | None |
| `row.delete_empty` | `v1` | `streaming` | Empty strict object | None |
| `row.deduplicate` | `v1` | `bounded_spool` | `DeduplicateParameters` | None |
| `row.sort` | `v1` | `bounded_spool` | `SortParameters` | Row order changes only |

#### Scenario: Catalog completeness
- **WHEN** runtime capability discovery runs for the V1 composition root
- **THEN** all ten logical IDs above are present once at `v1` and no excluded operation is registered

### Requirement: Exact logical parameter structures
The ten operations SHALL use these exact parameter objects.

| Structure | Fields | Validation |
|---|---|---|
| `ProjectParameters` | `column_ids:ColumnID[]` | Non-empty; unique; all exist |
| `RenameParameters` | `renames:RenameItem[]` | Non-empty; unique `column_id`; resulting headers non-empty and unique case-insensitively |
| `RenameItem` | `column_id:ColumnID`, `new_header:string` | Header normalized to NFC |
| `FilterParameters` | `expression:ExpressionNode` | Valid constant-only tree |
| `TrimParameters` | `column_ids:ColumnID[]` | Non-empty string columns; trims Unicode leading/trailing whitespace only |
| `NormalizeParameters` | `targets:NormalizeTarget[]`, `invalid_action:"error"|"keep"` | Non-empty unique targets; default invalid action `error` |
| `NormalizeTarget` | `column_id:ColumnID`, `data_type:DataTypeID`, `semantic_type:string?`, `format:string?` | Registered type; optional semantic type must resolve to same data type |
| `ReplaceParameters` | `column_ids:ColumnID[]`, `match:TypedLiteral`, `replacement:TypedLiteral` | Non-empty; all three types match each target column |
| `RegexReplaceParameters` | `column_ids:ColumnID[]`, `pattern:string`, `replacement:string`, `replace_all:boolean` | String columns; non-empty RE2 pattern; maximum 4096 bytes |
| `DeduplicateParameters` | `key_column_ids:ColumnID[]`, `keep:"first"|"last"|"error"` | Empty keys mean column `#` when present, otherwise every column |
| `SortParameters` | `keys:SortKey[]` | Non-empty ordered keys; every column exists and is comparable |
| `SortKey` | `column_id:ColumnID`, `direction:"asc"|"desc"`, `nulls:"first"|"last"` | Each field required |

`row.delete_empty` SHALL delete a row only when every cell in the row is either
canonical null or an empty string. A literal hyphen, zero, or `false` is not
empty.

`row.deduplicate` SHALL use exact state and MAY remain entirely in memory only
while the run budget permits. Every keep mode is `bounded_spool`; a spill must
preserve stable source order, exact keys, cancellation, and cleanup. `keep=first`
retains the first row, `keep=last` retains the last row, and `keep=error` reports
the first duplicate locator.

`row.sort` SHALL be stable. It compares keys in request order through registered
data-type ordering; NFC strings compare by Unicode code point, case-sensitively.
Rows equal on every key retain input order. The implementation SHALL use bounded
external sorting when the full state cannot fit within the run memory budget.

#### Scenario: Delete-empty whole-row safety
- **WHEN** one cell is null or empty but any other cell contains a non-empty canonical value
- **THEN** `row.delete_empty` preserves the row

### Requirement: Operation failure detail catalog
Logical operation failures other than resource exhaustion SHALL return API code
`operation_failed` with an `ErrorDetail` key `operation_reason` whose value is
one of: `invalid_parameters`, `column_not_found`, `duplicate_output_header`,
`type_mismatch`, `invalid_regex`, or `dirty_value`. Resource exhaustion SHALL
return the top-level API code `resource_limit_exceeded`, not
`operation_failed` with `operation_reason=resource_limit_exceeded`. The error
SHALL identify the operation call ID; row-time failures SHALL include the first
source locator.

#### Scenario: Stable operation failure
- **WHEN** normalization encounters the same invalid value under `invalid_action=error`
- **THEN** it returns `operation_failed`, reason `dirty_value`, and the same first source locator
