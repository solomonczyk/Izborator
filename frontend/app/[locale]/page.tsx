import { redirect } from 'next/navigation'
import { locales } from '@/i18n'

// Генерируем статические параметры для всех локалей
export function generateStaticParams() {
  return locales.map((locale) => ({ locale }))
}

// Главная страница для локали редиректит на каталог
export default async function LocaleHomePage({
  params,
}: {
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params
  redirect(`/${locale}/catalog`)
}

