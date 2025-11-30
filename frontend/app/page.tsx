import { redirect } from 'next/navigation'

// Корневая страница редиректит на дефолтную локаль
// Middleware next-intl обработает это и добавит локаль автоматически
export default function HomePage() {
  redirect('/en')
}
