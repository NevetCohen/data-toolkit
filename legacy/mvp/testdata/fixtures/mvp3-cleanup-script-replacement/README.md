# MVP 3 cleanup-script replacement fixtures

Synthetic immutable reference inputs and exact external-behaviour assertions for task 12.2. They are derived from `docs/mvp3-cleanup-script-behavior.md` and the approved read-only script `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py` (SHA-256 `2d13534e0c58df08db9ae9cff94c08b2701d21fd48887180762b2343e7636dfe`). No production record or comparison output is copied.

`normal-input.json` and the ID seed start with a UTF-8 BOM. The normal case contains Hebrew, emoji, Unicode trimming, two discovered `Data` arrays, stable-order acceptance, filter-before-ID-validation, duplicate/missing/lowercase/numeric/leading-zero IDs, aliases, unknown CSV fields, whitespace around `#`, all address shapes used by the case, and all-zero primary-house suppression. The focused test makes a disposable no-final-LF CSV copy to assert that edge without mutating a fixture.

Assertions cover normal, IDs-only, dry-run, zero-accepted/default-header, exact-`#` maximum-index, and error cases. Source/output collision is deliberately not run because the reference is unsafe; task 12.6 must record Data Toolkit rejection as an intentional gap.
