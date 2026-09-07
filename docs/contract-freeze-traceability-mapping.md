# Contract-freeze traceability mapping

Scope: OpenSpec change `implement-revised-data-toolkit-v1`. This is the task 1.4 mapping for every one of the 193 `Requirement` and `Scenario` headings in the nine active capability specifications. `file-engine/` is an empty directory and has no `spec.md`.

Anchors are `spec:line`. Task ranges contain exact task IDs. Test tasks are separate from implementation tasks where the plan provides a focused test.

## OS-readiness-v1.1 change-scoped precedence

For this change, the later approved OS-readiness-v1.1 decision supersedes the
revised source brief line 32 full-cleanup-replacement wording. Exactly four
mapped behaviors are V1 coverage through the fixed ten-operation catalog and
fourteen behaviors are `outside_v1`; no replacement or equivalence claim is
made. The source brief remains unchanged because it is outside this run's
authorized write scope.

## application-api

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Common JSON envelope | application-api:3 | 5.1, 5.2 | 5.3 |
| Scenario | Exclusive result branch | application-api:19 | 5.2 | 5.3 |
| Requirement | Exact application API | application-api:23 | 5.14 | 5.15 |
| Scenario | Thin interface client | application-api:40 | 5.14 | 5.15 |
| Requirement | Capability discovery contract | application-api:44 | 5.7, 5.8 | 5.15 |
| Scenario | Runtime-only discovery | application-api:59 | 5.8 | 5.15 |
| Requirement | Operation description contract | application-api:63 | 5.9 | 5.15 |
| Scenario | Ambiguous operation version | application-api:74 | 5.9 | 5.15 |
| Requirement | Configuration validation contract | application-api:78 | 5.10 | 5.15 |
| Scenario | Unknown configuration field | application-api:91 | 5.10 | 5.15 |
| Requirement | Inspection API contract | application-api:95 | 5.11 | 5.15 |
| Scenario | Inspection needs a decision | application-api:108 | 5.11 | 5.15 |
| Requirement | Workflow validation API contract | application-api:112 | 5.12 | 5.16 |
| Scenario | Validation does not read data rows | application-api:131 | 5.12 | 5.16 |
| Requirement | Run API contract | application-api:135 | 5.13, 5.14 | 5.16, 5.17 |
| Scenario | Validate-run drift | application-api:161 | 5.13 | 5.16 |
| Scenario | Finalized failed run | application-api:165 | 5.14, 8.14, 8.17 | 5.17, 8.18 |
| Requirement | Stable API error catalog | application-api:169 | 5.4-5.6 | 5.17 |
| Scenario | Stable programmatic failure | application-api:195 | 5.4-5.6 | 5.17 |
| Requirement | CLI command and exit mapping | application-api:199 | 15.2-15.9 | 15.10 |
| Scenario | JSON stdout on non-zero exit | application-api:225 | 15.8, 15.9 | 15.10 |

## cli-json-surface

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Required local commands | cli-json-surface:3 | 15.2-15.7 | 15.10 |
| Scenario | Command inventory | cli-json-surface:12 | 15.2-15.7 | 15.10 |
| Requirement | Runtime-derived discovery | cli-json-surface:16 | 15.2, 15.3 | 15.12 |
| Scenario | Newly registered operation | cli-json-surface:21 | 15.2, 15.3 | 15.12 |
| Requirement | JSON stdout isolation | cli-json-surface:25 | 15.1, 15.8 | 15.10 |
| Scenario | Validation error | cli-json-surface:30 | 15.8 | 15.10 |
| Requirement | Stable exit status | cli-json-surface:34 | 15.9 | 15.10 |
| Scenario | Unsupported operation exit | cli-json-surface:39 | 15.9 | 15.10 |
| Requirement | Thin interface boundary | cli-json-surface:43 | 15.1 | 15.12 |
| Scenario | API substitution test | cli-json-surface:48 | 15.1 | 15.12 |
| Requirement | Workflow-template loading and CLI patches | cli-json-surface:52 | 15.6, 15.7, 15.14-15.17 | 15.18-15.20 |
| Scenario | Protected configuration patch | cli-json-surface:61 | 15.15, 15.17 | 15.20 |

