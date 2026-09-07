# Contract freeze baseline

This record is the pre-implementation evidence for OpenSpec change
`implement-revised-data-toolkit-v1`. It covers Section 1 tasks only and does
not claim that the foundation implements V1.

## Environment and worktree

- Repository: `C:\Users\nevet\data_toolkit`
- Branch: `main`
- Go: `go1.26.4 windows/amd64`
- OpenSpec change: `implement-revised-data-toolkit-v1` (`spec-driven`, `ready`)
- OpenSpec task progress before this work: `0/319`
- Pre-existing worktree edits were preserved.

## Task 1.1 and 1.2 baseline commands

| Command | Started | Ended | Exit | Result |
| --- | --- | --- | ---: | --- |
| `openspec validate implement-revised-data-toolkit-v1 --strict` | `2026-08-11T14:24:20.2374103+03:00` | `2026-08-11T14:24:22.3136457+03:00` | 0 | `Change 'implement-revised-data-toolkit-v1' is valid` |
| `go test ./...` | `2026-08-11T14:24:27.8777105+03:00` | `2026-08-11T14:25:32.1784953+03:00` | 0 | All packages passed; several had no test files. |
| `go vet ./...` | `2026-08-11T14:25:39.5691622+03:00` | `2026-08-11T14:25:45.5328208+03:00` | 0 | No output. |
| `go build ./cmd/data-toolkit` | `2026-08-11T14:25:52.0095823+03:00` | `2026-08-11T14:25:54.0440133+03:00` | 0 | No output. |

The test run passed, including `internal/application`, `internal/architecture`,
`internal/config`, `internal/contract`, `internal/core/datatype`,
`internal/core/table`, `internal/interfaces/cli`, and
`internal/interfaces/tui`.

## Task 1.5 current-reference inventory

| Subject | Classification | Exact current references |
| --- | --- | --- |
| Cell provenance | Active stale runtime dependency | `internal/core/cell/cell.go:11`, `:22`, and `:32`; active test use at `internal/core/table/table_test.go:44`. |
| Retired combined-engine boundary | Active stale package boundary | `cmd/data-toolkit/main.go:11` imports `internal/fileengine/format`; `internal/architecture/dependencies_test.go:20` still names `internal/fileengine`. Literal `combined engine` has no active match. |
| Old CLI request flags | No active runtime match | No current runtime match for `--request`, `--input`, `--output`, or `--workflow`; `internal/interfaces/cli/command.go:51` still exposes the foundation positional `config validate <path>` command. Planned contract references are not runtime references. |
| TUI | Active stale scaffold | `internal/interfaces/tui/client.go:1` and `internal/interfaces/tui/client_test.go:16`; V1 excludes TUI in `PROJECT.md:60`. |
| TXT | Active stale config and scaffold | `configs/default.yaml:34`, `internal/config/types.go:55`, and `internal/fileengine/formats/txt/doc.go:1`; V1 excludes TXT in `PROJECT.md:60`. |
| Native Google format | No active runtime native-format implementation | `PROJECT.md:51` and `docs/architecture.md:64` correctly state that Google Sheets is not a Go format. No active runtime `FormatGoogle` or `google_sheets` match was found. |
| `boolean.not_scope` | Active stale configuration | `configs/default.yaml:48`, `configs/schema/config-v1.schema.json:84`, `internal/config/types.go:72`, and `internal/config/resolve.go:51`. The revised Infrastructure spec requires removal. |
| `interfaces.recent_workflows` | Active stale configuration | `configs/default.yaml:59`, `configs/schema/config-v1.schema.json:105`, `internal/config/types.go:85`, and `internal/config/decode.go:109`. The revised Infrastructure spec requires removal. |

Historical references under `legacy/mvp` remain isolated as evidence and are not
part of this current-reference inventory.

## Task 1.8 cleanup-script behavior inventory

The behavioral source was read statically and was not executed or imported:
`C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py`
with SHA-256
`2D13534E0C58DF08DB9AE9CFF94C08B2701D21FD48887180762B2343E7636DFE`.

