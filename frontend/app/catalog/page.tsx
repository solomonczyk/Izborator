// app/catalog/page.tsx
import React from "react";

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
};

type BrowseResponse = {
  items: BrowseProduct[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
};

// ⚠️ Подправь URL API под своё окружение (localhost:8080 или nginx /api-прокси)
const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

async function fetchCatalog(query: string): Promise<BrowseResponse> {
  const url = new URL("/api/v1/products/browse", API_BASE);
  if (query) url.searchParams.set("query", query);
  url.searchParams.set("page", "1");
  url.searchParams.set("per_page", "20");

  const res = await fetch(url.toString(), {
    // SSR запрос
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch catalog: ${res.status}`);
  }

  return res.json();
}

export default async function CatalogPage({
  searchParams,
}: {
  searchParams?: Promise<{ q?: string }>;
}) {
  // В Next.js 16 searchParams - это Promise, нужно await
  const params = await searchParams;
  const q = params?.q || "";
  
  let data: BrowseResponse;
  let error: string | null = null;
  
  try {
    data = await fetchCatalog(q);
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

  return (
    <main className="min-h-screen bg-slate-50">
      <div className="max-w-5xl mx-auto px-4 py-8">
        <h1 className="text-2xl font-semibold mb-4">
          Каталог {q ? `— поиск по "${q}"` : ""}
        </h1>
        
        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800 font-medium">Ошибка загрузки каталога:</p>
            <p className="text-red-600 text-sm">{error}</p>
            <p className="text-red-600 text-xs mt-2">
              Убедитесь, что API сервер запущен на {API_BASE}
            </p>
          </div>
        )}

        {/* Поисковая строка (пока без живого обновления, чистый GET через URL) */}
        <form className="mb-6" method="GET">
          <input
            type="text"
            name="q"
            defaultValue={q}
            placeholder="Найти товар…"
            className="w-full border rounded-lg px-3 py-2 shadow-sm"
          />
        </form>

        {data.items.length === 0 ? (
          <p className="text-slate-500">Ничего не найдено.</p>
        ) : (
          <>
            <p className="text-sm text-slate-600 mb-4">
              Найдено: {data.total} товаров
            </p>
            <ul className="grid gap-4 md:grid-cols-2">
              {data.items.map((p) => (
                <li
                  key={p.id}
                  className="bg-white rounded-xl shadow-sm border p-4 flex gap-4"
                >
                  {p.image_url && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={p.image_url}
                      alt={p.name}
                      className="w-24 h-24 object-contain rounded-md border"
                    />
                  )}
                  <div className="flex-1">
                    <h2 className="font-medium text-sm mb-1 line-clamp-2">
                      {p.name}
                    </h2>
                    {p.brand && (
                      <p className="text-xs text-slate-500 mb-1">
                        Бренд: {p.brand}
                      </p>
                    )}
                    {typeof p.min_price === "number" && (
                      <p className="font-semibold text-sm">
                        {p.min_price === p.max_price
                          ? `${p.min_price} ${p.currency || "RSD"}`
                          : `от ${p.min_price} до ${p.max_price} ${
                              p.currency || "RSD"
                            }`}
                      </p>
                    )}
                    {typeof p.shops_count === "number" && (
                      <p className="text-xs text-slate-500">
                        Магазинов: {p.shops_count}
                      </p>
                    )}

                    <a
                      href={`/product/${p.id}`}
                      className="inline-block mt-2 text-xs text-blue-600 hover:underline"
                    >
                      Смотреть цены →
                    </a>
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
    </main>
  );
}

