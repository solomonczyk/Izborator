import { CategoryCard, type CategoryCardProps } from '@/components/category-card'

type FloatingCategoryCloudProps = {
  categories: CategoryCardProps[]
  maxVisible?: number
}

const positions = [
  'left-[8%] top-[6%] -translate-x-1/2',
  'left-1/2 top-[2%] -translate-x-1/2',
  'right-[8%] top-[8%] translate-x-1/2',
  'left-[6%] top-1/2 -translate-x-1/2 -translate-y-1/2',
  'right-[6%] top-1/2 translate-x-1/2 -translate-y-1/2',
  'left-[10%] bottom-[6%] -translate-x-1/2',
  'left-1/2 bottom-[2%] -translate-x-1/2',
  'right-[10%] bottom-[6%] translate-x-1/2',
]

export function FloatingCategoryCloud({
  categories,
  maxVisible = 8,
}: FloatingCategoryCloudProps) {
  const visible = categories.slice(0, maxVisible)

  if (visible.length === 0) {
    return null
  }

  return (
    <>
      <div className="mt-10 grid grid-cols-2 gap-4 md:hidden">
        {visible.map((category) => (
          <CategoryCard key={category.id} {...category} />
        ))}
      </div>
      <div className="pointer-events-none absolute inset-0 hidden md:block">
        {visible.map((category, index) => (
          <div
            key={category.id}
            className={`pointer-events-auto absolute ${positions[index % positions.length]} w-[220px]`}
          >
            <CategoryCard {...category} />
          </div>
        ))}
      </div>
    </>
  )
}
