import Link from "next/link"
import { IBM_Plex_Mono, Rubik } from "next/font/google"

const rubik = Rubik({
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "700"],
  display: "swap",
})

const plexMono = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "600"],
  display: "swap",
})

const sections = [
  {
    id: "electronics",
    title: "???????????",
    blurb: "???????, ????? ??? ? ???????, ??????? ?????????? ? ???? ??????.",
    href: "/catalog?type=good&category=electronics",
    mode: "GOODS",
    accent: "#22d3ee",
    accentSoft: "rgba(34, 211, 238, 0.35)",
    progress: 68,
    stats: [
      { label: "???????", value: "24" },
      { label: "?????", value: "4.8k" },
      { label: "??????", value: "0.8s" },
    ],
  },
  {
    id: "food",
    title: "??? ? ???????",
    blurb: "??????????? ??? ???????: ?? ????????? ???????? ?? ??????? ?????.",
    href: "/catalog?type=good&category=food",
    mode: "GOODS",
    accent: "#34d399",
    accentSoft: "rgba(52, 211, 153, 0.35)",
    progress: 54,
    stats: [
      { label: "????", value: "12" },
      { label: "????", value: "2.1k" },
      { label: "????", value: "0.6s" },
    ],
  },
  {
    id: "fashion",
    title: "???? ? ?????",
    blurb: "????????? ? ???????, ????????? ? ????? ??????? ??????.",
    href: "/catalog?type=good&category=fashion",
    mode: "GOODS",
    accent: "#a78bfa",
    accentSoft: "rgba(167, 139, 250, 0.35)",
    progress: 72,
    stats: [
      { label: "???????", value: "18" },
      { label: "???????", value: "64" },
      { label: "?????", value: "3.2" },
    ],
  },
  {
    id: "home",
    title: "??? ? ???",
    blurb: "????????????, ??? ?????? ?????? ????????? ????? ???????? ????.",
    href: "/catalog?type=good&category=home",
    mode: "GOODS",
    accent: "#f97316",
    accentSoft: "rgba(249, 115, 22, 0.35)",
    progress: 46,
    stats: [
      { label: "????????", value: "11" },
      { label: "??????", value: "740" },
      { label: "?????", value: "0.9" },
    ],
  },
  {
    id: "sport",
    title: "????? ? ?????",
    blurb: "???????? ????????: ??????????, ???????????, ??????????????.",
    href: "/catalog?type=good&category=sport",
    mode: "GOODS",
    accent: "#38bdf8",
    accentSoft: "rgba(56, 189, 248, 0.35)",
    progress: 61,
    stats: [
      { label: "??????????", value: "15" },
      { label: "?????", value: "118" },
      { label: "????", value: "2.4" },
    ],
  },
  {
    id: "auto",
    title: "????-????",
    blurb: "????????, ??? ?????? ? ??????? ???????? ? ?????? ??????.",
    href: "/catalog?type=good&category=auto",
    mode: "GOODS",
    accent: "#f43f5e",
    accentSoft: "rgba(244, 63, 94, 0.35)",
    progress: 57,
    stats: [
      { label: "????", value: "19" },
      { label: "????????", value: "3.6" },
      { label: "????", value: "0.7s" },
    ],
  },
  {
    id: "services",
    title: "??????",
    blurb: "?????? ???????? ???????: ?????, ???????, ???? ? ???????.",
    href: "/catalog?type=service",
    mode: "SERVICE",
    accent: "#f59e0b",
    accentSoft: "rgba(245, 158, 11, 0.35)",
    progress: 80,
    stats: [
      { label: "????????", value: "27" },
      { label: "?????????", value: "5" },
      { label: "????", value: "0.5s" },
    ],
  },
  {
    id: "finance",
    title: "?????????? ??????",
    blurb: "??????, ??????? ???????, ???????? ? ???????? ???????.",
    href: "/catalog?type=service&category=finance",
    mode: "SERVICE",
    accent: "#14b8a6",
    accentSoft: "rgba(20, 184, 166, 0.35)",
    progress: 63,
    stats: [
      { label: "????????", value: "9" },
      { label: "????", value: "0.3" },
      { label: "?????", value: "1.1k" },
    ],
  },
]

