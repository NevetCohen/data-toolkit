# Data Toolkit V1 architecture

## Status

`implement-revised-data-toolkit-v1` is the active implementation-ready OpenSpec
contract. It updates specifications and documentation only; no Go product
implementation is claimed by this document.

## Subsystems and data flow

```text
CLI / Codex Adapter
        |
   Application API
        |
   Orchestrator
        |
Reader Engine -> Logical Engine -> Writer Engine
        \          |             /
          Infrastructure
```

Infrastructure provides immutable `Schema + RowStream + Value`, identities,
registries, configuration, resource reservations, reports, and receipts.

Reader inspection occurs before workflow execution. During a run, the
Orchestrator validates the request, resolves configuration, creates a
deterministic plan, then invokes `reader.snapshot` and `reader.read`. Reader
holds a shared-read lock, creates an immutable snapshot and hash, decodes to the
public data model, and performs schema and dirty-input checks before values reach
the Logical Engine.

Logical Engine executes registered user operations exactly in requested order.
V1 includes ten logical operations, including keyed deduplication and stable
multi-key sort. Work is governed by reservations and bounded spool/spill;
unrecoverable pressure is `resource_limit_exceeded` and cannot publish output.

Writer renders to a staging artifact, applies optional Excel style, closes and
reopens it, runs mandatory format/schema checks and requested assertions, then
atomically publishes under collision policy. `writer.finalize` runs after
workspace creation on success, failure, or cancellation. It releases resources
in reverse order, applies artifact retention/cleanup policy, creates the final
report, and appends exactly one receipt. Finalized failures return that complete
report through the Application API; only pre-workspace failures use a bare
error envelope.

## Workflow files and configuration

The CLI accepts `WorkflowTemplateDocument` files. Typed variables use exact
value references such as `{"$var": 1}` and are substituted before strict
workflow validation and digest calculation. The CLI also accepts repeatable
JSON-pointer configuration patches. Precedence is defaults, user config, CLI
patches, workflow overrides, output overrides; protected paths cannot be
changed through CLI patches. See [workflow JSON instructions](codex/workflow-json-format.md).

## Codex and Google Sheets

Codex owns procedural adaptation, not runtime catalogs. It discovers capabilities
and descriptors from the executable. The sixth project Skill handles Google
Sheets through local XLSX only: it verifies that required Google Drive connector
tools are callable, exports a source Sheet or imports a verified target XLSX,
performs exact paginated title-collision search and bounded readbacks, cleans
local intermediates, and never deletes a Drive file.
Google Sheets is not a Reader or Writer `FormatID` and Go has no connector
credentials or network integration.
