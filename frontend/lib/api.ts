// API utilities for backend requests

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8081"
const TENANT_ID =
  process.env.NEXT_PUBLIC_TENANT_ID || process.env.TENANT_ID || "default"

export async function apiFetch(
  path: string | URL,
  options?: RequestInit,
): Promise<Response> {
  const url =
    typeof path === "string" ? new URL(path, API_BASE) : new URL(path.toString())
  url.searchParams.set("tenant_id", TENANT_ID)
  return fetch(url.toString(), options)
}

export interface CategoryNode {
  id: string
  slug: string
  code: string
  name?: string
  name_sr: string
  name_sr_lc: string
  level: number
  is_active: boolean
  sort_order: number
  children?: CategoryNode[]
}

export interface City {
  id: string
  slug: string
  name_sr: string
  region_sr?: string
  sort_order: number
  is_active: boolean
}

export type HomeHero = {
  title: string
  subtitle?: string
  searchPlaceholder: string
  showTypeToggle: boolean
  showCitySelect: boolean
  defaultType: "all" | "good" | "service"
}

export type HomeCategoryCard = {
  id: string
  title: string
  hint?: string
  icon_key?: string
  href: string
  priority?: "primary" | "secondary"
  weight?: number
  domain?: "good" | "service" | "all"
  analytics_id?: string
}

export type HomeModel = {
  version: "1"
  tenant_id: string
  locale: string
  hero: HomeHero
  categoryCards: HomeCategoryCard[]
}

export type HomeMeta = {
  version: "1"
  tenant_id: string
  locale: string
  cards_count: number
  showTypeToggle: boolean
  showCitySelect: boolean
  defaultType: "all" | "good" | "service"
}

export async function fetchCategoriesTree(locale?: string): Promise<CategoryNode[]> {
  try {
    const url = locale
      ? `/api/v1/categories/tree?lang=${locale}`
      : "/api/v1/categories/tree"

    const res = await apiFetch(url, {
      cache: "no-store",
      next: { revalidate: 0 },
    })

    if (!res.ok) {
      console.warn(`Failed to fetch categories: ${res.status}`)
      return []
    }

    const data = await res.json()

    if (Array.isArray(data)) {
      return data
    }
    if (data && typeof data === "object") {
      return [data]
    }
    console.warn("Categories API returned unexpected format:", data)
    return []
  } catch (error) {
    console.warn(
      "Categories API unavailable, returning empty array:",
      error instanceof Error ? error.message : String(error),
    )
    return []
  }
}

export async function fetchCities(): Promise<City[]> {
  const res = await apiFetch("/api/v1/cities", {
    next: { revalidate: 3600 },
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch cities: ${res.status}`)
  }

  const data = await res.json()

  if (!Array.isArray(data)) {
    console.warn("Cities API returned non-array:", data)
    return []
  }

  return data
}

export async function fetchHomeModel(params: {
  locale?: string
}): Promise<HomeModel | null> {
  try {
    const url = new URL("/api/v1/home", API_BASE)
    if (params.locale) {
      url.searchParams.set("locale", params.locale)
    }

    const res = await apiFetch(url, {
      cache: "no-store",
      next: { revalidate: 0 },
    })

    if (!res.ok) {
      console.warn(`Failed to fetch home model: ${res.status}`)
      return null
    }

    const data = (await res.json()) as Partial<HomeModel>
    if (!data || data.version !== "1") {
      console.warn(`Home model version mismatch: ${data?.version ?? "unknown"}`)
      return null
    }
    return data as HomeModel
  } catch (error) {
    console.warn(
      "Home model API unavailable:",
      error instanceof Error ? error.message : String(error),
    )
    return null
  }
}

export async function fetchHomeMeta(params: {
  locale?: string
}): Promise<HomeMeta | null> {
  try {
    const url = new URL("/api/v1/home/meta", API_BASE)
    if (params.locale) {
      url.searchParams.set("locale", params.locale)
    }

    const res = await apiFetch(url, {
      next: { revalidate: 60 },
    })

    if (!res.ok) {
      console.warn(`Failed to fetch home meta: ${res.status}`)
      return null
    }

    const data = (await res.json()) as Partial<HomeMeta>
    if (!data || data.version !== "1") {
      console.warn(`Home meta version mismatch: ${data?.version ?? "unknown"}`)
      return null
    }
    return data as HomeMeta
  } catch (error) {
    console.warn(
      "Home meta API unavailable:",
      error instanceof Error ? error.message : String(error),
    )
    return null
  }
}

export function flattenCategories(
  categories: CategoryNode[] | CategoryNode | null | undefined,
): Array<{ value: string; label: string; level: number }> {
  const result: Array<{ value: string; label: string; level: number }> = []

  if (!categories) {
    return result
  }

  const categoriesArray: CategoryNode[] = Array.isArray(categories)
    ? categories
    : [categories]

  function traverse(nodes: CategoryNode[], level: number = 0) {
    if (!nodes || !Array.isArray(nodes)) {
      return
    }

    for (const node of nodes) {
      if (node.is_active !== false) {
        const indent = "  ".repeat(level)
        result.push({
          value: node.slug,
          label: `${indent}${node.name || node.name_sr}`,
          level: node.level || level,
        })

        if (node.children && node.children.length > 0) {
          traverse(node.children, level + 1)
        }
      }
    }
  }

  traverse(categoriesArray)
  return result
}
