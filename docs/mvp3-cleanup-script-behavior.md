# MVP 3 cleanup-script behavior contract

## Purpose and scope

This document records the externally observable behavior of the read-only reference script used by OpenSpec task 12.1:

`C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py`

The inspected script was 14,567 bytes, last modified at `2026-07-06T15:12:02.3170331+03:00`, with SHA-256:

`2d13534e0c58df08db9ae9cff94c08b2701d21fd48887180762b2343e7636dfe`

The Python script is the behavior source of truth for MVP 3. This document describes what it does, including accidental and unsafe behavior. It does not decide which behavior the toolkit should reproduce. Task 12.3 maps the behavior to reusable operations, and task 12.6 must record intentional incompatibilities as acceptance gaps.

No source or comparison file was modified during this inspection. Direct evidence was gathered by read-only record comparison and by isolated probes in a temporary directory.

## Command-line contract

The script accepts these options:

| Option | Alias | Required | Meaning |
| --- | --- | --- | --- |
| `--json-path` | `-JsonPath` | Yes | Input file searched for `Data` arrays. |
| `--entity-ids-path` | `-EntityIdsPath` | Yes | ID-list output path. Existing contents are inspected but ultimately replaced on a successful non-dry run. |
| `--csv-path` | `-CsvPath` | Unless `--ids-only` | CSV to create or append. |
| `--dry-run` | `-DryRun` | No | Execute parsing and accounting without writing output files. |
| `--filter-field` | `-FilterField` | Paired | Exact top-level record field to compare. |
| `--filter-value` | `-FilterValue` | Paired | Exact comparison value after record-side text normalization. |
| `--ids-only` | `-IdsOnly` | No | Write only the ID list; do not read or write CSV content. |

The CLI rejects a missing CSV path outside ID-only mode. It also rejects a filter field without a truthy filter value, or a truthy filter value without a field. Argument errors are `argparse` usage failures. Processing exceptions are not converted to a structured error result; they escape and produce a traceback and nonzero process exit.

If `--ids-only` and `--csv-path` are both provided, the CSV parent is checked and the path appears in the report, but no CSV is created or read.

## Processing order

For each run, observable behavior follows this order:

1. Require the JSON path to be an existing file and resolve that input path.
2. Require output parent directories to exist. The script never creates them.
3. Require a CSV path unless ID-only mode is active.
4. Read and classify the existing ID-list file, if present.
5. Outside ID-only mode, read the CSV header and derive the next `#` index from existing rows.
6. Outside dry-run mode, create temporary row and ID files in their respective output directories.
7. Stream every located `Data` array in discovery order and every array element in source order.
8. For each element, apply record-shape checking, filtering, ID validation, and within-run ID deduplication in that exact order.
9. Derive CSV fields from the existing or default header and write one ID line for each accepted record.
10. After the entire input parses successfully, create a missing/default CSV header if necessary, append staged rows to the CSV, and replace the ID-list path with the staged current-run list.
11. Remove ordinary temporary files in `finally` and return the text summary.

Filtering therefore precedes ID validation and deduplication. A nonmatching dictionary record is counted as filtered even if its ID would have been invalid. A non-dictionary array element is counted as missing/invalid ID without applying the filter. Duplicates are counted only among valid IDs that passed the filter.

## Input discovery and JSON parsing

### `Data` discovery

The reader does not validate a conventional top-level JSON envelope. It searches text for the case-sensitive regular expression:

`"Data"\s*:\s*\[`

Consequences:

- A `Data` array can be nested; the script still treats it as an input document.
- Prefix text, envelope members, and trailing text outside located arrays are not validated.
- Multiple `Data` arrays, including arrays in concatenated JSON documents, are processed in textual discovery order.
- A top-level array, a differently cased key, or `Data` whose value is not an array is not recognized.
- Before a match is found, only the last 256 characters of a growing unmatched prefix are retained.
- The nominal read chunk is 4 MiB. A single large record can grow the buffer beyond one chunk.

