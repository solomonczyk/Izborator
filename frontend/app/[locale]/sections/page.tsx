import { Link } from "@/navigation"
import { Manrope, Rubik } from "next/font/google"

const manrope = Manrope({
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "600", "700"],
  display: "swap",
})

const rubik = Rubik({
  subsets: ["latin", "cyrillic"],
  weight: ["500", "600", "700"],
  display: "swap",
})

const sections = [
  {
    id: "electronics",
    title: "Электроника",
    blurb: "Гаджеты, умный дом и техника, которые собираются в одну историю.",
    href: "/catalog?type=good&category=electronics",
    accent: "#f07c4a",
    accentSoft: "rgba(240, 124, 74, 0.35)",
    background: "linear-gradient(120deg, #e6eaee 0%, #f2ebe4 45%, #f6d6c4 100%)",
  },
  {
    id: "food",
    title: "Еда и напитки",
    blurb: "Гастрономия с продуманными наборами, сезонными предложениями и вкусом.",
    href: "/catalog?type=good&category=food",
    accent: "#e39b5f",
    accentSoft: "rgba(227, 155, 95, 0.35)",
    background: "linear-gradient(120deg, #e7ecea 0%, #f2efe7 50%, #f7dfc4 100%)",
  },
  {
    id: "fashion",
    title: "Мода и обувь",
    blurb: "Коллекции и капсулы, которые можно собрать по настроению и сезону.",
    href: "/catalog?type=good&category=fashion",
    accent: "#b08fe6",
    accentSoft: "rgba(176, 143, 230, 0.35)",
    background: "linear-gradient(120deg, #e9edf1 0%, #efe9f3 45%, #e3d2f5 100%)",
  },
  {
    id: "home",
    title: "Дом и сад",
    blurb: "Уют и функциональность: всё, чтобы пространство работало на вас.",
    href: "/catalog?type=good&category=home",
    accent: "#f28f57",
    accentSoft: "rgba(242, 143, 87, 0.35)",
    background: "linear-gradient(120deg, #ecebea 0%, #f3ede5 45%, #f4d7c1 100%)",
  },
  {
    id: "sport",
    title: "Спорт и отдых",
    blurb: "Экипировка, путешествия и восстановление для активных сценариев.",
    href: "/catalog?type=good&category=sport",
    accent: "#5fb2e6",
    accentSoft: "rgba(95, 178, 230, 0.35)",
    background: "linear-gradient(120deg, #e6eef3 0%, #ecf2f1 45%, #cfe2f4 100%)",
  },
  {
    id: "auto",
    title: "Авто-мото",
    blurb: "Каталоги, где детали, сервис и подборки работают вместе.",
    href: "/catalog?type=good&category=auto",
    accent: "#f56b88",
    accentSoft: "rgba(245, 107, 136, 0.35)",
    background: "linear-gradient(120deg, #e9eaef 0%, #f2e7eb 45%, #f7cdd6 100%)",
  },
  {
    id: "services",
    title: "Услуги",
    blurb: "Чистая механика сервиса: время, маршрут, команда и результат.",
    href: "/catalog?type=service",
    accent: "#f3a63a",
    accentSoft: "rgba(243, 166, 58, 0.35)",
    background: "linear-gradient(120deg, #ecebe7 0%, #f4eee2 45%, #f6e1c0 100%)",
  },
  {
    id: "finance",
    title: "Финансовые услуги",
    blurb: "Потоки, которые считают, защищают и ускоряют решения.",
    href: "/catalog?type=service&category=finance",
    accent: "#3db6a5",
    accentSoft: "rgba(61, 182, 165, 0.35)",
    background: "linear-gradient(120deg, #e4eeea 0%, #eaf2ef 45%, #cceae1 100%)",
  },
]

