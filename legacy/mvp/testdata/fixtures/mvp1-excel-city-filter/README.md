# MVP 1 Excel city-filter fixture

This immutable, synthetic fixture is the contract for filtering `ישוב == בת ים`.
It was derived on 2026-07-19 from the approved read-only workbook:

- Source: `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\all_final.xlsx`
- Source size: 7,065,978 bytes.
- Source SHA-256: `F7D3861B458E548127ED8CCEA4C69B70CCDB5EEC0B464BC9F70F6D9B8E5C399C`.
- Observed workbook contract: `all_final.xlsx`, sheet `all_final`, seven headers, and `ישוב` at physical column 6.

`all_final.xlsx` contains the observed workbook/sheet/header structure, header
styling, frozen header row, and an autofilter. All data values are deterministic
synthetic tokens; no source row, identity, address, email, or phone value was
copied. Its five rows contain three matching `בת ים` rows in source-order-like
positions 3, 4, and 6 (Excel row numbers), separated by two non-matches.

`expected-assertions.json` is the exact oracle for later MVP 1 work: source and
fixture hashes, sheet and header identity, city-column position, input rows,
matching row positions, and the complete expected synthetic values. Consumers
must snapshot the fixture before running and assert it remains unchanged.
