## Why

The active V1 contract still combines source reading and target writing in one
combined adapter boundary, accepts inline workflow requests, and leaves the lifecycle,
configuration patches, sort, and Google Sheets handling incomplete. These
gaps make ownership, deterministic planning, and Codex operation ambiguous
before implementation begins.

## What Changes

- Define seven V1 subsystems: Infrastructure, Reader Engine, Logical Engine,
  Writer Engine, Orchestrator, CLI, and Codex Adapter.
- Replace the active combined capability with separate Reader and Writer
  registries, operation owners, format lists, and lifecycle contracts.
- Make workflow commands load a strict `WorkflowTemplateDocument` JSON file;
  add typed variables, `--var`, validated `--config-set`, protected config paths,
  documented precedence, and deterministic digests.
- Keep `Schema + RowStream + Value` public, make all deduplication modes
  bounded-spool capable, add stable deterministic `row.sort@v1`, and set the
  V1 logical catalog to ten operations.
- Add a sixth Codex Skill which bridges Google Sheets through local XLSX and
  checked Google Drive connector calls; it does not make Google Sheets a Go
  format or add runtime credentials.
- Add normative memory reservation, spill, finalization, receipt, fixture, and
  wide-benchmark requirements.
- Constrain cleanup-script-related V1 acceptance to coverage evidence for the
  four fixture behaviors expressible through the fixed ten-operation catalog;
  record the other fourteen as `outside_v1` for a future change that this
  proposal MUST NOT create. V1 makes no full replacement or equivalence claim
  for `Invoke-RawMemberCleanup.Streaming.py`.

### OS-readiness-v1.1 scoped precedence

For this change, the later approved OS-readiness-v1.1 decision supersedes the
revised source brief's line 32 full-cleanup-replacement wording. The contract
covers exactly four mapped behaviors through the fixed ten-operation V1 catalog
and records fourteen behaviors as `outside_v1`; it makes no replacement or
equivalence claim. The source brief remains unchanged because it is outside
this run's authorized write scope.

## Impact

Affected capability specs are infrastructure contracts, application API, CLI
JSON surface, transform workflow, logical operations, Reader Engine, Writer
Engine, Codex Adapter, and V1 acceptance. The project brief, architecture
documentation, `PROJECT.md`, implementation tasks, and a new Codex workflow
JSON instruction are aligned with the contract. This change is documentation
and specification work only: no Go source or source data is modified.
