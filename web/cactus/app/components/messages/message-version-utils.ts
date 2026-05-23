export interface MessageWorkflowVersionInfo {
  workflow_name?: string | null
  workflow_version_name?: string | null
  workflow_version_number?: number | null
}

export interface WorkflowVersionLabelParts {
  name: string
  number: string
}

export function getWorkflowVersionLabelParts(info: MessageWorkflowVersionInfo): WorkflowVersionLabelParts {
  return {
    name: info.workflow_version_name?.trim() ?? '',
    number: typeof info.workflow_version_number === 'number'
      ? `v${info.workflow_version_number}`
      : '',
  }
}

export function formatWorkflowVersionLabel(info: MessageWorkflowVersionInfo): string {
  const { name, number } = getWorkflowVersionLabelParts(info)

  if (name && number) return `${name} ${number}`
  if (number) return number
  if (name) return name
  return '-'
}

export function formatMessageWorkflowSubtitle(info: MessageWorkflowVersionInfo): string {
  const workflowName = info.workflow_name?.trim()
  const versionLabel = formatWorkflowVersionLabel(info)

  if (workflowName && versionLabel !== '-') return `${workflowName} (${versionLabel})`
  if (workflowName) return workflowName
  return versionLabel === '-' ? '' : versionLabel
}