The report's `JsonDocuments` value is the highest discovered array number observed through a yielded element, not a reliable count of all located arrays. A trailing empty array is not observed by `run_cleanup`. An input containing only `{"Data":[]}` raises `ValueError: No top-level Data arrays found ...` and produces no outputs.

### Element decoding

Each array element is decoded with Python's default `json.JSONDecoder`:

- Object key order is immaterial to the transformations.
- JSON integers become Python integers; JSON floating-point numbers become binary floats. Raw numeric lexemes are not preserved.
- Strings, booleans, arrays, and objects placed in output fields are converted later with Python `str`, not canonical JSON serialization.
- A UTF-8 BOM is accepted because the input uses `utf-8-sig`.
- Invalid UTF-8 raises a decoding error.
- A truncated array between records raises `ValueError: Unexpected end of file inside Data array #N`.
- A truncated or malformed current record re-raises `JSONDecodeError`.
- A scalar or array element is skipped and increments `MissingOrInvalidEntityIds`.

Input order is preserved across all accepted elements and across multiple located arrays.

## Text normalization

Every record field used by the script is normalized independently:

- `null` or a missing field becomes the empty string.
- Every other value becomes `str(value).strip()`.
- Python `strip` removes leading and trailing Unicode whitespace, but not internal whitespace.
- No case folding, Unicode normalization, punctuation normalization, or internal whitespace collapse occurs.
- Non-string values expose Python string formatting. For example, booleans become `True` or `False`, and nested lists/dictionaries use Python representations rather than JSON text.

The direct 92,922-row evidence set contained at least one such edge-trim change in 92,874 accepted records across the fields used by the script.

## Filter semantics

With no field and no value, every dictionary record passes. With a filter:

- The field name is exact and case-sensitive.
- Only a top-level field is addressed.
- The record value is normalized as described above.
- The CLI filter value is compared exactly and is not trimmed by the script.
- Missing and `null` record values normalize to empty text.

The public CLI prevents a nonempty field from being paired with an empty value. Direct calls to `run_cleanup` are less strict because they distinguish `None` from an empty value inside `record_matches_filter`.

## Entity-ID validation and deduplication

The accepted ID source is exactly the record field `EntityID`. Source fields such as `entityID` or `entity_id` are ignored even though similarly named CSV headers are supported.

After text normalization, an ID is accepted only when it fully matches Python regular expression `\d+`:

- Leading and trailing whitespace is removed.
- Leading zeroes are preserved.
- Signs, decimal points, separators, and embedded whitespace are rejected.
- Python regular-expression digit semantics are Unicode-aware; the check is not limited explicitly to ASCII digits.
- Numeric JSON integers may pass after conversion to text. Floating-point values such as `123.0` fail.

Accepted IDs are deduplicated by exact normalized string within the current run. The stable first occurrence is retained and later occurrences increment `DuplicateJsonEntityIds`. The in-memory `json_seen` set grows with the number of accepted unique IDs, so record parsing is streaming but ID deduplication is not memory-bounded independently of input cardinality.

Existing IDs do not participate in input-row deduplication. Re-running against an existing CSV can append the same IDs again.

## CSV schema and field transformations

When a CSV does not exist, is zero bytes, or has an empty/whitespace-only first physical line, the default header is:

| Position | Header |
| --- | --- |
| 1 | `#` |
| 2 | `entityID` |
| 3 | `שם פרטי` |
| 4 | `שם משפחה` |
| 5 | `כתובת` |
| 6 | `מס' טלפון` |
| 7 | `מס' טלפון נייד` |
| 8 | `מייל` |

For an existing CSV with a nonblank first line, that line is parsed as the authoritative header and preserved. The script does not validate that it is a known schema. Each header is trimmed only for output mapping:

| Trimmed output header | Source/transformation |
| --- | --- |
| `#` | Sequential decimal index for the accepted record. |
| `EntityID`, `entityID`, `entity_id`, `entity id`, `מספר מתפקד` | Normalized source `EntityID`. |
| `שם פרטי` | Normalized source `FirstName`. |
| `שם משפחה` | Normalized source `LastName`. |
| `כתובת` | Composed address described below. |
| `מס' טלפון` | Always empty, even when the source contains a phone. |
| `מס' טלפון נייד` | Always empty, even when the source contains a mobile phone. |
| `מייל`, `אימייל`, `Email`, `email` | Normalized source `Email`. |
| Any other header | Always empty. |

