import { describe, expect, it } from 'vitest'
import {
  distributeEqualTraffic,
  distributePerVersionTraffic,
  normalizeFixedTrafficInput,
} from '../app/components/dag/traffic-utils'

describe('workflow traffic settings', () => {
  it('splits equal traffic with remainder on first rows', () => {
    expect(distributeEqualTraffic([1, 2, 3])).toEqual([
      { version_id: 1, weight: 34 },
      { version_id: 2, weight: 33 },
      { version_id: 3, weight: 33 },
    ])
  })

  it('splits remaining traffic between share versions after fixed weights', () => {
    expect(distributePerVersionTraffic([
      { version_id: 3, mode: 'share' },
      { version_id: 2, mode: 'fixed', weight: 70 },
      { version_id: 1, mode: 'share' },
    ])).toEqual([
      { version_id: 3, weight: 15 },
      { version_id: 2, weight: 70 },
      { version_id: 1, weight: 15 },
    ])
  })

  it('uses shared remainder before taking the rest from other fixed rows', () => {
    expect(normalizeFixedTrafficInput(
      [
        { version_id: 1, mode: 'fixed', weight: 40 },
        { version_id: 2, mode: 'fixed', weight: 30 },
        { version_id: 3, mode: 'share' },
      ],
      1,
      80,
    )).toEqual([
      { version_id: 1, mode: 'fixed', weight: 80 },
      { version_id: 2, mode: 'fixed', weight: 20 },
      { version_id: 3, mode: 'share' },
    ])
  })

  it('uses available shared remainder before taking traffic from other fixed rows', () => {
    expect(normalizeFixedTrafficInput(
      [
        { version_id: 1, mode: 'fixed', weight: 50 },
        { version_id: 2, mode: 'fixed', weight: 40 },
        { version_id: 3, mode: 'share' },
      ],
      2,
      50,
    )).toEqual([
      { version_id: 1, mode: 'fixed', weight: 50 },
      { version_id: 2, mode: 'fixed', weight: 50 },
      { version_id: 3, mode: 'share' },
    ])
  })

  it('uses partial shared remainder before reducing other fixed rows', () => {
    expect(normalizeFixedTrafficInput(
      [
        { version_id: 1, mode: 'fixed', weight: 40 },
        { version_id: 2, mode: 'fixed', weight: 40 },
        { version_id: 3, mode: 'share' },
      ],
      2,
      70,
    )).toEqual([
      { version_id: 1, mode: 'fixed', weight: 30 },
      { version_id: 2, mode: 'fixed', weight: 70 },
      { version_id: 3, mode: 'share' },
    ])
  })

  it('takes excess traffic proportionally from other fixed rows after shared remainder is used', () => {
    expect(normalizeFixedTrafficInput(
      [
        { version_id: 1, mode: 'fixed', weight: 30 },
        { version_id: 2, mode: 'fixed', weight: 30 },
        { version_id: 3, mode: 'fixed', weight: 30 },
        { version_id: 4, mode: 'share' },
      ],
      3,
      60,
    )).toEqual([
      { version_id: 1, mode: 'fixed', weight: 20 },
      { version_id: 2, mode: 'fixed', weight: 20 },
      { version_id: 3, mode: 'fixed', weight: 60 },
      { version_id: 4, mode: 'share' },
    ])
  })
})
