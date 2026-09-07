# Cleanup-script replacement fixtures

These fixtures are a sanitized, static behavioral inventory for OpenSpec tasks
1.9 and 1.10. They were derived from the recorded task 1.8 inventory only; the
legacy script is neither imported nor executed by these fixtures.

Each manifest entry has exactly one externally visible behavior. Inputs contain
only synthetic numeric identifiers in the `1001`--`1009` range and generic
field names. They contain no production rows, names, addresses, telephone
numbers, email addresses, or identifiers.

`manifest.json` is the source of truth. `operation_ids` lists only exact V1
logical operation IDs. An entry with `outside_v1` deliberately has no logical
mapping; it records the fixed-catalog or workflow-boundary reason instead of
inventing an operation.

Run the structural fixture check from the repository root:

```powershell
& .\testdata\contracts\cleanup-script-replacement\validate.ps1
```

The check validates the manifest shape, the one-behavior rule, fixture paths,
exact catalog IDs, and the required split of exactly four V1-mapped fixtures
and fourteen `outside_v1` fixtures. The latter remain evidence for a future
OpenSpec change; this change does not create it. The check intentionally does
not execute the legacy script or the unfinished Data Toolkit runtime.