## codex-adapter

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Runtime-first Codex sequence | codex-adapter:3 | 16.1, 16.3, 16.5, 16.8 | 16.2, 16.4, 16.7, 16.9 |
| Scenario | Supported natural-language task | codex-adapter:10 | 16.1 | 16.2 |
| Requirement | No duplicate capability catalog | codex-adapter:14 | 16.1, 16.3-16.12 | 16.23 |
| Scenario | Runtime catalog changes | codex-adapter:18 | 16.5 | 16.23 |
| Requirement | Clarification instead of guessing | codex-adapter:22 | 16.1, 16.5 | 16.2, 16.7 |
| Scenario | Duplicate headers | codex-adapter:27 | 16.3 | 16.4 |
| Requirement | Explicit unsupported-capability behavior | codex-adapter:31 | 16.1, 16.5 | 16.2, 16.7 |
| Scenario | Cross-source join request | codex-adapter:36 | 16.1 | 16.2 |
| Requirement | Machine-error self-correction | codex-adapter:40 | 16.10 | 16.11 |
| Scenario | Unknown request field | codex-adapter:45 | 16.10 | 16.11 |
| Requirement | Versioned adapter fixtures | codex-adapter:49 | 16.1-16.12 | 16.2, 16.4, 16.7, 16.9, 16.11, 16.15, 16.21 |
| Scenario | Skill regression | codex-adapter:54 | 16.1-16.12 | 16.25, 16.26 |
| Requirement | Exact Codex Skill inventory | codex-adapter:58 | 16.1, 16.3, 16.5, 16.8, 16.10, 16.12 | 16.22 |
| Scenario | Skill family discovery | codex-adapter:72 | 16.1-16.12 | 16.22 |
| Requirement | Root Skill contract | codex-adapter:76 | 16.1 | 16.2 |
| Scenario | Unsupported join | codex-adapter:88 | 16.1 | 16.2 |
| Requirement | Inspection Skill contract | codex-adapter:92 | 16.3 | 16.4 |
| Scenario | Ambiguous worksheet | codex-adapter:99 | 16.3 | 16.4 |
| Requirement | Workflow Builder Skill contract | codex-adapter:103 | 16.5, 16.6 | 16.7 |
| Scenario | Unknown operation parameter | codex-adapter:121 | 16.5 | 16.7 |
| Requirement | Run Skill contract | codex-adapter:125 | 16.8 | 16.9 |
| Scenario | Successful run handoff | codex-adapter:135 | 16.8 | 16.9 |
| Requirement | Google Sheets bridge Skill contract | codex-adapter:139 | 16.12-16.20 | 16.15, 16.21 |
| Scenario | Missing Drive connector | codex-adapter:177 | 16.12, 16.13 | 16.15 |
| Scenario | Google destination succeeds | codex-adapter:181 | 16.16-16.20 | 16.21 |
| Scenario | Remote title collision | codex-adapter:185 | 16.16, 16.17 | 16.21 |
| Requirement | Recovery Skill contract | codex-adapter:189 | 16.10 | 16.11 |
| Scenario | One correction limit | codex-adapter:200 | 16.10 | 16.11 |
| Requirement | Skill input and output fixture schemas | codex-adapter:204 | 16.1-16.12 | 16.24 |
| Scenario | Handoff regression | codex-adapter:213 | 16.1-16.12 | 16.24 |

