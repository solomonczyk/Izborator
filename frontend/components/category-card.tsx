import type { ReactNode } from 'react'
import { Link } from '@/navigation'

export type CategoryCardProps = {
  id: string
  title: string
  hint?: string
  icon?: ReactNode
  href: string
  priority?: 'primary' | 'secondary'
  analyticsId?: string
}

export function CategoryCard({
  id,
  title,
  hint,
  icon,
  href,
  priority = 'secondary',
  analyticsId,
}: CategoryCardProps) {
  const ariaLabel = hint ? `${title}. ${hint}` : title

  return (
    <Link
      href={href}
      tabIndex={0}
      aria-label={ariaLabel}
      data-card-id={id}
      data-priority={priority}
      data-analytics-id={analyticsId}
      className="group flex w-full items-center gap-3 rounded-2xl border border-slate-200 bg-white/90 p-4 shadow-sm md:hover:border-slate-300 md:hover:shadow-md md:hover:[--card-scale:var(--scale-hover)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-200 focus-visible:ring-offset-2 active:[--card-scale:var(--scale-press)]"
      style={{
        transform: 'scale(var(--card-scale, 1))',
        transition:
          'transform var(--motion-fast) var(--ease-out-soft), box-shadow var(--motion-fast) var(--ease-out-soft), border-color var(--motion-fast) var(--ease-out-soft)',
      }}
    >
      {icon ? (
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100 text-slate-500 transition-colors md:group-hover:text-slate-700">
          {icon}
        </div>
      ) : null}
      <div className="text-left">
        <div className="text-base font-semibold text-slate-900">{title}</div>
        {hint ? <div className="text-xs text-slate-500">{hint}</div> : null}
      </div>
    </Link>
  )
}
