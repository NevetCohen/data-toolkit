# Cross-source comparison reference fixtures

These immutable, non-personal-data fixtures support OpenSpec task 10.1 and
later cross-source operation tests.

- `legacy-source.csv`, `legacy-compare.csv`, and `legacy-static-values.json`
  are minimal inputs that expose the read-only reference's positional equality,
  external whitespace trim, Boolean formula, and emitted-side behavior.
- `semantic-cases.json` is the target semantic truth matrix. It is independent
  of CSV/XLSX layout and intentionally requires source-order symmetry and
  scoped NOT complements that the legacy reference itself does not provide.
- `reference-manifest.sha256` records the reference code/module hash evidence
  and the source directory attributes observed during this task. It is not a
  copied reference artifact.

Tests that consume any file in this directory must snapshot it before running
and assert that it remains unchanged afterward.