## infrastructure-contracts

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Canonical streaming table model | infrastructure-contracts:3 | 2.4-2.13 | 2.7, 2.15 |
| Scenario | Distinct null and text values | infrastructure-contracts:9 | 2.6 | 2.7 |
| Scenario | Exact numeric value | infrastructure-contracts:13 | 2.4, 2.5 | 2.7 |
| Scenario | Diagnostic location without cell provenance | infrastructure-contracts:17 | 2.8, 3.5 | 3.13 |
| Requirement | Complete schema metadata | infrastructure-contracts:21 | 2.9, 2.10 | 2.15 |
| Scenario | Duplicate source headers | infrastructure-contracts:26 | 2.9 | 11.12 |
| Requirement | Strict V1 configuration | infrastructure-contracts:30 | 4.1-4.15 | 4.18, 4.19 |
| Scenario | Removed foundation field | infrastructure-contracts:36 | 4.15 | 4.19 |
| Scenario | Registered semantic type reference | infrastructure-contracts:40 | 4.4 | 4.19 |
| Requirement | Immutable configuration precedence | infrastructure-contracts:44 | 4.16, 4.17 | 4.21 |
| Scenario | Output override wins | infrastructure-contracts:51 | 4.16 | 4.21 |
| Requirement | Protected CLI configuration patches | infrastructure-contracts:55 | 4.11, 4.16 | 4.20 |
| Scenario | Protected descendant patch | infrastructure-contracts:68 | 4.11, 4.16 | 4.20 |
| Requirement | Versioned application envelopes | infrastructure-contracts:72 | 5.1, 5.2 | 5.3 |
| Scenario | Request correlation | infrastructure-contracts:77 | 5.1 | 5.3 |
| Requirement | Structured deterministic errors | infrastructure-contracts:81 | 3.7, 5.4-5.6 | 5.17 |
| Scenario | Unknown operation | infrastructure-contracts:86 | 5.5, 5.6 | 5.17 |
| Requirement | Compact append-only run receipt | infrastructure-contracts:90 | 6.1-6.7 | 6.8-6.10 |
| Scenario | Failed run receipt | infrastructure-contracts:99 | 6.1, 6.7 | 6.10 |
| Scenario | Secret redaction | infrastructure-contracts:103 | 3.7, 6.7 | 6.8 |
| Requirement | Primitive identities and enums | infrastructure-contracts:107 | 2.1, 2.2 | 2.3 |
| Scenario | Unknown enum value | infrastructure-contracts:137 | 2.2 | 2.3 |
| Requirement | Exact canonical data structures | infrastructure-contracts:141 | 2.4-2.13 | 2.15 |
| Scenario | Row/schema mismatch | infrastructure-contracts:183 | 2.10-2.12 | 2.15 |
| Requirement | Exact descriptors and reporting structures | infrastructure-contracts:187 | 3.1-3.11 | 3.12, 3.13 |
| Scenario | Receipt compactness | infrastructure-contracts:265 | 6.1, 6.7 | 6.9 |
| Requirement | Exact configuration structures | infrastructure-contracts:284 | 4.1-4.14 | 4.18, 4.22 |
| Scenario | Complete default configuration | infrastructure-contracts:353 | 4.14 | 4.18, 4.22 |
| Requirement | Memory-budget enforcement | infrastructure-contracts:357 | 12.15-12.18, 12.25, 18.14 | 18.15 |
| Scenario | Unique keyed deduplication exceeds memory | infrastructure-contracts:366 | 12.15-12.18, 18.14 | 18.15 |

