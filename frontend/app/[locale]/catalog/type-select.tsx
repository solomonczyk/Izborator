'use client'

import React from "react"

type TypeOption = {
  value: string
  label: string
  className?: string
}

type TypeSelectProps = {
  id: string
  name: string
  defaultValue: string
  tenantId: string
  className?: string
  options: TypeOption[]
}

export function TypeSelect({ id, name, defaultValue, tenantId, className, options }: TypeSelectProps) {
  const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const selected = event.target.value
    const facetsType = selected === "service" ? "services" : "goods"
    const encodedTenantId = encodeURIComponent(tenantId)
    fetch(`/api/v1/products/facets?type=${facetsType}&tenant_id=${encodedTenantId}`).catch(() => {})
  }

  return (
    <select
      id={id}
      name={name}
      defaultValue={defaultValue}
      className={className}
      onChange={handleChange}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value} className={option.className}>
          {option.label}
        </option>
      ))}
    </select>
  )
}
