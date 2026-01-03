type TypeToggleProps = {
  defaultValue?: "all" | "good" | "service"
  labels: {
    all: string
    goods: string
    services: string
  }
  ariaLabel?: string
  name?: string
}

export function TypeToggle({
  defaultValue = "all",
  labels,
  ariaLabel = "Type",
  name = "type",
}: TypeToggleProps) {
  const options = [
    { value: "all", label: labels.all },
    { value: "good", label: labels.goods },
    { value: "service", label: labels.services },
  ]

  return (
    <div
      className="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-white/90 p-1 text-sm"
      role="radiogroup"
      aria-label={ariaLabel}
    >
      {options.map((option) => {
        const inputValue = option.value === "all" ? "" : option.value

        return (
          <label key={option.value} className="cursor-pointer">
            <input
              className="peer sr-only"
              type="radio"
              name={name}
              value={inputValue}
              defaultChecked={defaultValue === option.value}
            />
            <span className="rounded-full px-3 py-1 text-slate-600 transition-colors peer-checked:bg-blue-600 peer-checked:text-white">
              {option.label}
            </span>
          </label>
        )
      })}
    </div>
  )
}
