// app/[locale]/catalog/page.tsx
import React from "react"
import { Link } from '@/navigation'
import { getTranslations } from 'next-intl/server'
import { fetchCategoriesTree, fetchCities, flattenCategories, type CategoryNode, type City } from '@/lib/api'
import { ProductCard } from '@/components/product-card'
import { LanguageSwitcher } from '@/components/language-switcher'
import { TypeSelect } from "./type-select"

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
  shop_names?: string[]
  type?: 'good' | 'service'
  service_metadata?: {
    duration?: string
    master_name?: string
    service_area?: string
  }
  is_deliverable?: boolean
  is_onsite?: boolean
}

type BrowseResponse = {
  items: BrowseProduct[]
  page: number
  per_page: number
  total: number
  total_pages: number
}

type FacetDefinition = {
  semantic_type: string
  facet_type: string
  values?: string[]
}

type FacetSchemaResponse = {
  domain: string
  tenant_id?: string
  facets: FacetDefinition[]
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8081"

async function fetchCatalog(params: {
  query?: string
  category?: string
  brand?: string
  city?: string
  type?: string
  minPrice?: string
  maxPrice?: string
  minDuration?: string
  maxDuration?: string
  page?: string
  perPage?: string
  sort?: string
  lang?: string
}): Promise<BrowseResponse> {
  const url = new URL("/api/v1/products/browse", API_BASE)

  if (params.query) url.searchParams.set("query", params.query)
  if (params.category) url.searchParams.set("category", params.category)
  if (params.brand) url.searchParams.set("brand", params.brand)
  if (params.city) url.searchParams.set("city", params.city)
  if (params.type) url.searchParams.set("type", params.type)
  if (params.minPrice) url.searchParams.set("min_price", params.minPrice)
  if (params.maxPrice) url.searchParams.set("max_price", params.maxPrice)
  if (params.minDuration) url.searchParams.set("min_duration", params.minDuration)
  if (params.maxDuration) url.searchParams.set("max_duration", params.maxDuration)
  if (params.sort) url.searchParams.set("sort", params.sort)
  if (params.lang) url.searchParams.set("lang", params.lang)

  const page = params.page ? parseInt(params.page, 10) : 1
  const perPage = params.perPage ? parseInt(params.perPage, 10) : 20

  url.searchParams.set("page", page.toString())
  url.searchParams.set("per_page", perPage.toString())

  const res = await fetch(url.toString(), {
    next: { revalidate: 0 }, // Временно отключен кэш для теста
    cache: 'no-store', // Принудительно не кэшировать
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch catalog: ${res.status}`)
  }

  return res.json()
}

async function fetchFacets(params: { type: "goods" | "services"; tenantId: string }): Promise<FacetSchemaResponse> {
  const url = new URL("/api/v1/products/facets", API_BASE)
  url.searchParams.set("type", params.type)
  url.searchParams.set("tenant_id", params.tenantId)

  const res = await fetch(url.toString(), {
    next: { revalidate: 300 },
  })

  if (!res.ok) {
    throw new Error(`Failed to fetch facets: ${res.status}`)
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
    brand?: string
    city?: string
    type?: string
    min_price?: string
    max_price?: string
    min_duration?: string
    max_duration?: string
    page?: string
    per_page?: string
    sort?: string
  }>
}) {
  const ssrStart = Date.now()
  const { locale } = await params
  const resolvedSearchParams = await searchParams
  const t = await getTranslations({ locale })
  
  const query = resolvedSearchParams?.q || ""
  const category = resolvedSearchParams?.category || ""
  const brand = resolvedSearchParams?.brand || ""
  const city = resolvedSearchParams?.city || ""
  const productType = resolvedSearchParams?.type || ""
  const tenantId = process.env.NEXT_PUBLIC_TENANT_ID || process.env.TENANT_ID || "default"
  const minPrice = resolvedSearchParams?.min_price || ""
  const maxPrice = resolvedSearchParams?.max_price || ""
  const minDuration = resolvedSearchParams?.min_duration || ""
  const maxDuration = resolvedSearchParams?.max_duration || ""
  const page = resolvedSearchParams?.page || "1"
  const perPage = resolvedSearchParams?.per_page || "20"
  const sort = resolvedSearchParams?.sort || "price_asc"

  const facetsType = productType === "service" ? "services" : "goods"
  const catalogPromise = fetchCatalog({
    query,
    category,
    brand,
    city,
    type: productType,
    minPrice,
    maxPrice,
    minDuration,
    maxDuration,
    page,
    perPage,
    sort,
    lang: locale,
  })
  const facetsPromise = fetchFacets({ type: facetsType, tenantId })

  let facetSet = new Set<string>()
  let facetsCount = 0
  let brandOptions: string[] = []

  try {
    const facetsResponse = await facetsPromise
    const facets = Array.isArray(facetsResponse.facets) ? facetsResponse.facets : []
    facetsCount = facets.length
    facetSet = new Set(facets.map((facet) => facet.semantic_type))
    const brandFacet = facets.find((facet) => facet.semantic_type === "brand")
    brandOptions = Array.isArray(brandFacet?.values) ? brandFacet.values : []
    console.log('Facets loaded:', facetsType, facets)
  } catch (err) {
    console.error('Facets error:', err)
  }

  // Загружаем категории и города
  const shouldShowCategory = facetSet.has("category")
  const shouldShowLocation = facetSet.has("location")
  let categories: CategoryNode[] = []
  let cities: City[] = []
  let categoriesError: string | null = null
  let citiesError: string | null = null

  const [categoriesResult, citiesResult] = await Promise.allSettled([
    shouldShowCategory ? fetchCategoriesTree(locale) : Promise.resolve([]),
    shouldShowLocation ? fetchCities() : Promise.resolve([]),
  ])

  if (shouldShowCategory) {
    if (categoriesResult.status === "fulfilled") {
      categories = Array.isArray(categoriesResult.value) ? categoriesResult.value : []
      console.log('Categories loaded:', categories.length, categories)
    } else {
      categoriesError = categoriesResult.reason instanceof Error ? categoriesResult.reason.message : "Failed to load categories"
      categories = [] // Убеждаемся, что это массив
      console.error('Categories error:', categoriesResult.reason)
    }
  }

  if (shouldShowLocation) {
    if (citiesResult.status === "fulfilled") {
      cities = Array.isArray(citiesResult.value) ? citiesResult.value : []
      console.log('Cities loaded:', cities.length, cities)
    } else {
      citiesError = citiesResult.reason instanceof Error ? citiesResult.reason.message : "Failed to load cities"
      cities = [] // Убеждаемся, что это массив
      console.error('Cities error:', citiesResult.reason)
    }
  }

  const flatCategories = shouldShowCategory ? flattenCategories(categories) : []
  if (shouldShowCategory) {
    console.log('Flat categories:', flatCategories.length, flatCategories)
  }

  let data: BrowseResponse
  let error: string | null = null

  try {
    data = await catalogPromise
    console.log('Catalog data loaded:', {
      total: data.total,
      itemsCount: data.items.length,
      page: data.page,
      perPage: data.per_page,
      items: data.items.map(i => ({ name: i.name, shops_count: i.shops_count, shop_names: i.shop_names }))
    })
  } catch (err) {
    error = err instanceof Error ? err.message : t('common.error_unknown')
    console.error('Catalog error:', err)
    data = {
      items: [],
      page: 1,
      per_page: 20,
      total: 0,
      total_pages: 0,
    }
  }

  // Создаём базовые параметры для URL (без page, чтобы пагинация могла менять только страницу)
  const ssrMs = Date.now() - ssrStart
  const warnMs = Number(process.env.CATALOG_SSR_WARN_MS ?? 1500)
  const warnFacets = Number(process.env.CATALOG_FACETS_WARN_COUNT ?? 20)
  const isWarn = ssrMs > warnMs || facetsCount > warnFacets
  const logPayload = {
    event: "catalog_ssr",
    level: isWarn ? "warn" : "info",
    path: "/catalog",
    locale,
    type: productType || "all",
    facets_type: facetsType,
    tenant_id: tenantId,
    facets_count: facetsCount,
    ms: ssrMs,
  }
  if (isWarn) {
    console.warn(JSON.stringify(logPayload))
  } else {
    console.log(JSON.stringify(logPayload))
  }

  const baseParams = new URLSearchParams()
  if (query) baseParams.set("q", query)
  if (category) baseParams.set("category", category)
  if (brand) baseParams.set("brand", brand)
  if (city) baseParams.set("city", city)
  if (minPrice) baseParams.set("min_price", minPrice)
  if (maxPrice) baseParams.set("max_price", maxPrice)
  if (minDuration) baseParams.set("min_duration", minDuration)
  if (maxDuration) baseParams.set("max_duration", maxDuration)
  if (sort && sort !== "price_asc") baseParams.set("sort", sort)
  if (perPage !== "20") baseParams.set("per_page", perPage)

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900 relative">
      {/* Language Switcher - в правом верхнем углу */}
      <div className="absolute top-4 right-4 md:top-6 md:right-6 z-10">
        <LanguageSwitcher />
      </div>

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

        {(categoriesError || citiesError) && (
          <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            {categoriesError && (
              <p className="text-yellow-800 text-sm">⚠️ {categoriesError}</p>
            )}
            {citiesError && (
              <p className="text-yellow-800 text-sm">⚠️ {citiesError}</p>
            )}
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
            <div className="grid grid-cols-1 md:grid-cols-6 gap-4">
              {/* Тип (товар/услуга) */}
              <div>
                <label
                  htmlFor="type"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  {t('catalog.type')}
                </label>
                <TypeSelect
                  id="type"
                  name="type"
                  defaultValue={productType}
                  tenantId={tenantId}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                  options={[
                    { value: "", label: t('catalog.type_all'), className: "text-slate-400" },
                    { value: "good", label: t('catalog.type_goods'), className: "text-slate-900" },
                    { value: "service", label: t('catalog.type_services'), className: "text-slate-900" },
                  ]}
                />
              </div>

              {/* Категория */}
              {facetSet.has("category") && (
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
                  {flatCategories.length > 0 ? (
                    flatCategories.map((cat) => (
                      <option key={cat.value} value={cat.value} className="text-slate-900">
                        {cat.label}
                      </option>
                    ))
                  ) : (
                    <option value="" disabled className="text-slate-400">
                      {categoriesError || 'Loading...'}
                    </option>
                  )}
                </select>
                </div>
              )}

              {/* Город */}
              {facetSet.has("brand") && (
              <div>
                <label
                  htmlFor="brand"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  Brand
                </label>
                <select
                  id="brand"
                  name="brand"
                  defaultValue={brand}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="" className="text-slate-400">All brands</option>
                  {brandOptions.length > 0 ? (
                    brandOptions.map((brandOption) => (
                      <option key={brandOption} value={brandOption} className="text-slate-900">
                        {brandOption}
                      </option>
                    ))
                  ) : (
                    <option value="" disabled className="text-slate-400">
                      Loading...
                    </option>
                  )}
                </select>
                </div>
              )}

              {facetSet.has("location") && (
              <div>
                <label
                  htmlFor="city"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  {t('catalog.city') || 'Город'}
                </label>
                <select
                  id="city"
                  name="city"
                  defaultValue={city}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="" className="text-slate-400">{t('catalog.all_cities') || 'Все города'}</option>
                  {cities.length > 0 ? (
                    cities
                      .filter((c) => c.is_active)
                      .sort((a, b) => a.sort_order - b.sort_order)
                      .map((city) => (
                        <option key={city.id} value={city.slug} className="text-slate-900">
                          {city.name_sr}
                        </option>
                      ))
                  ) : (
                    <option value="" disabled className="text-slate-400">
                      {citiesError || 'Loading...'}
                    </option>
                  )}
                </select>
                </div>
              )}

              {/* Минимальная цена */}
              {productType === "service" && facetSet.has("duration") && (
                <>
                  <div>
                    <label
                      htmlFor="min_duration"
                      className="block text-sm font-medium mb-2 text-slate-900"
                    >
                      Duration from (min)
                    </label>
                    <input
                      type="number"
                      id="min_duration"
                      name="min_duration"
                      defaultValue={minDuration}
                      placeholder="0"
                      min="0"
                      style={{ color: '#0f172a' }}
                      className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-slate-400 text-slate-900"
                    />
                  </div>

                  <div>
                    <label
                      htmlFor="max_duration"
                      className="block text-sm font-medium mb-2 text-slate-900"
                    >
                      Duration to (min)
                    </label>
                    <input
                      type="number"
                      id="max_duration"
                      name="max_duration"
                      defaultValue={maxDuration}
                      placeholder="240"
                      min="0"
                      style={{ color: '#0f172a' }}
                      className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-slate-400 text-slate-900"
                    />
                  </div>
                </>
              )}

              {facetSet.has("price") && (
                <>
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
                </>
              )}
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
              {(query || category || brand || city || productType || minPrice || maxPrice || minDuration || maxDuration || sort !== "price_asc") && (
                <Link
                  href="catalog"
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
                {/* Debug info */}
                <span className="ml-2 text-xs text-slate-400">
                  (Items in array: {data.items.length})
                </span>
              </p>
            </div>

            <ul className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {data.items.map((p) => (
                <ProductCard
                  key={p.id}
                  id={p.id}
                  name={p.name}
                  brand={p.brand}
                  category={p.category}
                  image_url={p.image_url}
                  min_price={p.min_price}
                  max_price={p.max_price}
                  currency={p.currency}
                  shops_count={p.shops_count}
                  shop_names={p.shop_names}
                  type={p.type}
                  service_metadata={p.service_metadata}
                  is_deliverable={p.is_deliverable}
                  is_onsite={p.is_onsite}
                  locale={locale}
                />
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

