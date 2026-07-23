# MVP 2 representative fixture

This anonymized immutable fixture was derived from the approved source contract:

- Source: `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\ext_members_from_12072026_to_14072026_no_full_phone_2026-07-14-12-35-21.json`
- Observed on 2026-07-19: 13,559,851 bytes, 4,665 rows under `/Data`.
- SHA-256: `7BB9B97F201F0882140A8C7C9EAEE458B11237A0A859CED71C75565379F0F53B`.
- Required source fields: `FirstName`, `LastName`, `EntityID`, `TZ`, `PnimCity`, and `Email`.
- Required output order: `שם פרטי`, `שם מלא`, `entityID`, `TZ`, `ישוב`, `מייל`.
- `שם מלא` is the trimmed concatenation of non-empty `FirstName` and `LastName`.

The fixture uses non-real identities and `.invalid` email addresses. `expected.csv` is the exact semantic output: header order, row order, values, null rendering, Hebrew, and emoji are assertions.
