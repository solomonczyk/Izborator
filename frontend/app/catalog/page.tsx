// app/catalog/page.tsx
import React from "react";
import Link from "next/link";
import { ProductCard } from '@/components/product-card'

type BrowseProduct = {
  id: string;
  name: string;
  brand?: string;
  category?: string;
  image_url?: string;
  min_price?: number;
  max_price?: number;
  currency?: string;
  shops_count?: number;
  shop_names?: string[];
  type?: 'good' | 'service';
  service_metadata?: {
    duration?: string;
    master_name?: string;
    service_area?: string;
  };
  is_deliverable?: boolean;
  is_onsite?: boolean;
};

type BrowseResponse = {
  items: BrowseProduct[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
};

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8081";

async function fetchCatalog(params: {
  query?: string;
  category?: string;
  minPrice?: string;
  maxPrice?: string;
  page?: string;
  perPage?: string;
  sort?: string;
}): Promise<BrowseResponse> {
  const url = new URL("/api/v1/products/browse", API_BASE);

  if (params.query) url.searchParams.set("query", params.query);
  if (params.category) url.searchParams.set("category", params.category);
  if (params.minPrice) url.searchParams.set("min_price", params.minPrice);
  if (params.maxPrice) url.searchParams.set("max_price", params.maxPrice);
  if (params.sort) url.searchParams.set("sort", params.sort);

  const page = params.page ? parseInt(params.page, 10) : 1;
  const perPage = params.perPage ? parseInt(params.perPage, 10) : 20;

  url.searchParams.set("page", page.toString());
  url.searchParams.set("per_page", perPage.toString());

  const res = await fetch(url.toString(), {
    next: { revalidate: 180 }, // Кэшируем на 3 минуты
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch catalog: ${res.status}`);
  }

  return res.json();
}

// Компонент для пагинации
function Pagination({
  currentPage,
  totalPages,
  baseParams,
}: {
  currentPage: number;
  totalPages: number;
  baseParams: URLSearchParams;
}) {
  if (totalPages <= 1) return null;

  const pages: number[] = [];
  const showPages = 5; // Показываем максимум 5 страниц

  let startPage = Math.max(1, currentPage - Math.floor(showPages / 2));
  const endPage = Math.min(totalPages, startPage + showPages - 1);

  if (endPage - startPage < showPages - 1) {
    startPage = Math.max(1, endPage - showPages + 1);
  }

  for (let i = startPage; i <= endPage; i++) {
    pages.push(i);
  }

  const createPageUrl = (page: number): string => {
    const params = new URLSearchParams(baseParams.toString());
    params.set("page", page.toString());
    return `/catalog?${params.toString()}`;
  };

  return (
    <nav className="flex items-center justify-center gap-2 mt-8">
      {currentPage > 1 && (
        <Link
          href={createPageUrl(currentPage - 1)}
          className="px-3 py-2 border-2 border-slate-400 rounded-lg hover:bg-slate-100 hover:border-slate-500"
        >
          ← Предыдущая
        </Link>
      )}

      {startPage > 1 && (
        <>
          <Link
            href={createPageUrl(1)}
            className="px-3 py-2 border-2 border-slate-400 rounded-lg hover:bg-slate-100 hover:border-slate-500"
          >
            1
          </Link>
          {startPage > 2 && <span className="px-2">...</span>}
        </>
      )}

      {pages.map((page) => (
        <Link
          key={page}
          href={createPageUrl(page)}
          className={`px-3 py-2 border rounded-lg ${
            page === currentPage
              ? "bg-blue-600 text-white border-blue-600 border-2"
              : "hover:bg-slate-100 border-2 border-slate-400"
          }`}
        >
          {page}
        </Link>
      ))}

      {endPage < totalPages && (
        <>
          {endPage < totalPages - 1 && <span className="px-2">...</span>}
          <Link
            href={createPageUrl(totalPages)}
            className="px-3 py-2 border-2 border-slate-400 rounded-lg hover:bg-slate-100 hover:border-slate-500"
          >
            {totalPages}
          </Link>
        </>
      )}

      {currentPage < totalPages && (
        <Link
          href={createPageUrl(currentPage + 1)}
          className="px-3 py-2 border-2 border-slate-400 rounded-lg hover:bg-slate-100 hover:border-slate-500"
        >
          Следующая →
        </Link>
      )}
    </nav>
  );
}

export default async function CatalogPage({
  searchParams,
}: {
  searchParams?: Promise<{
    q?: string;
    category?: string;
    min_price?: string;
    max_price?: string;
    page?: string;
    per_page?: string;
    sort?: string;
  }>;
}) {
  const params = await searchParams;
  const query = params?.q || "";
  const category = params?.category || "";
  const minPrice = params?.min_price || "";
  const maxPrice = params?.max_price || "";
  const page = params?.page || "1";
  const perPage = params?.per_page || "20";
  const sort = params?.sort || "price_asc";

  let data: BrowseResponse;
  let error: string | null = null;

  try {
    data = await fetchCatalog({
      query,
      category,
      minPrice,
      maxPrice,
      page,
      perPage,
      sort,
    });
  } catch (err) {
    error = err instanceof Error ? err.message : "Неизвестная ошибка";
    data = {
      items: [],
      page: 1,
      per_page: 20,
      total: 0,
      total_pages: 0,
    };
  }

  // Создаём базовые параметры для URL (без page, чтобы пагинация могла менять только страницу)
  const baseParams = new URLSearchParams();
  if (query) baseParams.set("q", query);
  if (category) baseParams.set("category", category);
  if (minPrice) baseParams.set("min_price", minPrice);
  if (maxPrice) baseParams.set("max_price", maxPrice);
  if (sort && sort !== "price_asc") baseParams.set("sort", sort);
  if (perPage !== "20") baseParams.set("per_page", perPage);

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-semibold mb-6 text-slate-900">
          Каталог {query ? `— поиск по "${query}"` : ""}
        </h1>

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800 font-medium">Ошибка загрузки каталога:</p>
            <p className="text-red-600 text-sm">{error}</p>
            <p className="text-red-600 text-xs mt-2">
              Убедитесь, что API сервер запущен на {API_BASE}
            </p>
          </div>
        )}

        {/* Фильтры и поиск */}
        <div className="bg-white rounded-xl shadow-sm border-2 border-slate-300 p-6 mb-6">
          <form method="GET" action="/catalog" className="space-y-4">
            {/* Поисковая строка */}
            <div>
              <label htmlFor="search" className="block text-sm font-medium mb-2 text-slate-900">
                Поиск
              </label>
              <input
                type="text"
                id="search"
                name="q"
                defaultValue={query}
                placeholder="Найти товар…"
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
                  Категория
                </label>
                <select
                  id="category"
                  name="category"
                  defaultValue={category}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="" className="text-slate-400">Все категории</option>
                  <option value="phones" className="text-slate-900">Телефоны</option>
                  <option value="laptops" className="text-slate-900">Ноутбуки</option>
                  <option value="tablets" className="text-slate-900">Планшеты</option>
                  <option value="accessories" className="text-slate-900">Аксессуары</option>
                </select>
              </div>

              {/* Минимальная цена */}
              <div>
                <label
                  htmlFor="min_price"
                  className="block text-sm font-medium mb-2 text-slate-900"
                >
                  Цена от (RSD)
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
                  Цена до (RSD)
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
                  Сортировка
                </label>
                <select
                  id="sort"
                  name="sort"
                  defaultValue={sort}
                  className="w-full border-2 border-slate-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-slate-900 bg-white"
                >
                  <option value="price_asc" className="text-slate-900">Цена: по возрастанию</option>
                  <option value="price_desc" className="text-slate-900">Цена: по убыванию</option>
                  <option value="name_asc" className="text-slate-900">Название: A-Z</option>
                  <option value="name_desc" className="text-slate-900">Название: Z-A</option>
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
                Применить фильтры
              </button>
              {(query || category || minPrice || maxPrice || sort !== "price_asc") && (
                <Link
                  href="/catalog"
                  className="px-4 py-2 bg-slate-200 text-slate-700 rounded-lg hover:bg-slate-300 focus:outline-none focus:ring-2 focus:ring-slate-500"
                >
                  Сбросить
                </Link>
              )}
            </div>
          </form>
        </div>

        {/* Результаты */}
        {data.items.length === 0 ? (
          <div className="bg-white rounded-xl shadow-sm border-2 border-slate-300 p-8 text-center">
            <p className="text-slate-500 text-lg">Ничего не найдено.</p>
            <p className="text-slate-400 text-sm mt-2">
              Попробуйте изменить параметры поиска или фильтры.
            </p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between mb-4">
              <p className="text-sm text-slate-700">
                Найдено: <span className="font-semibold text-slate-900">{data.total}</span>{" "}
                товаров
                {data.total_pages > 1 && (
                  <span className="ml-2">
                    (страница {data.page} из {data.total_pages})
                  </span>
                )}
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
                  locale="sr"
                />
              ))}
            </ul>

            {/* Пагинация */}
            <Pagination
              currentPage={data.page}
              totalPages={data.total_pages}
              baseParams={baseParams}
            />
          </>
        )}
      </div>
    </main>
  );
}
