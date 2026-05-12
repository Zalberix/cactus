import type { WorkerSettingsSchemaSummary } from '~/composables/useWorkers'

export type StepKind = 'task' | 'control'

export interface StepCatalogPayload {
  stepType: StepKind
  workTypeId?: number
  workTypeCode?: string
  workerSettingsSchemaId?: number
  name?: string
  schemas?: WorkerSettingsSchemaSummary[]
}

export interface StepAddPayload extends StepCatalogPayload {
  position: { x: number, y: number }
}

export interface StepCatalogItem {
  name: string
  category: string
  stepType: StepKind
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

export function schemaChoiceOptions(
  schemas: WorkerSettingsSchemaSummary[] | undefined,
): WorkerSettingsSchemaSummary[] {
  return (schemas ?? []).filter(schema => schema.worker_count > 0)
}

export function selectedSchemaIdForStep(
  stepType: StepKind,
  schemas: WorkerSettingsSchemaSummary[] | undefined,
  selectedSchemaId?: number,
): number | undefined {
  if (stepType !== 'task') return undefined
  if (selectedSchemaId) return selectedSchemaId

  const choices = schemaChoiceOptions(schemas)
  return choices.length === 1 ? choices[0]?.id : undefined
}

export function shouldPromptForSchemaChoice(
  stepType: StepKind,
  schemas: WorkerSettingsSchemaSummary[] | undefined,
  selectedSchemaId?: number,
): boolean {
  if (stepType !== 'task' || selectedSchemaId) return false
  return schemaChoiceOptions(schemas).length > 1
}
