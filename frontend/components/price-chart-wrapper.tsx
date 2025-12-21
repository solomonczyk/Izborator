'use client'

import { useEffect, useState } from 'react'
import { useTranslations } from 'next-intl'
import { PriceChart } from './price-chart'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8081'

interface PricePoint {
  price: number
  timestamp: string
  shop_id: string
}

interface PriceChartData {
  product_id: string
  period: string
  from: string
  to: string
  shops: Record<string, PricePoint[]>
  shop_names?: Record<string, string>
  stats: {
    min_price: number
    max_price: number
    avg_price: number
    price_change: number
    first_price: number
    last_price: number
    first_date: string
    last_date: string
  }
}

interface PriceChartWrapperProps {
  productId: string
  locale?: string
}

export function PriceChartWrapper({ productId, locale = 'en' }: PriceChartWrapperProps) {
  const t = useTranslations('product')
  const [data, setData] = useState<PriceChartData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [period, setPeriod] = useState<string>('month')

  useEffect(() => {
    async function fetchPriceHistory() {
      setLoading(true)
      setError(null)
      
      try {
        const url = new URL(`/api/v1/products/${productId}/price-history`, API_BASE)
        url.searchParams.set('period', period)
        url.searchParams.set('lang', locale)

        const res = await fetch(url.toString(), {
          next: { revalidate: 900 }, // Кэшируем на 15 минут (история меняется редко)
        })

        if (!res.ok) {
          throw new Error(t('price_history_error'))
        }

        const result = await res.json()
        setData(result)
      } catch (err) {
        setError(err instanceof Error ? err.message : t('error_load_failed'))
      } finally {
        setLoading(false)
      }
    }

    fetchPriceHistory()
  }, [productId, period, locale, t])

  const periods = [
    { value: 'day', label: t('price_history_period_day') },
    { value: 'week', label: t('price_history_period_week') },
    { value: 'month', label: t('price_history_period_month') },
    { value: 'year', label: t('price_history_period_year') },
  ]

  return (
    <div className="mt-8">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-semibold text-slate-900">{t('price_history')}</h2>
        <div className="flex gap-2">
          {periods.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className={`px-4 py-2 rounded-lg border-2 text-sm font-medium transition-colors ${
                period === p.value
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-slate-700 border-slate-300 hover:border-slate-400'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {loading && (
        <div className="bg-white p-6 rounded-lg border-2 border-slate-200">
          <div className="text-center py-12 text-slate-600">{t('price_history_loading')}</div>
        </div>
      )}

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {!loading && !error && (!data || !data.shops || Object.keys(data.shops).length === 0) && (
        <div className="bg-white p-6 rounded-lg border-2 border-slate-200">
          <div className="text-center py-12 text-slate-600">
            {t('price_history_no_data')}
          </div>
        </div>
      )}

      {!loading && !error && data && data.shops && Object.keys(data.shops).length > 0 && (
        <PriceChart data={data} locale={locale} />
      )}
    </div>
  )
}

