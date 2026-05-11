export interface StepCatalogItem {
  name: string
  category: string
  stepType: 'task' | 'control'
  code?: string
  kind?: string
  available: boolean
}

export function filterStepCatalog<T extends StepCatalogItem>(items: T[], query: string): T[] {
  const search = query.trim().toLowerCase()
  if (!search) return items

  return items.filter((item) => {
    const fields = [
      item.name,
      item.category,
      item.stepType,
      item.code,
      item.kind,
    ]

    return fields.some(field => field?.toLowerCase().includes(search))
  })
}
