// components/product-card.tsx
// Универсальная плитка для отображения товаров и услуг

import React from 'react'
import Image from 'next/image'
import Link from 'next/link'

export interface ServiceMetadata {
  duration?: string
  master_name?: string
  service_area?: string
}

export interface ProductCardProps {
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
  service_metadata?: ServiceMetadata
  is_deliverable?: boolean
  is_onsite?: boolean
  locale?: string
}

// Форматирование цены с разделителями тысяч
function formatPrice(price: number): string {
  return price.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
}

// Иконки для категорий услуг (расширенный список)
function getServiceIcon(category?: string): string {
  if (!category) return '🛠️'
  
  const categoryLower = category.toLowerCase()
  
  // Красота и здоровье
  if (categoryLower.includes('frizerski') || categoryLower.includes('frizerska') || categoryLower.includes('šišanje')) return '✂️'
  if (categoryLower.includes('kozmetički') || categoryLower.includes('kozmetika') || categoryLower.includes('manikir')) return '💅'
  if (categoryLower.includes('masaža') || categoryLower.includes('masaž') || categoryLower.includes('masaža')) return '💆'
  if (categoryLower.includes('zubarska') || categoryLower.includes('stomatolog') || categoryLower.includes('dentist')) return '🦷'
  if (categoryLower.includes('teretana') || categoryLower.includes('fitness') || categoryLower.includes('gym')) return '💪'
  if (categoryLower.includes('joga') || categoryLower.includes('yoga')) return '🧘'
  
  // Еда и напитки
  if (categoryLower.includes('restoran') || categoryLower.includes('kafe') || categoryLower.includes('cafe')) return '🍽️'
  if (categoryLower.includes('pizza') || categoryLower.includes('pica')) return '🍕'
  if (categoryLower.includes('kafić') || categoryLower.includes('bar')) return '☕'
  
  // Размещение
  if (categoryLower.includes('hotel') || categoryLower.includes('apartman') || categoryLower.includes('smestaj')) return '🏨'
  if (categoryLower.includes('prenočište') || categoryLower.includes('hostel')) return '🛏️'
  
  // Транспорт
  if (categoryLower.includes('prevoz') || categoryLower.includes('transport') || categoryLower.includes('taxi')) return '🚗'
  if (categoryLower.includes('selidbe') || categoryLower.includes('selidba')) return '📦'
  
  // Ремонт и обслуживание
  if (categoryLower.includes('popravka') || categoryLower.includes('servis') || categoryLower.includes('auto servis')) return '🔧'
  if (categoryLower.includes('majstor') || categoryLower.includes('električar') || categoryLower.includes('vodoinstalater')) return '🔨'
  if (categoryLower.includes('čišćenje') || categoryLower.includes('cleaning')) return '🧹'
  
  // Образование
  if (categoryLower.includes('obuka') || categoryLower.includes('kursevi') || categoryLower.includes('edukacija')) return '📚'
  if (categoryLower.includes('jezik') || categoryLower.includes('language')) return '🌐'
  
  // Фото и видео
  if (categoryLower.includes('fotografisanje') || categoryLower.includes('fotograf') || categoryLower.includes('photo')) return '📸'
  if (categoryLower.includes('video') || categoryLower.includes('produkcija')) return '🎥'
  
  // Недвижимость
  if (categoryLower.includes('nekretnine') || categoryLower.includes('real estate') || categoryLower.includes('stan')) return '🏠'
  if (categoryLower.includes('arhitekta') || categoryLower.includes('projektovanje')) return '🏗️'
  
  // Юридические и финансовые
  if (categoryLower.includes('advokat') || categoryLower.includes('lawyer')) return '⚖️'
  if (categoryLower.includes('knjigovodja') || categoryLower.includes('računovodstvo')) return '📊'
  
  // Развлечения
  if (categoryLower.includes('zabava') || categoryLower.includes('event') || categoryLower.includes('proslava')) return '🎉'
  
  return '🛠️'
}

