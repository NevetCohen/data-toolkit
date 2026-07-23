## Why

The current repository is organized around a partially implemented MVP change,
while the next phase requires a stable product foundation for Data Toolkit V1.
The project needs explicit extension templates, package boundaries, a shared
application API, and complete non-secret configuration before V1 capabilities
are developed.

## What Changes

- **BREAKING** Replace the active MVP-oriented runtime layout with a clean V1
  foundation organized around canonical core data, logical operations, file
  formats and file operations, orchestration, application services, and
  interface clients.
- Archive `build-data-toolkit-mvp` without syncing its delta specifications and
  preserve its partial implementation as an isolated nested legacy module that
  the V1 runtime cannot import.
- Define compile-time extension templates as a versioned descriptor, behavior
  interface, constructor, validation, and explicit registry entry.
- Add canonical data-type, cell, column, table, and row-stream contracts.
- Add logical-operation and file-format registries that own implementations and
  eliminate kind-based dispatch from the orchestrator.
- Add capability-specific file operation contracts for inspection, mapping,
  reading, writing, styling, and validation, plus shared snapshot, staging,
  publication, and lock services.
- Add one interface-neutral application API for capability discovery,
  configuration and workflow validation, mapping, planning, and execution.
- Add a complete schema-validated V1 default configuration and a generic
  versioned workflow operation envelope.
- Scaffold V1 format and interface packages without implementing end-user data
  capabilities.

## Capabilities

### New Capabilities

- `v1-foundation`: Product layout, legacy isolation, dependency boundaries, and
  buildable V1 scaffolding.
- `canonical-data`: Extensible canonical data types, cells, columns, tables,
  provenance, and row streams.
- `extension-contracts`: Logical-operation, file-format, and file-operation
  descriptors, interfaces, validation, and registries.
- `application-api`: Interface-neutral capabilities, validation, mapping,
  planning, run, event, and result contracts.
- `configuration`: Complete V1 defaults, schema, precedence, workflow envelope,
  and secret boundaries.

### Modified Capabilities

None. The former MVP delta specifications are superseded and will be archived
without being synchronized as main specifications.

## Impact

- Replaces the active Go module layout beneath `internal/` and the executable
  composition root.
- Introduces a nested read-only reference module beneath `legacy/mvp`.
- Changes the workflow contract to the V1 generic operation envelope.
- Replaces MVP-oriented architecture documentation and project metadata.
- Adds no server, runtime plugin loader, arbitrary workflow code, or complete
  JSON, CSV, Excel, TXT, Google, or logical-operation implementation.