## logical-operations

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Explicit logical operation registry | logical-operations:3 | 10.5, 10.8, 12.1-12.25 | 10.12, 12.27 |
| Scenario | Duplicate logical operation | logical-operations:8 | 10.5, 10.8 | 10.12 |
| Requirement | Validate before consuming rows | logical-operations:12 | 10.1-10.3 | 10.4 |
| Scenario | Unknown parameter | logical-operations:16 | 10.5, 10.8 | 10.12 |
| Requirement | Projection and rename operations | logical-operations:20 | 10.5, 10.6, 12.1, 12.2 | 10.7 |
| Scenario | Project Hebrew columns | logical-operations:25 | 10.6 | 10.7 |
| Requirement | Constant-only Boolean filter | logical-operations:29 | 10.1-10.3, 10.8-10.10 | 10.4, 10.11 |
| Scenario | Nested filter expression | logical-operations:34 | 10.1-10.3, 10.10 | 10.4 |
| Scenario | Column comparison rejected | logical-operations:38 | 10.2, 10.3 | 10.4 |
| Requirement | Cleaning and normalization catalog | logical-operations:42 | 12.3-12.25 | 12.7, 12.19, 12.26, 12.27 |
| Scenario | Empty-field safety | logical-operations:47 | 12.12, 12.13 | 12.13 |
| Scenario | Stable deduplication | logical-operations:51 | 12.14-12.18 | 12.19 |
| Requirement | Immutable streaming execution | logical-operations:55 | 2.4-2.13, 10.6, 10.10 | 12.28 |
| Scenario | Changed value | logical-operations:60 | 12.6, 12.9, 12.11 | 12.7, 12.9, 12.11 |
| Scenario | Operation chain | logical-operations:64 | 9.16 | 12.28 |
| Requirement | Exact operation descriptor | logical-operations:68 | 3.3, 10.5, 10.8, 12.1-12.25 | 10.12, 12.27 |
| Scenario | Incomplete descriptor | logical-operations:87 | 3.3 | 10.12 |
| Requirement | Exact typed literal and expression structures | logical-operations:91 | 10.1-10.3 | 10.4 |
| Scenario | Invalid tagged expression | logical-operations:112 | 10.2, 10.3 | 10.4 |
| Requirement | Complete V1 logical operation catalog | logical-operations:116 | 1.10, 10.5, 10.8, 12.1-12.29 | 12.27, 17.9-17.12 |
| Scenario | Catalog completeness | logical-operations:133 | 10.5, 10.8, 12.1-12.25 | 12.29 |
| Requirement | Exact logical parameter structures | logical-operations:137 | 10.1-10.3, 10.5, 10.8, 12.1-12.25 | 10.12, 12.27 |
| Scenario | Delete-empty whole-row safety | logical-operations:170 | 12.12, 12.13 | 12.13 |
| Requirement | Operation failure detail catalog | logical-operations:174 | 3.7, 12.7-12.18 | 12.7, 12.26 |
| Scenario | Stable operation failure | logical-operations:184 | 3.7, 12.17 | 12.7 |

## reader-engine

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Explicit Reader Engine registry | reader-engine:3 | 7.2, 7.3 | 7.4 |
| Scenario | Descriptor and implementation mismatch | reader-engine:8 | 7.2, 7.3 | 7.4 |
| Requirement | Deterministic inspection | reader-engine:12 | 7.5-7.10, 11.2-11.4, 13.2, 14.3-14.7 | 7.11-7.13, 11.12, 13.11, 14.13 |
| Scenario | Ambiguous source structure | reader-engine:19 | 7.5-7.10, 11.2 | 7.12 |
| Requirement | Immutable source snapshot and canonical read | reader-engine:23 | 7.14, 7.16-7.20 | 7.15, 7.18, 7.21 |
| Scenario | Source remains unchanged | reader-engine:30 | 7.17 | 7.18 |
| Scenario | Blocking dirty value | reader-engine:34 | 7.20 | 7.21 |
| Requirement | Deterministic source-format contracts | reader-engine:38 | 11.2-11.6, 13.2-13.7, 14.3-14.8 | 11.12, 13.11, 14.13 |
| Scenario | Unsupported nested JSON source | reader-engine:45 | 13.2, 13.7 | 13.11 |
| Requirement | Exact reader request and descriptor structures | reader-engine:49 | 7.1, 7.2 | 7.4 |
| Scenario | Non-local source path | reader-engine:79 | 7.1 | 7.4 |
| Requirement | Exact inspection result structures | reader-engine:83 | 7.5-7.10 | 7.11-7.13 |
| Scenario | Format detail exclusivity | reader-engine:124 | 7.5, 7.8-7.10 | 7.11 |
| Requirement | Exact immutable snapshot structure | reader-engine:128 | 7.16, 7.17 | 7.18 |
| Scenario | Snapshot before row access | reader-engine:138 | 7.17 | 7.18 |
| Requirement | Reader Engine operation catalog | reader-engine:142 | 7.3 | 7.4 |
| Scenario | Lifecycle operation in workflow | reader-engine:152 | 7.22 | 7.22 |

