# Changelog

## v0.2.0 - 2026-09-07

This is a pre-1.0 repository release. `v0.2.0` names the repository release,
not the `v1` workflow schema. It establishes the revised V1 contract and the
canonical data layer. It is not a claim that Data Toolkit V1 is ready for
end-user workflows.

### Included

- Replaced the former foundation change with the active OpenSpec change
  `implement-revised-data-toolkit-v1`. It contains 18 task chapters and 319
  implementation tasks.
- Added the contract-freeze record and 193-row traceability material for the
  nine capability specifications.
- Added sanitized cleanup-fixture coverage. Exactly four behaviors map to the
  fixed V1 logical catalog. Fourteen behaviors are recorded as `outside_v1`;
  this release makes no cleanup-script replacement or equivalence claim.
- Implemented and tested the completed canonical data work: validated
  identities and enums, immutable values, canonical types, cells, columns,
  schemas, rows, streams, descriptors, locators, findings, validation results,
  exception summaries, and published-output records.
- Documented the strict workflow JSON procedure and added the Google Sheets
  XLSX bridge Skill. The bridge is agent-side only and does not add a Go Google
  Sheets format or runtime network integration.
- Rewrote the project, README, and architecture documentation around the
  revised seven-subsystem V1 contract.

### Verified for this release

```powershell
openspec validate implement-revised-data-toolkit-v1 --strict
.\testdata\contracts\cleanup-script-replacement\validate.ps1
go test ./...
go vet ./...
go build ./cmd/data-toolkit
.\scripts\setup-agent-skills.ps1
```

All commands passed with Go `1.26.4` on Windows. The fixture validator reported
18 unique fixtures, with exactly four V1 mappings and 14 `outside_v1` entries.

### Not included

OpenSpec records 38 completed tasks and 281 open tasks. In particular, the
complete run report, configuration, runtime API, Reader, Logical, Writer, CLI,
and acceptance workflows remain to be implemented and validated. Do not use
this release as a completed V1 product.

### Repository hygiene

`.codex_artifacts/` contains generated planning and rendered-document output.
It is ignored and is not part of the release.
