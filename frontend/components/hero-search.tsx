import { SearchForm } from '@/components/search-form'

type HeroSearchProps = {
  title: string
  subtitle?: string
  showTypeToggle?: boolean
  defaultType?: "all" | "good" | "service"
}

export function HeroSearch({
  title,
  subtitle,
  showTypeToggle = true,
  defaultType = "all",
}: HeroSearchProps) {
  return (
    <div className="w-full max-w-3xl text-center">
      <h1 className="text-4xl font-bold text-slate-900 md:text-5xl lg:text-6xl">
        {title}
      </h1>
      {subtitle ? (
        <p className="mt-4 text-lg text-slate-600 md:text-2xl">{subtitle}</p>
      ) : null}
      <div className="mt-10">
        <SearchForm showTypeToggle={showTypeToggle} defaultType={defaultType} />
      </div>
    </div>
  )
}
