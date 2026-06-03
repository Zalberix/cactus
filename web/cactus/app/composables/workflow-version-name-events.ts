export const workflowVersionNameUpdatedEvent = 'cactus:workflow-version-name-updated'

export interface WorkflowVersionNameUpdatedDetail {
  versionId: number
  name: string
  versionNumber?: number
}

export function notifyWorkflowVersionNameUpdated(detail: WorkflowVersionNameUpdatedDetail) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(workflowVersionNameUpdatedEvent, { detail }))
}
