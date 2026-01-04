import { HeroSearch } from "@/components/hero-search";
import { FloatingCategoryCloud } from "@/components/floating-category-cloud";
import type { HomeModel } from "@/types/home";

const ORBIT_BAND = 24;
const ORBIT_BAND_PX = ORBIT_BAND * 4;
const DEBUG_ORBIT =
  process.env.NODE_ENV !== "production" &&
  process.env.NEXT_PUBLIC_DEBUG_CLOUD === "1";

async function getHomeModel(): Promise<HomeModel> {
  return {
    version: "2",
    tenant_id: "default",
    locale: "en",
    hero: {
      title: "Find goods and services",
      subtitle: "Search or browse categories",
      searchPlaceholder: "What are you looking for?",
      showTypeToggle: true,
      showCitySelect: false,
      defaultType: "all",
    },
    featuredCategories: [],
  };
}

export default async function HomePage() {
  await getHomeModel();

  const handleHeroSubmit = (query: string) => {
    console.log(query);
  };

  return (
    <main className="min-h-screen">
      {/* Header — если уже подключается глобально, тут не нужен */}

      {/* Scene container */}
      <section className="relative min-h-screen overflow-hidden">
        {/* Dead zones (не кликаются и не перекрывают центр) */}
        <div aria-hidden="true" className="absolute inset-0 pointer-events-none">
          {/* TODO: FloatingCategoryCloud will live here */}
        </div>

        {/* Layout frame */}
        <div className="relative mx-auto min-h-screen max-w-6xl px-4">
          {/* Orbit zone (пространство вокруг Safe Center под будущее облако) */}
          <div
            aria-hidden="true"
            className="absolute inset-0 pointer-events-none"
          >
            <FloatingCategoryCloud />
            {DEBUG_ORBIT ? (
              <div className="absolute inset-0">
              {/* Top band */}
              <div
                className="absolute left-0 right-0 top-0 bg-sky-200/30 opacity-0 md:opacity-100"
                style={{ height: ORBIT_BAND_PX }}
              />
              {/* Bottom band */}
              <div
                className="absolute left-0 right-0 bottom-0 bg-sky-200/30 opacity-0 md:opacity-100"
                style={{ height: ORBIT_BAND_PX }}
              />
              {/* Left band */}
              <div
                className="absolute left-0 bg-sky-200/30 opacity-0 md:opacity-100"
                style={{
                  top: ORBIT_BAND_PX,
                  bottom: ORBIT_BAND_PX,
                  width: ORBIT_BAND_PX,
                }}
              />
              {/* Right band */}
              <div
                className="absolute right-0 bg-sky-200/30 opacity-0 md:opacity-100"
                style={{
                  top: ORBIT_BAND_PX,
                  bottom: ORBIT_BAND_PX,
                  width: ORBIT_BAND_PX,
                }}
              />
              {/* Center keep-out (dead zone вокруг Safe Center) */}
              <div
                className="absolute bg-rose-200/20 opacity-0 md:opacity-100"
                style={{
                  left: ORBIT_BAND_PX,
                  right: ORBIT_BAND_PX,
                  top: ORBIT_BAND_PX,
                  bottom: ORBIT_BAND_PX,
                }}
              />
            </div>
            ) : null}
          </div>

          {/* Safe Center */}
          <div className="relative z-10 flex min-h-screen items-center justify-center">
            <div className="w-full max-w-2xl">
              <HeroSearch onSubmit={handleHeroSubmit} />
            </div>
          </div>
        </div>
      </section>

      {/* Footer — если уже подключается глобально, тут не нужен */}
    </main>
  );
}
