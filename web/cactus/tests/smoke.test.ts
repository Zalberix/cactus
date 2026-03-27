import { describe, it, expect } from 'vitest'

describe('smoke test', () => {
  it('vitest runs successfully', () => {
    expect(true).toBe(true)
  })

  it('basic math works', () => {
    expect(1 + 1).toBe(2)
  })
})
