import fs from "node:fs"
import path from "node:path"

const root = process.cwd()
const schemaPath = path.join(root, "backend", "internal", "domainpack", "domain_pack_v1.json")
const uiPath = path.join(root, "frontend", "app", "[locale]", "catalog", "page.tsx")

function readJson(filePath) {
  try {
    return JSON.parse(fs.readFileSync(filePath, "utf8"))
  } catch (err) {
    console.error(`Failed to read JSON: ${filePath}`)
    console.error(err instanceof Error ? err.message : err)
    process.exit(1)
  }
}

if (!fs.existsSync(schemaPath)) {
  console.error(`Facet schema not found: ${schemaPath}`)
  process.exit(1)
}
if (!fs.existsSync(uiPath)) {
  console.error(`UI catalog page not found: ${uiPath}`)
  process.exit(1)
}

const schema = readJson(schemaPath)
const uiSource = fs.readFileSync(uiPath, "utf8")

const domains = Object.keys(schema || {})
if (domains.length === 0) {
  console.error("Facet schema is empty or invalid.")
  process.exit(1)
}

const facetTypes = new Set()
for (const domain of domains) {
  const facets = schema?.[domain]?.facets
  if (!Array.isArray(facets)) {
    console.error(`Facet schema is missing facets array for ${domain}`)
    process.exit(1)
  }
  for (const facet of facets) {
    if (facet && typeof facet.semantic_type === "string") {
      facetTypes.add(facet.semantic_type)
    }
  }
}

const missing = []
for (const semanticType of facetTypes) {
  const token = `facetSet.has("${semanticType}")`
  if (!uiSource.includes(token)) {
    missing.push(semanticType)
  }
}

if (missing.length > 0) {
  console.error("Facet/UI contract check failed.")
  console.error("UI is missing facetSet checks for:", missing.join(", "))
  process.exit(1)
}

console.log("Facet/UI contract check passed.")
