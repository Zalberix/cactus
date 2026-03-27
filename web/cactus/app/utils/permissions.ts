export type Permission =
  | 'org:read' | 'org:write'
  | 'user:read' | 'user:write'
  | 'role:read' | 'role:write'
  | 'system:read' | 'system:write'
  | 'workflow:read' | 'workflow:write'
  | 'worker:read' | 'worker:write'
  | 'message:read' | 'message:write'

export interface PermissionGroup {
  domain: string
  label: string
  permissions: { value: Permission; label: string }[]
}

export const PERMISSION_GROUPS: PermissionGroup[] = [
  {
    domain: 'org',
    label: 'Organizations',
    permissions: [
      { value: 'org:read', label: 'View organizations' },
      { value: 'org:write', label: 'Manage organizations' },
    ],
  },
  {
    domain: 'user',
    label: 'Users',
    permissions: [
      { value: 'user:read', label: 'View users' },
      { value: 'user:write', label: 'Manage users' },
    ],
  },
  {
    domain: 'role',
    label: 'Roles',
    permissions: [
      { value: 'role:read', label: 'View roles' },
      { value: 'role:write', label: 'Manage roles' },
    ],
  },
  {
    domain: 'system',
    label: 'Systems',
    permissions: [
      { value: 'system:read', label: 'View systems' },
      { value: 'system:write', label: 'Manage systems' },
    ],
  },
  {
    domain: 'workflow',
    label: 'Workflows',
    permissions: [
      { value: 'workflow:read', label: 'View workflows' },
      { value: 'workflow:write', label: 'Manage workflows' },
    ],
  },
  {
    domain: 'worker',
    label: 'Workers',
    permissions: [
      { value: 'worker:read', label: 'View workers' },
      { value: 'worker:write', label: 'Manage workers' },
    ],
  },
  {
    domain: 'message',
    label: 'Messages',
    permissions: [
      { value: 'message:read', label: 'View messages' },
      { value: 'message:write', label: 'Send messages' },
    ],
  },
]

export const ALL_PERMISSIONS: Permission[] = PERMISSION_GROUPS.flatMap(
  g => g.permissions.map(p => p.value),
)
