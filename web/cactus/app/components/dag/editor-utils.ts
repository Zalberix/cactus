export interface ReadOnlyVersionLike {
  is_active?: boolean
  run_count?: number
}

export function isVersionReadOnly(version: ReadOnlyVersionLike | null | undefined): boolean {
  return Boolean(version?.is_active && (version.run_count ?? 0) > 0)
}

export function editorSurfaceForStep(data: { stepType?: string, controlKind?: string }) {
  if (data.controlKind === 'start') return 'none'
  return data.stepType === 'control' ? 'step-panel' : 'node-editor'
}
