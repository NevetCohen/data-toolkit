## ADDED Requirements

### Requirement: V1 product package layout
The V1 runtime SHALL use explicit packages for canonical core data, logical
operations, file formats and file operations, orchestration, application
services, configuration, reporting, and interface clients.

#### Scenario: Root module is built
- **WHEN** the root module is tested and built
- **THEN** only the V1 packages and executable participate in the root module dependency graph

### Requirement: Legacy MVP isolation
The partial MVP implementation MUST remain available as a nested legacy module
and the V1 runtime MUST NOT import it.

#### Scenario: V1 dependency scan
- **WHEN** architecture tests scan every V1 Go import
- **THEN** no import resolves to the legacy module or a legacy package

### Requirement: Scaffold-only capability state
The V1 foundation SHALL create format and interface package scaffolds without
advertising unimplemented end-user capabilities.

#### Scenario: Initial capability discovery
- **WHEN** a caller requests V1 capabilities before product capabilities are registered
- **THEN** core data types are reported while logical operations and file formats remain empty

### Requirement: Buildable product foundation
The V1 root module SHALL format, test, vet, and build with Go 1.26.4 without
requiring credentials or external source files.

#### Scenario: Foundation quality gate
- **WHEN** the repository quality commands run in a clean environment
- **THEN** the root tests, vet, and executable build complete without using live Google credentials or approved data sources
