import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const projectRoot = resolve(__dirname, '..')

const readProjectFile = (path: string) =>
  readFileSync(resolve(projectRoot, path), 'utf8')

describe('DAG viewer styles', () => {
  it('loads Vue Flow core styles globally for every page that renders DagCanvas', () => {
    const appStyles = readProjectFile('app/assets/css/tailwind.css')

    expect(appStyles).toContain('@vue-flow/core/dist/style.css')
  })

  it('uses a slightly gray editor canvas background', () => {
    const canvas = readProjectFile('app/components/dag/DagCanvas.vue')

    expect(canvas).toContain('class="bg-slate-100"')
    expect(canvas).toContain('<Background :gap="GRID_SIZE" color="#cbd5e1" />')
  })
})
