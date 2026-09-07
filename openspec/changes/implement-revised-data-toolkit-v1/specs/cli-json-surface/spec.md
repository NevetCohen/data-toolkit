## ADDED Requirements

### Requirement: Required local commands
The English CLI SHALL expose `capabilities --json`, `describe --kind operation
--id <id> --json`, `config validate --request <file> --json`, `inspect --request
<file> --json`, `workflow validate --file <workflow.json> [--config <config>]
[--var <id>=<value>] [--config-set <json-pointer>=<json-value>] --json`, and
`run --file <workflow.json> [--config <config>] [--var <id>=<value>]
[--config-set <json-pointer>=<json-value>] --expected-workflow-digest <digest>
--json` through the shared application API.

#### Scenario: Command inventory
- **WHEN** CLI help is requested
- **THEN** all required command families are present and no TUI, server, MCP, or remote-execution command is advertised as V1

### Requirement: Runtime-derived discovery
The `capabilities` and `describe` commands SHALL serialize sorted descriptors
from the compiled runtime registries and SHALL NOT read a duplicated capability
catalog from documentation or configuration.

#### Scenario: Newly registered operation
- **WHEN** the composition root registers a valid operation
- **THEN** `capabilities` and `describe` expose it without changes to CLI command logic

### Requirement: JSON stdout isolation
With `--json`, each command SHALL write exactly one complete versioned JSON
result to stdout. Human diagnostics, progress, and unexpected runtime detail
SHALL go to stderr.

#### Scenario: Validation error
- **WHEN** `workflow validate --json` rejects a request
- **THEN** stdout remains valid machine-readable JSON and stderr contains no duplicate JSON result

### Requirement: Stable exit status
The CLI SHALL return exit code zero only for a successful requested operation
and SHALL map validation, capability, lock, execution, cancellation, and
publication failures to documented non-zero exit behavior.

#### Scenario: Unsupported operation exit
- **WHEN** workflow validation finds an unsupported operation
- **THEN** the command returns the documented validation or capability exit class and a structured JSON error

### Requirement: Thin interface boundary
CLI code SHALL parse flags and files, call the application API, and render
results without owning product rules, registries, workflow planning, or file
adapter behavior.

#### Scenario: API substitution test
- **WHEN** a fake application API is provided to a CLI command test
- **THEN** the command behavior is testable without real adapters or duplicated business logic

### Requirement: Workflow-template loading and CLI patches
The CLI SHALL load only a strict `WorkflowTemplateDocument` for workflow
validation and run; it SHALL NOT accept a long command that describes logical
operations. It SHALL accept repeatable `--var <id>=<value>` and `--config-set
<json-pointer>=<json-value>` flags, plus optional `--config <path>`. Variable
bindings are parsed according to declared type; object and array values use JSON
syntax. Config patches are passed unchanged to the application layer, which
enforces protected paths and full configuration validation.

#### Scenario: Protected configuration patch
- **WHEN** `--config-set` targets a protected pointer, its parent, or its child
- **THEN** validation returns a structured configuration error before source access
