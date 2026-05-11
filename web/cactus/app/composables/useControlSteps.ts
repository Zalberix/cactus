export interface ControlStepDefinition {
  kind: 'condition' | 'switch' | 'delay'
  name: string
  category: string
  icon: string
  color: string
  settingsSchema: Record<string, unknown>
  handles: Array<{ id: string, label: string, color: string }>
}

export function useControlSteps() {
  return [
    {
      kind: 'condition',
      name: 'IF Condition',
      category: 'Logic',
      icon: 'git-branch',
      color: '#ef4444',
      settingsSchema: {
        type: 'object',
        required: ['left', 'operator'],
        properties: {
          left: { type: 'string', description: 'Value 1' },
          operator: { type: 'string', enum: ['eq', 'neq', 'gt', 'gte', 'lt', 'lte', 'contains', 'not_contains', 'exists', 'not_exists'] },
          right: { type: 'string', description: 'Value 2' },
        },
      },
      handles: [
        { id: 'true', label: 'true', color: '#22c55e' },
        { id: 'false', label: 'false', color: '#ef4444' },
      ],
    },
    {
      kind: 'switch',
      name: 'Switch',
      category: 'Logic',
      icon: 'split',
      color: '#f59e0b',
      settingsSchema: {
        type: 'object',
        properties: {
          expression: { type: 'string', description: 'Value to match' },
          cases: { type: 'array', items: { type: 'string' } },
        },
      },
      handles: [
        { id: 'default', label: 'default', color: '#6b7280' },
      ],
    },
    {
      kind: 'delay',
      name: 'Delay',
      category: 'Logic',
      icon: 'clock',
      color: '#6366f1',
      settingsSchema: {
        type: 'object',
        required: ['duration'],
        properties: {
          duration: { type: 'string', description: 'Delay duration, for example 10s or 5m' },
        },
      },
      handles: [
        { id: 'success', label: 'success', color: '#94a3b8' },
      ],
    },
  ] satisfies ControlStepDefinition[]
}
