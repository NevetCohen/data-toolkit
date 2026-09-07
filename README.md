# Data Toolkit V1

Data Toolkit is a deterministic local data-transformation product. The active
OpenSpec contract is `implement-revised-data-toolkit-v1`; it freezes the V1
workflow and safety model before the Go implementation begins.

V1 is organized as Infrastructure, Reader Engine, Logical Engine, Writer
Engine, Orchestrator, CLI, and Codex Adapter. The Reader snapshots and validates
immutable source data; the Logical Engine applies exactly the requested ordered
operations; the Writer stages, reopens, validates, atomically publishes, and
finalizes every run with one receipt.

Workflow commands use strict JSON template files rather than a long command
line. See [workflow JSON instructions](docs/codex/workflow-json-format.md) for
variables, config patches, PowerShell quoting, validate/run examples, and stop
conditions.

Google Sheets is supported through a Codex Skill that bridges a verified local
XLSX with the available Google Drive connector. It is not a native Go format or
network integration.

## Verify the contract

```powershell
openspec validate implement-revised-data-toolkit-v1 --strict
```

The legacy MVP remains isolated under `legacy/mvp`. Current Go foundation code
is not the complete product implementation described by the active contract.

See [CHANGELOG.md](CHANGELOG.md) for the `v0.2.0` scope, verification results,
and known limits.
