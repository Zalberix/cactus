package rbac

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

// AuthService — интерфейс для хеширования паролей (реализует auth.Service).
type AuthService interface {
	HashPassword(password string) (string, error)
}

// Service — бизнес-логика управления организациями, пользователями и ролями.
type Service struct {
	store       Storage
	authService AuthService
}

// NewService создаёт новый RBAC сервис.
func NewService(store Storage, authService AuthService) *Service {
	return &Service{store: store, authService: authService}
}

// --- Organization methods ---

// CreateOrg создаёт новую организацию.
func (s *Service) CreateOrg(ctx context.Context, req CreateOrgRequest) (db.Organization, error) {
	return s.store.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: req.Name,
		Code: req.Code,
	})
}

// ListOrgs возвращает постраничный список организаций.
func (s *Service) ListOrgs(ctx context.Context, pq PaginationQuery) ([]db.Organization, int64, error) {
	pq.Defaults()
	orgs, err := s.store.ListOrganizations(ctx, db.ListOrganizationsParams{
		Limit:  pq.Limit(),
		Offset: pq.Offset(),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list organizations: %w", err)
	}
	total, err := s.store.CountOrganizations(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count organizations: %w", err)
	}
	return orgs, total, nil
}

// GetOrg возвращает организацию по ID.
func (s *Service) GetOrg(ctx context.Context, id int32) (db.Organization, error) {
	return s.store.GetOrganizationByID(ctx, id)
}

// UpdateOrg обновляет организацию.
func (s *Service) UpdateOrg(ctx context.Context, id int32, req UpdateOrgRequest) (db.Organization, error) {
	return s.store.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:   id,
		Name: req.Name,
		Code: req.Code,
	})
}

// DeleteOrg мягко удаляет организацию.
func (s *Service) DeleteOrg(ctx context.Context, id int32) error {
	return s.store.SoftDeleteOrganization(ctx, id)
}

// --- User methods ---

// CreateUser создаёт нового пользователя (только для admin, D-02).
// Пароль хешируется через bcrypt.
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (db.User, error) {
	hashed, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return db.User{}, fmt.Errorf("hash password: %w", err)
	}
	return s.store.CreateUser(ctx, db.CreateUserParams{
		LastName:  req.LastName,
		FirstName: req.FirstName,
		Patronymic: pgtype.Text{
			String: req.Patronymic,
			Valid:  req.Patronymic != "",
		},
		Email:    req.Email,
		Password: hashed,
		ResetPasswordAfterLogin: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	})
}

// ListUsersByOrg возвращает постраничный список пользователей организации.
func (s *Service) ListUsersByOrg(ctx context.Context, orgID int32, pq PaginationQuery) ([]db.User, int64, error) {
	pq.Defaults()
	orgPg := pgtype.Int4{Int32: orgID, Valid: true}
	users, err := s.store.ListUsersByOrgID(ctx, db.ListUsersByOrgIDParams{
		OrganizationID: orgPg,
		Limit:          pq.Limit(),
		Offset:         pq.Offset(),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	total, err := s.store.CountUsersByOrgID(ctx, orgPg)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	return users, total, nil
}

// UpdateUser обновляет данные пользователя.
func (s *Service) UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (db.User, error) {
	return s.store.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		LastName:  req.LastName,
		FirstName: req.FirstName,
		Patronymic: pgtype.Text{
			String: req.Patronymic,
			Valid:  req.Patronymic != "",
		},
		Email: req.Email,
	})
}

// DeleteUser мягко удаляет пользователя.
func (s *Service) DeleteUser(ctx context.Context, id int32) error {
	return s.store.SoftDeleteUser(ctx, id)
}

// --- Role methods ---

