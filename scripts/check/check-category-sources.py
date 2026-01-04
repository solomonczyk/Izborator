#!/usr/bin/env python3
import argparse
import json
import sys
from collections import Counter
from pathlib import Path


def load_json(path: Path):
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate category source mappings.")
    parser.add_argument(
        "--canonical",
        default="docs/category_tree/canonical_tree_v1.json",
        help="Path to canonical tree JSON.",
    )
    parser.add_argument(
        "--mapping",
        default="docs/category_tree/mappings/eponuda_v1.json",
        help="Path to source mapping JSON.",
    )
    parser.add_argument(
        "--source",
        default="docs/category_tree/sources/eponuda/raw/tree_v1.json",
        help="Path to raw source tree JSON.",
    )
    parser.add_argument(
        "--google-mapping",
        default="docs/category_tree/mappings/google_v1.json",
        help="Path to Google taxonomy mapping JSON.",
    )
    parser.add_argument(
        "--google-derived",
        default="docs/category_tree/sources/google/taxonomy_l1.json",
        help="Path to derived Google taxonomy L1 JSON.",
    )
    args = parser.parse_args()

    canonical_path = Path(args.canonical)
    mapping_path = Path(args.mapping)
    source_path = Path(args.source)
    google_mapping_path = Path(args.google_mapping)
    google_derived_path = Path(args.google_derived)

    canonical = load_json(canonical_path)
    mapping = load_json(mapping_path)
    source = load_json(source_path)
    google_mapping = load_json(google_mapping_path)
    google_derived = load_json(google_derived_path)

    canonical_l1 = {item["id"] for item in canonical.get("categories", [])}
    canonical_l2 = set()
    canonical_l3 = set()
    for item in canonical.get("categories", []):
        for child in item.get("children", []):
            canonical_l2.add(child.get("id"))
            for grandchild in child.get("children", []):
                canonical_l3.add(grandchild.get("id"))
    source_ids = {item["id"] for item in source.get("categories", [])}

    l1_map = mapping.get("l1_map", [])
    mapped_source_ids = [entry.get("source_id") for entry in l1_map]
    mapped_set = {value for value in mapped_source_ids if value}

    missing = sorted(source_ids - mapped_set)
    extra = sorted(mapped_set - source_ids)
    duplicates = [key for key, count in Counter(mapped_source_ids).items() if count > 1]
    target_level_map = {
        "L1": canonical_l1,
        "L2": canonical_l2,
        "L3": canonical_l3,
    }
    missing_targets = []
    invalid_levels = []
    for entry in l1_map:
        if entry.get("action") != "map":
            continue
        target_id = entry.get("target_id")
        target_level = entry.get("target_level") or "L1"
        if target_level not in target_level_map:
            invalid_levels.append(target_level)
            continue
        if target_id and target_id not in target_level_map[target_level]:
            missing_targets.append(target_id)

    google_l1 = google_derived.get("l1", [])
    google_map = google_mapping.get("l1_map", [])
    google_slugs = {item.get("slug") for item in google_l1 if item.get("slug")}
    mapped_google_slugs = {
        entry.get("source_slug")
        for entry in google_map
        if entry.get("source_slug")
    }
    missing_google = sorted(google_slugs - mapped_google_slugs)
    extra_google = sorted(mapped_google_slugs - google_slugs)
    invalid_actions = sorted(
        {
            entry.get("action")
            for entry in google_map
            if entry.get("action") not in {"review", "map", "drop"}
        }
    )
    if google_mapping.get("version") and google_derived.get("version"):
        version_match = google_mapping.get("version") == google_derived.get("version")
    else:
        version_match = True

    errors = []
    if missing:
        errors.append(f"Missing mappings for: {', '.join(missing)}")
    if extra:
        errors.append(f"Mapping contains unknown source ids: {', '.join(extra)}")
    if duplicates:
        errors.append(f"Duplicate source ids in mapping: {', '.join(duplicates)}")
    if invalid_levels:
        errors.append(f"Unknown target levels: {', '.join(sorted(set(invalid_levels)))}")
    if missing_targets:
        errors.append(f"Unknown target ids: {', '.join(sorted(set(missing_targets)))}")
    if missing_google:
        errors.append(f"Missing Google mapping for: {', '.join(missing_google)}")
    if extra_google:
        errors.append(f"Google mapping contains unknown slugs: {', '.join(extra_google)}")
    if invalid_actions:
        errors.append(f"Google mapping has invalid actions: {', '.join(invalid_actions)}")
    if not version_match:
        errors.append("Google mapping version does not match derived taxonomy version")

    if errors:
        for err in errors:
            print(f"ERROR: {err}")
        return 1

    print("OK: category source mapping is consistent.")
    print(f"Source L1 categories: {len(source_ids)}")
    print(f"Mapped entries: {len(l1_map)}")
    print(f"Canonical L1 categories: {len(canonical_l1)}")
    print(f"Canonical L2 categories: {len(canonical_l2)}")
    print(f"Google L1 categories: {len(google_l1)}")
    print(f"Google mapping entries: {len(google_map)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
