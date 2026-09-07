## ADDED Requirements

### Requirement: Runtime-first Codex sequence
The Codex Adapter SHALL instruct Codex to call capability discovery, describe
candidate operations, inspect an unfamiliar or ambiguous source when needed,
resolve required decisions, construct and validate a request, run it, and
interpret the final report. It SHALL not require inspection before every
otherwise complete repeat run.

#### Scenario: Supported natural-language task
- **WHEN** a user asks for a task expressible by registered capabilities
- **THEN** Codex uses the Data Toolkit binary and does not create a one-off data-processing script

### Requirement: No duplicate capability catalog
The Codex Skills SHALL describe procedure and decision rules but SHALL NOT embed
an authoritative list of operation IDs, format IDs, or parameter schemas.

#### Scenario: Runtime catalog changes
- **WHEN** a registered capability is added or removed
- **THEN** Codex follows the new runtime discovery result without a skill catalog update

### Requirement: Clarification instead of guessing
The Codex Adapter SHALL require clarification when inspection or request
validation identifies ambiguous sheet, delimiter, header, root path, duplicate
column name, output target, or user intent that changes the result.

#### Scenario: Duplicate headers
- **WHEN** inspection reports duplicate headers and the requested column cannot be identified uniquely
- **THEN** Codex asks the user to select a stable column ID before execution

### Requirement: Explicit unsupported-capability behavior
Codex SHALL report a capability gap when the runtime cannot express the task and
SHALL NOT invent descriptors, parameters, or successful results. It SHALL not
silently bypass the toolkit with new processing code.

#### Scenario: Cross-source join request
- **WHEN** the user requests a V1-excluded join
- **THEN** Codex reports that the capability is unavailable and does not submit an invalid workflow

### Requirement: Machine-error self-correction
Codex SHALL use structured validation errors for at most one corrected request
attempt when the user's intent is already clear; it SHALL ask for clarification
when correction would require a user decision.

#### Scenario: Unknown request field
- **WHEN** validation rejects a mechanically removable unknown field
- **THEN** Codex corrects the request once and revalidates without changing user intent

### Requirement: Versioned adapter fixtures
The adapter SHALL include golden request and response fixtures for discovery,
clarification, validation, execution, invented capability, unknown field,
override precedence, and report interpretation.

#### Scenario: Skill regression
- **WHEN** the skill or CLI contract changes
- **THEN** its fixture suite detects any divergence from the supported sequence and safety rules

### Requirement: Exact Codex Skill inventory
The Codex Adapter SHALL contain exactly the following six project-local Skills.
These Skills define procedure only; operation IDs, format IDs, and parameter
schemas SHALL be read from `capabilities` and `describe` at runtime.

| Skill | Location | Trigger | Primary output | Handoff |
|---|---|---|---|---|
| `data-toolkit` | `.agents/skills/data-toolkit/SKILL.md` | User requests local transformation of JSON, CSV, Excel, or a Google Spreadsheet | Selected adapter path, unsupported-capability result, or user clarification | `data-toolkit-google-sheets`, `data-toolkit-inspect`, `data-toolkit-build-workflow`, or stop |
| `data-toolkit-inspect` | `.agents/skills/data-toolkit-inspect/SKILL.md` | Source is unfamiliar or source structure is unresolved | `InspectionReport` plus resolved decisions or a focused question | `data-toolkit-build-workflow` |
| `data-toolkit-build-workflow` | `.agents/skills/data-toolkit-build-workflow/SKILL.md` | Intent and required source decisions are known | Template file path, exact bindings/patches, validated resolved request, and both digests | `data-toolkit-run` |
| `data-toolkit-run` | `.agents/skills/data-toolkit-run/SKILL.md` | A validated template handoff and matching digest exist | Published output plus concise `RunReport`, or one structured failure | `data-toolkit-recover` only when eligible |
| `data-toolkit-recover` | `.agents/skills/data-toolkit-recover/SKILL.md` | First attempt failed with a mechanically repairable structured error | One corrected validated request, or a focused question/stop result | `data-toolkit-run` once or stop |
| `data-toolkit-google-sheets` | `.agents/skills/data-toolkit-google-sheets/SKILL.md` | Source or final destination is a Google Spreadsheet | Local XLSX bridge result, verified remote Spreadsheet link, or a capability-gap stop | Local Skills before or after the bridge, as applicable |

