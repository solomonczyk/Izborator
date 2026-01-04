#!/usr/bin/env python3
import argparse
import json
import re
from pathlib import Path


def slugify(value: str) -> str:
    value = value.lower()
    value = value.replace("&", "and")
    value = re.sub(r"[^a-z0-9]+", "-", value)
    value = re.sub(r"-{2,}", "-", value).strip("-")
    return value


def main() -> int:
    parser = argparse.ArgumentParser(description="Extract L1 categories from Google taxonomy.")
    parser.add_argument(
        "--input",
        default="docs/category_tree/sources/google/raw/taxonomy.en-US.txt",
        help="Path to taxonomy txt file.",
    )
    parser.add_argument(
        "--output",
        default="docs/category_tree/sources/google/taxonomy_l1.json",
        help="Path to output JSON file.",
    )
    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.output)

    raw = input_path.read_text(encoding="utf-8").splitlines()
    version = ""
    l1_titles = []
    seen = set()

    for line in raw:
        line = line.strip()
        if not line:
            continue
        if line.startswith("#"):
            if "Google_Product_Taxonomy_Version:" in line:
                version = line.split(":", 1)[-1].strip()
            continue
        title = line.split(" > ", 1)[0].strip()
        if title and title not in seen:
            seen.add(title)
            l1_titles.append(title)

    payload = {
        "source": "google",
        "version": version or "unknown",
        "l1": [{"title": title, "slug": slugify(title)} for title in l1_titles],
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
