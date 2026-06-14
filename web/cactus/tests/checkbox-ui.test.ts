import { readdirSync, readFileSync, statSync } from 'node:fs'
import { relative, resolve, sep } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('checkbox UI usage', () => {
  it('uses the shared Checkbox component instead of native checkbox inputs', () => {
    const appRoot = resolve(__dirname, '../app')
    const allowedFile = ['components', 'ui', 'checkbox', 'Checkbox.vue'].join(sep)
    const nativeCheckboxFiles = vueFiles(appRoot)
      .filter(file => !relative(appRoot, file).endsWith(allowedFile))
      .filter(file => /\btype\s*=\s*["']checkbox["']/.test(readFileSync(file, 'utf8')))
      .map(file => relative(appRoot, file).replaceAll('\\', '/'))

    expect(nativeCheckboxFiles).toEqual([])
  })

  it('uses the modelValue API for the shared Checkbox component', () => {
    const appRoot = resolve(__dirname, '../app')
    const oldApiFiles = vueFiles(appRoot)
      .filter(file => {
        const source = readFileSync(file, 'utf8')
        return /<Checkbox[\s\S]*?(?::checked|@update:checked|v-model:checked)/.test(source)
      })
      .map(file => relative(appRoot, file).replaceAll('\\', '/'))

    expect(oldApiFiles).toEqual([])
  })
})

function vueFiles(root: string): string[] {
  const entries = readdirSync(root)
  const files: string[] = []

  for (const entry of entries) {
    const path = resolve(root, entry)
    const stat = statSync(path)
    if (stat.isDirectory()) {
      files.push(...vueFiles(path))
    }
    else if (entry.endsWith('.vue')) {
      files.push(path)
    }
  }

  return files
}