The existing header order, duplicate headers, header spelling, and arbitrary unknown columns determine the emitted row shape. Values are written with Python's CSV writer, including ordinary CSV quoting when required. Text output is UTF-8 without a newly added BOM and new generated rows use LF terminators.

### Sequential `#` behavior

For an existing CSV, the first parsed row is always treated as the header. Remaining wholly blank rows are ignored for index accounting.

- If the parsed headers contain an exact `#`, the maximum positive digit-only value in that column becomes `CsvLastIndexBeforeRun`.
- If no positive digit-only value is found, the count of nonblank existing data rows is used.
- If there is no exact `#` header, the nonblank data-row count is used.
- Header whitespace matters for index discovery: ` # ` is recognized by output mapping but not by `headers.index("#")`.
- If at least one positive index exists, row count is ignored even when the maximum is smaller than the number of existing rows. New indices can therefore collide with existing values.
- Negative, signed, and decimal indices are ignored. Some non-ASCII strings accepted by `str.isdigit()` may still fail Python integer conversion and abort the run.

Accepted records receive consecutive indices beginning at the derived value plus one. Skipped, filtered, invalid, and duplicate records consume no index.

### Existing CSV edge cases

- A missing or empty CSV receives the default header even when a nonempty `Data` array yields zero accepted records.
- An existing nonblank-header CSV with zero accepted records remains unchanged.
- A file whose first physical line is blank is treated as needing the default header. On a successful run the script opens it with `w`, truncating any later pre-existing rows before appending accepted rows.
- A data file with no real header has its first data row interpreted as a header.
- Existing row widths and malformed business schemas are not validated.
- If an existing nonempty CSV lacks a final LF and rows are appended, the script first adds `os.linesep` and then copies LF-terminated staged rows. On Windows this can create mixed CRLF/LF endings.
- CSV rows are staged, but the existing CSV is not replaced atomically. It is appended in place after parsing succeeds.

## Address composition

Address components use this precedence:

| Component | First candidate | Fallback candidate |
| --- | --- | --- |
| Street | `PnimStreet` | `AddStreet` |
| City | `PnimCity` | `AddCity` |
| House number | `PnimHouseNum` | `AddressHouseNum` |

The first normalized nonempty candidate wins. For the house number, usefulness is checked only after candidate selection. A chosen house value is discarded when it consists entirely of one or more zeroes. Therefore a primary value such as `000` suppresses the house number and also prevents a valid fallback `AddressHouseNum` from being used. Values such as `0.0` or `000A` are considered useful.

Composition rules are:

1. Join street and useful house with one ASCII space. If there is no street, the useful house stands alone.
2. If the resulting street/house text and city are nonempty, append comma-space plus city, except when normalized street equals normalized city exactly.
3. When street equals city, the city is not repeated; a house still remains after the street.
4. If there is no street/house text, return city.
5. If every component is empty or suppressed, return the empty string.

Comparison is case-sensitive and performs no geographic, punctuation, or Unicode-equivalence normalization.

The direct evidence set exercised these output shapes:

| Address shape | Rows |
| --- | ---: |
| Street + house + city | 69,290 |
| Street + city | 5,661 |
| Street + house, with duplicate city suppressed or absent | 4,502 |
| Street only, with duplicate city suppressed or absent | 8,277 |
| House + city | 1,686 |
| House only | 0 |
| City only | 3,498 |
| Empty | 8 |

Within that set, 17,433 chosen house values were all-zero and suppressed, and 12,779 normalized street/city pairs suppressed duplicate city output.

## ID-list input and output behavior

An existing ID file is read as UTF-8 with optional BOM. For accounting only:

- CR/LF is removed.
- Outer whitespace is removed.
- One trailing comma plus trailing whitespace is removed.
- Blank results, including a comma-only line, are ignored.
- Non-digit results are collected as invalid.
- Valid IDs are deduplicated in first-occurrence order.

