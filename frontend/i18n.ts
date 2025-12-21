import { getRequestConfig } from 'next-intl/server'

// Поддерживаемые языки
export const locales = ['en', 'sr', 'ru', 'hu', 'zh'] as const
export type Locale = (typeof locales)[number]

export default getRequestConfig(async ({ locale }) => {
  let resolvedLocale = locale as Locale

  // Проверяем, что язык поддерживается
  if (!locales.includes(resolvedLocale)) {
    // Если локаль не поддерживается, используем английский
    resolvedLocale = 'en'
  }

  return {
    locale: resolvedLocale,
    messages: (await import(`./messages/${resolvedLocale}.json`)).default
  }
})
