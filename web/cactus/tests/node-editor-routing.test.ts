import { describe, expect, it } from 'vitest'
import { editorSurfaceForStep } from '../app/components/dag/editor-utils'

describe('node editor routing', () => {
  it('opens NodeEditor for every non-start step', () => {
    expect(editorSurfaceForStep({ stepType: 'task' })).toBe('node-editor')
    expect(editorSurfaceForStep({ stepType: 'control', controlKind: 'condition' })).toBe('node-editor')
    expect(editorSurfaceForStep({ stepType: 'control', controlKind: 'switch' })).toBe('node-editor')
    expect(editorSurfaceForStep({ stepType: 'control', controlKind: 'delay' })).toBe('node-editor')
    expect(editorSurfaceForStep({ stepType: 'control', controlKind: 'start' })).toBe('none')
  })
})