`ExistingEntityIdsBeforeRun` is the count of unique valid existing IDs. `InvalidEntityIdLinesRemoved` is the number of nonblank invalid lines. Neither list changes which JSON records are accepted.

On a successful non-dry run, the existing ID file is replaced rather than appended or merged:

- Normal CSV mode writes `EntityID,` followed by LF for each accepted current-run record.
- ID-only mode writes `EntityID` followed by LF, without the comma.
- Existing valid IDs that are absent from the current run are removed.
- Existing duplicates and invalid lines are removed as a side effect of replacement.
- A nonempty `Data` array with no accepted records replaces the ID file with an empty file.
- An empty `Data` array errors before replacement.

The ID file is staged in its destination directory and published with `os.replace`, but no durability sync or post-write validation is performed.

## Dry-run behavior

Dry-run mode performs path checks, reads existing output metadata, parses the complete input, applies all transformations, and computes all counters. It does not create, append, truncate, or replace either output.

The report remains phrased as if the hypothetical current-run IDs were the post-run result: `RowsAppended` and `EntityIdsAfterRun` both equal the number that would be accepted, even though files are unchanged. The script first prints `Dry run: no files were changed.` and then prints every result entry as `Key: value`.

## Result counters

The successful result dictionary and stdout summary preserve this insertion order:

| Field | Observed meaning |
| --- | --- |
| `JsonDocuments` | Highest `Data` array number observed through a yielded element. Empty-only input reports an error instead. |
| `RawRecords` | Number of decoded array elements, including non-dictionaries and later-skipped records. |
| `DuplicateJsonEntityIds` | Valid, filter-passing repeated IDs after their first accepted occurrence. |
| `MissingOrInvalidEntityIds` | Non-dictionary elements plus filter-passing dictionaries with missing or non-digit `EntityID`. |
| `FilteredOutRecords` | Dictionary records whose normalized named field does not equal the exact filter value. |
| `ExistingEntityIdsBeforeRun` | Unique valid IDs read from the pre-run ID file. |
| `InvalidEntityIdLinesRemoved` | Invalid nonblank lines found in the pre-run ID file. |
| `CsvLastIndexBeforeRun` | Derived maximum positive `#`, or nonblank existing data-row count fallback. Zero in ID-only mode. |
| `RowsAppended` | Accepted current-run records. In ID-only or dry-run mode the name is still used. |
| `EntityIdsAfterRun` | Exactly `RowsAppended`, not an observed file count and not existing plus new IDs. |
| `IdsOnly` | Boolean mode value. |
| `FilterField` | Supplied field or empty string. |
| `FilterValue` | Supplied value or empty string. |
| `CsvPath` | Supplied CSV path string, even when ID-only mode does not use it; empty if omitted. |
| `EntityIdsPath` | Supplied ID output path string. |

The resolved JSON input path is not included in the result.

## Failure, publication, and safety edges

The script has no explicit source-immutability or output-identity guard:

- It does not reject a CSV or ID output path that aliases the JSON input.
- It does not reject CSV and ID paths that alias each other.
- Such aliases can append to, truncate, or replace an input or another output after reading completes.
- It does not snapshot the source or verify that it stayed unchanged during the read.

The Data Toolkit must retain its stronger invariant and reject source/output identity collisions; reproducing these unsafe effects is not acceptable equivalence.

Publication is not transactional across both outputs:

- Parsing failures happen before ordinary target writes, and staged temporary files are removed during normal exception unwinding.
- CSV header creation/truncation and row append occur before ID-list replacement.
- A failure during or after CSV mutation but before ID replacement can leave outputs inconsistent.
- A process crash can leave `.raw-member-cleanup-*.tmp` files because `finally` does not run.
- Existing CSV contents are not backed up, validated, or atomically replaced.
- No output row-count, schema, encoding, or content validation runs after publication.

Other externally observable errors include missing input files, missing output parent directories, inaccessible files, CSV decoding/parsing failures, Unicode errors, JSON errors, and operating-system failures during append or replace. They are not normalized into stable domain error codes.

