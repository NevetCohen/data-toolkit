## ADDED Requirements

### Requirement: MVP fixtures are immutable and reproducible
The implementation SHALL derive read-only representative fixtures from the approved sources, record expected outputs and assertions, and leave every original source unchanged.

#### Scenario: Acceptance run does not change source
- **WHEN** any MVP acceptance workflow runs
- **THEN** source identity and content checks confirm that no source file was modified

### Requirement: Excel city-filter workflow
The toolkit SHALL create a new Excel output containing rows whose `ישוב` value equals `בת ים` from `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\all_final.xlsx`.

#### Scenario: Filter Bat Yam rows
- **WHEN** the approved city-filter workflow runs against its fixture
- **THEN** every output row satisfies `ישוב == בת ים`, expected columns and stable order are preserved, output styling is valid, and the source is unchanged

### Requirement: JSON field-extraction workflow
The toolkit SHALL create a new CSV from `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\ext_members_from_12072026_to_14072026_no_full_phone_2026-07-14-12-35-21.json` containing exactly `שם פרטי`, `שם מלא`, `entityID`, `TZ`, `ישוב`, and `מייל` in the approved order.

#### Scenario: Extract six exact fields
- **WHEN** the approved JSON extraction workflow runs against its fixture
- **THEN** the CSV contains exactly the six declared headers, the expected stable row count and values, valid UTF-8, and no source modification

### Requirement: Cleanup-script replacement workflow
The implementation SHALL treat `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py` as the behavioral source of truth, derive explicit fixtures and equivalence assertions, and express the behavior through reusable toolkit operations.

#### Scenario: Replacement matches behavioral fixtures
- **WHEN** the replacement workflow and the existing script process the same approved fixture
- **THEN** all declared output values, row inclusion/exclusion decisions, normalization effects, order, and exception behavior satisfy the recorded equivalence assertions

### Requirement: Complex comparison reference coverage
The logical-operation tests SHALL inspect the read-only `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\golang\csv_compare` reference and SHALL include fixtures for positive symmetric conditions and scoped NOT conditions.

#### Scenario: Symmetric and scoped conditions are covered
- **WHEN** the logical-operation acceptance suite runs
- **THEN** it verifies source-order independence for `AND`, `OR`, and `XOR` and distinct expected results for `source_1`, `source_2`, and `both` NOT scopes

### Requirement: Baseline performance gates
The first verified implementation SHALL record runtime and peak-memory baselines for all three MVP fixtures on the local machine and SHALL establish regression thresholds that honor the configured memory budget.

#### Scenario: Baseline is recorded
- **WHEN** an MVP workflow first passes functional acceptance on the target machine
- **THEN** its fixture identity, runtime, peak memory, row count, output hashes or equivalent assertions, and selected threshold are recorded

