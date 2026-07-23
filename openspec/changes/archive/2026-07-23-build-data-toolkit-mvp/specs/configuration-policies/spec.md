## ADDED Requirements

### Requirement: Schema-validated configuration
The system SHALL load versioned YAML configuration, accept equivalent JSON, reject unknown invalid values, and report the exact invalid key before execution.

#### Scenario: Invalid collision action
- **WHEN** configuration specifies a collision action outside `overwrite`, `alternate_name`, or `block`
- **THEN** configuration loading fails with the invalid path and allowed values

### Requirement: Deterministic configuration precedence
Resolved settings SHALL apply built-in defaults, user configuration, workflow overrides, and per-output overrides in that order and SHALL record the non-secret resolved values in the execution report.

#### Scenario: Per-output delimiter wins
- **WHEN** configuration defines a TXT delimiter and one output overrides it
- **THEN** that output uses the override while other TXT outputs use the configured default

### Requirement: Approved display defaults
Initial configuration SHALL define true-null rendering as `"-"`, decimal display precision as three places, date display as `02/02/26`, time display as `14:34`, phone display as `052-6105412`, and text output encoding as UTF-8.

#### Scenario: Defaults without workflow overrides
- **WHEN** a workflow formats values without overrides
- **THEN** the approved configured display defaults are used

### Requirement: Global aliases and named styles
Configuration SHALL define globally reusable logical column aliases, named table styles, and one default table style, and SHALL reject duplicate alias identities or missing default styles.

#### Scenario: Unknown style
- **WHEN** an output requests a table style not present in configuration
- **THEN** preflight rejects the output before file creation

### Requirement: Cleaning and Boolean defaults
Configuration SHALL define the default mixed-type action, deduplication retain policy, exception-report detail, and cross-source `not_scope`, whose initial value is `both`.

#### Scenario: NOT uses configured default
- **WHEN** an interface resolves a user request with no selected NOT universe
- **THEN** it uses the configured `not_scope` value and initially sends `both`

### Requirement: Output defaults and collision policy
Configuration SHALL define the default output directory, filename template, table style, TXT delimiter, and collision action, with per-output overrides supported by the workflow contract.

#### Scenario: Existing output is blocked
- **WHEN** the resolved collision action is `block` and the target already exists
- **THEN** the run stops before executing data operations and reports the target conflict

### Requirement: Runtime safety settings
Configuration SHALL define maximum memory, file-lock retry count, interval and timeout, temporary workspace location, failed-run retention, and recent-workflow retention.

#### Scenario: Memory limit is resolved
- **WHEN** a run starts without a workflow memory override
- **THEN** the orchestrator enforces and reports the configured safe local memory budget

### Requirement: Secrets remain external
Ordinary configuration, workflows, logs, and reports MUST NOT contain OAuth access tokens, refresh tokens, client secrets, or other credentials.

#### Scenario: Google adapter obtains credentials
- **WHEN** a Google Sheets workflow starts
- **THEN** the adapter obtains credentials from an external provider and redacts credential material from all reports and errors

