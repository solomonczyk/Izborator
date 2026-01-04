import Link from "next/link";

export type CategoryCardProps = {
  title: string;
  hint?: string;
  href: string;
};

export function CategoryCard({ title, hint, href }: CategoryCardProps) {
  return (
    <Link
      href={href}
      className="group block rounded-2xl border border-slate-200 bg-white p-4 text-left shadow-sm transition-shadow hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-400 focus-visible:ring-offset-2"
      aria-label={title}
    >
      <div className="text-base font-semibold text-slate-900">{title}</div>
      {hint ? (
        <div className="mt-1 text-sm text-slate-600">{hint}</div>
      ) : null}
    </Link>
  );
}
