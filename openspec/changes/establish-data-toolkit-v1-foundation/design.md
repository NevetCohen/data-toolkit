## Context

Data Toolkit has a validated but incomplete MVP implementation whose package
layout, typed operation union, adapter interface, and orchestrator pipeline were
optimized for incremental delivery of three acceptance workflows. V1 now needs
the product foundation itself: stable extension contracts, explicit component
registries, a shared application API, complete configuration, and a directory
layout that can accept capabilities without adding architectural layers.

The existing implementation is useful reference material but is not the V1
runtime baseline. Approved source datasets remain immutable and are not moved or
processed by this change. Go 1.26.4 remains the required toolchain.

## Goals / Non-Goals

**Goals:**

- Establish a buildable V1 Go module with explicit core, logical, file-engine,
  orchestrator, application, and interface boundaries.
- Make data types, logical operations, file formats, and file operations
  extensible through versioned descriptors, interfaces, constructors,
  validation, and explicit registries.
- Provide a generic V1 workflow operation envelope and one application API for
  every interface.
- Provide complete schema-validated, non-secret defaults with deterministic
  precedence.
- Preserve the partial MVP implementation as an isolated nested module and
  remove it from the V1 dependency graph.
- Prove extension behavior with test-only implementations rather than product
  features.

**Non-Goals:**

- Implementing JSON, CSV, Excel, TXT, Google Sheets, or end-user logical
  operations.
- Delivering an MVP workflow, vertical slice, TUI flow, or production Codex
  integration.
- Maintaining runtime compatibility with the partial MVP executable or its
  workflow schema.
- Loading runtime plugins, using implicit `init` registration, or executing
  arbitrary user code.
- Adding an HTTP service or cross-process protocol.

## Decisions

### 1. Replace the runtime baseline and preserve the MVP as a nested module

The current `cmd`, `internal`, `docs`, `testdata`, module manifest, and module
checksums will move beneath `legacy/mvp`. That directory will contain its own
`go.mod`, so root `go test ./...` and root builds do not traverse it. A V1
architecture test will reject imports containing `/legacy/`.

The active `build-data-toolkit-mvp` OpenSpec change will be archived with
`--skip-specs`. This retains its planning evidence without making superseded MVP
requirements the main specifications. Deleting the old code was rejected
because it would remove useful implementation evidence; retaining it in the V1
runtime was rejected because it would create two active architectures.

### 2. Use packages that mirror product ownership

The V1 module uses:

- `internal/core/{datatype,cell,column,table}` for canonical data;
- `internal/logical/operation` for format-neutral logical execution;
- `internal/fileengine/{format,operation,formats/...}` for external data
  boundaries and shared file actions;
- `internal/orchestrator` for validation, planning, and registry dispatch;
- `internal/application` for composition and the interface-neutral service;
- `internal/interfaces/{cli,tui}` for clients of that service;
- `internal/config` and `internal/report` for cross-cutting policy and results.

Canonical identity types live in the parent `internal/core` package so child
packages can share identities without importing one another cyclically.

### 3. Define one explicit template shape for extension families

An extension template consists of:

1. a stable identifier and versioned descriptor;
2. a behavior interface;
3. an explicit constructor;
4. deterministic validation;
5. registration in an instance-owned registry.

Registries reject empty identities, invalid descriptors, duplicate identities,
and incompatible versions. They return sorted capability descriptions. The
composition root receives registry instances explicitly. Global mutable
registries and `init` functions were rejected because they hide dependencies
and make tests order-dependent.

### 4. Keep canonical values immutable and columns metadata-only

A canonical value carries a data-type identity and canonical encoded content.
The data-type handler owns parsing, validation, comparison, and rendering.
Cells carry values and provenance. Columns carry identity, header, subheader,
resolved type, type authority, optional physical range, and an opaque source
binding that logical code preserves but never interprets.

