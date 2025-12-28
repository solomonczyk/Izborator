// API utilities для работы с backend

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8081"

export interface CategoryNode {
  id: string
  slug: string
  code: string
  name?: string      // Переведенное название (в зависимости от locale)
  name_sr: string   // Сербское название (для обратной совместимости)
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
export async function fetchCategoriesTree(locale?: string): Promise<CategoryNode[]> {
  try {
    // Передаем locale в query параметре для получения переведенных названий
    const url = locale 
      ? `${API_BASE}/api/v1/categories/tree?lang=${locale}`
      : `${API_BASE}/api/v1/categories/tree`
    
    const res = await fetch(url, {
      cache: 'no-store', // Не кэшируем - категории могут изменяться
      next: { revalidate: 0 }, // Отключаем revalidation
    })

    if (!res.ok) {
      console.warn(`Failed to fetch categories: ${res.status}`)
      return [] // Возвращаем пустой массив вместо ошибки
    }

    const data = await res.json()
    
    // API может вернуть объект или массив - нормализуем в массив
    if (Array.isArray(data)) {
      return data
    } else if (data && typeof data === 'object') {
      // Если это объект (одиночная категория), оборачиваем в массив
      return [data]
    } else {
      console.warn('Categories API returned unexpected format:', data)
      return []
    }
  } catch (error) {
    // Во время сборки API может быть недоступен - это нормально
    // Возвращаем пустой массив, чтобы сборка не падала
    console.warn('Categories API unavailable, returning empty array:', error instanceof Error ? error.message : String(error))
    return []
  }
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
export function flattenCategories(categories: CategoryNode[] | CategoryNode | null | undefined): Array<{ value: string; label: string; level: number }> {
  const result: Array<{ value: string; label: string; level: number }> = []

  // Проверяем, что categories существует
  if (!categories) {
    return result
  }
  
  // Если это не массив, но объект - оборачиваем в массив
  const categoriesArray: CategoryNode[] = Array.isArray(categories) ? categories : [categories]

  function traverse(nodes: CategoryNode[], level: number = 0) {
    // Дополнительная проверка на случай, если nodes не массив
    if (!nodes || !Array.isArray(nodes)) {
      return
    }

    for (const node of nodes) {
      // Проверяем is_active, если поле существует, иначе считаем активным
      if (node.is_active !== false) {
        const indent = "  ".repeat(level)
        result.push({
          value: node.slug,
          label: `${indent}${node.name || node.name_sr}`, // Используем переведенное название, если есть
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