// CreateRole создаёт роль с набором прав в одной транзакции.
func (s *Service) CreateRole(ctx context.Context, orgID int32, req CreateRoleRequest) (db.Role, error) {
	var role db.Role
	err := s.store.WithTx(ctx, func(q *db.Queries) error {
		var txErr error
		role, txErr = q.CreateRole(ctx, db.CreateRoleParams{
			OrganizationID: pgtype.Int4{Int32: orgID, Valid: true},
			Name:           req.Name,
			Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
			IsSystem:       false,
		})
		if txErr != nil {
			return fmt.Errorf("create role: %w", txErr)
		}
		for _, slug := range req.Permissions {
			perm, txErr := q.GetPermissionBySlug(ctx, slug)
			if txErr != nil {
				return fmt.Errorf("get permission %q: %w", slug, txErr)
			}
			if txErr = q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{
				PermissionID: perm.ID,
				RoleID:       role.ID,
			}); txErr != nil {
				return fmt.Errorf("add permission %q to role: %w", slug, txErr)
			}
		}
		return nil
	})
	return role, err
}

// ListRolesByOrg возвращает роли организации.
func (s *Service) ListRolesByOrg(ctx context.Context, orgID int32) ([]db.Role, error) {
	return s.store.ListRolesByOrgID(ctx, pgtype.Int4{Int32: orgID, Valid: true})
}

// ListRolesByUser возвращает роли пользователя.
func (s *Service) ListRolesByUser(ctx context.Context, userID int32) ([]db.Role, error) {
	return s.store.ListRolesByUserID(ctx, userID)
}

// UpdateRole обновляет роль и заменяет набор прав в транзакции.
func (s *Service) UpdateRole(ctx context.Context, id int32, req UpdateRoleRequest) (db.Role, error) {
	var role db.Role
	err := s.store.WithTx(ctx, func(q *db.Queries) error {
		var txErr error
		role, txErr = q.UpdateRole(ctx, db.UpdateRoleParams{
			ID:          id,
			Name:        req.Name,
			Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		})
		if txErr != nil {
			return fmt.Errorf("update role: %w", txErr)
		}

		// Получаем текущие права
		currentPerms, txErr := q.ListPermissionsByRoleID(ctx, id)
		if txErr != nil {
			return fmt.Errorf("list current permissions: %w", txErr)
		}

		// Удаляем все текущие права
		for _, p := range currentPerms {
			if txErr = q.RemovePermissionFromRole(ctx, db.RemovePermissionFromRoleParams{
				PermissionID: p.ID,
				RoleID:       id,
			}); txErr != nil {
				return fmt.Errorf("remove permission: %w", txErr)
			}
		}

		// Добавляем новые права
		for _, slug := range req.Permissions {
			perm, txErr := q.GetPermissionBySlug(ctx, slug)
			if txErr != nil {
				return fmt.Errorf("get permission %q: %w", slug, txErr)
			}
			if txErr = q.AddPermissionToRole(ctx, db.AddPermissionToRoleParams{
				PermissionID: perm.ID,
				RoleID:       id,
			}); txErr != nil {
				return fmt.Errorf("add permission %q: %w", slug, txErr)
			}
		}
		return nil
	})
	return role, err
}

// DeleteRole удаляет роль (hard delete — таблица role не имеет deleted_at).
func (s *Service) DeleteRole(ctx context.Context, id int32) error {
	return s.store.DeleteRole(ctx, id)
}

// --- Role assignment ---

// AssignRole назначает роль пользователю.
func (s *Service) AssignRole(ctx context.Context, roleID, userID int32) error {
	return s.store.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		RoleID: roleID,
		UserID: userID,
	})
}

// RemoveRole снимает роль с пользователя.
func (s *Service) RemoveRole(ctx context.Context, roleID, userID int32) error {
	return s.store.RemoveRoleFromUser(ctx, db.RemoveRoleFromUserParams{
		RoleID: roleID,
		UserID: userID,
	})
}

// ListRolePermissions возвращает права роли.
func (s *Service) ListRolePermissions(ctx context.Context, roleID int32) ([]db.Permission, error) {
	return s.store.ListPermissionsByRoleID(ctx, roleID)
}
