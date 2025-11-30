// app/[locale]/catalog/page.tsx
import React from "react"
import { Link } from '@/navigation'
import { getTranslations } from 'next-intl/server'

type BrowseProduct = {
  id: string
  name: string
  brand?: string
  category?: string
  image_url?: string
  min_price?: number
  max_price?: number
  currency?: string
  shops_count?: number
}

type BrowseResponse = {
  items: BrowseProduct[]
  page: number
  per_page: number
  total: number
  total_pages: number
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8081"

// Форматирование цены с разделителями тысяч
function formatPrice(price: number): string {
  return price.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ")
}

async function fetchCatalog(params: {
  query?: string
  category?: string
  minPrice?: string
  maxPrice?: string
  page?: string
  perPage?: string
  sort?: string
  lang?: string
}): Promise<BrowseResponse> {
  const url = new URL("/api/v1/products/browse", API_BASE)

  if (params.query) url.searchParams.set("query", params.query)
  if (params.category) url.searchParams.set("category", params.category)
  if (params.minPrice) url.searchParams.set("min_price", params.minPrice)
  if (params.maxPrice) url.searchParams.set("max_price", params.maxPrice)
  if (params.sort) url.searchParams.set("sort", params.sort)
  if (params.lang) url.searchParams.set("lang", params.lang)

  const page = params.page ? parseInt(params.page, 10) : 1
  const perPage = params.perPage ? parseInt(params.perPage, 10) : 20

  url.searchParams.set("page", page.toString())
  url.searchParams.set("per_page", perPage.toString())

  const res = await fetch(url.toString(), {
    next: { revalidate: 180 }, // Кэшируем на 3 минуты
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch catalog: ${res.status}`)
  }

  return res.json()
}

import { Pagination } from './pagination'

export default async function CatalogPage({
  params,
  searchParams,
}: {
  params: Promise<{ locale: string }>
  searchParams?: Promise<{
    q?: string
    category?: string
    min_price?: string
    max_price?: string
    page?: string
    per_page?: string
    sort?: string
  }>
}) {
  const { locale } = await params
  const resolvedSearchParams = await searchParams
  const t = await getTranslations({ locale })
  
  const query = resolvedSearchParams?.q || ""
  const category = resolvedSearchParams?.category || ""
  const minPrice = resolvedSearchParams?.min_price || ""
  const maxPrice = resolvedSearchParams?.max_price || ""
  const page = resolvedSearchParams?.page || "1"
  const perPage = resolvedSearchParams?.per_page || "20"
  const sort = resolvedSearchParams?.sort || "price_asc"

  let data: BrowseResponse
  let error: string | null = null

  try {
    data = await fetchCatalog({
      query,
      category,
      minPrice,
      maxPrice,
      page,
      perPage,
      sort,
      lang: locale,
    })
  } catch (err) {
    error = err instanceof Error ? err.message : t('common.error_unknown')
    data = {
      items: [],
      page: 1,
      per_page: 20,
      total: 0,
      total_pages: 0,
    }
  }

  // Создаём базовые параметры для URL (без page, чтобы пагинация могла менять только страницу)
  const baseParams = new URLSearchParams()
  if (query) baseParams.set("q", query)
  if (category) baseParams.set("category", category)
  if (minPrice) baseParams.set("min_price", minPrice)
  if (maxPrice) baseParams.set("max_price", maxPrice)
  if (sort && sort !== "price_asc") baseParams.set("sort", sort)
  if (perPage !== "20") baseParams.set("per_page", perPage)

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-semibold mb-6 text-slate-900">
          {t('catalog.title')} {query ? `— ${t('catalog.search_for')} "${query}"` : ""}
        </h1>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800 font-medium">{t('catalog.error_load')}:</p>
            <p className="text-red-600 text-sm">{error}</p>
            <p className="text-red-600 text-xs mt-2">
              {t('catalog.error_api_hint', { apiBase: API_BASE })}
            </p>
          </div>
        )}

        {/* Фильтры и поиск */}
        <div className="bg-white rounded-xl shadow-sm border-2 border-slate-300 p-6 mb-6">
          <form method="GET" action={`/${locale}/catalog`} className="space-y-4">
            {/* Поисковая строка */}
            <div>
              <label htmlFor="search" className="block text-sm font-medium mb-2 text-slate-900">
                {t('catalog.search')}
              </label>
              <input
                type="text"
                id="search"
                name="q"
                defaultValue={query}
                placeholder={t('catalog.search_placeholder')}
                style={{ color: '#0f172a' }}
                className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-slate-400 text-slate-900"
              />
            </div>

            {/* Фильтры в одну строку */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* Категория */}
              <div>
                <label
                  htmlFor="category"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  {t('catalog.category')}
                </label>
                <select
                  id="category"
                  name="category"
                  defaultValue={category}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="" className="text-slate-400">{t('catalog.all_categories')}</option>
                  <option value="phones" className="text-slate-900">{t('catalog.category_phones')}</option>
                  <option value="laptops" className="text-slate-900">{t('catalog.category_laptops')}</option>
                  <option value="tablets" className="text-slate-900">{t('catalog.category_tablets')}</option>
                  <option value="accessories" className="text-slate-900">{t('catalog.category_accessories')}</option>
                </select>
              </div>

              {/* Минимальная цена */}
              <div>
                <label
                  htmlFor="min_price"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  {t('catalog.price_from')}
                </label>
                <input
                  type="number"
                  id="min_price"
                  name="min_price"
                  defaultValue={minPrice}
                  placeholder="0"
                  min="0"
                  style={{ color: '#0f172a' }}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-slate-400 text-slate-900"
                />
              </div>

              {/* Максимальная цена */}
              <div>
                <label
                  htmlFor="max_price"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  {t('catalog.price_to')}
                </label>
                <input
                  type="number"
                  id="max_price"
                  name="max_price"
                  defaultValue={maxPrice}
                  placeholder="1000000"
                  min="0"
                  style={{ color: '#0f172a' }}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-slate-400 text-slate-900"
                />
              </div>

              {/* Сортировка */}
              <div>
                <label htmlFor="sort" className="block text-sm font-medium mb-2 text-slate-900">
                  {t('catalog.sort')}
                </label>
                <select
                  id="sort"
                  name="sort"
                  defaultValue={sort}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="price_asc" className="text-slate-900">{t('catalog.sort_price_asc')}</option>
                  <option value="price_desc" className="text-slate-900">{t('catalog.sort_price_desc')}</option>
                  <option value="name_asc" className="text-slate-900">{t('catalog.sort_name_asc')}</option>
                  <option value="name_desc" className="text-slate-900">{t('catalog.sort_name_desc')}</option>
                </select>
              </div>
            </div>

            {/* Скрытые параметры для сохранения при смене фильтров */}
            {page !== "1" && <input type="hidden" name="page" value="1" />}
            {perPage !== "20" && (
              <input type="hidden" name="per_page" value={perPage} />
            )}

            {/* Кнопки */}
            <div className="flex gap-2">
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {t('catalog.apply_filters')}
              </button>
              {(query || category || minPrice || maxPrice || sort !== "price_asc") && (
                <Link
                  href={`/${locale}/catalog`}
                  className="px-4 py-2 bg-slate-200 text-slate-700 rounded-lg hover:bg-slate-300 focus:outline-none focus:ring-2 focus:ring-slate-500"
                >
                  {t('catalog.reset')}
                </Link>
              )}
            </div>
          </form>
        </div>

        {/* Результаты */}
        {data.items.length === 0 ? (
          <div className="bg-white rounded-xl shadow-sm border-2 border-slate-300 p-8 text-center">
            <p className="text-slate-500 text-lg">{t('catalog.no_results')}</p>
            <p className="text-slate-400 text-sm mt-2">
              {t('catalog.no_results_hint')}
            </p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between mb-4">
              <p className="text-sm text-slate-700">
                {t('catalog.found')}: <span className="font-semibold text-slate-900">{data.total}</span>{" "}
                {t('catalog.items')}
                {data.total_pages > 1 && (
                  <span className="ml-2">
                    ({t('catalog.page')} {data.page} {t('catalog.of')} {data.total_pages})
                  </span>
                )}
              </p>
            </div>

            <ul className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {data.items.map((p) => (
                <li
                  key={p.id}
                  className="bg-white rounded-xl shadow-sm border-2 border-slate-300 p-4 hover:shadow-md hover:border-blue-400 transition-all"
                >
                  <Link href={`/${locale}/product/${p.id}`} className="flex gap-4">
                    {p.image_url && (
                      <img
                        src={p.image_url}
                        alt={p.name}
                        className="w-24 h-24 object-contain rounded-md border-2 border-slate-300 bg-white flex-shrink-0"
                      />
                    )}
                    <div className="flex-1 min-w-0">
                      <h2 className="font-medium text-sm mb-1 line-clamp-2 hover:text-blue-600 text-slate-900">
                        {p.name}
                      </h2>
                      {p.brand && (
                        <p className="text-xs text-slate-600 mb-1">
                          {p.brand}
                        </p>
                      )}
                      {typeof p.min_price === "number" && (
                        <p className="font-semibold text-sm mt-2 text-slate-900">
                          {p.min_price === p.max_price
                            ? `${formatPrice(p.min_price)} ${p.currency || "RSD"}`
                            : `${t('catalog.from')} ${formatPrice(p.min_price)} ${t('catalog.to')} ${formatPrice(p.max_price || p.min_price)} ${
                                p.currency || "RSD"
                              }`}
                        </p>
                      )}
                      {typeof p.shops_count === "number" && (
                        <p className="text-xs text-slate-600 mt-1">
                          {t('catalog.shops_count')}: {p.shops_count}
                        </p>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>

            {/* Пагинация */}
            <Pagination
              currentPage={data.page}
              totalPages={data.total_pages}
              baseParams={baseParams}
              locale={locale}
            />
          </>
        )}
      </div>
    </main>
  )
}

