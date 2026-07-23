## ADDED Requirements

### Requirement: Complete versioned defaults
The repository SHALL contain a complete non-secret V1 default configuration and
a strict matching schema covering display, aliases, semantic types, styles,
outputs, cleaning, exceptions, Boolean defaults, runtime safety, and interface
history.

#### Scenario: Load embedded defaults
- **WHEN** the application loads configuration without a user file
- **THEN** every required V1 setting resolves from the embedded default document and passes schema and semantic validation

### Requirement: Strict YAML and JSON validation
Configuration SHALL accept equivalent YAML and JSON, reject unknown fields and
invalid enum values, and return field-specific errors.

#### Scenario: Unknown configuration key
- **WHEN** a user configuration contains an undeclared key
- **THEN** validation fails and identifies that key before planning

### Requirement: Deterministic configuration precedence
Effective settings SHALL apply default, user, workflow, and output layers in
ascending precedence without mutating lower layers.

#### Scenario: Output collision override
- **WHEN** every layer supplies a collision setting
- **THEN** the output setting wins while the original default, user, and workflow documents remain unchanged

### Requirement: Unique reusable identities
Configuration SHALL reject duplicate alias, semantic-type, and style identities
and SHALL require the selected default style to exist.

#### Scenario: Duplicate alias with different case
- **WHEN** aliases contain `Phones` and `phones`
- **THEN** semantic validation rejects the duplicate normalized identity

### Requirement: Generic V1 operation envelope
The V1 workflow contract SHALL represent each logical operation with `id`,
`kind`, `inputs`, `parameters`, and `overrides`, and SHALL delegate parameter
validation to the registered executor.

#### Scenario: Future operation shape
- **WHEN** a future operation uses new parameter fields
- **THEN** the shared workflow structure remains unchanged and the operation validates its parameter object

### Requirement: Secret and code exclusion
Configuration and workflow documents MUST NOT contain credentials, compiled
extension registrations, plugin paths, or arbitrary executable code.

#### Scenario: Credential-like field
- **WHEN** a strict configuration document includes an undeclared token or client-secret field
- **THEN** validation rejects the document and no credential value enters a result or event
