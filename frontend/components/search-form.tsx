// components/search-form.tsx
// Клиентский компонент поисковой строки
'use client'

import React from 'react'
import { useTranslations } from 'next-intl'
import { CitySelect } from '@/components/city-select'
import { TypeToggle } from '@/components/type-toggle'

type SearchFormProps = {
  showTypeToggle?: boolean
  defaultType?: "all" | "good" | "service"
  showCitySelect?: boolean
  cityOptions?: Array<{ value: string; label: string }>
  defaultCity?: string
  searchPlaceholder?: string
}

export function SearchForm({
  showTypeToggle = false,
  defaultType = "all",
  showCitySelect = false,
  cityOptions = [],
  defaultCity = "",
  searchPlaceholder,
}: SearchFormProps) {
  const t = useTranslations('home')
  const tCatalog = useTranslations('catalog')

  return (
    <form 
      method="GET" 
      action={`/catalog`}
      className="w-full max-w-3xl mx-auto"
    >
      <div className="relative">
        <input
          type="text"
          name="q"
          placeholder={searchPlaceholder ?? t('search_placeholder')}
          className="w-full px-6 py-5 text-lg border-2 border-slate-300 rounded-2xl shadow-lg focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition-all"
          autoFocus
        />
        <button
          type="submit"
          className="absolute right-2 top-1/2 -translate-y-1/2 px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 active:bg-blue-800 transition-colors font-medium shadow-md"
        >
          {t('search_button')}
        </button>
      </div>
      {showTypeToggle || showCitySelect ? (
        <div className="mt-4 flex flex-wrap items-center justify-center gap-3">
          {showTypeToggle ? (
            <TypeToggle
              ariaLabel={tCatalog('type')}
              defaultValue={defaultType}
              labels={{
                all: tCatalog('type_all'),
                goods: tCatalog('type_goods'),
                services: tCatalog('type_services'),
              }}
            />
          ) : null}
          {showCitySelect ? (
            <CitySelect
              label={tCatalog('city')}
              allLabel={tCatalog('all_cities')}
              options={cityOptions}
              defaultValue={defaultCity}
            />
          ) : null}
        </div>
      ) : null}
    </form>
  )
}