#### Scenario: Skill family discovery
- **WHEN** project-local Skills are enumerated
- **THEN** all six names and locations above exist and no additional Data Toolkit execution Skill competes with their responsibilities

### Requirement: Root Skill contract
`data-toolkit` SHALL accept the user's natural-language intent, source path,
desired output, and any explicit preferences already supplied. It SHALL run
`capabilities --json`, reject V1-excluded tasks, decide whether inspection is
needed, detect a Google Spreadsheet source or target before local processing,
and hand off without constructing operation parameters itself.

Its terminal outputs are exactly one of: `unsupported`, `needs_clarification`,
or a handoff containing runtime descriptors, user intent, source/output paths,
and resolved preferences. It SHALL stop before any write when source and output
normalize to the same path.

#### Scenario: Unsupported join
- **WHEN** runtime discovery and V1 scope show that the requested task requires a join
- **THEN** the root Skill returns `unsupported` without inventing a workflow or invoking a data-processing script

### Requirement: Inspection Skill contract
`data-toolkit-inspect` SHALL accept a `SourceRequest`, capabilities digest, and
optional already-resolved decisions. It SHALL describe the selected format when
needed, call `inspect`, and return the report unchanged plus a separate ordered
list of decisions supplied by the user. It SHALL not edit the source or build a
transform request.

#### Scenario: Ambiguous worksheet
- **WHEN** inspection requires `select_sheet`
- **THEN** the Skill asks one focused question using the reported options and does not hand off until the answer is resolved

### Requirement: Workflow Builder Skill contract
`data-toolkit-build-workflow` SHALL accept user intent, runtime descriptors,
source inspection or known source contract, output target, and resolved
preferences. For each candidate operation it SHALL call `describe`, construct
only fields allowed by the returned schemas, call `workflow validate`, and
return the validated request plus both digests.

It SHALL stop with `needs_clarification` when correction would change the user's
requested result. It SHALL stop with `unsupported` when no registered operation
expresses the intent.

The Workflow Builder Skill SHALL read the repository instruction
`docs/codex/workflow-json-format.md` before constructing a template. It SHALL
write a strict `WorkflowTemplateDocument`, invoke `workflow validate --file`,
and pass only repeatable `--var` and `--config-set` values that have been
resolved from user choices. It SHALL preserve the returned workflow digest and
effective-configuration digest for the Run Skill.

#### Scenario: Unknown operation parameter
- **WHEN** a desired parameter is absent from the runtime descriptor schema
- **THEN** the Skill reports a capability gap instead of adding the field

### Requirement: Run Skill contract
`data-toolkit-run` SHALL accept the exact validated workflow-template file path,
optional user-config path, ordered `--var` bindings, ordered `--config-set`
patches, expected workflow digest, effective-configuration digest,
capabilities digest, and user authorization implicit in the active request. It
SHALL call `run --file` with the same template and bindings used by validation,
plus the expected workflow digest. It SHALL verify request correlation, surface
the published output, validations, exceptions, source hash, terminal failure
when present, and receipt ID, and perform no independent data transformation.

#### Scenario: Successful run handoff
- **WHEN** `run` returns a succeeded `RunReport`
- **THEN** the Skill reports the exact published path and validation summary and does not invoke recovery

### Requirement: Google Sheets bridge Skill contract
`data-toolkit-google-sheets` SHALL be the only V1 Codex procedure that handles
a Google Spreadsheet. Google Sheets SHALL NOT be a Go `FormatID`, runtime
format provider, or toolkit network integration. Before processing it SHALL
check that the callable Google Drive connector operations needed for the chosen
direction are available: metadata lookup, export or native spreadsheet import,
spreadsheet metadata and bounded cell readback, exact collision search, and
update/move. If they are unavailable, the
Skill SHALL stop before local processing, report `capability_gap`, and offer
connector enablement or the manual XLSX fallback. It SHALL not use network
workarounds, `clasp`, or a static claim that a connector is installed.

