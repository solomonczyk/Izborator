// Root layout не должен делать редирект
// Middleware next-intl сам обрабатывает редиректы
export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return children
}
