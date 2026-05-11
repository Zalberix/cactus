export interface TrafficWeight {
  version_id: number
  weight: number
}

export type TrafficMode = 'share' | 'fixed'

export interface TrafficVersionSetting {
  version_id: number
  mode: TrafficMode
  weight?: number
}

export function distributeEqualTraffic(versionIds: number[]): TrafficWeight[] {
  if (versionIds.length === 0) return []

  const base = Math.floor(100 / versionIds.length)
  const remainder = 100 - base * versionIds.length

  return versionIds.map((version_id, index) => ({
    version_id,
    weight: base + (index < remainder ? 1 : 0),
  }))
}

export function clampTrafficWeight(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.max(0, Math.min(100, Math.round(value)))
}

export function distributePerVersionTraffic(settings: TrafficVersionSetting[]): TrafficWeight[] {
  const fixedTotal = settings.reduce(
    (sum, item) => sum + (item.mode === 'fixed' ? clampTrafficWeight(item.weight ?? 0) : 0),
    0,
  )
  const shareItems = settings.filter(item => item.mode === 'share')
  const remaining = Math.max(0, 100 - fixedTotal)
  const base = shareItems.length > 0 ? Math.floor(remaining / shareItems.length) : 0
  let remainder = shareItems.length > 0 ? remaining - base * shareItems.length : 0

  return settings.map((item) => {
    if (item.mode === 'fixed') {
      return { version_id: item.version_id, weight: clampTrafficWeight(item.weight ?? 0) }
    }

    const weight = base + (remainder > 0 ? 1 : 0)
    if (remainder > 0) remainder--
    return { version_id: item.version_id, weight }
  })
}

export function normalizeFixedTrafficInput(
  settings: TrafficVersionSetting[],
  editedVersionId: number,
  nextWeight: number,
): TrafficVersionSetting[] {
  const previousEdited = settings.find(item => item.version_id === editedVersionId)
  const previousWeight = previousEdited?.mode === 'fixed' ? clampTrafficWeight(previousEdited.weight ?? 0) : 0
  const requestedWeight = clampTrafficWeight(nextWeight)
  const increase = Math.max(0, requestedWeight - previousWeight)
  const previousFixedTotal = settings.reduce(
    (sum, item) => sum + (item.mode === 'fixed' ? clampTrafficWeight(item.weight ?? 0) : 0),
    0,
  )
  const freeTraffic = Math.max(0, 100 - previousFixedTotal)
  let fixedTrafficToTake = Math.max(0, increase - freeTraffic)

  const next = settings.map((item) => {
    if (item.version_id !== editedVersionId) return { ...item }
    return { version_id: item.version_id, mode: 'fixed' as const, weight: requestedWeight }
  })

  const otherFixedItems = next
    .filter(item => item.version_id !== editedVersionId && item.mode === 'fixed')
    .map(item => ({
      item,
      current: clampTrafficWeight(item.weight ?? 0),
      baseReduction: 0,
      fraction: 0,
    }))
    .filter(entry => entry.current > 0)

  const otherFixedTotal = otherFixedItems.reduce((sum, entry) => sum + entry.current, 0)
  if (fixedTrafficToTake > 0 && otherFixedTotal > 0) {
    const targetReduction = Math.min(fixedTrafficToTake, otherFixedTotal)
    let assignedReduction = 0

    for (const entry of otherFixedItems) {
      const exactReduction = (entry.current / otherFixedTotal) * targetReduction
      entry.baseReduction = Math.floor(exactReduction)
      entry.fraction = exactReduction - entry.baseReduction
      assignedReduction += entry.baseReduction
    }

    let remainingReduction = targetReduction - assignedReduction
    const byRemainder = [...otherFixedItems].sort((a, b) => {
      if (b.fraction !== a.fraction) return b.fraction - a.fraction
      return a.item.version_id - b.item.version_id
    })
    for (const entry of byRemainder) {
      if (remainingReduction <= 0) break
      if (entry.baseReduction >= entry.current) continue
      entry.baseReduction++
      remainingReduction--
    }

    for (const entry of otherFixedItems) {
      entry.item.weight = entry.current - entry.baseReduction
    }
  }

  const fixedTotal = next.reduce(
    (sum, item) => sum + (item.mode === 'fixed' ? clampTrafficWeight(item.weight ?? 0) : 0),
    0,
  )
  if (fixedTotal > 100) {
    const edited = next.find(item => item.version_id === editedVersionId)
    if (edited && edited.mode === 'fixed') {
      edited.weight = clampTrafficWeight((edited.weight ?? 0) - (fixedTotal - 100))
    }
  }

  return next
}