export default function SectionsPage() {
  return (
    <main
      className={`${rubik.className} relative h-screen overflow-y-scroll snap-y snap-mandatory bg-[#0b0d12] text-white`}
    >
      <style jsx global>{`
        @keyframes rise {
          from { opacity: 0; transform: translateY(16px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes drift {
          from { transform: translateY(0); }
          50% { transform: translateY(-12px); }
          to { transform: translateY(0); }
        }
        @keyframes spinSlow {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>

      <div className="pointer-events-none fixed inset-0 -z-10">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(148,163,184,0.15),transparent_40%),radial-gradient(circle_at_80%_10%,rgba(56,189,248,0.12),transparent_35%),radial-gradient(circle_at_50%_80%,rgba(15,23,42,0.9),transparent_55%)]" />
        <div className="absolute inset-0 opacity-40 mix-blend-soft-light [background-image:linear-gradient(rgba(255,255,255,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.04)_1px,transparent_1px)] [background-size:64px_64px]" />
      </div>

      <nav className="hidden lg:flex fixed left-8 top-1/2 -translate-y-1/2 flex-col gap-4 z-20">
        {sections.map((s, idx) => (
          <a
            key={s.id}
            href={`#${s.id}`}
            className="group flex items-center gap-3 text-xs uppercase tracking-[0.35em] text-white/40 hover:text-white"
          >
            <span className={`${plexMono.className} text-[11px]`}>{`0${idx + 1}`}</span>
            <span className="h-px w-10 bg-white/20 group-hover:bg-white" />
          </a>
        ))}
      </nav>

      {sections.map((s, idx) => (
        <section
          key={s.id}
          id={s.id}
          style={{
            ["--accent" as string]: s.accent,
            ["--accent-soft" as string]: s.accentSoft,
          }}
          className="relative h-screen snap-start flex items-center justify-center px-6"
        >
          <div className="absolute inset-0 -z-10">
            <div className="absolute -left-20 top-10 h-64 w-64 rounded-full blur-3xl" style={{ background: "var(--accent-soft)" }} />
            <div className="absolute right-0 bottom-0 h-72 w-72 rounded-full blur-[120px]" style={{ background: "var(--accent)" }} />
          </div>

          <div className="grid w-full max-w-6xl gap-10 lg:grid-cols-[1.15fr_0.85fr] items-center">
            <div className="space-y-6">
              <div className={`${plexMono.className} text-xs uppercase tracking-[0.4em] text-white/40 motion-safe:animate-[rise_0.8s_ease-out]`}>
                ?????? {idx + 1} / {sections.length}
              </div>
              <h1 className="text-4xl sm:text-6xl font-semibold leading-[0.95] motion-safe:animate-[rise_0.9s_ease-out]">
                {s.title}
              </h1>
              <p className="text-white/70 text-lg leading-relaxed max-w-xl motion-safe:animate-[rise_1s_ease-out]">
                {s.blurb}
              </p>

              <div className="flex flex-wrap items-center gap-3">
                <span
                  className={`${plexMono.className} rounded-full border border-white/15 bg-white/5 px-3 py-1 text-[11px] uppercase tracking-[0.2em] text-white/70`}
                >
                  {s.mode}
                </span>
                <span className="text-xs text-white/50">?????? ???????</span>
              </div>

              <div className="flex flex-wrap gap-3">
                <Link
                  href={s.href}
                  className="inline-flex items-center justify-center rounded-full bg-white text-slate-950 px-6 py-3 text-sm font-semibold shadow-[0_10px_30px_rgba(255,255,255,0.2)] hover:shadow-[0_12px_38px_rgba(255,255,255,0.25)]"
                >
                  ??????? ???????
                </Link>
                <a
                  href={idx < sections.length - 1 ? `#${sections[idx + 1].id}` : `#${sections[0].id}`}
                  className="inline-flex items-center justify-center rounded-full border border-white/20 px-6 py-3 text-sm font-semibold text-white/80 hover:text-white"
                >
                  ?????
                </a>
              </div>
            </div>

            <div className="relative">
              <div className="rounded-[32px] border border-white/10 bg-white/5 p-6 backdrop-blur">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full" style={{ background: "var(--accent)" }} />
                    <span className={`${plexMono.className} text-[11px] uppercase tracking-[0.35em] text-white/60`}>
                      Module
                    </span>
                  </div>
                  <span className={`${plexMono.className} text-xs text-white/50`}>{`IDX-${idx + 1}`}</span>
                </div>

                <div className="mt-6 flex flex-col items-center gap-6">
                  <div className="relative w-full max-w-[240px]">
                    <div className="aspect-square rounded-full border border-white/15 p-6">
                      <div
                        className="absolute inset-6 rounded-full border border-white/10"
                        style={{ background: "radial-gradient(circle at 30% 30%, rgba(255,255,255,0.12), transparent 55%)" }}
                      />
                      <div
                        className="absolute inset-0 rounded-full motion-safe:animate-[spinSlow_16s_linear_infinite]"
                        style={{
                          background: `conic-gradient(from 210deg, var(--accent), rgba(255,255,255,0.1), transparent 55%, var(--accent))`,
                          opacity: 0.6,
                        }}
                      />
                      <div className="absolute left-1/2 top-1/2 h-2 w-20 -translate-y-1/2 origin-left rotate-[20deg]" style={{ background: "var(--accent)" }} />
                      <div className="absolute left-1/2 top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white" />
                    </div>
                  </div>

                  <div className="w-full space-y-3">
                    <div className="flex items-center justify-between text-xs text-white/60">
                      <span className={`${plexMono.className} uppercase tracking-[0.2em]`}>Progress</span>
                      <span className={`${plexMono.className}`}>{s.progress}%</span>
                    </div>
                    <div className="h-2 w-full rounded-full bg-white/10">
                      <div
                        className="h-full rounded-full"
                        style={{ width: `${s.progress}%`, background: "var(--accent)" }}
                      />
                    </div>
                  </div>

                  <div className="grid w-full grid-cols-3 gap-3">
                    {s.stats.map((stat) => (
                      <div key={stat.label} className="rounded-2xl border border-white/10 bg-white/5 p-3">
                        <div className={`${plexMono.className} text-xs text-white/50 uppercase tracking-[0.2em]`}>
                          {stat.label}
                        </div>
                        <div className="text-lg font-semibold mt-1" style={{ color: "var(--accent)" }}>
                          {stat.value}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="absolute -right-6 -bottom-6 hidden lg:block">
                <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 text-xs text-white/60 shadow-lg motion-safe:animate-[drift_5s_ease-in-out_infinite]">
                  ??????? ???? ??? ????????? ??????
                </div>
              </div>
            </div>
          </div>
        </section>
      ))}
    </main>
  )
}
