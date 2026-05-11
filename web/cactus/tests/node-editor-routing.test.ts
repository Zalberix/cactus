import { describe, expect, it } from 'vitest'
import { editorSurfaceForStep } from '../app/components/dag/editor-utils'

describe('node editor routing', () => {
  it('opens NodeEditor for task and StepPanel for control', () => {
    expect(editorSurfaceForStep({ stepType: 'task' })).toBe('node-editor')
    expect(editorSurfaceForStep({ stepType: 'control', controlKind: 'condition' })).toBe('step-panel')
  })
})
