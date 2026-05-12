import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const projectRoot = resolve(__dirname, '..')

const readProjectFile = (path: string) =>
  readFileSync(resolve(projectRoot, path), 'utf8')

describe('profile language controls', () => {
  it('uses Russian as the default locale', () => {
    const nuxtConfig = readProjectFile('nuxt.config.ts')

    expect(nuxtConfig).toMatch(/defaultLocale:\s*'ru'/)
  })

  it('places the language switcher in the user dropdown above logout', () => {
    const layout = readProjectFile('app/layouts/default.vue')
    const languageSwitcherPosition = layout.indexOf('<DropdownMenuRadioGroup')
    const logoutPosition = layout.indexOf("t('auth.logout')")

    expect(layout).toContain('DropdownMenuRadioGroup')
    expect(layout).toContain('DropdownMenuRadioItem')
    expect(layout).toContain("t('common.language')")
    expect(layout).toContain('@update:model-value="setLocale"')
    expect(languageSwitcherPosition).toBeGreaterThan(-1)
    expect(logoutPosition).toBeGreaterThan(-1)
    expect(languageSwitcherPosition).toBeLessThan(logoutPosition)
  })
})

describe('Russian locale coverage', () => {
  it('includes a Russian translation for the role description label', () => {
    const russianMessages = JSON.parse(readProjectFile('i18n/locales/ru.json'))

    expect(russianMessages.roles.description).toBe('Описание')
  })

  it('includes every common key used by the English locale', () => {
    const englishMessages = JSON.parse(readProjectFile('i18n/locales/en.json'))
    const russianMessages = JSON.parse(readProjectFile('i18n/locales/ru.json'))

    expect(Object.keys(russianMessages.common).sort()).toEqual(
      Object.keys(englishMessages.common).sort(),
    )
  })

  it('includes every editor key used by the English locale', () => {
    const englishMessages = JSON.parse(readProjectFile('i18n/locales/en.json'))
    const russianMessages = JSON.parse(readProjectFile('i18n/locales/ru.json'))

    expect(Object.keys(russianMessages.editor).sort()).toEqual(
      Object.keys(englishMessages.editor).sort(),
    )
  })
})
