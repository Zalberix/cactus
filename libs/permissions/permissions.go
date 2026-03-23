package permissions

// Permission — тип slug права доступа.
type Permission string

const (
	// Организация
	OrgRead  Permission = "org:read"
	OrgWrite Permission = "org:write"

	// Пользователи
	UserRead  Permission = "user:read"
	UserWrite Permission = "user:write"

	// Роли
	RoleRead  Permission = "role:read"
	RoleWrite Permission = "role:write"

	// Системы
	SystemRead  Permission = "system:read"
	SystemWrite Permission = "system:write"

	// Рабочие процессы
	WorkflowRead  Permission = "workflow:read"
	WorkflowWrite Permission = "workflow:write"

	// Работники
	WorkerRead  Permission = "worker:read"
	WorkerWrite Permission = "worker:write"

	// Сообщения
	MessageRead  Permission = "message:read"
	MessageWrite Permission = "message:write"
)

// All возвращает полный список всех прав доступа.
func All() []Permission {
	return []Permission{
		OrgRead, OrgWrite,
		UserRead, UserWrite,
		RoleRead, RoleWrite,
		SystemRead, SystemWrite,
		WorkflowRead, WorkflowWrite,
		WorkerRead, WorkerWrite,
		MessageRead, MessageWrite,
	}
}

// String реализует fmt.Stringer.
func (p Permission) String() string {
	return string(p)
}
