// app/product/[id]/page.tsx
import React from "react";
import Link from "next/link";
import { PriceChartWrapper } from "@/components/price-chart-wrapper";

type ProductPrice = {
  product_id: string;
  shop_id: string;
  shop_name: string;
  price: number;
  currency: string;
  url: string;
  in_stock: boolean;
  updated_at: string;
};

type ProductResponse = {
  id: string;
  name: string;
  brand?: string;
  category?: string;
  image_url?: string;
  specs?: Record<string, string>;
  prices: ProductPrice[];
};

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:3002";

async function fetchProduct(id: string): Promise<ProductResponse> {
  const url = new URL(`/api/v1/products/${id}`, API_BASE);

  const res = await fetch(url.toString(), {
    next: { revalidate: 300 }, // Кэшируем на 5 минут
  });

  if (!res.ok) {
    if (res.status === 404) {
      throw new Error("Товар не найден");
    }
    throw new Error(`Ошибка загрузки товара: ${res.status}`);
  }

  return res.json();
}

export default async function ProductPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let product: ProductResponse | null = null;
  let error: string | null = null;

  try {
    product = await fetchProduct(id);
  } catch (err) {
    error = err instanceof Error ? err.message : "Неизвестная ошибка";
  }

  if (error || !product) {
    return (
      <main className="min-h-screen bg-slate-50">
        <div className="max-w-5xl mx-auto px-4 py-8">
          <Link
            href="/catalog"
            className="inline-block mb-4 text-blue-600 hover:underline"
          >
            ← Назад к каталогу
          </Link>
          <div className="bg-red-50 border border-red-200 rounded-lg p-6">
            <h1 className="text-2xl font-semibold text-red-800 mb-2">
              Ошибка загрузки товара
            </h1>
            <p className="text-red-600">{error || "Товар не найден"}</p>
          </div>
        </div>
      </main>
    );
  }

  // Сортируем цены: сначала в наличии, потом по цене (от меньшей к большей)
  const sortedPrices = [...product.prices].sort((a, b) => {
    if (a.in_stock !== b.in_stock) {
      return a.in_stock ? -1 : 1;
    }
    return a.price - b.price;
  });

  return (
    <main className="min-h-screen bg-slate-50">
      <div className="max-w-5xl mx-auto px-4 py-8">
        <Link
          href="/catalog"
          className="inline-block mb-6 text-blue-600 hover:underline"
        >
          ← Назад к каталогу
        </Link>

        <div className="bg-white rounded-xl shadow-sm border p-6 mb-6">
          <div className="flex gap-6 flex-col md:flex-row">
            {product.image_url && (
              <div className="flex-shrink-0">
                <img
                  src={product.image_url}
                  alt={product.name}
                  className="w-full md:w-64 h-64 object-contain rounded-lg border bg-white"
                />
              </div>
            )}

            <div className="flex-1">
              <h1 className="text-3xl font-semibold mb-2">{product.name}</h1>

              {product.brand && (
                <p className="text-slate-600 mb-1">
                  <span className="font-medium">Бренд:</span> {product.brand}
                </p>
              )}

              {product.category && (
                <p className="text-slate-600 mb-4">
                  <span className="font-medium">Категория:</span>{" "}
                  {product.category}
                </p>
              )}

              {product.specs && Object.keys(product.specs).length > 0 && (
                <div className="mt-4 pt-4 border-t">
                  <h2 className="font-semibold mb-2">Характеристики:</h2>
                  <dl className="grid grid-cols-1 md:grid-cols-2 gap-2">
                    {Object.entries(product.specs).map(([key, value]) => (
                      <div key={key}>
                        <dt className="text-sm text-slate-500">{key}:</dt>
                        <dd className="text-sm font-medium">{value}</dd>
                      </div>
                    ))}
                  </dl>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm border p-6">
          <h2 className="text-2xl font-semibold mb-4">
            Цены в магазинах ({product.prices.length})
          </h2>

          {sortedPrices.length === 0 ? (
            <p className="text-slate-500">Цены пока не доступны</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-4 font-semibold">Магазин</th>
                    <th className="text-right py-2 px-4 font-semibold">Цена</th>
                    <th className="text-center py-2 px-4 font-semibold">
                      Наличие
                    </th>
                    <th className="text-center py-2 px-4 font-semibold">
                      Обновлено
                    </th>
                    <th className="text-center py-2 px-4 font-semibold">
                      Действие
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {sortedPrices.map((price, idx) => (
                    <tr
                      key={`${price.shop_id}-${idx}`}
                      className="border-b hover:bg-slate-50"
                    >
                      <td className="py-3 px-4">{price.shop_name}</td>
                      <td className="py-3 px-4 text-right font-semibold">
                        {price.price.toLocaleString("ru-RU")} {price.currency}
                      </td>
                      <td className="py-3 px-4 text-center">
                        {price.in_stock ? (
                          <span className="text-green-600 font-medium">
                            В наличии
                          </span>
                        ) : (
                          <span className="text-red-600 font-medium">
                            Нет в наличии
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-center text-sm text-slate-500">
                        {new Date(price.updated_at).toLocaleDateString("ru-RU", {
                          day: "2-digit",
                          month: "2-digit",
                          year: "numeric",
                        })}
                      </td>
                      <td className="py-3 px-4 text-center">
                        {price.url && (
                          <a
                            href={price.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:underline text-sm"
                          >
                            Перейти →
                          </a>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <PriceChartWrapper productId={id} />
      </div>
    </main>
  );
}

