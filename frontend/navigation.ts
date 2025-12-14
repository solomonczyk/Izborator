import { createNavigation } from 'next-intl/navigation'
import { locales } from './i18n'

// Единая точка для навигации с учётом локалей
export const { Link, redirect, usePathname, useRouter } = createNavigation({
  locales: [...locales] as string[],
  localePrefix: 'always'
})

