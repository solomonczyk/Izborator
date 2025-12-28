'use client'

import { useLocale } from 'next-intl'
import { usePathname } from '@/navigation'
import { Link } from '@/navigation'
import { locales, type Locale } from '@/i18n'
import { useState, useRef, useEffect } from 'react'

// Названия языков для отображения
const languageNames: Record<Locale, string> = {
  sr: 'Српски',
  en: 'English',
  ru: 'Русский',
  hu: 'Magyar',
  zh: '中文',
}

// Флаги/иконки для языков
const languageFlags: Record<Locale, string> = {
  sr: '🇷🇸',
  en: '🇬🇧',
  ru: '🇷🇺',
  hu: '🇭🇺',
  zh: '🇨🇳',
}

export function LanguageSwitcher() {
  const locale = useLocale() as Locale
  const pathname = usePathname()
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Закрываем dropdown при клике вне его
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-4 py-2 bg-white border-2 border-slate-300 rounded-lg hover:border-blue-500 hover:shadow-md transition-all"
        aria-label="Switch language"
        aria-expanded={isOpen}
      >
        <span className="text-xl">{languageFlags[locale]}</span>
        <span className="text-sm font-medium text-slate-700">{languageNames[locale]}</span>
        <svg
          className={`w-4 h-4 text-slate-500 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute top-full right-0 mt-2 bg-white border-2 border-slate-300 rounded-lg shadow-lg z-50 min-w-[160px]">
          {locales.map((loc) => (
            <Link
              key={loc}
              href={pathname}
              locale={loc}
              onClick={() => setIsOpen(false)}
              className={`w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-slate-50 transition-colors ${
                loc === locale ? 'bg-blue-50 text-blue-600' : 'text-slate-700'
              } ${loc === locales[0] ? 'rounded-t-lg' : ''} ${loc === locales[locales.length - 1] ? 'rounded-b-lg' : ''}`}
            >
              <span className="text-xl">{languageFlags[loc]}</span>
              <span className="text-sm font-medium">{languageNames[loc]}</span>
              {loc === locale && (
                <svg className="w-4 h-4 ml-auto text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

