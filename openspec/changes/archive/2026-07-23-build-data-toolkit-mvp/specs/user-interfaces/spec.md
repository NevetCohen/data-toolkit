## ADDED Requirements

### Requirement: Shared interface contract
Terminal commands, the interactive TUI, and Codex integration SHALL use the same versioned workflow, mapping, validation, and result contracts and SHALL NOT implement independent data logic.

#### Scenario: Saved workflow runs from two interfaces
- **WHEN** the same saved workflow runs from a terminal command and the TUI
- **THEN** both interfaces submit equivalent normalized requests and receive equivalent results

### Requirement: English user interfaces
All interface labels, prompts, help, errors, schemas, and technical interface documentation SHALL be in English while preserving user data in any supported Unicode language.

#### Scenario: Hebrew source headers
- **WHEN** a source contains Hebrew column headers
- **THEN** the English interface displays the headers unchanged as user data

### Requirement: Mapping helper command
The terminal interface SHALL provide a command that shows source column names and optionally extended type and format counts for selected columns.

#### Scenario: Request selected-column details
- **WHEN** a user requests extended mapping for one column
- **THEN** the command returns its identity, inferred type counts, observed formats, and cleanliness findings

### Requirement: Terminal commands without aliases
Terminal commands SHALL permit a column reference by original source header without requiring a global logical alias to be assigned first.

#### Scenario: Filter by original header
- **WHEN** a command filters an unmapped source using its original `ישוב` header
- **THEN** mapping resolves the header and the workflow proceeds without a pre-existing alias

### Requirement: Interactive workflow flow
The TUI SHALL support source intake, basic or extended mapping, adding sources, defining outputs and queries, editing output metadata and operations, explicit Boolean grouping and NOT scope, approval, execution review, and saving or loading workflows.

#### Scenario: User builds a two-source NOT filter
- **WHEN** the user creates a two-source filter containing `NOT`
- **THEN** the TUI presents source scope, applies `both` when no alternative is selected, and submits explicit `not_scope`

### Requirement: Recent and pasted workflows
The interface SHALL retain a configurable number of recent workflows and SHALL accept schema-validated pasted YAML or JSON workflow text.

#### Scenario: Invalid pasted workflow
- **WHEN** pasted workflow text fails schema validation
- **THEN** the interface shows the exact validation errors and does not execute it

### Requirement: Codex discovery and use
Codex-facing README and skills SHALL expose supported operations and schemas, direct Codex to invoke the toolkit when it supports a task, and assign output validation to the toolkit rather than to generated one-off scripts.

#### Scenario: Supported Codex request
- **WHEN** Codex receives a data task expressible by the installed workflow schema
- **THEN** it constructs and validates a workflow and invokes the toolkit instead of creating a task-specific data script

