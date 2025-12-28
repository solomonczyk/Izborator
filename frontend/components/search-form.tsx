// components/search-form.tsx
// Клиентский компонент поисковой строки
'use client'

import React from 'react'

export function SearchForm({ locale }: { locale: string }) {
  return (
    <form 
      method="GET" 
      action={`/${locale}/catalog`}
      className="w-full max-w-3xl mx-auto"
    >
      <div className="relative">
        <input
          type="text"
          name="q"
          placeholder={locale === 'sr' 
            ? 'Претражи производе и услуге...' 
            : 'Search for products and services...'}
          className="w-full px-6 py-5 text-lg border-2 border-slate-300 rounded-2xl shadow-lg focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-200 transition-all"
          autoFocus
        />
        <button
          type="submit"
          className="absolute right-2 top-1/2 -translate-y-1/2 px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 active:bg-blue-800 transition-colors font-medium shadow-md"
        >
          {locale === 'sr' ? 'Претрага' : 'Search'}
        </button>
      </div>
    </form>
  )
}