For a source Spreadsheet, the Skill SHALL: obtain Drive metadata and verify the
native Google Sheets MIME type; export through `google_drive_export_file` as
XLSX; use only the absolute `workspace_path` returned by that operation; pass
that local XLSX through inspect, build, validate, and run; and remove the exact
local intermediate in a finally-equivalent cleanup path. It SHALL never delete
the source Drive file. An export exceeding the connector's 10 MB limit is a
capability gap with the same manual fallback.

For a destination Spreadsheet, the local Writer Engine SHALL first have
published a validated XLSX using the resolved display and style configuration.
Before import, the Skill SHALL resolve the exact title and call
`google_drive_search` with an exact native-Sheets MIME, title, parent-folder,
and `trashed=false` filter, following every returned page token. `block` SHALL
stop when any exact collision exists. `alternate_name` SHALL choose the first
available deterministic suffix ` (2)`, ` (3)`, and so on, rechecking the exact
candidate before import. The Skill SHALL then import it with
`google_drive_import_spreadsheet` and
`upload_mode:"native_google_sheets"`; verify conversion, native MIME type,
Spreadsheet ID, tabs, and metadata; move it with `google_drive_update_file`
only after parent readback; perform structural and bounded display readback; and
then remove only the local XLSX. Remote collision policy is limited to `block`
and `alternate_name`; remote overwrite is not V1 behavior. The Skill SHALL
derive the title from `google_sheets.title_template` (default
`{output_stem}`), use `google_sheets.destination_folder` when configured, and
honor the validated `google_sheets.collision` policy.

#### Scenario: Missing Drive connector
- **WHEN** a source or destination is a Google Spreadsheet and a required Drive operation is not callable
- **THEN** the Skill stops before any transformation, reports the missing capability, and offers manual XLSX export or verified XLSX upload as applicable

#### Scenario: Google destination succeeds
- **WHEN** the local XLSX run, native import, parent move, and readbacks succeed
- **THEN** the Skill reports the verified Spreadsheet link and receipt, deletes only its local XLSX artifact, and leaves the Drive Spreadsheet intact

#### Scenario: Remote title collision
- **WHEN** an exact native Google Sheet with the resolved title already exists in the destination folder
- **THEN** `block` stops before import, while `alternate_name` searches deterministically for the first available suffixed title before importing

### Requirement: Recovery Skill contract
`data-toolkit-recover` SHALL accept the original request, one `APIError`, and an
attempt count. It SHALL run only when `attempt_count=0`, `retryable=true`, and
the correction is mechanical: remove an unknown optional field, add an explicit
registered version, normalize a syntactic request representation, or select the
single unambiguous option already present in error details.

It SHALL ask the user rather than retry for sheet, delimiter, root path, column,
collision, dirty-input policy, or any result-changing choice. It SHALL never run
after a second failure.

#### Scenario: One correction limit
- **WHEN** the corrected request fails
- **THEN** recovery stops and reports the second structured error without another mutation or run

### Requirement: Skill input and output fixture schemas
Every golden Skill fixture SHALL contain `skill`, `input`, `runtime_responses`,
`expected_commands`, `expected_output_kind`, and `expected_output`. Fixtures
SHALL cover every handoff, stop condition, and forbidden side effect. Runtime
responses are recorded fixtures, not a static capability catalog in Skill text.
Google bridge fixtures SHALL additionally cover a missing connector, successful
export and import, destination folder, collision and title/style policy, upload
failure, local cleanup, and the prohibition on Drive-file deletion.

#### Scenario: Handoff regression
- **WHEN** a Skill change alters a command, handoff payload, output kind, or stop condition
- **THEN** the golden fixture fails before the Codex evaluation suite is run
