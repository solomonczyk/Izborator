import { Geist, Geist_Mono } from 'next/font/google'
import './globals.css'

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
})

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
})

// Root layout для Next.js App Router
// next-intl использует [locale]/layout.tsx для локализованных маршрутов
export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const motionDisabled = process.env.NEXT_PUBLIC_DISABLE_MOTION === 'true'

  return (
    <html suppressHydrationWarning>
      <body
        data-motion={motionDisabled ? 'off' : 'on'}
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        {children}
      </body>
    </html>
  )
}
