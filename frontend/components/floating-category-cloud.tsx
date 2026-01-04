import { CategoryCard } from "@/components/category-card";
import type { FeaturedCategory } from "@/types/home";

const DEBUG_CLOUD =
  process.env.NODE_ENV !== "production" &&
  process.env.NEXT_PUBLIC_DEBUG_CLOUD === "1";

const POSITION_CLASSES = [
  "left-4 top-4 w-56",
  "left-1/2 top-6 w-56 -translate-x-1/2",
  "right-4 top-8 w-56",
  "left-0 top-32 w-56",
  "right-0 bottom-32 w-56",
  "left-6 bottom-10 w-56",
  "left-1/2 bottom-6 w-56 -translate-x-1/2",
  "right-6 bottom-8 w-56",
];

type FloatingCategoryCloudProps = {
  categories: FeaturedCategory[];
  maxVisible?: number;
};

const SKELETON_ITEMS = Array.from({ length: 8 }, (_, index) => ({
  id: `skeleton-${index + 1}`,
  positionClass: POSITION_CLASSES[index],
}));

export function FloatingCategoryCloud({
  categories,
  maxVisible = 8,
}: FloatingCategoryCloudProps) {
  const hasCategories = Array.isArray(categories) && categories.length > 0;

  const sorted = [...categories].sort(
    (a, b) => (a.order || 0) - (b.order || 0),
  );
  const visible = sorted.slice(0, Math.min(maxVisible, POSITION_CLASSES.length));
  const debugClass = DEBUG_CLOUD ? "ring-1 ring-slate-300/50" : "";

  return (
    <>
      <div className="absolute inset-0 hidden md:block">
        {hasCategories
          ? visible.map((category, index) => (
              <div
                key={category.category_id}
                className={`absolute ${POSITION_CLASSES[index]} pointer-events-auto ${debugClass}`}
              >
                <CategoryCard title={category.title} href={category.href} />
              </div>
            ))
          : SKELETON_ITEMS.map((item) => (
              <div
                key={item.id}
                className={`absolute ${item.positionClass} pointer-events-none ${debugClass}`}
              >
                <div className="h-20 w-full rounded-2xl border border-slate-200 bg-slate-100/70" />
              </div>
            ))}
      </div>
      <div className="grid grid-cols-2 gap-4 px-4 pb-8 pt-4 md:hidden">
        {hasCategories
          ? visible.map((category) => (
              <CategoryCard
                key={category.category_id}
                title={category.title}
                href={category.href}
              />
            ))
          : SKELETON_ITEMS.map((item) => (
              <div
                key={item.id}
                className="h-20 rounded-2xl border border-slate-200 bg-slate-100/70"
              />
            ))}
      </div>
    </>
  );
}
