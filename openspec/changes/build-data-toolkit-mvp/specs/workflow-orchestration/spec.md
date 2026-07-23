## ADDED Requirements

### Requirement: Validate and map before planning
The orchestrator SHALL validate the workflow contract and resolved configuration, map all sources, resolve approval and policy decisions, and check output targets before producing a final execution plan.

#### Scenario: Output collision found in preflight
- **WHEN** mapping succeeds but an output collision resolves to `block`
- **THEN** the orchestrator returns the conflict without running logical operations

### Requirement: Explicit ordered execution plan
The orchestrator SHALL produce an ordered plan that identifies each operation, responsible engine, inputs, outputs, resource constraints, repairs, and validations.

#### Scenario: JSON to styled Excel
- **WHEN** a workflow projects JSON fields, renames one field, and requests Excel output
- **THEN** the plan orders adapter read, logical projection/rename, adapter write/style, and output validation explicitly

### Requirement: Deterministic bounded execution
The orchestrator SHALL enforce the resolved memory budget, stable operation order, bounded spooling where required, and cancellation without publishing incomplete successful outputs.

#### Scenario: Blocking sort exceeds memory buffer
- **WHEN** a sort cannot complete within its in-memory allocation
- **THEN** the orchestrator uses a deterministic temporary spool or fails under policy without exceeding the resolved budget

### Requirement: Run status and completion reporting
The orchestrator SHALL report that work has started after preflight succeeds and SHALL return a final result containing outputs and/or query answers, validations, warnings, exceptions, and detailed errors.

#### Scenario: Query and file output complete together
- **WHEN** one workflow produces a file and a query answer
- **THEN** the final result includes both with their validation and execution details

### Requirement: Verification before success
The orchestrator SHALL treat adapter and logical validations as required plan steps and SHALL NOT mark an output or workflow successful when a required validation fails.

#### Scenario: Row-count assertion fails
- **WHEN** a required output row-count assertion fails
- **THEN** the workflow is reported unsuccessful and the output is not published as a successful final artifact

### Requirement: Immutable-source enforcement
The orchestrator SHALL reject any operation or adapter plan that attempts to write to a declared source path or Google source identity.

#### Scenario: Output equals source path
- **WHEN** an output resolves to the same path as a source
- **THEN** preflight rejects the workflow regardless of overwrite configuration

### Requirement: Exception order and provenance
When an exception report is requested, the orchestrator SHALL prioritize stable record order, complete provenance, and deterministic reasons over maximum throughput.

#### Scenario: Mixed-type rows are excluded
- **WHEN** incompatible rows are excluded under an exception policy
- **THEN** the report lists them in stable source order with source, sheet, row, column, observed value class, and reason

