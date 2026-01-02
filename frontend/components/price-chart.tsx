'use client'

import { useTranslations } from 'next-intl'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

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

interface PriceChartProps {
  data: PriceChartData
  locale?: string
}

interface ChartDataPoint {
  date: string
  [shopId: string]: string | number
}

export function PriceChart({ data, locale = 'en' }: PriceChartProps) {
  const t = useTranslations('product')
  // Преобразуем данные для графика
  // Группируем по датам и собираем цены всех магазинов
  const chartData: ChartDataPoint[] = []
  const dateMap = new Map<string, ChartDataPoint>()

  // Определяем локаль для форматирования дат
  const dateLocale = locale === 'sr' ? 'sr-RS' : locale === 'ru' ? 'ru-RU' : locale === 'hu' ? 'hu-HU' : locale === 'zh' ? 'zh-CN' : 'en-US'
  
  // Проходим по всем магазинам и их точкам
  Object.entries(data.shops).forEach(([shopId, points]) => {
    points.forEach((point) => {
      const date = new Date(point.timestamp).toLocaleDateString(dateLocale, {
        month: 'short',
        day: 'numeric',
      })

      if (!dateMap.has(date)) {
        dateMap.set(date, { date })
      }

      const dayData = dateMap.get(date)!
      dayData[shopId] = point.price
    })
  })

  // Преобразуем в массив и сортируем по дате
  const sortedData = Array.from(dateMap.values())
  sortedData.sort((a, b) => {
    // Создаём временные даты для сравнения из строки даты
    const dateA = new Date(a.date)
    const dateB = new Date(b.date)
    return dateA.getTime() - dateB.getTime()
  })
  chartData.push(...sortedData)

  // Цвета для линий магазинов
  const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

  const shopIds = Object.keys(data.shops)
  const shopNames = shopIds.map((id) => {
    // Используем shop_name из данных, если есть, иначе используем shop_id
    return data.shop_names?.[id] || `${t('shop')} ${id.slice(0, 8)}`
  })

  return (
    <div>

      {/* Статистика */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white p-4 rounded-lg border-2 border-slate-200">
          <div className="text-sm text-slate-600 mb-1">{t('price_history_min')}</div>
          <div className="text-xl font-semibold text-slate-900">
            {data.stats.min_price.toLocaleString(dateLocale)} RSD
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg border-2 border-slate-200">
          <div className="text-sm text-slate-600 mb-1">{t('price_history_max')}</div>
          <div className="text-xl font-semibold text-slate-900">
            {data.stats.max_price.toLocaleString(dateLocale)} RSD
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg border-2 border-slate-200">
          <div className="text-sm text-slate-600 mb-1">{t('price_history_avg')}</div>
          <div className="text-xl font-semibold text-slate-900">
            {Math.round(data.stats.avg_price).toLocaleString(dateLocale)} RSD
          </div>
        </div>
        <div className="bg-white p-4 rounded-lg border-2 border-slate-200">
          <div className="text-sm text-slate-600 mb-1">{t('price_history_change')}</div>
          <div
            className={`text-xl font-semibold ${
              data.stats.price_change >= 0 ? 'text-red-600' : 'text-green-600'
            }`}
          >
            {data.stats.price_change >= 0 ? '+' : ''}
            {data.stats.price_change.toFixed(1)}%
          </div>
        </div>
      </div>

      {/* График */}
      <div className="bg-white p-6 rounded-lg border-2 border-slate-200">
        {chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis
                dataKey="date"
                stroke="#64748b"
                style={{ fontSize: '12px' }}
              />
              <YAxis
                stroke="#64748b"
                style={{ fontSize: '12px' }}
                tickFormatter={(value) => `${value.toLocaleString(dateLocale)}`}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#fff',
                  border: '1px solid #e2e8f0',
                  borderRadius: '8px',
                }}
                formatter={(value?: number) => {
                  if (typeof value !== 'number') {
                    return ['', t('price')]
                  }
                  return [`${value.toLocaleString(dateLocale)} RSD`, t('price')]
                }}
              />
              <Legend />
              {shopIds.map((shopId, index) => (
                <Line
                  key={shopId}
                  type="monotone"
                  dataKey={shopId}
                  stroke={colors[index % colors.length]}
                  strokeWidth={2}
                  name={shopNames[index]}
                  dot={{ r: 4 }}
                  activeDot={{ r: 6 }}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <div className="text-center py-12 text-slate-600">
            {t('price_history_no_data')}
          </div>
        )}
      </div>
    </div>
  )
}