export function ProductCard({
  id,
  name,
  brand,
  category,
  image_url,
  min_price,
  max_price,
  currency = 'RSD',
  shops_count,
  shop_names = [],
  type = 'good',
  service_metadata,
  is_deliverable = true,
  is_onsite = false,
  locale = 'sr',
}: ProductCardProps) {
  const isService = type === 'service'
  const hasPrice = typeof min_price === 'number'
  
  // Форматирование цены
  const priceDisplay = hasPrice
    ? isService
      ? // Для услуг всегда показываем "от..."
        `${locale === 'sr' ? 'од' : 'from'} ${formatPrice(min_price!)} ${currency}`
      : // Для товаров: точная цена или диапазон
        min_price === max_price || !max_price
        ? `${formatPrice(min_price!)} ${currency}`
        : `${formatPrice(min_price!)} - ${formatPrice(max_price)} ${currency}`
    : null

  return (
    <li className={`bg-white rounded-xl shadow-sm border-2 p-4 hover:shadow-md transition-all ${
      isService 
        ? 'border-indigo-300 hover:border-indigo-400' 
        : 'border-slate-300 hover:border-blue-400'
    }`}>
      <div className="flex flex-col gap-3">
        {/* Основной контент - кликабельный */}
        <Link href={`/product/${id}`} className="flex gap-4">
          {/* Изображение или иконка */}
          <div className="flex-shrink-0">
          {isService ? (
            // Для услуг: иконка категории или фото салона
            <div className="w-24 h-24 rounded-lg border-2 border-indigo-200 bg-gradient-to-br from-indigo-50 to-purple-50 flex items-center justify-center shadow-sm">
              {image_url ? (
                <Image
                  src={image_url}
                  alt={name}
                  width={96}
                  height={96}
                  className="w-full h-full object-cover rounded-lg"
                  unoptimized
                />
              ) : (
                <span className="text-5xl">{getServiceIcon(category)}</span>
              )}
            </div>
          ) : (
            // Для товаров: фото товара
            image_url ? (
              <Image
                src={image_url}
                alt={name}
                width={96}
                height={96}
                className="w-24 h-24 object-contain rounded-lg border-2 border-slate-300 bg-white shadow-sm"
                unoptimized
              />
            ) : (
              <div className="w-24 h-24 rounded-lg border-2 border-slate-300 bg-slate-100 flex items-center justify-center">
                <span className="text-slate-400 text-xs text-center px-1">Нема слике</span>
              </div>
            )
          )}
          </div>

          {/* Контент */}
          <div className="flex-1 min-w-0">
          <h2 className="font-medium text-sm mb-1 line-clamp-2 hover:text-blue-600 text-slate-900">
            {name}
          </h2>

          {/* Бренд (только для товаров) */}
          {!isService && brand && (
            <p className="text-xs text-slate-600 mb-1">{brand}</p>
          )}

          {/* Метаданные услуги */}
          {isService && service_metadata && (
            <div className="text-xs text-indigo-700 mb-1 space-y-0.5">
              {service_metadata.duration && (
                <p className="flex items-center gap-1">
                  <span>⏱️</span>
                  <span>{service_metadata.duration}</span>
                </p>
              )}
              {service_metadata.master_name && (
                <p className="flex items-center gap-1">
                  <span>👤</span>
                  <span>{service_metadata.master_name}</span>
                </p>
              )}
              {service_metadata.service_area && (
                <p className="flex items-center gap-1">
                  <span>📍</span>
                  <span>{service_metadata.service_area}</span>
                </p>
              )}
            </div>
          )}

          {/* Цена */}
          {priceDisplay && (
            <p className={`font-semibold text-base mt-2 ${
              isService ? 'text-indigo-700' : 'text-slate-900'
            }`}>
              {priceDisplay}
            </p>
          )}

          {/* Информация о магазинах/провайдерах */}
          {typeof shops_count === 'number' && shops_count > 0 && (
            <p className="text-xs text-slate-600 mt-1">
              {isService
                ? `${shops_count} ${shops_count === 1 ? 'провајдер' : 'провајдера'}`
                : `${shops_count} ${shops_count === 1 ? 'продавница' : 'продавница'}`}
            </p>
          )}

          {/* Бейджи */}
          <div className="flex flex-wrap gap-1.5 mt-2">
            {isService ? (
              // Бейджи для услуг
              <>
                {is_onsite && (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-800 border border-emerald-200">
                    <span className="mr-1">🚗</span>
                    <span>{locale === 'sr' ? 'Вози до вас' : 'Onsite'}</span>
                  </span>
                )}
                {service_metadata?.service_area && (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 border border-indigo-200">
                    <span className="mr-1">📍</span>
                    <span>{service_metadata.service_area}</span>
                  </span>
                )}
              </>
            ) : (
              // Бейджи для товаров
              <>
                {is_deliverable && (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
                    <span className="mr-1">🚚</span>
                    <span>{locale === 'sr' ? 'Достава' : 'Delivery'}</span>
                  </span>
                )}
              </>
            )}
          </div>
          </div>
        </Link>

        {/* Кнопки действий - отдельно от основного контента */}
        <div className="mt-1">
          {isService ? (
            // Кнопка "Записаться" для услуг
            <Link
              href={`/product/${id}`}
              className="inline-flex items-center justify-center w-full px-4 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 active:bg-indigo-800 transition-colors shadow-sm"
            >
              <span className="mr-2">📅</span>
              <span>{locale === 'sr' ? 'Записати се' : 'Book Appointment'}</span>
            </Link>
          ) : (
            // Кнопка "В магазин" для товаров
            <Link
              href={`/product/${id}`}
              className="inline-flex items-center justify-center w-full px-4 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors shadow-sm"
            >
              <span className="mr-2">🛒</span>
              <span>{locale === 'sr' ? 'У продавницу' : 'Go to Shop'}</span>
            </Link>
          )}
        </div>
      </div>
    </li>
  )
}

