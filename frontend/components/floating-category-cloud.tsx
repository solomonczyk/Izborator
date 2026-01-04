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

export function FloatingCategoryCloud({
  categories,
  maxVisible = 8,
}: FloatingCategoryCloudProps) {
  if (!categories || categories.length === 0) {
    return null;
  }

  const sorted = [...categories].sort(
    (a, b) => (a.order || 0) - (b.order || 0),
  );
  const visible = sorted.slice(0, Math.min(maxVisible, POSITION_CLASSES.length));
  const debugClass = DEBUG_CLOUD ? "ring-1 ring-slate-300/50" : "";

  return (
    <>
      <div
        aria-hidden="true"
        className="absolute inset-0 pointer-events-none hidden md:block"
      >
        {visible.map((category, index) => (
          <div
            key={category.category_id}
            className={`absolute ${POSITION_CLASSES[index]} pointer-events-auto ${debugClass}`}
          >
            <CategoryCard title={category.title} href={category.href} />
          </div>
        ))}
      </div>
      <div className="grid grid-cols-2 gap-4 px-4 pb-8 pt-4 md:hidden">
        {visible.map((category) => (
          <CategoryCard
            key={category.category_id}
            title={category.title}
            href={category.href}
          />
        ))}
      </div>
    </>
  );
}
