# Data Toolkit V1 Foundation Architecture

## Status

`establish-data-toolkit-v1-foundation` is the active product baseline. It
creates the contracts and package boundaries required for V1 development. It
does not implement end-user operations or file formats.

The former MVP runtime is preserved under `legacy/mvp` in a nested Go module.
Root commands do not traverse it, and an architecture test rejects imports from
V1 into legacy code.

## Dependency graph

```text
cmd/data-toolkit
        |
internal/interfaces/{cli,tui}
        |
internal/application
        |
internal/orchestrator
      /        \
logical      fileengine
      \        /
      core + report

config -> embedded configs assets
```

Dependencies point inward. Core owns only canonical identities and models.
Logical operations do not access files. File formats do not contain logical
transformations. The orchestrator validates dependencies and dispatches
registered implementations; it does not switch on operation or format kinds.
The application layer is the only API used by interface packages.

## Extension template

Every extension family has the same structural contract:

1. stable identity and versioned descriptor;
2. behavior interface;
3. explicit constructor;
4. deterministic validation;
5. explicit instance-owned registry.

Registries reject empty or duplicate identities, unsupported versions, invalid
descriptors, and capability mismatches. Registration order does not affect
capability output. Extensions are compiled into the composition root. Runtime
plugin discovery, package-level `init` registration, and executable
configuration are forbidden.

### Canonical data types

A data-type handler parses text into an immutable canonical `Value`, validates
the value, compares two values deterministically, and renders a value under
explicit display options. Null, string, Boolean, exact integer, exact decimal,
date, time, JSON object, and JSON array use this same registry contract.

### Logical operations

Each executor provides a descriptor, parameter schema, validation, and
`Execute`. Workflows use a generic operation envelope:

```yaml
id: normalized_members
kind: future.normalize
inputs: [members]
parameters: {}
overrides: {}
```

The operation registry stores the executable implementation and validates its
JSON parameters before execution. The orchestrator resolves only operation
identities and dependencies.

### File formats and file operations

A file-format provider declares and implements only the capabilities it owns:
inspect, map, read, write, style, or validate. Registration fails if declared
and implemented capabilities differ. The JSON, CSV, Excel, TXT, and Google
packages advertise no capabilities in this change.

Snapshot, lock, staging, output validation, publication, and cleanup are
format-neutral contracts under `internal/fileengine/operation`.

## Canonical data flow

`table.Schema` contains column metadata and opaque source bindings.
`table.RowStream` carries the actual values incrementally. A `cell.Cell`
contains an immutable canonical value plus source, sheet, row, and column
provenance. Columns never materialize all their values.

## Application API

The in-process API exposes:

- `Capabilities`
- `ValidateConfig`
- `ValidateWorkflow`
- `Map`
- `Plan`
- `Run`

CLI and TUI call this API directly. Future Codex integration will use the same
API or its JSON/YAML contract. V1 foundation opens no network listener.

Runs emit monotonically increasing validation, planning, operation, and
terminal events through an optional observer. Plans preserve the deterministic
workflow order and reject forward or missing dependencies.

## Configuration

`configs/default.yaml` is the complete non-secret V1 default. The embedded
`configs/schema/config-v1.schema.json` rejects unknown fields and invalid
values before decoding to typed Go configuration.

Precedence is:

```text
defaults -> user config -> workflow overrides -> output overrides
```

Resolution returns a copy and never mutates a lower-precedence layer.
Credentials and extension registrations are intentionally outside the config.

## Adding future product capabilities

A new capability requires its own OpenSpec change. Its implementation supplies
the appropriate interface and constructor, registers it explicitly in the
composition root, adds contract and product tests, and updates capability
documentation. Adding a logical operation must not require an orchestrator
edit; adding a format must not require a logical-engine edit.
