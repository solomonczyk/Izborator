'use client'

import { useEffect, useRef } from 'react'
import { CategoryCard, type CategoryCardProps } from '@/components/category-card'

type FloatingCategoryCloudProps = {
  categories: CategoryCardProps[]
  maxVisible?: number
  isLoading?: boolean
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
  isLoading = false,
}: FloatingCategoryCloudProps) {
  const visible = categories.slice(0, maxVisible)
  const skeletons = Array.from({ length: maxVisible }, (_, index) => ({
    id: `skeleton-${index}`,
  }))
  const items = isLoading ? skeletons : visible
  const cardRefs = useRef<Array<HTMLDivElement | null>>([])

  useEffect(() => {
    if (!window.matchMedia('(hover: hover) and (pointer: fine)').matches) {
      return
    }
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return
    }

    const rootStyles = getComputedStyle(document.documentElement)
    const maxMove = parseFloat(rootStyles.getPropertyValue('--move-md')) || 12
    const rotateMax = parseFloat(rootStyles.getPropertyValue('--rotate-xs')) || 2
    const radius = 120
    let rafId = 0
    let pointerX = 0
    let pointerY = 0
    let hasPointer = false

    const update = () => {
      rafId = 0
      cardRefs.current.forEach((card) => {
        if (!card) {
          return
        }
        if (!hasPointer) {
          card.style.setProperty('--proximity-x', '0px')
          card.style.setProperty('--proximity-y', '0px')
          card.style.setProperty('--proximity-rotate', '0deg')
          return
        }
        const rect = card.getBoundingClientRect()
        const cx = rect.left + rect.width / 2
        const cy = rect.top + rect.height / 2
        const dx = pointerX - cx
        const dy = pointerY - cy
        const dist = Math.hypot(dx, dy)

        if (dist > 0 && dist < radius) {
          const strength = (radius - dist) / radius
          const move = Math.min(maxMove, maxMove * strength)
          const nx = dx / dist
          const ny = dy / dist
          card.style.setProperty('--proximity-x', `${-nx * move}px`)
          card.style.setProperty('--proximity-y', `${-ny * move}px`)
          card.style.setProperty('--proximity-rotate', `${nx * rotateMax}deg`)
        } else {
          card.style.setProperty('--proximity-x', '0px')
          card.style.setProperty('--proximity-y', '0px')
          card.style.setProperty('--proximity-rotate', '0deg')
        }
      })
    }

    const scheduleUpdate = () => {
      if (rafId) {
        return
      }
      rafId = window.requestAnimationFrame(update)
    }

    const handlePointerMove = (event: PointerEvent) => {
      pointerX = event.clientX
      pointerY = event.clientY
      hasPointer = true
      scheduleUpdate()
    }

    const handlePointerLeave = () => {
      hasPointer = false
      scheduleUpdate()
    }

    window.addEventListener('pointermove', handlePointerMove, { passive: true })
    window.addEventListener('pointerleave', handlePointerLeave)

    return () => {
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerleave', handlePointerLeave)
      if (rafId) {
        window.cancelAnimationFrame(rafId)
      }
    }
  }, [])

  if (!isLoading && visible.length === 0) {
    return null
  }

  return (
    <>
      <div className="mt-10 grid grid-cols-2 gap-4 md:hidden">
        {isLoading
          ? skeletons.map((skeleton) => (
              <div
                key={skeleton.id}
                className="h-[96px] rounded-2xl border border-slate-200 bg-white/80 shadow-sm animate-pulse"
                aria-hidden="true"
              />
            ))
          : visible.map((category) => <CategoryCard key={category.id} {...category} />)}
      </div>
      <div className="absolute inset-0 hidden md:block">
        {items.map((category, index) => (
          <div
            key={category.id}
            ref={(node) => {
              cardRefs.current[index] = node
            }}
            className={`absolute ${positions[index % positions.length]} w-[220px]`}
            style={{
              transform:
                'translate(var(--proximity-x, 0px), var(--proximity-y, 0px)) rotate(var(--proximity-rotate, 0deg))',
              transition: 'transform var(--motion-base) var(--ease-out-soft)',
              willChange: 'transform',
            }}
          >
            {isLoading ? (
              <div
                className="h-[112px] rounded-2xl border border-slate-200 bg-white/80 shadow-sm animate-pulse"
                aria-hidden="true"
              />
            ) : (
              <CategoryCard {...category} />
            )}
          </div>
        ))}
      </div>
    </>
  )
}
