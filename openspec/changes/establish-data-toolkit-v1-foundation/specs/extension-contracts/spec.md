## ADDED Requirements

### Requirement: Explicit extension template
Each extension family SHALL use a versioned descriptor, behavior interface,
constructor, deterministic validation, and explicit instance-owned registry.

#### Scenario: Duplicate extension identity
- **WHEN** two extensions with the same normalized identity are registered
- **THEN** the second registration fails deterministically without replacing the first

### Requirement: Logical operation executor registry
The logical-operation registry SHALL store executable implementations and the
orchestrator SHALL dispatch through the registry without switching on operation
kind.

#### Scenario: Execute a test operation
- **WHEN** a test registers a new executor and submits its generic operation envelope
- **THEN** the orchestrator validates and executes it without modification to orchestrator source

### Requirement: Operation-owned parameters
Each logical operation SHALL validate its own JSON parameters and reject invalid
parameters before consuming input rows.

#### Scenario: Invalid custom parameters
- **WHEN** a registered test operation receives parameters that violate its contract
- **THEN** execution fails before the input stream is read

### Requirement: Capability-composed file formats
A file format SHALL declare its capabilities and implement the corresponding
inspect, map, read, write, style, or validate interfaces.

#### Scenario: Capability declaration mismatch
- **WHEN** a provider declares a capability whose interface it does not implement
- **THEN** format registration fails before the provider can be selected

### Requirement: Shared file operation contracts
Snapshot, locking, staging, validation, publication, and cleanup SHALL be
format-neutral file-operation contracts that can be composed independently of
format implementations.

#### Scenario: New format uses publication contract
- **WHEN** a future format implementation adds output support
- **THEN** it can use the shared staging and publication contracts without reimplementing their interfaces

### Requirement: No implicit runtime extensions
The V1 runtime MUST NOT discover extensions through package initialization,
dynamic plugins, arbitrary code in workflows, or configuration-supplied code.

#### Scenario: Composition root starts
- **WHEN** the application service is constructed
- **THEN** only extensions passed explicitly to its registries are available
