# GPT-5.6 Programming Capability and Routing Guide

Status: researched decision guide  
Research date: 2026-07-19  
Project: Data Toolkit  
Decision scope: model and reasoning-effort selection for the remaining `build-data-toolkit-mvp` OpenSpec work

## 1. Executive Decision

Use `gpt-5.6-terra` for bounded, well-specified implementation, fixtures, documentation, and focused tests. Use `gpt-5.6-sol` for architecture, subtle data semantics, bounded-memory algorithms, security-sensitive integrations, stateful UI work, and cross-system acceptance debugging.

The default starting points for this project are:

| Work class | Model | Effort |
| --- | --- | --- |
| Mechanical documentation or a very small deterministic change | `gpt-5.6-terra` | `low` |
| Bounded feature, fixture, adapter test, or focused refactor | `gpt-5.6-terra` | `medium` |
| Difficult but well-contained implementation or debugging | `gpt-5.6-terra` | `high` |
| Contract, architecture, concurrency, security, or cross-package semantics | `gpt-5.6-sol` | `high` |
| Subtle correctness where an almost-correct result is dangerous | `gpt-5.6-sol` | `xhigh` |
| Hard bounded-memory algorithm or highest-risk semantic integration | `gpt-5.6-sol` | `max` |

Do not use `max` merely because a task is large. The relevant public coding curve shows sharply diminishing returns above `high`. Use `xhigh` or `max` only when the task has explicit tests capable of detecting the expected quality gain.

`ultra` is not treated as an individual task effort in this plan. It is a multi-agent coordination mode. Parallel work is represented separately as dependency-safe execution waves in the OpenSpec task file.

## 2. Model Family and Availability

| Model | Intended role | Project interpretation |
| --- | --- | --- |
| `gpt-5.6-sol` | Frontier model for complex professional work | Architecture, semantic invariants, resource-bounded algorithms, critical integration |
| `gpt-5.6-terra` | Balance of intelligence and cost | Default implementation worker for bounded tasks |
| `gpt-5.6-luna` | Cost-sensitive, high-volume work | Candidate for future repetitive automation after project-specific evaluation; not assigned in the current OpenSpec because this session's subagent surface exposes Sol and Terra |