export default function SectionsPage() {
  return (
    <main
      className={`${manrope.className} relative h-screen overflow-y-scroll snap-y snap-mandatory scroll-smooth text-slate-900`}
    >
      {sections.map((s, idx) => {
        const progress = Math.round(((idx + 1) / sections.length) * 100)
        return (
          <section
            key={s.id}
            id={s.id}
            style={{
              background: s.background,
              ["--accent" as string]: s.accent,
              ["--accent-soft" as string]: s.accentSoft,
            }}
            className="relative h-screen snap-start overflow-hidden"
          >
            <div className="absolute inset-0">
              <div
                className="absolute -left-20 -top-16 h-72 w-72 rounded-full blur-3xl"
                style={{ background: "var(--accent-soft)" }}
              />
              <div
                className="absolute right-0 bottom-0 h-80 w-80 rounded-full blur-[120px]"
                style={{ background: "var(--accent)" }}
              />
              <div className="absolute inset-0 opacity-40 [background-image:radial-gradient(circle_at_20%_20%,rgba(255,255,255,0.7),transparent_55%),radial-gradient(circle_at_80%_30%,rgba(255,255,255,0.5),transparent_60%)]" />
            </div>

            <div className="relative z-10 flex h-full flex-col justify-between px-6 py-10 md:px-12">
              <header className="flex items-center justify-between text-xs uppercase tracking-[0.35em] text-slate-500">
                <span className={rubik.className}>Izborator</span>
                <span className={rubik.className}>Раздел {idx + 1} / {sections.length}</span>
              </header>

              <div className="relative flex flex-1 items-center justify-center">
                <div className="relative w-full max-w-3xl">
                  <div className="rounded-[36px] border border-white/70 bg-white/70 p-8 shadow-[0_30px_80px_rgba(15,23,42,0.18)] backdrop-blur-xl md:p-12">
                    <div className="text-xs uppercase tracking-[0.4em] text-slate-500">
                      Категория
                    </div>
                    <h1 className={`${rubik.className} mt-4 text-3xl font-semibold leading-tight text-slate-900 md:text-5xl`}>
                      {s.title}
                    </h1>
                    <p className="mt-4 text-base text-slate-600 md:text-lg">
                      {s.blurb}
                    </p>
                    <div className="mt-8 flex flex-wrap gap-3">
                      <Link
                        href={s.href}
                        className="rounded-full bg-slate-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-slate-900/25"
                      >
                        Открыть каталог
                      </Link>
                      <a
                        href={idx < sections.length - 1 ? `#${sections[idx + 1].id}` : `#${sections[0].id}`}
                        className="rounded-full border border-white/70 bg-white/60 px-6 py-3 text-sm font-semibold text-slate-800"
                      >
                        Далее
                      </a>
                    </div>
                  </div>

                  <div className="pointer-events-none absolute -left-24 top-6 hidden xl:block">
                    <div className="w-[260px] rounded-[28px] border border-white/70 bg-white/70 p-4 shadow-[0_24px_60px_rgba(15,23,42,0.18)] backdrop-blur motion-safe:animate-[float_6s_ease-in-out_infinite]">
                      <div
                        className="h-32 rounded-2xl"
                        style={{ background: `linear-gradient(140deg, rgba(255,255,255,0.9), var(--accent-soft))` }}
                      />
                      <div className="mt-3 text-xs text-slate-500">Витрина</div>
                      <div className="text-sm font-semibold text-slate-900">Топ-подборка</div>
                    </div>
                  </div>

                  <div className="pointer-events-none absolute -right-28 top-0 hidden xl:block">
                    <div className="w-[240px] rounded-[24px] border border-white/70 bg-white/65 p-4 shadow-[0_20px_50px_rgba(15,23,42,0.16)] backdrop-blur motion-safe:animate-[float_5s_ease-in-out_infinite]">
                      <div className="flex gap-2">
                        <span className="h-2.5 w-2.5 rounded-full bg-rose-300" />
                        <span className="h-2.5 w-2.5 rounded-full bg-amber-300" />
                        <span className="h-2.5 w-2.5 rounded-full bg-emerald-300" />
                      </div>
                      <div
                        className="mt-3 h-24 rounded-2xl"
                        style={{ background: `linear-gradient(160deg, rgba(255,255,255,0.9), var(--accent-soft))` }}
                      />
                      <div className="mt-3 text-xs text-slate-500">Сценарий</div>
                      <div className="text-sm font-semibold text-slate-900">Свежие решения</div>
                    </div>
                  </div>

                  <div className="pointer-events-none absolute right-10 -bottom-16 hidden lg:block">
                    <div className="w-[200px] rounded-[22px] border border-white/70 bg-white/70 p-4 shadow-[0_20px_50px_rgba(15,23,42,0.16)] backdrop-blur motion-safe:animate-[float_7s_ease-in-out_infinite]">
                      <div className="text-xs uppercase tracking-[0.3em] text-slate-500">Поток</div>
                      <div className="mt-2 text-2xl font-semibold text-slate-900">{progress}%</div>
                      <div className="mt-3 h-2 w-full rounded-full bg-slate-200">
                        <div
                          className="h-full rounded-full"
                          style={{ width: `${progress}%`, background: "var(--accent)" }}
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <footer className="flex items-center justify-between text-xs text-slate-500">
                <span className="hidden sm:inline">Нажмите «Далее» или прокрутите</span>
                <a
                  href={idx < sections.length - 1 ? `#${sections[idx + 1].id}` : `#${sections[0].id}`}
                  className="rounded-full bg-slate-900 px-5 py-2 text-[11px] font-semibold uppercase tracking-[0.3em] text-white"
                >
                  Далее →
                </a>
                <span>Прогресс {progress}%</span>
              </footer>
            </div>
          </section>
        )
      })}
    </main>
  )
}
