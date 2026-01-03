type CityOption = {
  value: string
  label: string
}

type CitySelectProps = {
  options: CityOption[]
  allLabel: string
  label?: string
  name?: string
  defaultValue?: string
}

export function CitySelect({
  options,
  allLabel,
  label,
  name = "city",
  defaultValue = "",
}: CitySelectProps) {
  return (
    <label className="flex flex-col gap-1 text-left text-xs text-slate-500">
      {label ? <span>{label}</span> : null}
      <select
        name={name}
        defaultValue={defaultValue}
        className="h-10 rounded-full border border-slate-200 bg-white px-3 text-sm text-slate-700"
      >
        <option value="">{allLabel}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  )
}