## transform-workflow

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Single-source single-output request | transform-workflow:3 | 9.1, 9.7 | 9.20 |
| Scenario | Second source rejected | transform-workflow:9 | 9.7 | 9.20 |
| Requirement | Registry-based workflow validation | transform-workflow:13 | 9.7, 9.16 | 9.20 |
| Scenario | Invented capability | transform-workflow:18 | 9.7 | 9.20 |
| Requirement | Minimal explicit output assertions | transform-workflow:22 | 9.5, 9.6 | 11.10 |
| Scenario | Row predicate assertion fails | transform-workflow:31 | 3.5, 3.8, 8.9, 9.6 | 8.18, 11.10 |
| Requirement | Ordered thin orchestration | transform-workflow:35 | 9.14-9.16 | 9.20, 9.21 |
| Scenario | Fixed engine ownership | transform-workflow:41 | 9.16 | 9.21 |
| Requirement | Safe transform lifecycle | transform-workflow:45 | 8.6-8.18, 9.14-9.19 | 8.18, 9.20 |
| Scenario | Successful run | transform-workflow:56 | 8.13, 8.14, 8.17 | 8.18 |
| Scenario | Mid-run failure | transform-workflow:60 | 8.14-8.17 | 8.18 |
| Requirement | Cancellation and transient progress | transform-workflow:64 | 8.14, 8.15, 9.17, 9.18 | 8.18, 9.20 |
| Scenario | Cancellation during streaming | transform-workflow:69 | 8.14, 9.17 | 8.18 |
| Requirement | Complete final report | transform-workflow:73 | 8.14, 8.17, 9.19 | 8.18 |
| Scenario | Report after blocked dirty input | transform-workflow:80 | 7.20, 8.14, 8.17 | 8.18 |
| Requirement | Exact Transform Workflow structures | transform-workflow:84 | 9.1-9.4, 9.8-9.13 | 9.12 |
| Scenario | Strict operation call | transform-workflow:122 | 9.2, 9.7 | 9.20 |
| Requirement | Exact validation request structures | transform-workflow:126 | 9.5, 9.6 | 11.10 |
| Scenario | Unknown assertion field | transform-workflow:139 | 9.5, 9.6 | 11.10 |
| Requirement | Exact deterministic execution plan | transform-workflow:143 | 9.13-9.16 | 9.20 |
| Scenario | Plan reproducibility | transform-workflow:150 | 9.13, 9.14 | 9.20 |

## v1-acceptance

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | CSV vertical-slice gate | v1-acceptance:3 | 11.2-11.11, 11.13 | 11.14-11.17 |
| Scenario | Vertical slice success | v1-acceptance:11 | 11.13 | 11.14-11.17 |
| Requirement | Excel city-filter acceptance | v1-acceptance:16 | 14.3-14.12 | 17.1-17.4 |
| Scenario | City-filter workflow | v1-acceptance:22 | 14.9-14.12 | 17.1-17.4 |
| Requirement | JSON field-extraction acceptance | v1-acceptance:26 | 13.2-13.10 | 17.5-17.8 |
| Scenario | JSON-to-CSV workflow | v1-acceptance:31 | 13.8-13.10 | 17.5-17.8 |
| Requirement | Cleanup-script scoped coverage acceptance | v1-acceptance:35 | 1.10, 12.1-12.29 | 17.9-17.12 |
| Scenario | Cleanup V1 scoped coverage | v1-acceptance:41 | 12.1-12.29 | 17.10, 17.11 |
| Requirement | Codex reliability suite | v1-acceptance:45 | 16.1-16.25 | 18.1-18.5 |
| Scenario | Luna 5.6 target | v1-acceptance:50 | 16.1-16.25 | 18.3, 18.5 |
| Requirement | Baseline-derived performance thresholds | v1-acceptance:54 | 18.6-18.13, 18.16 | 18.17 |
| Scenario | Performance regression | v1-acceptance:60 | 18.16 | 18.17 |
| Requirement | Wide-fixture memory and spill gate | v1-acceptance:64 | 12.15-12.18, 12.25, 18.8-18.14 | 18.15 |
| Scenario | Bounded sort under pressure | v1-acceptance:86 | 12.20-12.25, 18.14 | 12.26, 18.15 |
| Scenario | Batch prototype fidelity | v1-acceptance:90 | 18.12, 18.13 | 18.13 |
| Requirement | Workflow template and CLI fixture gate | v1-acceptance:94 | 9.8-9.13, 15.14-15.17 | 9.12, 15.18-15.20 |
| Scenario | Protected configuration override | v1-acceptance:100 | 4.11, 4.16, 15.17 | 4.20, 15.20 |
| Requirement | Reader, Writer, and logical lifecycle gate | v1-acceptance:104 | 7.14-7.22, 8.6-8.19, 9.14-9.19 | 8.18, 9.20, 11.16 |
| Scenario | Writer failure finalization | v1-acceptance:112 | 8.14-8.17 | 8.18 |
| Requirement | Google Sheets agent-bridge gate | v1-acceptance:116 | 16.12-16.20 | 16.15, 16.21 |
| Scenario | Missing Google connector | v1-acceptance:123 | 16.12, 16.13 | 16.15 |
| Requirement | Extension and architecture acceptance | v1-acceptance:127 | 2.16, 18.18, 18.19 | 18.18, 18.19 |
| Scenario | Test-only extension | v1-acceptance:133 | 2.16, 18.18 | 18.18 |
| Requirement | Repository quality gates | v1-acceptance:137 | 18.20-18.22 | 18.23, 18.24 |
| Scenario | Final implementation gate | v1-acceptance:142 | 18.20-18.22 | 18.23, 18.24 |

