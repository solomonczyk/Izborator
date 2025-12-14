import createMiddleware from 'next-intl/middleware'
import { locales } from './i18n'

export default createMiddleware({
  // Поддерживаемые языки
  locales: [...locales] as string[],
  
  // Язык по умолчанию
  defaultLocale: 'en',
  
  // Префикс локали в URL (например, /en/catalog)
  localePrefix: 'always'
})

export const config = {
  // Матчинг для всех путей кроме API, статических файлов и изображений
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)']
}