Rows carry cells; columns do not contain a slice of all values. Row streams are
the data transport boundary. Embedding values in columns was rejected because
it conflicts with the intended 500,000-row scale and bounded-memory execution.

### 5. Register operation implementations, not declarations alone

Each logical operation implements `Descriptor`, `Validate`, and `Execute`.
Workflow operations use the generic fields `id`, `kind`, `inputs`,
`parameters`, and `overrides`. The registry stores the executor and owns
lookup. The orchestrator resolves dependencies and calls the selected executor;
it contains no switch over operation kinds.

Parameter bytes remain JSON so each operation can own its parameter schema.
The foundation validates generic envelope rules and delegates parameter
validation to the registered executor. A central typed union was rejected
because every new operation would require edits to shared contracts.

### 6. Compose file formats from capability-specific interfaces

A format provider exposes a descriptor and implements any supported capability
interfaces: inspect, map, read, write, style, and validate. Its declared
capabilities must match the interfaces it implements. The format registry
validates this at registration and rejects unsupported capability requests
before execution.

Snapshot, lock, staging, publication, and cleanup are format-neutral file
operation contracts. Format folders are committed scaffolds with documentation,
not implementations that falsely advertise capabilities.

### 7. Expose one in-process application API

The service exposes capability discovery, configuration validation, workflow
validation, mapping, planning, and running. It accepts an observer for ordered
events and returns typed results. CLI and TUI packages depend only on this API.
Codex will eventually invoke the same executable JSON surface.

The orchestrator can execute test-registered logical operations and can route
mapping to test-registered file formats, proving dispatch without shipping
product capabilities. No network API is introduced.

### 8. Make the default configuration a versioned product artifact

`configs/default.yaml` and `configs/schema/config-v1.schema.json` are embedded
by a small `configs` asset package. The internal config package decodes strict
YAML or JSON, validates the schema and semantic identities, and resolves
built-ins, user settings, workflow overrides, and output overrides in that
order.

The document includes display, aliases, semantic types, styles, output,
cleaning, exceptions, Boolean defaults, runtime limits, locks, temporary
retention, and interface history. Credentials and compiled registrations are
not configuration.

## Risks / Trade-offs

- **[The foundation temporarily removes usable MVP behavior]** → Preserve the
  old implementation as a buildable nested module and document the V1 runtime
  as scaffold-only.
- **[Fine-grained core packages create import complexity]** → Keep shared
  identities in `internal/core` and enforce a one-way dependency graph.
- **[Generic JSON parameters reduce compile-time checking]** → Require every
  registered executor to validate its own parameters before execution and test
  invalid parameter rejection.
- **[Scaffold format folders may be mistaken for support]** → Do not register
  them and label each package as a V1 placeholder.
- **[Default configuration and Go types can drift]** → Embed the authoritative
  YAML/schema pair and add default-load, round-trip, and semantic validation
  tests.
- **[The nested legacy module may accidentally become a dependency]** → Add
  import-boundary tests and omit any root workspace file that includes it.

## Migration Plan

1. Create and strictly validate all V1 OpenSpec artifacts.
2. Archive `build-data-toolkit-mvp` with `--skip-specs`.
3. Move the partial runtime into the `legacy/mvp` nested module and verify its
   files remain available.
4. Create the new root Go module and V1 package tree.
5. Implement core contracts, registries, orchestrator, application API,
   configuration, and interface scaffolds.
6. Add extension, boundary, configuration, CLI, TUI, and build tests.
7. Update project and architecture documentation to identify the V1 foundation
   as the only active runtime.
8. Run formatting, root tests, vet, build, and strict OpenSpec validation.

Rollback is a Git revert of the foundation change. Approved external sources
are never touched, and the legacy module preserves the prior runtime snapshot.

## Open Questions

None. Concrete V1 operations, formats, workflows, and interface experiences
will be specified in subsequent OpenSpec changes against this foundation.
