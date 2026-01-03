import { getTranslations } from 'next-intl/server'
import { HeroSearch } from '@/components/hero-search'
import { LanguageSwitcher } from '@/components/language-switcher'

export const dynamic = 'force-dynamic'

export default async function HomePage({
  params,
}: {
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'home' })

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 to-white">
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
          <div className="relative flex w-full max-w-[720px] items-center justify-center md:min-w-[520px] min-h-[320px] md:min-h-[360px] lg:min-h-[420px]">
            <HeroSearch title={t('title')} subtitle={t('subtitle')} />
          </div>

          <div className="mt-14 w-full max-w-6xl">
            <div className="flex h-[360px] items-center justify-center rounded-3xl border border-dashed border-slate-200 bg-white/70 text-sm text-slate-500">
              FloatingCategoryCloud placeholder
            </div>
          </div>
        </div>
      </main>

      <footer className="pb-8 text-center text-xs text-slate-400">
        {t('footer_text')}
      </footer>
    </div>
  )
}