## writer-engine

| Kind | Title | Source anchor | Implementation tasks | Focused test tasks |
|---|---|---:|---|---|
| Requirement | Explicit Writer Engine registry | writer-engine:3 | 8.3, 8.4 | 8.5 |
| Scenario | Descriptor and implementation mismatch | writer-engine:9 | 8.3, 8.4 | 8.5 |
| Requirement | Staged, reopened, validated output | writer-engine:13 | 8.6-8.9, 8.13 | 8.18 |
| Scenario | Output validation fails | writer-engine:20 | 8.8, 8.9, 8.13 | 8.18 |
| Requirement | Publication and terminal finalization | writer-engine:24 | 8.10-8.17 | 8.18 |
| Scenario | Cancellation during streaming | writer-engine:34 | 8.14, 8.15 | 8.18 |
| Requirement | Exact writer request and descriptor structures | writer-engine:38 | 8.1-8.4 | 8.5 |
| Scenario | Output equals source | writer-engine:67 | 8.2 | 8.18 |
| Requirement | Exact staged artifact and validation structures | writer-engine:71 | 8.6, 8.9 | 8.18 |
| Scenario | Publication input | writer-engine:84 | 8.6, 8.13 | 8.18 |
| Requirement | Writer Engine operation catalog | writer-engine:88 | 8.4 | 8.5 |
| Scenario | Writer lifecycle operation in workflow | writer-engine:99 | 8.19 | 8.19 |

## Task 1.6: implementation-task contract-decision scan

All implementation tasks were scanned for literal and semantic variants of unresolved contract-decision language: define schema, choose fields, decide behavior, determine, select, specify, establish, and design.

No true unresolved contract-decision task was found. The only literal hits are:

- `1.6` itself, which is the audit task rather than an implementation task.
- `13.2`, which says `select_root_path` decisions but implements the already-fixed inspection decision ID and contract.
- `16.17`, which implements the already-fixed `block` and deterministic `alternate_name` policies; it does not ask an implementer to choose one.

Other superficially decision-like tasks are normative, not undecided: `7.7` implements fixed `InspectionDecision` codes; `8.7` applies a resolved named style; `8.10-8.12` implement the three fixed local collision policies; `9.14` builds the fixed phase order; `12.15-12.17` implement all prescribed dedup modes; `15.2-15.9` implement fixed commands and exits; and `16.13-16.20` implement the fixed Google bridge procedure.

## Task 1.7: existing normative implementation stop rule

Task 1.7 is already fulfilled and must not create a duplicate rule. `design.md:184-185` states: “An implementer SHALL stop for an OpenSpec correction rather than invent a missing field, operation, error, or Skill contract.” The surrounding normative contract map makes the specifications authoritative over implementation tasks. This covers the task’s missing field, enum, operation, API code, and Skill-contract cases; strict revalidation remains the documented active-contract gate.

## Shared run-budget implementation coverage

`Memory-budget enforcement` is traceable to bounded-spool implementations
(`12.15-12.18`, `12.25`) and task `18.14`, which atomically implements shared
run-budget reservation accounting plus Reader, Logical, and Writer hooks.
Task `18.15` is the focused spill-before-limit and no-publication fixture gate.
