# Critical Path Roadmap (2 weeks, v1)

**Scope:** goods + services only  
**Out of scope:** new verticals (healthcare/education/construction), new parsers, UI polish

---

## Goals (must hit)

1. Semantic validation runtime (goods + services)  
2. Quality metrics + CI gates enforcement  
3. UI facets from semantic schema (no string-match)

---

## Constraints (do not break)

- No new verticals until critical path is complete.  
- No new parsers; work with existing pipeline only.  
- UI logic must be driven by semantic schema, not category string heuristics.  
- Quality thresholds must be config-driven (not hardcoded).

---

## Week 1 (Days 1–5) — Validation + Metrics foundation

**Day 1 — Validation contract & data signals**
- Tasks: confirm semantic registry + domain packs; define present_semantic extraction rules per table.  
- Exit: documented mapping of semantic → DB fields + validation result format agreed.

**Day 2 — present_semantic emission**
- Tasks: emit `present_semantic[]` from normalize stage (goods + services); write once, reused by validate/metrics/UI.  
- Exit: runtime logs or stored output show semantic presence per object.

**Day 3 — Semantic Validation Result**
- Tasks: implement validation using domain pack rules; output explainable result.  
- Exit: non‑valid objects report missing semantics (e.g., missing location).

**Day 4 — Metrics aggregation**
- Tasks: compute valid_rate, semantic_coverage, normalization_success per domain.  
- Exit: metrics report per domain (goods/services).

**Day 5 — CI Quality Gates**
- Tasks: wire CI to fail on gates (config‑driven thresholds).  
- Exit: CI can fail for goods/services based on domain metrics.

**Checkpoint W1:** validation + metrics + CI gates for goods/services; no UI changes yet.

---

## Week 2 (Days 6–10) — UI facets from schema

**Day 6 — Facet schema output**
- Tasks: define facet schema response from backend (semantic_type + facet_type).  
- Exit: API exposes facet schema independent of category strings.

**Day 7 — UI: remove string‑match**
- Tasks: replace category/slug heuristics with facet schema in UI.  
- Exit: UI builds filters from schema only.

**Day 8 — UI integration + regressions**
- Tasks: verify browse + filters for goods/services; fix regressions.  
- Exit: filters stable without domain-specific string logic.

**Day 9 — Domain pack alignment**
- Tasks: verify UI + validation align with domain packs and CI gates.  
- Exit: domain packs drive behavior end‑to‑end.

**Day 10 — Documentation & handoff**
- Tasks: update docs + runbooks; finalize checkpoints.  
- Exit: roadmap complete, ready for new verticals.

**Checkpoint W2:** UI facets driven by schema; string‑match removed; CI gates enforced.
