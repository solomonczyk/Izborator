// app/[locale]/page.tsx
// Главная страница с глобальным поиском (Hero Section)
import React from 'react'
import { Link } from '@/navigation'
import { getTranslations } from 'next-intl/server'
import { fetchCategoriesTree, type CategoryNode } from '@/lib/api'
import { SearchForm } from '@/components/search-form'
import { LanguageSwitcher } from '@/components/language-switcher'

// Делаем страницу динамической, чтобы избежать ошибок при статической генерации
export const dynamic = 'force-dynamic'

// Компонент быстрых категорий
function QuickCategories({ 
  categories, 
  title
}: { 
  categories: CategoryNode[]
  title: string
}) {
  // Показываем все родительские категории (нет ограничения по количеству)
  const quickCategories = categories
  
  if (quickCategories.length === 0) {
    return null
  }

  return (
    <div className="w-full max-w-5xl mx-auto mt-8">
      <h2 className="text-xl font-semibold text-slate-800 mb-4 text-center">
        {title}
      </h2>
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 gap-4">
        {quickCategories.map((category) => (
            <Link
              key={category.id}
              href={`/catalog?category=${category.slug}`}
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
              {category.name || category.name_sr}
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
  const t = await getTranslations({ locale, namespace: 'home' })

  // Загружаем категории для быстрого доступа
  let categories: CategoryNode[] = []
  let categoriesError: string | null = null

  try {
    categories = await fetchCategoriesTree(locale)
  } catch (err) {
    categoriesError = err instanceof Error ? err.message : 'Failed to load categories'
    console.error('Failed to fetch categories:', err)
  }

  // Показываем только родительские категории (level 1) на главной странице
  // Дочерние категории будут доступны при клике на родительскую или в каталоге
  const allCategories = categories.filter(cat => {
    // Показываем только корневые категории (без parent_id, level 1)
    // children может быть пустым массивом или undefined - это нормально
    return cat.level === 1 || !cat.children || cat.children.length === 0
  })

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-50 to-white">
      {/* Language Switcher - в правом верхнем углу */}
      <div className="absolute top-4 right-4 md:top-6 md:right-6 z-10">
        <LanguageSwitcher />
      </div>

      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 py-16 md:py-24">
        <div className="text-center mb-12">
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-slate-900 mb-4">
            {t('title')}
          </h1>
          <p className="text-xl md:text-2xl text-slate-600 mb-8">
            {t('subtitle')}
          </p>
        </div>

        {/* Поисковая строка */}
        <SearchForm />

        {/* Быстрые категории */}
        {allCategories.length > 0 && (
          <QuickCategories categories={allCategories} title={t('popular_categories')} />
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
            {t('footer_text')}
          </p>
        </div>
      </div>
    </main>
  )
}
