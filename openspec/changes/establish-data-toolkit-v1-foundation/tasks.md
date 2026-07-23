## 1. Supersede the MVP Baseline

- [x] 1.1 Archive `build-data-toolkit-mvp` with `--skip-specs` and verify it is no longer active.
- [x] 1.2 Move the partial MVP module, documentation, and fixtures into an isolated `legacy/mvp` nested Go module.
- [x] 1.3 Create the Go 1.26.4 V1 root module and the approved product package tree.

## 2. Canonical Core

- [x] 2.1 Implement shared canonical identities and the immutable data-type value, handler, descriptor, and registry contracts.
- [x] 2.2 Register and test the canonical V1 core data types through the same data-type template.
- [x] 2.3 Implement validated cell provenance and column metadata, range, type-authority, and opaque source-binding contracts.
- [x] 2.4 Implement canonical tables, rows, and deterministic row-stream contracts without column value materialization.

## 3. Extension Contracts

- [x] 3.1 Implement the generic workflow operation envelope plus logical-operation descriptor, executor, validation, and registry.
- [x] 3.2 Implement file-format descriptors, capability-specific interfaces, and a registry that verifies declared capabilities.
- [x] 3.3 Implement format-neutral snapshot, lock, staging, publication, validation, and cleanup operation contracts.
- [x] 3.4 Add unregistered JSON, CSV, Excel, TXT, and Google format scaffolds with no advertised runtime capability.

## 4. Configuration

- [x] 4.1 Add the embedded complete `configs/default.yaml` and strict `config-v1` JSON schema.
- [x] 4.2 Implement strict YAML/JSON loading, semantic validation, and field-specific configuration errors.
- [x] 4.3 Implement immutable default, user, workflow, and output precedence resolution.

## 5. Orchestration and Application API

- [x] 5.1 Implement deterministic workflow validation, dependency planning, registry dispatch, ordered events, and test-operation execution.
- [x] 5.2 Implement registry-based file mapping and capability rejection before provider invocation.
- [x] 5.3 Implement the interface-neutral application service for capabilities, config/workflow validation, mapping, planning, and run.

## 6. Interfaces and Executable

- [x] 6.1 Implement CLI scaffolding for deterministic capability discovery and configuration validation.
- [x] 6.2 Implement a TUI client scaffold that delegates capability and configuration calls to the application service.
- [x] 6.3 Wire the V1 executable composition root without product operations, file formats, credentials, or a network listener.

## 7. Architecture, Documentation, and Quality Gates

- [x] 7.1 Add external extension tests proving custom data types, operations, and formats require registration only and reject invalid or duplicate definitions.
- [x] 7.2 Add dependency-boundary tests that reject V1 imports of legacy code and outward layer violations.
- [x] 7.3 Update project and architecture documentation for the V1 foundation and legacy boundary.
- [x] 7.4 Run formatting, root tests, vet, executable build, and strict OpenSpec validation.
