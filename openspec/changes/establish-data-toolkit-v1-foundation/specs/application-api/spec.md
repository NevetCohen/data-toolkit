## ADDED Requirements

### Requirement: Shared application service
The application layer SHALL expose capability discovery, configuration
validation, workflow validation, mapping, planning, and run methods through one
interface-neutral Go API.

#### Scenario: CLI and TUI use the service
- **WHEN** CLI and TUI scaffold clients request capability discovery
- **THEN** both receive results from the same application service without implementing registry logic

### Requirement: Capability discovery
Capability results SHALL list registered data types, logical operations, file
formats, and file capabilities in deterministic order.

#### Scenario: Registrations occur in different order
- **WHEN** equivalent registries are populated in different insertion orders
- **THEN** capability discovery returns identical sorted results

### Requirement: Registry-based mapping
The application mapping method SHALL select a registered file-format mapper and
reject missing formats or capabilities before invoking it.

#### Scenario: Format lacks mapping
- **WHEN** a caller requests mapping from a format without the mapping capability
- **THEN** the API returns a capability error without invoking another format

### Requirement: Deterministic planning and execution
The orchestrator SHALL validate generic workflow operation envelopes, create a
stable dependency order, and execute registered logical operations through
their executors.

#### Scenario: Operation dependency chain
- **WHEN** one registered operation consumes the output identity of a preceding operation
- **THEN** planning and ordered run execute both exactly once in dependency order

### Requirement: Ordered application events
Application runs SHALL emit monotonically sequenced validation, planning,
execution, and completion events through an optional observer.

#### Scenario: Run observer is installed
- **WHEN** a valid test workflow runs with an observer
- **THEN** the observer receives stable increasing sequence numbers and one terminal event

### Requirement: No network service
The V1 foundation SHALL expose no HTTP server or remote execution endpoint.

#### Scenario: Executable starts
- **WHEN** the V1 executable starts for capability or configuration commands
- **THEN** it performs the command locally without opening a listening socket