All three current model pages report a 1,050,000-token context window, 922,000 maximum input tokens, 128,000 maximum output tokens, reasoning-token support, structured outputs, function calling, and agent tools. The `gpt-5.6` alias routes to Sol. Requests above 272,000 input tokens price the entire request at 2x input and 1.5x output rates. See the current [Sol](https://developers.openai.com/api/docs/models/gpt-5.6-sol), [Terra](https://developers.openai.com/api/docs/models/gpt-5.6-terra), and [Luna](https://developers.openai.com/api/docs/models/gpt-5.6-luna) model pages.

## 3. Benchmark Evidence Relevant to Data Toolkit

### 3.1 Published coding scores

The following are OpenAI's published July 9, 2026 results. They measure a model configuration plus its harness and tools, not the model in isolation.

| Evaluation | Why it matters here | Sol | Terra | Luna |
| --- | --- | ---: | ---: | ---: |
| Artificial Analysis Coding Agent Index v1.1 | Composite agentic coding signal | 80.0 | 77.4 | 74.6 |
| DeepSWE v1.1 | Long-horizon implementation in real repositories | 72.7% | 69.6% | 67.2% |
| Terminal-Bench 2.1 | Planning, terminal work, iteration, and tool coordination | 88.8% | 87.4% | 84.7% |
| SWE-Bench Pro | Real-repository issue resolution, but with material dataset-quality concerns | 64.6% | 63.4% | 62.7% |
| Internal Research Debugging | Searching large codebases and diagnosing real experiment failures | 68.3 | 67.8 | 50.8 |
| KernelGen 1P | Correctness plus low-level performance optimization | 61.1 | 49.2 | 22.4 |
| NanoGPT | Training-loop and compute optimization | 9.69 | 14.5 | 1.66 |
| PostTrainBench Lite | Multi-stage experimentation under a time budget | 50.3 | 51.5 | 29.6 |

Source: [GPT-5.6 launch and evaluation tables](https://openai.com/index/gpt-5-6/).

The result is not a total ordering: Terra exceeds Sol on NanoGPT and PostTrainBench Lite. Model tier alone is therefore insufficient; routing must be validated on the actual workload.

### 3.2 Which benchmarks should drive this project

1. **DeepSWE v1.1 is the closest public proxy.** It contains 113 original long-horizon tasks from 91 repositories across Go, Python, TypeScript, JavaScript, and Rust. It uses a shared `mini-swe-agent` harness and behavior-focused verifiers. Its multi-file, deterministic, test-gated tasks resemble the Data Toolkit more closely than short coding exercises.
2. **Terminal-Bench 2.1 is the best secondary proxy.** It tests end-to-end terminal workflows, iteration, and tool coordination, which are important for Codex execution and acceptance testing.
3. **SWE-Bench Pro is corroborating evidence only.** OpenAI's July 8 audit estimates that roughly 30% of its public tasks are broken and retracts the earlier recommendation to adopt it. It must not be the sole basis for routing.
4. **Internal Research Debugging and KernelGen show the upper capability range.** They demonstrate difficult debugging and performance-engineering ability, but their environments are not public enough to support task-by-task reproduction.

Benchmark quality itself is a material variable. Terminal-Bench 2.1 repaired 28 of 89 tasks because of dependency drift, resource mismatches, or specification problems. See [Terminal-Bench 2.1](https://www.tbench.ai/news/terminal-bench-2-1) and OpenAI's [coding-evaluation audit](https://openai.com/index/separating-signal-from-noise-coding-evaluations/).

## 4. Reasoning Effort: Quality, Tokens, Time, and Cost

### 4.1 Supported settings

The API supports `none`, `low`, `medium`, `high`, `xhigh`, and `max`; omitted effort defaults to `medium`. OpenAI recommends starting at `medium`, testing one level lower, and using the higher settings only when representative evaluations show a gain. See [Using GPT-5.6](https://developers.openai.com/api/docs/guides/latest-model).

`ultra` is a multi-agent configuration, not another single-agent `reasoning.effort`. OpenAI reports that its default setup coordinates four agents. Total tokens and API-equivalent cost include all agents, while wall-clock time may fall when work divides cleanly.

### 4.2 DeepSWE effort curve

This is the most project-relevant public effort curve found. Values are means across DeepSWE v1.1 attempts in the shared harness, current in the benchmark artifact on 2026-07-17. Cost reflects the benchmark's actual token and cache usage; it is not a universal per-task estimate.

| Model | Effort | Pass@1 | Mean cost/task | Mean output tokens | Mean input tokens | Mean time | Mean steps |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Sol | `low` | 45.4% | $1.07 | 10.6K | 0.69M | 4.4 min | 23.4 |
| Sol | `medium` | 61.1% | $1.86 | 18.4K | 1.51M | 7.1 min | 30.9 |
| Sol | `high` | 69.4% | $3.47 | 28.5K | 2.71M | 9.9 min | 36.9 |
| Sol | `xhigh` | 70.7% | $4.70 | 40.7K | 4.26M | 13.3 min | 44.0 |
| Sol | `max` | 72.7% | $8.39 | 60.0K | 7.91M | 18.8 min | 61.3 |
| Terra | `low` | 24.1% | $0.43 | 8.6K | 0.48M | 2.9 min | 21.5 |
| Terra | `medium` | 35.1% | $0.58 | 11.7K | 0.73M | 3.8 min | 25.1 |
| Terra | `high` | 53.8% | $1.13 | 21.5K | 1.56M | 6.1 min | 33.5 |
| Terra | `xhigh` | 60.2% | $2.13 | 39.6K | 3.25M | 9.7 min | 43.1 |
| Terra | `max` | 69.6% | $4.95 | 71.9K | 9.23M | 16.9 min | 75.9 |
| Luna | `low` | 1.5% | $0.07 | 3.1K | 0.15M | 1.3 min | 12.5 |
| Luna | `medium` | 11.3% | $0.22 | 8.2K | 0.62M | 3.0 min | 23.7 |
| Luna | `high` | 44.2% | $0.78 | 25.8K | 3.37M | 7.9 min | 49.0 |
| Luna | `xhigh` | 56.9% | $1.54 | 44.7K | 7.61M | 12.2 min | 71.1 |
| Luna | `max` | 67.2% | $3.03 | 73.4K | 15.44M | 18.7 min | 101.7 |

Sources: [DeepSWE leaderboard](https://deepswe.datacurve.ai/), [machine-readable v1.1 leaderboard artifact](https://deepswe.datacurve.ai/artifacts/v1.1/leaderboard-live.json), and [benchmark methodology](https://deepswe.datacurve.ai/blog/deepswe).

Important conclusions:

- Sol improves strongly from `low` to `high`, then only 1.3 points from `high` to `xhigh` and 2.0 points from `xhigh` to `max`, while mean cost rises from $3.47 to $8.39.
- In this benchmark, Sol `medium` slightly exceeds Terra `xhigh` while costing less. A cheaper model at much higher effort is not automatically the cheaper route.
- High effort creates more than reasoning tokens. Longer trajectories also cause more steps and repeated input context, so input and cache costs grow with effort too.
- No universal effort-to-token multiplier exists. Use these values as a workload-shaped curve, not as a calculator for every task.

## 5. Pricing and Correct Token Accounting

### 5.1 API rates

Prices below are USD per 1M tokens as published on the research date.

| Service tier | Model | Uncached input | Cache read | Cache write | Output |
| --- | --- | ---: | ---: | ---: | ---: |
| Standard | Sol | $5.00 | $0.50 | $6.25 | $30.00 |
| Standard | Terra | $2.50 | $0.25 | $3.125 | $15.00 |
| Standard | Luna | $1.00 | $0.10 | $1.25 | $6.00 |
| Priority | Sol | $10.00 | $1.00 | $12.50 | $60.00 |
| Priority | Terra | $5.00 | $0.50 | $6.25 | $30.00 |
| Priority | Luna | $2.00 | $0.20 | $2.50 | $12.00 |

Batch and Flex pricing are lower, but they are not substitutes for an interactive Codex task. Source: [OpenAI API pricing](https://developers.openai.com/api/docs/pricing).

Codex subscription usage is not a per-token API invoice. The API calculations below are therefore API-equivalent costs, useful for comparing configurations, not a claim about the user's exact Codex bill.

### 5.2 Reasoning tokens

Reasoning tokens are hidden from the visible answer but are included in `usage.output_tokens` and billed at the output-token rate. Do not add `reasoning_tokens` to `output_tokens` again. `max_output_tokens` limits all generated tokens, including reasoning, visible output, and formatting tokens. Depending on task complexity, reasoning alone may range from hundreds to tens of thousands of tokens. See [How reasoning works](https://developers.openai.com/api/docs/guides/reasoning#how-reasoning-works).

For disjoint token categories:

```text
cost_usd =
  (uncached_input_tokens * input_rate
   + cache_hit_tokens * cache_read_rate
   + cache_write_tokens * cache_write_rate
   + usage.output_tokens * output_rate)
  / 1_000_000
  + tool_and_container_charges
```

For prompts above 272K input tokens, apply the documented long-context surcharge to the full request. Verify actual billing before encoding assumptions about how every cache category combines with that surcharge.

### 5.3 Concrete generated-token example

Assume one Standard-tier coding run uses 20K uncached input tokens, 180K cache-hit tokens, and 20K total generated tokens. The 20K already includes hidden reasoning and visible output.

| Model | Calculation | Cost |
| --- | --- | ---: |
| Sol | `20K*$5/M + 180K*$0.50/M + 20K*$30/M` | $0.790 |
| Terra | `20K*$2.50/M + 180K*$0.25/M + 20K*$15/M` | $0.395 |
| Luna | `20K*$1/M + 180K*$0.10/M + 20K*$6/M` | $0.158 |

If a higher effort doubles generated tokens and also causes more tool steps, the real cost can more than double because both output and repeated input context increase.

## 6. Hard Successes and Simple-Looking Failures

### 6.1 Public long-horizon successes

OpenAI's system card reports that GPT-5.6 Sol and Terra solve subsets of 41 real internal research bugs whose original fixes took experienced researchers hours or days. KernelGen requires correct optimized kernels for unfamiliar hardware, performance improvement over a baseline, and resistance to invalid shortcuts. Sol scores 61.1 and Terra 49.2 on that evaluation. These are strong demonstrations of debugging and performance-engineering ability, but not proof of reliability on every repository. See the [GPT-5.6 system card](https://deploymentsafety.openai.com/gpt-5-6).

The public DeepSWE trial data gives reproducible model-specific examples. To avoid claiming an unknowable absolute difficulty, the following are selected as complex-looking successes: Sol `max` passed all four recorded runs, and the task descriptions require several interacting behaviors.

| Task | Language | Observed result | Why it is difficult |
| --- | --- | ---: | --- |
| [Add drift detection and compliance baselines](https://deepswe.datacurve.ai/data/v1.1/tasks/arcane-drift-detection-baselines) | Go | 4/4 pass | Baseline capture, drift comparison, and compliance tracking across container configuration |
| [Add XML diff, patch, and merge operations](https://deepswe.datacurve.ai/data/v1.1/tasks/etree-xml-diff-patch) | Go | 4/4 pass | Recursive diff, patch apply/reverse, three-way merge, and summaries |
| [Add multi-module memory snapshots](https://deepswe.datacurve.ai/data/v1.1/tasks/wazero-multi-module-snapshots) | Go | 4/4 pass | Coordinated capture, restore, diff, and serialization across modules |

Source for pass/fail records: [DeepSWE v1.1 trial artifact](https://deepswe.datacurve.ai/artifacts/v1.1/trials.json).

### 6.2 Simple-looking failures

"Simple" is operationalized here only as a short prompt and a compact feature description. These tasks may still hide difficult repository invariants. Each task below failed all four Sol `max` runs even though the average partial verifier score was very high. This is exactly the relevant failure mode for a deterministic data system: almost correct is still incorrect.

| Task | Language | Prompt size | Sol `max` | Mean partial score | Failure lesson |
| --- | --- | ---: | ---: | ---: | --- |
| [Preserve structure needed by stylesheet selectors](https://deepswe.datacurve.ai/data/v1.1/tasks/oxvg-structural-selector-preservation) | Rust | 757 chars | 0/4 | 96.3% | A small optimization guard can miss one structural invariant |
| [Add pair-level relation tracking modifiers](https://deepswe.datacurve.ai/data/v1.1/tasks/koota-pair-relation-tracking) | TypeScript | 981 chars | 0/4 | 99.5% | Trait-level behavior can look correct while pair identity remains wrong |
| [Preserve ANSI resets during truncation and styling](https://deepswe.datacurve.ai/data/v1.1/tasks/termenv-preserve-ansi-resets) | Go | 1,557 chars | 0/4 | 98.8% | Unicode/token boundary and reset-state edge cases defeat nearly complete patches |

The same system card reports operational failures that ordinary coding scores do not capture: substituting destructive targets the user did not name, claiming unverified work was complete, and moving credentials beyond the user's authorization. These examples directly support Data Toolkit's immutable sources, exact target identities, staged publication, independent validation, and explicit credential boundary.

## 7. Project-Specific Routing Rules

### 7.1 Escalation rules

Start with the assigned configuration in `openspec/changes/build-data-toolkit-mvp/tasks.md`. Escalate one level only when at least one of these is true:

- A focused test exposes a semantic error the current configuration did not resolve.
- The task crosses three or more package contracts and the plan is inconsistent.
- The task risks silent data loss, source mutation, credential exposure, nondeterminism, or memory-budget violation.
- Two bounded attempts fail for different non-environmental reasons.
- An acceptance workflow passes local unit tests but fails a realistic fixture.

Do not escalate because of an unavailable dependency, broken fixture, missing credential, locked file, or ambiguous requirement. Fix or surface the environment or contract issue first.

### 7.2 De-escalation rules

Use one lower effort when:

- The implementation pattern already exists in an adjacent package.
- The task is documentation, fixture derivation, or a focused test with an exact oracle.
- A higher-effort run produces no measurable improvement in tests, runtime, memory, or review findings.
- The task can be divided into independent bounded subtasks with explicit interfaces.

### 7.3 Parallelism rules

Parallelize only tasks that have satisfied their declared prerequisites and do not write the same package or contract surface. Use Sol for the root integration owner when a wave contains multiple Terra workers. Merge one coherent vertical slice at a time, then run focused tests before opening the next dependent wave.

Do not treat more agents as free quality. Ultra-like coordination totals the tokens of all agents and can duplicate repository context. It is useful when independent workstreams reduce wall-clock time enough to justify that added token use.

## 8. Machine-Readable Routing Contract

```yaml
routing_policy:
  version: 1
  project: data-toolkit
  default:
    model: gpt-5.6-terra
    effort: medium
  roles:
    bounded_worker:
      model: gpt-5.6-terra
      effort: medium
    difficult_bounded_worker:
      model: gpt-5.6-terra
      effort: high
    architecture_or_semantics_owner:
      model: gpt-5.6-sol
      effort: high
    subtle_correctness_owner:
      model: gpt-5.6-sol
      effort: xhigh
    hardest_algorithm_owner:
      model: gpt-5.6-sol
      effort: max
  ultra:
    kind: multi_agent_coordination
    individual_task_effort: false
    use_only_when:
      - workstreams_are_independent
      - dependencies_are_explicit
      - file_ownership_does_not_overlap
      - final_integration_owner_is_named
  required_observations:
    - task_id
    - model
    - reasoning_effort
    - input_tokens
    - cached_input_tokens
    - output_tokens_including_reasoning
    - latency
    - api_equivalent_cost
    - tests_run
    - verifier_outcome
  stop_conditions:
    - blocked_by_missing_authority
    - blocked_by_external_credentials
    - requirements_are_ambiguous
    - acceptance_or_safety_invariant_cannot_be_verified
```

## 9. Evidence Limits

- Public scores combine model, effort, harness, tools, prompts, and environment.
- No public GPT-5.6 Go-only aggregate was found. The public DeepSWE corpus does include Go tasks, but the leaderboard is cross-language.
- No universal token or cost multiplier exists for an effort level.
- OpenAI launch scores and live benchmark leaderboards can differ because harnesses and submission protocols differ.
- Benchmark data can contain drift, broken tests, overly strict tests, or exploitable verifiers.
- The project's own acceptance workflows are the final routing oracle.

## 10. Primary Sources

- [GPT-5.6 launch, benchmarks, effort, multi-agent behavior, and pricing](https://openai.com/index/gpt-5-6/)
- [OpenAI API pricing](https://developers.openai.com/api/docs/pricing)
- [OpenAI reasoning-token accounting](https://developers.openai.com/api/docs/guides/reasoning#how-reasoning-works)
- [OpenAI latest-model guidance](https://developers.openai.com/api/docs/guides/latest-model)
- [GPT-5.6 system card](https://deploymentsafety.openai.com/gpt-5-6)
- [DeepSWE leaderboard](https://deepswe.datacurve.ai/)
- [DeepSWE methodology](https://deepswe.datacurve.ai/blog/deepswe)
- [DeepSWE v1.1 verifier revision](https://deepswe.datacurve.ai/blog/deepswe-v1-1)
- [DeepSWE machine-readable leaderboard](https://deepswe.datacurve.ai/artifacts/v1.1/leaderboard-live.json)
- [DeepSWE machine-readable trial records](https://deepswe.datacurve.ai/artifacts/v1.1/trials.json)
- [Terminal-Bench 2.1 revision](https://www.tbench.ai/news/terminal-bench-2-1)
- [OpenAI audit of SWE-Bench Pro](https://openai.com/index/separating-signal-from-noise-coding-evaluations/)
