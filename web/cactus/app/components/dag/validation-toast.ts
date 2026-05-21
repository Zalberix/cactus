import type { ValidationIssue } from '~/composables/useVersions'

export function validationToastDescription(
  errors: ValidationIssue[],
  fallback: string,
): string {
  const firstMessage = errors
    .map(error => error.message?.trim())
    .find((message): message is string => Boolean(message))

  return firstMessage ?? fallback
}
