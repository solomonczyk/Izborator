// app/[locale]/page.tsx
// Главная страница с глобальным поиском (Hero Section)
import React from 'react'
import { Link } from '@/navigation'
import { fetchCategoriesTree, type CategoryNode } from '@/lib/api'
import { SearchForm } from '@/components/search-form'

// Компонент быстрых категорий
function QuickCategories({ 
  categories, 
  locale 
}: { 
  categories: CategoryNode[]
  locale: string 
}) {
  // Берем первые 8 категорий (или меньше, если их меньше)
  const quickCategories = categories.slice(0, 8)
  
  if (quickCategories.length === 0) {
    return null
  }

  return (
    <div className="w-full max-w-5xl mx-auto mt-8">
      <h2 className="text-xl font-semibold text-slate-800 mb-4 text-center">
        {locale === 'sr' ? 'Популарне категорије' : 'Popular Categories'}
      </h2>
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 gap-4">
        {quickCategories.map((category) => (
          <Link
            key={category.id}
            href={`/${locale}/catalog?category=${category.slug}`}
            className="bg-white rounded-xl border-2 border-slate-300 p-4 hover:border-blue-400 hover:shadow-md transition-all text-center group"
          >
            <div className="text-3xl mb-2 group-hover:scale-110 transition-transform">
              {/* Иконки для популярных категорий */}
              {category.code === 'phones' || category.slug.includes('telefon') ? '📱' :
               category.code === 'laptops' || category.slug.includes('laptop') ? '💻' :
               category.code === 'tablets' || category.slug.includes('tablet') ? '📱' :
               category.slug.includes('frizerski') || category.slug.includes('beauty') ? '✂️' :
               category.slug.includes('zubarska') || category.slug.includes('dental') ? '🦷' :
               category.slug.includes('masaža') || category.slug.includes('massage') ? '💆' :
               category.slug.includes('servis') || category.slug.includes('repair') ? '🔧' :
               category.slug.includes('prevoz') || category.slug.includes('transport') ? '🚗' :
               '🛍️'}
            </div>
            <p className="text-sm font-medium text-slate-700 group-hover:text-blue-600 transition-colors">
              {category.name_sr_lc || category.name_sr}
            </p>
          </Link>
        ))}
      </div>
    </div>
  )
}

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params

  // Загружаем категории для быстрого доступа
  let categories: CategoryNode[] = []
  let categoriesError: string | null = null

  try {
    categories = await fetchCategoriesTree()
  } catch (err) {
    categoriesError = err instanceof Error ? err.message : 'Failed to load categories'
    console.error('Failed to fetch categories:', err)
  }

  // Преобразуем дерево категорий в плоский список для быстрого доступа
  const allCategories = categories.flatMap(cat => {
    const result: CategoryNode[] = [cat]
    if (cat.children && cat.children.length > 0) {
      result.push(...cat.children)
    }
    return result
  })

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-50 to-white">
      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-slate-900 mb-4">
            {locale === 'sr' 
              ? 'Нађи све што ти треба' 
              : 'Find Everything You Need'}
          </h1>
          <p className="text-xl md:text-2xl text-slate-600 mb-8">
            {locale === 'sr'
              ? 'Претражи производе и услуге из целог интернета'
              : 'Search for products and services across the entire internet'}
          </p>
        </div>

        {/* Поисковая строка */}
        <SearchForm locale={locale} />

        {/* Быстрые категории */}
        {allCategories.length > 0 && (
          <QuickCategories categories={allCategories} locale={locale} />
        )}

        {categoriesError && (
          <div className="mt-8 text-center">
            <p className="text-sm text-yellow-600">
              ⚠️ {categoriesError}
            </p>
          </div>
        )}

        {/* Дополнительная информация */}
        <div className="mt-16 text-center">
          <p className="text-slate-500 text-sm">
            {locale === 'sr'
              ? 'Агрегирамо цене из више продавница и провајдера услуга'
              : 'We aggregate prices from multiple shops and service providers'}
          </p>
        </div>
      </div>
    </main>
  )
}
