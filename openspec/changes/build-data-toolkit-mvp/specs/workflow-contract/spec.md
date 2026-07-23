## ADDED Requirements

### Requirement: Versioned workflow documents
The system SHALL accept workflow documents in schema-validated YAML and JSON representations that normalize to the same typed contract and carry an explicit supported schema version.

#### Scenario: Equivalent YAML and JSON
- **WHEN** equivalent supported workflow definitions are submitted as YAML and JSON
- **THEN** the system produces equivalent normalized workflow objects and execution plans

#### Scenario: Unsupported workflow version
- **WHEN** a workflow declares an unsupported schema version
- **THEN** the system rejects it before mapping or execution with an explicit version error

### Requirement: Complete workflow declaration
The workflow contract SHALL declare sources, optional sheets, ordered operations, field mappings, outputs and/or queries, overrides, exception behavior, and validations without requiring free-form interpretation.

#### Scenario: Missing required operation parameter
- **WHEN** an operation omits a parameter required by its versioned schema
- **THEN** contract validation fails before any source is processed

### Requirement: Explicit Boolean expressions
The workflow contract SHALL encode conditions as expression trees with explicit grouping and SHALL include `not_scope` for every normalized cross-source expression that contains `NOT`.

#### Scenario: Interface defaults NOT scope
- **WHEN** a user creates a cross-source `NOT` expression without selecting a universe
- **THEN** the interface submits the normalized workflow with `not_scope: both`

#### Scenario: Ambiguous free-form precedence
- **WHEN** a request does not express Boolean grouping in the contract structure
- **THEN** the system rejects the request instead of inferring precedence from text

### Requirement: Output and query declarations
Each output SHALL declare its format, location or location policy, projected columns, style where applicable, and validations; each query SHALL declare a deterministic result shape.

#### Scenario: Two outputs use different overrides
- **WHEN** one workflow creates two outputs with different collision and naming overrides
- **THEN** the orchestrator resolves and reports each output independently

### Requirement: Execution result contract
The result contract SHALL return run status, produced files and/or query answers, resolved configuration, validation results, exceptions, warnings, and detailed errors without exposing credentials.

#### Scenario: Successful run with exceptions
- **WHEN** a workflow completes under an exception-report policy
- **THEN** the result includes successful outputs, validations, stable exception references, and the resolved non-secret settings

