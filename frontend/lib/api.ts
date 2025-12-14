// API utilities для работы с backend

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3002"

export interface CategoryNode {
  id: string
  slug: string
  code: string
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

/**
 * Загружает дерево категорий
 */
export async function fetchCategoriesTree(): Promise<CategoryNode[]> {
  const res = await fetch(`${API_BASE}/api/v1/categories/tree`, {
    next: { revalidate: 3600 }, // Кэшируем на 1 час
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch categories: ${res.status}`)
  }

  const data = await res.json()
  
  // Убеждаемся, что возвращаем массив
  if (!Array.isArray(data)) {
    console.warn('Categories API returned non-array:', data)
    return []
  }

  return data
}

/**
 * Загружает список активных городов
 */
export async function fetchCities(): Promise<City[]> {
  const res = await fetch(`${API_BASE}/api/v1/cities`, {
    next: { revalidate: 3600 }, // Кэшируем на 1 час
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch cities: ${res.status}`)
  }

  const data = await res.json()
  
  // Убеждаемся, что возвращаем массив
  if (!Array.isArray(data)) {
    console.warn('Cities API returned non-array:', data)
    return []
  }

  return data
}

/**
 * Преобразует дерево категорий в плоский список для select
 */
export function flattenCategories(categories: CategoryNode[] | null | undefined): Array<{ value: string; label: string; level: number }> {
  const result: Array<{ value: string; label: string; level: number }> = []

  // Проверяем, что categories - это массив
  if (!categories || !Array.isArray(categories)) {
    return result
  }

  function traverse(nodes: CategoryNode[], level: number = 0) {
    // Дополнительная проверка на случай, если nodes не массив
    if (!nodes || !Array.isArray(nodes)) {
      return
    }

    for (const node of nodes) {
      if (node.is_active) {
        const indent = "  ".repeat(level)
        result.push({
          value: node.slug,
          label: `${indent}${node.name_sr}`,
          level: node.level,
        })
        
        if (node.children && node.children.length > 0) {
          traverse(node.children, level + 1)
        }
      }
    }
  }

  traverse(categories)
  return result
}

