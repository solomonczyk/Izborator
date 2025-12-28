import createMiddleware from 'next-intl/middleware'
import { locales } from './i18n'

export default createMiddleware({
  // Поддерживаемые языки
  locales: [...locales] as string[],
  
  // Язык по умолчанию (сербский для сербского рынка)
  defaultLocale: 'sr',
  
  // Префикс локали в URL (например, /sr/catalog)
  localePrefix: 'always'
})

export const config = {
  // Матчинг для всех путей кроме API, статических файлов и изображений
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)']
}

