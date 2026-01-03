import { getTranslations } from 'next-intl/server'
import type { CategoryCardProps } from '@/components/category-card'
import { FloatingCategoryCloud } from '@/components/floating-category-cloud'
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
  const categoryCards: CategoryCardProps[] = [
    {
      id: 'electronics',
      title: 'Electronics',
      hint: 'Phones, laptops, gadgets',
      href: '/catalog?type=good&category=electronics',
      priority: 'primary',
    },
    {
      id: 'food',
      title: 'Food & Drinks',
      hint: 'Groceries and delivery',
      href: '/catalog?type=good&category=food',
      priority: 'primary',
    },
    {
      id: 'fashion',
      title: 'Fashion',
      hint: 'Clothes and shoes',
      href: '/catalog?type=good&category=fashion',
    },
    {
      id: 'home',
      title: 'Home & Garden',
      hint: 'Furniture and decor',
      href: '/catalog?type=good&category=home',
    },
    {
      id: 'sport',
      title: 'Sport & Leisure',
      hint: 'Outdoor and fitness',
      href: '/catalog?type=good&category=sport',
    },
    {
      id: 'auto',
      title: 'Auto',
      hint: 'Cars and accessories',
      href: '/catalog?type=good&category=auto',
    },
    {
      id: 'services',
      title: 'Services',
      hint: 'Repair, beauty, events',
      href: '/catalog?type=service',
    },
    {
      id: 'finance',
      title: 'Finance',
      hint: 'Insurance and banking',
      href: '/catalog?type=service&category=finance',
    },
  ]

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
          <div className="relative w-full md:min-h-[640px]">
            <div className="relative z-10 flex min-h-[520px] items-center justify-center">
              <div className="relative flex w-full max-w-[720px] items-center justify-center md:min-w-[520px] min-h-[320px] md:min-h-[360px] lg:min-h-[420px]">
                <HeroSearch title={t('title')} subtitle={t('subtitle')} />
              </div>
            </div>
            <FloatingCategoryCloud categories={categoryCards} />
          </div>
        </div>
      </main>

      <footer className="pb-8 text-center text-xs text-slate-400">
        {t('footer_text')}
      </footer>
    </div>
  )
}
