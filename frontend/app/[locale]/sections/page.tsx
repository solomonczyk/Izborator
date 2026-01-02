import Link from "next/link"

const sections = [
  { id: "electronics", title: "Электроника", href: "/catalog?type=good&category=electronics" },
  { id: "food", title: "Еда и напитки", href: "/catalog?type=good&category=food" },
  { id: "fashion", title: "Мода и обувь", href: "/catalog?type=good&category=fashion" },
  { id: "home", title: "Дом и сад", href: "/catalog?type=good&category=home" },
  { id: "sport", title: "Спорт и отдых", href: "/catalog?type=good&category=sport" },
  { id: "auto", title: "Авто-мото", href: "/catalog?type=good&category=auto" },
  { id: "services", title: "Услуги", href: "/catalog?type=service" },
  { id: "finance", title: "Финансовые услуги", href: "/catalog?type=service&category=finance" },
]

export default function SectionsPage() {
  return (
    <main className="h-screen overflow-y-scroll snap-y snap-mandatory">
      {sections.map((s, idx) => (
        <section
          key={s.id}
          id={s.id}
          className="h-screen snap-start flex items-center justify-center px-6"
        >
          <div className="w-full max-w-3xl rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
            <div className="text-sm text-slate-500 mb-2">Раздел {idx + 1} / {sections.length}</div>
            <h1 className="text-4xl font-bold text-slate-900 mb-4">{s.title}</h1>
            <p className="text-slate-600 mb-6">
              Тут будет краткое описание категории и подборка (можно вывести топ-товары/услуги).
            </p>

            <div className="flex gap-3">
              <Link
                href={s.href}
                className="inline-flex items-center justify-center rounded-lg bg-blue-600 px-5 py-2.5 text-white font-medium hover:bg-blue-700"
              >
                Открыть каталог
              </Link>
              <a
                href={idx < sections.length - 1 ? `#${sections[idx + 1].id}` : `#${sections[0].id}`}
                className="inline-flex items-center justify-center rounded-lg border border-slate-200 px-5 py-2.5 text-slate-900 font-medium hover:bg-slate-50"
              >
                Далее
              </a>
            </div>
          </div>
        </section>
      ))}
    </main>
  )
}
