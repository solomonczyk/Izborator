type HeroSearchProps = {
  title: string;
  subtitle?: string;
  searchPlaceholder?: string;
  submitLabel?: string;
  action: string;
};

export function HeroSearch({
  title,
  subtitle,
  searchPlaceholder,
  submitLabel = "Search",
  action,
}: HeroSearchProps) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="text-2xl font-semibold text-slate-900">{title}</div>
      {subtitle ? (
        <div className="mt-2 text-slate-600">{subtitle}</div>
      ) : null}

      <form className="mt-6 flex gap-3" action={action} method="GET">
        <label htmlFor="hero-search" className="sr-only">
          Search products and services
        </label>
        <input
          id="hero-search"
          name="q"
          type="search"
          placeholder={searchPlaceholder || "What are you looking for?"}
          className="h-14 flex-1 rounded-xl border border-slate-200 px-4 outline-none focus:border-slate-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2"
        />
        <button
          aria-label={submitLabel}
          type="submit"
          className="h-14 rounded-xl bg-slate-900 px-5 font-medium text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-600 focus-visible:ring-offset-2"
        >
          {submitLabel}
        </button>
      </form>
    </div>
  );
}
