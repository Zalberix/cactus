export function sortVersionsForDisplay<T extends { is_active: boolean, version_number: number }>(versions: T[]) {
  return [...versions].sort((a, b) => {
    if (a.is_active !== b.is_active) return a.is_active ? -1 : 1
    return b.version_number - a.version_number
  })
}
