type HeroSearchProps = {
  onSubmit?: (q: string) => void;
};

export function HeroSearch({ onSubmit }: HeroSearchProps) {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);
    const q = String(data.get("q") || "");
    onSubmit?.(q);
  };

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="text-2xl font-semibold text-slate-900">
        Найдите товары и услуги
      </div>
      <div className="mt-2 text-slate-600">
        Введите запрос, или выберите категорию вокруг
      </div>

      <form className="mt-6 flex gap-3" onSubmit={handleSubmit}>
        <label htmlFor="hero-search" className="sr-only">
          Поиск товаров и услуг
        </label>
        <input
          id="hero-search"
          name="q"
          type="search"
          placeholder="Что вы ищете?"
          className="h-14 flex-1 rounded-xl border border-slate-200 px-4 outline-none focus:border-slate-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2"
        />
        <button
          aria-label="Найти"
          type="submit"
          className="h-14 rounded-xl bg-slate-900 px-5 font-medium text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-600 focus-visible:ring-offset-2"
        >
          Найти
        </button>
      </form>
    </div>
  );
}
