import { readdirSync, readFileSync, statSync } from 'node:fs'
import { relative, resolve, sep } from 'node:path'
import { describe, expect, it } from 'vitest'

describe('select UI usage', () => {
  it('uses the shared Select component instead of native select elements', () => {
    const appRoot = resolve(__dirname, '../app')
    const allowedDir = ['components', 'ui', 'select'].join(sep)
    const nativeSelectFiles = vueFiles(appRoot)
      .filter(file => !relative(appRoot, file).startsWith(`${allowedDir}${sep}`))
      .filter(file => /<select[\s>]/.test(readFileSync(file, 'utf8')))
      .map(file => relative(appRoot, file).replaceAll('\\', '/'))

    expect(nativeSelectFiles).toEqual([])
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
