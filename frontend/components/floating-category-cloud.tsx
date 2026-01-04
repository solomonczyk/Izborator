import { CategoryCard } from "@/components/category-card";

const DEBUG_CLOUD =
  process.env.NODE_ENV !== "production" &&
  process.env.NEXT_PUBLIC_DEBUG_CLOUD === "1";

const MOCK_CATEGORIES = [
  {
    id: "electronics",
    title: "Electronics",
    hint: "Phones, laptops, gadgets",
    href: "/catalog?type=good&category=electronics",
    positionClass: "left-4 top-4 w-56",
  },
  {
    id: "food",
    title: "Food & Drink",
    hint: "Groceries and cafes",
    href: "/catalog?type=good&category=food",
    positionClass: "left-1/2 top-6 w-56 -translate-x-1/2",
  },
  {
    id: "fashion",
    title: "Fashion",
    hint: "Clothes and shoes",
    href: "/catalog?type=good&category=fashion",
    positionClass: "right-4 top-8 w-56",
  },
  {
    id: "home",
    title: "Home & Garden",
    hint: "Furniture and decor",
    href: "/catalog?type=good&category=home",
    positionClass: "left-0 top-32 w-56",
  },
  {
    id: "services",
    title: "Services",
    hint: "Repairs and delivery",
    href: "/catalog?type=service",
    positionClass: "right-0 bottom-32 w-56",
  },
  {
    id: "sport",
    title: "Sport & Leisure",
    hint: "Outdoor and fitness",
    href: "/catalog?type=good&category=sport",
    positionClass: "left-6 bottom-10 w-56",
  },
  {
    id: "auto",
    title: "Auto",
    hint: "Parts and accessories",
    href: "/catalog?type=good&category=auto",
    positionClass: "left-1/2 bottom-6 w-56 -translate-x-1/2",
  },
  {
    id: "finance",
    title: "Finance",
    hint: "Payments and loans",
    href: "/catalog?type=service&category=finance",
    positionClass: "right-6 bottom-8 w-56",
  },
] as const;

export function FloatingCategoryCloud() {
  const debugClass = DEBUG_CLOUD ? "ring-1 ring-slate-300/50" : "";

  return (
    <>
      <div
        aria-hidden="true"
        className="absolute inset-0 pointer-events-none hidden md:block"
      >
        {MOCK_CATEGORIES.map((category) => (
          <div
            key={category.id}
            className={`absolute ${category.positionClass} ${debugClass}`}
          >
            <CategoryCard
              title={category.title}
              hint={category.hint}
              href={category.href}
            />
          </div>
        ))}
      </div>
      <div className="grid grid-cols-2 gap-4 px-4 pb-8 pt-4 md:hidden">
        {MOCK_CATEGORIES.map((category) => (
          <CategoryCard
            key={category.id}
            title={category.title}
            hint={category.hint}
            href={category.href}
          />
        ))}
      </div>
    </>
  );
}