| Behavior | Source lines | Externally observable effect |
| --- | --- | --- |
| Text normalization | 19-34 | `None` becomes empty text and all extracted values are trimmed. |
| Address derivation | 37-58 | Street/city/house fallback fields are combined; empty and zero-only house numbers are discarded; repeated city/street text is avoided. |
| Header-driven value rendering | 61-79 | Recognized headers derive index, ID, name, address, and email; both phone columns and unknown headers become empty. |
| Exact filter | 82-87 | No filter accepts all records; incomplete arguments fail; supplied field/value is compared after normalization. |
| Input and parent validation | 90-99, 273-280 | Missing JSON input, missing output parent, or invalid CSV mode fails before processing. |
| Existing ID normalization | 102-123 | Blank ID lines are ignored; trailing commas are removed; nonnumeric values are discarded; duplicate existing IDs collapse to their first occurrence. |
| CSV header/index baseline | 126-162 | Existing first header is reused, otherwise a default header is emitted; the next `#` uses the maximum valid existing index or row count fallback. |
| Streaming Data-array parsing | 165-229 | Textual top-level `Data` arrays are scanned in 4 MiB chunks; array records are yielded in encounter order; malformed/truncated arrays fail; no array fails after processing. |
| Write staging and dry run | 232-261, 294-309 | Non-dry runs stage CSV and ID additions in local temporary files; dry runs write no artifacts. |
| Record admission and JSON deduplication | 311-349 | Non-object values, missing/invalid IDs, filter misses, and duplicate JSON IDs are counted and skipped; accepted IDs retain first encounter order. |
| Publication behavior | 351-373 | A missing CSV is initialized with its header; accepted CSV rows append to the existing file; the ID file is atomically replaced with only accepted IDs; `ids_only` omits CSV output. |
| Cleanup | 375-381 | Open handles close and remaining temporary files are removed on terminal paths reached by `finally`. |
| Report and CLI behavior | 383-440 | The report exposes counters, paths, filter settings, and mode; the CLI accepts long and PowerShell-style aliases and rejects invalid option combinations before calling the function. |

The inventory intentionally records behavior only. Mapping to the fixed V1
catalog is recorded separately with the sanitized fixtures.

## Traceability and readiness records

Task 1.3 has a complete 193-row base table in
`docs/contract-freeze-traceability.md`: 90 requirements and 103 scenarios, with
one unique source anchor per row across the nine capability specifications.

Task 1.4 maps every one of those 193 rows to at least one implementation task
and one focused test or verification task in
`docs/contract-freeze-traceability-mapping.md`. The mapping keeps the two task
roles distinct whenever the plan provides a distinct focused test.

Task 1.6 found no unresolved contract-decision task. The only literal or
decision-like phrases implement already fixed contracts, including
`select_root_path` in task 13.2 and the fixed collision policies in task 16.17.
The mapping record contains the complete scan and the shared Reader/Writer
run-budget implementation coverage.

Task 1.7 is already fulfilled normatively by `design.md:184-185`: an
implementer must stop for an OpenSpec correction rather than invent a missing
field, operation, error, or Skill contract. No duplicate stop rule was added.

## Task 1.9 and 1.10 cleanup-fixture result

### OS-readiness-v1.1 change-scoped precedence

The later approved OS-readiness-v1.1 decision supersedes this change's use of
the revised source brief line 32 full-cleanup-replacement wording. Its exact
scope is four mapped coverage behaviors through the fixed ten-operation V1
catalog and fourteen `outside_v1` behaviors; it makes no replacement or
equivalence claim. The source brief remains unchanged because it is outside
this run's authorized write scope.

The privacy-safe fixture suite is
`testdata/contracts/cleanup-script-replacement/`. Its structural validation
reports 18 unique single-behavior fixtures and accepts only the ten fixed V1
logical operation IDs.

Exactly four fixture behaviors map faithfully to `row.deduplicate`, `text.trim`,
`row.filter`, `table.project`, or `table.rename` as V1 coverage evidence.
Fourteen fixtures are explicitly
`outside_v1`: they require null-to-empty coercion, reader parse or fixed-buffer
behavior, derived/composed or constant columns, generated row numbers, a second
source or output rewrite/append, optional omitted output, cleanup-specific CLI
and report contracts, or temporary-file policy. These behaviors cannot be
represented by the fixed V1 logical catalog and the single-source,
single-new-output workflow. They are recorded for a future OpenSpec change that
this change MUST NOT create; V1 makes no full replacement or equivalence claim,
and no V1 operation was invented.