## Direct comparison evidence

The strongest direct comparison pair is under the read-only `all_060726` directory:

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `raw_GetContents 060726.json` | 303,006,347 | `b490efa67fe69a2859013603c1aee4119fd1957704c6a0ee2b2a5717e12e5bf5` |
| `all_060726.backup-before-phones-20260707-072706.csv` | 8,213,079 | `a76e6c25d1babdca0a77ea68e5938ac31bdfeffad9af3cb0d1ba54111dfdce21` |
| `entity_ids.txt` | 743,376 | `d436f7b63dc693cdedf0be8e08de99b6bdcdf0ee2375fe925ac68b85a7c62768` |

A read-only, record-by-record comparison using the inspected script's own transformation functions established:

- One observed `Data` array and 92,922 decoded records.
- 92,922 accepted rows in stable source order.
- Zero non-dictionary records, missing/invalid IDs, or duplicate JSON IDs.
- Exact default header equality.
- Zero CSV value, row-order, row-count, or ID-line mismatches.
- Zero extra CSV rows or ID lines.
- CSV uses UTF-8 without BOM, 92,923 LF terminators including the header, and a final LF.
- ID output uses UTF-8 without BOM, 92,922 LF terminators, comma-terminated IDs, and a final LF.

This pair is a direct full-size oracle for the default, unfiltered, normal CSV mode.

## `all_final` comparison-output boundary

The requirements identify the read-only `all_final` directory as existing comparison output. Its top-level artifacts represent a later multi-step pipeline, not a single direct invocation of this script.

Observed lineage evidence:

- `ext_members_from_12072026_to_14072026_no_full_phone_2026-07-14-12-35-21.json` contains 4,665 records in one observed `Data` array.
- `EntityID.txt` contains exactly the same 4,665 accepted IDs in source order, each comma-terminated, with no missing, invalid, or duplicate source IDs. This is direct evidence for normal-mode ID-list behavior.
- `mobile_phones_2026-07-14-12-40-13.csv` adds phone-fetch status and values. The cleanup script never fetches or populates phones.
- `דלתה 130726-140726.csv` has 4,665 rows and columns `EntityID`, `TZ`, `FirstName`, `LastName`, `Email`, `כתובת`, `PnimCity`, and `mobilePhone`. That schema and its populated phone/TZ fields are not emitted by the cleanup script's default header mapping.
- `all_130726.csv` has 108,317 rows and is byte-identical to the file of the same name under the `all_130726` directory.
- `all_final.csv` has 112,982 rows under another eight-column downstream schema, and `all_final.xlsx` is a further spreadsheet output. Neither is a direct script output oracle.

Accordingly, later fixture work may use `all_final` to establish pipeline provenance and selected ID/order facts, but must not attribute phone enrichment, TZ projection, final merging, deduplication policy, Excel conversion, or final schema naming to this script.

## Required edge coverage for later fixture work

Task 12.2 should derive minimal immutable fixtures that cover at least:

- UTF-8 BOM input, Hebrew/emoji text, Unicode edge trimming, and stable order.
- Default header creation and an arbitrary existing header with aliases, unknown columns, and whitespace around `#`.
- Each address-source fallback and every address output shape.
- All-zero primary house number suppressing a valid fallback house number.
- Missing, lowercase-only, numeric, leading-zero, invalid, and duplicate IDs.
- Existing ID-list blanks, one trailing comma, duplicates, invalid lines, and current-run replacement.
- Existing CSV index maximum, row-count fallback, no final newline, and a blank first line.
- Filter match/mismatch ordering relative to invalid IDs and duplicates.
- Normal, ID-only, and dry-run modes.
- Multiple discovered arrays, non-dictionary elements, empty `Data`, missing `Data`, malformed records, and truncated arrays.
- Zero accepted records from a nonempty array.
- Output/source and output/output alias rejection as a deliberate toolkit safety difference.
- Failure after staging but before complete two-output publication as an equivalence or intentional-gap decision.

No fixture or reusable operation is added by task 12.1.
