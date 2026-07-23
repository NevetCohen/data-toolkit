# Data Toolkit architecture

## Sources of truth

The approved product brief is the English `Final Specification` in:

`C:\Users\nevet\האחסון שלי\1 - אישי\Vaults\Programming_Vault\Programming\ארגז כלי דאטה\אפיון ראשוני.md`

Product context, repository instructions, OpenSpec artifacts, implementation
code, tests, fixtures, schemas, and runtime documentation now share one
repository:

`C:\Users\nevet\data_toolkit`

The active OpenSpec change is:

`C:\Users\nevet\data_toolkit\openspec\changes\build-data-toolkit-mvp`

The product brief and approved source datasets are read-only. Runtime and
acceptance workflows create new staged outputs and publish only validated
results.

## Target architecture

The target runtime is one modular Go executable with four internal parts and
three interfaces:

1. The logical engine (`operation`) owns deterministic, format-independent data
   transformations over canonical tables and row streams.
2. The file engine (`adapter`) owns JSON, CSV, Excel, TXT, and Google
   Sheets/Drive conversion, mapping, immutable snapshots, output staging,
   validation, and native output styling.
3. The workflow orchestrator (`orchestrator`) validates requests and policies,
   obtains mapping reports, builds ordered plans, selects engines, controls
   resources, verifies results, and reports warnings and exceptions.
4. The application boundary (`application`) is the composition root and shared
   application service. It wires built-in policy, adapters, operations, and the
   local pipeline without introducing a new data engine.

The three target interfaces are the Cobra terminal interface, the Bubble Tea
interactive TUI, and Codex integration. All three submit the same versioned
workflow contract and must not implement data logic independently.

## As-built capabilities

The following table distinguishes implemented behavior from target contracts.

| Area | Implemented now | Still partial or placeholder |
| --- | --- | --- |
| Logical engine | Canonical operation registry; streaming rename with collision validation; streaming Unicode-safe concatenation with null and targeted trim behavior; output projection during adapter writing | Boolean filtering, normalization, regex/text replacement, deduplication, deletion, sorting, splitting, and cross-source Boolean operations |
| File engine | Immutable local snapshots; staged publication and collision policies; JSON inspection/mapping/stream reading; CSV inspection/mapping/read/write/validation; Excel inspection, sheet selection, basic mapping, and row reading; external Google OAuth/client construction | JSON writing; Excel title/subheader/formula-warning/write/style/validation behavior; TXT adapter; Google discovery/read/write/style/validation |
| Orchestrator | Contract/config/registry validation; basic multi-source mapping; output collision and immutable-source checks; ordered plans; started/final results; JSON/CSV local vertical slice; bounded spool primitives | Full cleaning handoff, extended mapping policies, all logical operations, cross-source execution, complete format coverage, and all MVP workflows |
| Application service | `application.NewLocalService`, workflow-only validation, local execution, built-in registration of JSON/CSV/Excel and the operation registry | Google credential injection, complete adapter set, recent-workflow persistence, and approval-capable flows |
| Terminal interface | One executable at `cmd/data-toolkit`; `workflow validate`; `workflow run`; YAML/JSON loading; workflow-relative path resolution; optional user config; JSON result output and exit codes 0/1/2 | Config-validation, mapping, recent-workflow, and other commands required by full OpenSpec task 14.1 |
| Interactive TUI | Package boundary and pinned Bubble Tea dependencies | User flow and runtime implementation |
| Codex interface | Repository instructions and local OpenSpec/domain skills | Executable discovery documentation, schema/operation help, and interface-equivalence tests |

`docs/architecture.md` describes the target first, but the table above is the
authoritative as-built qualification until the remaining OpenSpec tasks pass.

## Package boundaries

- `contract` owns versioned request and result contracts.
- `config` resolves schema-validated policy without credentials.
- `table` owns canonical values, rows, tables, streams, and provenance.
- `operation` owns deterministic data logic and does not import adapters or the
  orchestrator.
- `adapter` maps external formats to and from canonical data.
- `orchestrator` plans and coordinates engines but contains no row-value
  transformations.
- `report` owns report ordering and diagnostic helpers.
- `application` composes the runtime and is imported only by interface layers.
- `cli` and `tui` are interface layers over the application service.

`internal/architecture/dependencies_test.go` enforces forbidden outward
dependencies, including the rule that inner layers do not import
`application`. Interface packages do not implement independent data logic.

## Google adapter boundary

Google Sheets and Drive support uses the Sheets API v4 and Drive API with
external, least-privilege OAuth credentials. `clasp` is not required, installed,
or invoked by Data Toolkit. Apps Script authoring, deployment, and management
remain outside the adapter contract.

## Test-data and repository boundaries

Immutable inputs live below `testdata/fixtures`. Runtime test outputs use
isolated directories below `testdata/generated`, are ignored by Git, and are
removed after each test. Acceptance tests snapshot each source with
`internal/testsupport` and assert that its size and SHA-256 digest remain
unchanged.

`.agents/skills` contains machine-local junctions and is ignored by Git.
`scripts/setup-agent-skills.ps1` recreates those junctions idempotently; the
actual OpenSpec skill sources under `.codex/skills` are versioned in this
repository.

## Local commands

Build the executable:

```powershell
go build ./cmd/data-toolkit
```

Validate or run a workflow:

```powershell
go run ./cmd/data-toolkit workflow validate .\workflow.yaml
go run ./cmd/data-toolkit workflow run .\workflow.yaml --config .\config.yaml
```

Run `scripts/dev.ps1` with one of `format`, `vet`, `unit`, `race`, `build`, or
`all`. These commands do not require live Google credentials.
