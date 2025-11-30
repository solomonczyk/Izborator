'use client'

import { useTranslations } from 'next-intl'
import { Link } from '@/navigation'

export function Pagination({
  currentPage,
  totalPages,
  baseParams,
  locale,
}: {
  currentPage: number
  totalPages: number
  baseParams: URLSearchParams
  locale: string
}) {
  const t = useTranslations('catalog')
  
  if (totalPages <= 1) return null

  const pages: number[] = []
  const showPages = 5 // Показываем максимум 5 страниц

  let startPage = Math.max(1, currentPage - Math.floor(showPages / 2))
  let endPage = Math.min(totalPages, startPage + showPages - 1)

  if (endPage - startPage < showPages - 1) {
    startPage = Math.max(1, endPage - showPages + 1)
  }

  for (let i = startPage; i <= endPage; i++) {
    pages.push(i)
  }

  const createPageUrl = (page: number): string => {
    const params = new URLSearchParams(baseParams.toString())
    params.set("page", page.toString())
    return `/${locale}/catalog?${params.toString()}`
  }

  return (
    <nav className="flex items-center justify-center gap-2 mt-8">
      {currentPage > 1 && (
        <Link
          href={createPageUrl(currentPage - 1)}
          className="px-3 py-2 border-2 border-slate-400 rounded-lg hover:bg-slate-100 hover:border-slate-500"
        >
          ← {t('prev')}
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
          {t('next')} →
        </Link>
      )}
    </nav>
  )
}

