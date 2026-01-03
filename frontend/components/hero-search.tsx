import { SearchForm } from '@/components/search-form'

type HeroSearchProps = {
  title: string
  subtitle?: string
  showTypeToggle?: boolean
  defaultType?: "all" | "good" | "service"
  showCitySelect?: boolean
  cityOptions?: Array<{ value: string; label: string }>
  defaultCity?: string
  searchPlaceholder?: string
}

export function HeroSearch({
  title,
  subtitle,
  showTypeToggle = true,
  defaultType = "all",
  showCitySelect = false,
  cityOptions,
  defaultCity,
  searchPlaceholder,
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
        <SearchForm
          showTypeToggle={showTypeToggle}
          defaultType={defaultType}
          showCitySelect={showCitySelect}
          cityOptions={cityOptions}
          defaultCity={defaultCity}
          searchPlaceholder={searchPlaceholder}
        />
      </div>
    </div>
  )
}
