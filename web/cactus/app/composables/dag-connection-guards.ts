import { useControlSteps } from './useControlSteps'

export function getControlOutcomeIds(data: { controlKind?: string }): string[] {
  const kind = data.controlKind
  if (!kind) return ['success']

  const definition = useControlSteps().find(control => control.kind === kind)
  return definition?.handles.map(handle => handle.id) ?? ['success']
}

export function canConnectSteps(args: {
  source?: string | null
  target?: string | null
  sourceHandle?: string | null
  nodes: Array<{ id: string, data: any }>
}) {
  if (!args.source || !args.target) return false
  if (args.source === args.target) return false

  const source = args.nodes.find(n => n.id === args.source)
  const target = args.nodes.find(n => n.id === args.target)
  if (!source || !target) return false
  if (target.data.controlKind === 'start') return false

  const handle = args.sourceHandle ?? 'success'
  if (source.data.controlKind === 'start') return handle === 'success'
  if (source.data.stepType === 'task') return handle === 'success'
  if (source.data.stepType === 'control') return getControlOutcomeIds(source.data).includes(handle)
  return false
}
