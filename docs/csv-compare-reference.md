# `csv_compare` comparison reference

## Scope

This document is the evidence record for OpenSpec task 10.1. It captures the
observable comparison behavior of the read-only reference at:

`C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\golang\csv_compare`

It is not an implementation specification for the logical engine. The
reference conflates format reading, interactive input, formula parsing,
pairing, and CSV writing. Later cross-source tasks must use the reusable
semantic contract below together with the versioned workflow contracts; they
must not copy reference format or publication behavior into logical
operations.

## Read-only evidence

The reference directory was read-only when inspected on 2026-07-19. The files
and SHA-256 values observed before and after this inspection are recorded in
`testdata/fixtures/cross-source-comparison-reference/reference-manifest.sha256`.
No reference file or source data was modified.

## Observable reference behavior

### Comparison arguments and formulas

- A run declares one or more numbered comparison arguments. Each argument is
  equality between a selected source column and either a selected comparison
  column (two-file mode) or a static value (static-value mode).
- Before equality, the reference applies `strings.TrimSpace` independently to
  the two values. A missing selected cell is treated as the empty string.
- The formula grammar uses numbered arguments, parentheses, `+` for OR, `*`
  for AND, and unary `~` for NOT. Its precedence is NOT, AND, OR.
- Formula parsing rejects an argument set other than exactly `1..cols`, invalid
  characters, and unbalanced parentheses. The current parser has no separate
  token type for repeated unary NOT; this is an implementation limitation, not
  a reusable semantic rule.

### Two-file mode

- The reference reads one data row from each source together and stops as soon
  as either reader ends. It therefore compares positional pairs only; it does
  not find a row elsewhere in the other source.
- A true formula selects the comparison-side row and comparison-side header.
  A false formula selects neither side. The source-side row is never emitted.
- Source order, unequal lengths, and row placement therefore change its result.
  This positional coupling is evidence about the legacy tool, not a permitted
  Data Toolkit cross-source semantic.

### Static-value mode

- The reference compares each selected source value to the correspondingly
  numbered static value, then emits matching source rows with the source
  header.

## Reusable semantic decisions for the toolkit

The following observations are reusable only after conversion into the existing
logical contracts:

| Reference observation | Toolkit treatment |
| --- | --- |
| Explicit numbered Boolean conditions and grouping | Reuse the versioned expression-tree contract; do not parse the legacy formula language as an engine API. |
| Equality predicates against mapped fields or constants | Reuse as explicit typed comparison nodes after mapping and clean-handoff validation. |
| Positive Boolean composition | `AND`, `OR`, and `XOR` are symmetric cross-source operations; source order must not alter the truth set when references and projection are remapped. |
| Unary complement | Reuse only with explicit normalized `not_scope` (`source_1`, `source_2`, or `both`); omission resolves to `both`. |
| Edge whitespace observed during comparison | Treat as a mapping/cleaning finding and explicit normalization plan, never as an implicit logical-operation trim. |
| Missing selected cell interpreted as empty text | Do not inherit silently. Canonical null and a literal empty string remain distinct; later operation validation decides any explicit null comparison. |

For `not_scope: both`, results must preserve `source_id`; branch projections
must use canonical null for absent source-only fields. Ambiguous combined
projections are rejected instead of inheriting the reference's comparison-side
row-only output.

## Excluded reference behavior

The logical engine must not copy any of the following:

- CSV/XLSX opening, sheet prompts, column-letter parsing, or Excel formula
  evaluation. These belong to adapters and adapter validation.
- Interactive prompts, flags, clipboard handling, date-stamped filenames, UTF-8
  BOM output, or output placement beside the input. These belong to interfaces,
  configuration, adapters, and orchestration.
- Full in-memory accumulation of matched rows, positional lock-step pairing,
  truncation at the shorter source, or comparison-side-only projection. Later
  tasks must use bounded cross-source indexes or partitions and explicit
  projections.
- Implicit whitespace trimming, missing-cell-to-empty coercion, or the legacy
  formula parser as a workflow input language.

## Fixture contract

`testdata/fixtures/cross-source-comparison-reference` contains two fixture
families:

1. `legacy-*` captures the minimal, format-free input evidence for the legacy
   positional two-file and static-value behavior.
2. `semantic-cases.json` defines the non-personal-data truth matrix and target
   assertions required by the OpenSpec cross-source requirements: symmetric
   positive conditions plus distinct `source_1`, `source_2`, and `both` NOT
   scopes.

The target assertions are deliberately not executable behavior of the legacy
tool. They are the required reusable behavior for tasks 10.2 through 10.6.
