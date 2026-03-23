package rbac

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// Storage — интерфейс хранилища для RBAC домена.
// Реализуется *store.Store через встроенный *db.Queries.
type Storage interface {
	// Organization CRUD
	CreateOrganization(ctx context.Context, arg db.CreateOrganizationParams) (db.Organization, error)
	GetOrganizationByID(ctx context.Context, id int32) (db.Organization, error)
	ListOrganizations(ctx context.Context, arg db.ListOrganizationsParams) ([]db.Organization, error)
	CountOrganizations(ctx context.Context) (int64, error)
	UpdateOrganization(ctx context.Context, arg db.UpdateOrganizationParams) (db.Organization, error)
	SoftDeleteOrganization(ctx context.Context, id int32) error

	// User CRUD
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	ListUsersByOrgID(ctx context.Context, arg db.ListUsersByOrgIDParams) ([]db.User, error)
	CountUsersByOrgID(ctx context.Context, organizationID pgtype.Int4) (int64, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	SoftDeleteUser(ctx context.Context, id int32) error

	// Role CRUD (role table has no deleted_at — uses hard delete)
	CreateRole(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
	GetRoleByID(ctx context.Context, id int32) (db.Role, error)
	ListRolesByOrgID(ctx context.Context, organizationID pgtype.Int4) ([]db.Role, error)
	UpdateRole(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
	DeleteRole(ctx context.Context, id int32) error

	// Permission-Role
	AddPermissionToRole(ctx context.Context, arg db.AddPermissionToRoleParams) error
	RemovePermissionFromRole(ctx context.Context, arg db.RemovePermissionFromRoleParams) error
	ListPermissionsByRoleID(ctx context.Context, roleID int32) ([]db.Permission, error)

	// Role-User
	AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) error
	RemoveRoleFromUser(ctx context.Context, arg db.RemoveRoleFromUserParams) error
	ListRolesByUserID(ctx context.Context, userID int32) ([]db.Role, error)

	// Permission lookup by slug
	GetPermissionBySlug(ctx context.Context, slug string) (db.Permission, error)

	// Transaction support
	WithTx(ctx context.Context, fn func(q *db.Queries) error) error
}
