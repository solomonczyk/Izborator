# QUALITY_CONTRACT_v1

Purpose: define the baseline quality contract enforced by CI for multi-domain ingestion.

## Scope (gated domains)
- goods
- services

Domains outside this list (e.g., unknown or exploratory) are not gated.

## Metrics (v1)
- valid_rate: share of objects that pass domain validation rules.
- semantic_coverage: 1 - (avg_missing / required_count).
- normalization_success: v1 fixed to 1.0 (placeholder).
- quality_score: 0.5*valid_rate + 0.3*semantic_coverage + 0.2*normalization_success.

## Required counts (v1)
- goods: required_count = 2 (title, price).
- services: required_count = 2 (title, duration|price).

## Quality gates (v1 thresholds)
- goods:
  - valid_rate >= 0.95
  - quality_score >= 0.85
- services:
  - valid_rate >= 0.80
  - quality_score >= 0.70

## CI behavior
- CI fails when logs contain: `scrapingstats: quality gate failed`.
- This converts WARN in runtime logs into hard fail in CI.

## Not a regression (v1)
- domain == "unknown" or any non-gated domain.
- exploratory data outside goods/services.

## Notes
- The contract is config-driven via `QualityGates` defaults.
- normalization_success will be replaced by real signals in later versions.
