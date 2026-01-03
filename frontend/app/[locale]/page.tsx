import { getTranslations } from 'next-intl/server'
import type { CategoryCardProps } from '@/components/category-card'
import { FloatingCategoryCloud } from '@/components/floating-category-cloud'
import { HeroSearch } from '@/components/hero-search'
import { LanguageSwitcher } from '@/components/language-switcher'
import { fetchCities, fetchHomeModel } from '@/lib/api'

export const dynamic = 'force-dynamic'

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'home' })
  const tenantId = process.env.NEXT_PUBLIC_TENANT_ID || process.env.TENANT_ID || 'default'
  const homeModel = await fetchHomeModel({ tenantId, locale })
  const isLoading = !homeModel
  const hero = homeModel?.hero ?? {
    title: t('title'),
    subtitle: t('subtitle'),
    searchPlaceholder: t('search_placeholder'),
    showTypeToggle: true,
    showCitySelect: false,
    defaultType: 'all' as const,
  }
  const cityOptions = hero.showCitySelect
    ? (await fetchCities()).map((city) => ({
        value: city.slug,
        label: city.name_sr,
      }))
    : []
  const categoryCards: CategoryCardProps[] =
    homeModel?.categoryCards.map((card) => ({
      id: card.id,
      title: card.title,
      hint: card.hint,
      href: card.href,
      priority: card.priority,
      weight: card.weight,
      analyticsId: card.analytics_id,
    })) ?? []

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 to-white">
      <a
        href="#home-search"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 rounded-lg bg-white px-4 py-2 text-sm font-medium text-slate-900 shadow"
      >
        Skip to search
      </a>
      <header className="absolute inset-x-0 top-0 z-10">
        <div className="mx-auto flex w-full max-w-7xl items-center justify-between px-4 py-4">
          <div className="text-xs font-semibold uppercase tracking-[0.28em] text-slate-500">
            Izborator
          </div>
          <LanguageSwitcher />
        </div>
      </header>

      <main className="min-h-screen">
        <div className="mx-auto flex min-h-screen w-full max-w-7xl flex-col items-center justify-center px-4 py-24">
          <div className="relative w-full md:min-h-[640px]">
            <div className="relative z-10 flex min-h-[520px] items-center justify-center">
              <div className="relative flex w-full max-w-[720px] items-center justify-center md:min-w-[520px] min-h-[320px] md:min-h-[360px] lg:min-h-[420px]">
                <HeroSearch
                  title={hero.title}
                  subtitle={hero.subtitle}
                  showTypeToggle={hero.showTypeToggle}
                  defaultType={hero.defaultType}
                  showCitySelect={hero.showCitySelect}
                  cityOptions={cityOptions}
                  searchPlaceholder={hero.searchPlaceholder}
                />
              </div>
            </div>
            <FloatingCategoryCloud categories={categoryCards} isLoading={isLoading} />
          </div>
        </div>
      </main>

      <footer className="pb-8 text-center text-xs text-slate-400">
        {t('footer_text')}
      </footer>
    </div>
  )
}
