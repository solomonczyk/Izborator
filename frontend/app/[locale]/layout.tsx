import { NextIntlClientProvider } from 'next-intl'
import { getMessages } from 'next-intl/server'
import { notFound } from 'next/navigation'
import { locales, type Locale } from '@/i18n'

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }))
}

export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode
  params: Promise<{ locale: string }>
}) {
  const { locale: resolvedLocale } = await params
  
  // Проверяем, что язык поддерживается
  if (!locales.includes(resolvedLocale as Locale)) {
    notFound()
  }

  // Загружаем сообщения для языка
  const messages = await getMessages()

  return (
    <NextIntlClientProvider locale={resolvedLocale} messages={messages}>
      {children}
    </NextIntlClientProvider>
  )
}

