package rbac

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zalberix/cactus/apps/core/internal/http/middleware"
	"github.com/zalberix/cactus/apps/core/internal/http/response"
	"github.com/zalberix/cactus/libs/permissions"
)

// Handler — HTTP-обработчики RBAC (организации, пользователи, роли).
type Handler struct {
	service *Service
}

// NewHandler создаёт новый RBAC Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все RBAC маршруты.
// Все маршруты защищены JWT auth middleware и permission middleware.
func (h *Handler) RegisterRoutes(r *gin.Engine, authMw gin.HandlerFunc, permChecker middleware.PermissionChecker) {
	v1 := r.Group("/api/v1", authMw)

	// Organizations
	v1.GET("/organizations", middleware.RequirePermission(permChecker, permissions.OrgRead), h.ListOrgs)
	v1.POST("/organizations", middleware.RequirePermission(permChecker, permissions.OrgWrite), h.CreateOrg)
	v1.GET("/organizations/:orgId", middleware.RequirePermission(permChecker, permissions.OrgRead), h.GetOrg)
	v1.PUT("/organizations/:orgId", middleware.RequirePermission(permChecker, permissions.OrgWrite), h.UpdateOrg)
	v1.DELETE("/organizations/:orgId", middleware.RequirePermission(permChecker, permissions.OrgWrite), h.DeleteOrg)

	// Users (nested under org)
	v1.GET("/organizations/:orgId/users", middleware.RequirePermission(permChecker, permissions.UserRead), h.ListUsers)
	v1.POST("/organizations/:orgId/users", middleware.RequirePermission(permChecker, permissions.UserWrite), h.CreateUser)
	v1.PUT("/users/:userId", middleware.RequirePermission(permChecker, permissions.UserWrite), h.UpdateUser)
	v1.DELETE("/users/:userId", middleware.RequirePermission(permChecker, permissions.UserWrite), h.DeleteUser)

	// Roles (nested under org)
	v1.GET("/organizations/:orgId/roles", middleware.RequirePermission(permChecker, permissions.RoleRead), h.ListRoles)
	v1.POST("/organizations/:orgId/roles", middleware.RequirePermission(permChecker, permissions.RoleWrite), h.CreateRole)
	v1.PUT("/roles/:roleId", middleware.RequirePermission(permChecker, permissions.RoleWrite), h.UpdateRole)
	v1.DELETE("/roles/:roleId", middleware.RequirePermission(permChecker, permissions.RoleWrite), h.DeleteRole)
	v1.GET("/roles/:roleId/permissions", middleware.RequirePermission(permChecker, permissions.RoleRead), h.ListRolePermissions)

	// Role assignment
	v1.POST("/roles/:roleId/users/:userId", middleware.RequirePermission(permChecker, permissions.RoleWrite), h.AssignRole)
	v1.DELETE("/roles/:roleId/users/:userId", middleware.RequirePermission(permChecker, permissions.RoleWrite), h.RemoveRole)
}

// parseID извлекает int32 ID из параметра URL.
func parseID(c *gin.Context, param string) (int32, bool) {
	raw := c.Param(param)
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		response.BadRequest(c, "INVALID_PARAM", "Неверный формат ID: "+param)
		return 0, false
	}
	return int32(id), true
}

// --- Organization handlers ---

// ListOrgs godoc
// GET /api/v1/organizations
func (h *Handler) ListOrgs(c *gin.Context) {
	var pq PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		response.BadRequest(c, "INVALID_QUERY", err.Error())
		return
	}
	pq.Defaults()

	orgs, total, err := h.service.ListOrgs(c.Request.Context(), pq)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка получения организаций")
		return
	}
	response.OKPaginated(c, orgs, total, pq.Page, pq.PerPage)
}

// CreateOrg godoc
// POST /api/v1/organizations
func (h *Handler) CreateOrg(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	org, err := h.service.CreateOrg(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка создания организации")
		return
	}
	response.Created(c, org)
}

// GetOrg godoc
// GET /api/v1/organizations/:orgId
func (h *Handler) GetOrg(c *gin.Context) {
	id, ok := parseID(c, "orgId")
	if !ok {
		return
	}

	org, err := h.service.GetOrg(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Организация не найдена")
		return
	}
	response.OK(c, org)
}

// UpdateOrg godoc
// PUT /api/v1/organizations/:orgId
func (h *Handler) UpdateOrg(c *gin.Context) {
	id, ok := parseID(c, "orgId")
	if !ok {
		return
	}
	var req UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	org, err := h.service.UpdateOrg(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка обновления организации")
		return
	}
	response.OK(c, org)
}

// DeleteOrg godoc
// DELETE /api/v1/organizations/:orgId
func (h *Handler) DeleteOrg(c *gin.Context) {
	id, ok := parseID(c, "orgId")
	if !ok {
		return
	}

	if err := h.service.DeleteOrg(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка удаления организации")
		return
	}
	response.OK(c, gin.H{"message": "Организация удалена"})
}

// --- User handlers ---

// ListUsers godoc
// GET /api/v1/organizations/:orgId/users
func (h *Handler) ListUsers(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}
	var pq PaginationQuery
	if err := c.ShouldBindQuery(&pq); err != nil {
		response.BadRequest(c, "INVALID_QUERY", err.Error())
		return
	}
	pq.Defaults()

	users, total, err := h.service.ListUsersByOrg(c.Request.Context(), orgID, pq)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка получения пользователей")
		return
	}

	result := toUserResponses(users)
	for i, u := range users {
		roles, err := h.service.ListRolesByUser(c.Request.Context(), u.ID)
		if err != nil {
			continue
		}
		refs := make([]UserRoleRef, len(roles))
		for j, r := range roles {
			refs[j] = UserRoleRef{ID: r.ID, Name: r.Name}
		}
		result[i].Roles = refs
	}

	response.OKPaginated(c, result, total, pq.Page, pq.PerPage)
}

// CreateUser godoc
// POST /api/v1/organizations/:orgId/users
// Per D-02: only admin can create users.
func (h *Handler) CreateUser(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrOrganizationNotFound) {
			response.NotFound(c, "Организация не найдена")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка создания пользователя")
		return
	}
	response.Created(c, toUserResponse(user))
}

// UpdateUser godoc
// PUT /api/v1/users/:userId
func (h *Handler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c, "userId")
	if !ok {
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка обновления пользователя")
		return
	}
	response.OK(c, toUserResponse(user))
}

// DeleteUser godoc
// DELETE /api/v1/users/:userId
func (h *Handler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c, "userId")
	if !ok {
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка удаления пользователя")
		return
	}
	response.OK(c, gin.H{"message": "Пользователь удалён"})
}

// --- Role handlers ---

// ListRoles godoc
// GET /api/v1/organizations/:orgId/roles
func (h *Handler) ListRoles(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}

	roles, err := h.service.ListRolesByOrg(c.Request.Context(), orgID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка получения ролей")
		return
	}

	type roleWithPerms struct {
		ID          int32    `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		IsSystem    bool     `json:"is_system"`
		Permissions []string `json:"permissions"`
	}

	result := make([]roleWithPerms, 0, len(roles))
	for _, r := range roles {
		perms, permErr := h.service.ListRolePermissions(c.Request.Context(), r.ID)
		if permErr != nil {
			perms = nil
		}
		slugs := make([]string, 0, len(perms))
		for _, p := range perms {
			slugs = append(slugs, p.Slug)
		}
		desc := ""
		if r.Description.Valid {
			desc = r.Description.String
		}
		result = append(result, roleWithPerms{
			ID:          r.ID,
			Name:        r.Name,
			Description: desc,
			IsSystem:    r.IsSystem,
			Permissions: slugs,
		})
	}
	response.OK(c, result)
}

// CreateRole godoc
// POST /api/v1/organizations/:orgId/roles
func (h *Handler) CreateRole(c *gin.Context) {
	orgID, ok := parseID(c, "orgId")
	if !ok {
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	role, err := h.service.CreateRole(c.Request.Context(), orgID, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка создания роли")
		return
	}
	response.Created(c, role)
}

// UpdateRole godoc
// PUT /api/v1/roles/:roleId
func (h *Handler) UpdateRole(c *gin.Context) {
	id, ok := parseID(c, "roleId")
	if !ok {
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_BODY", err.Error())
		return
	}

	role, err := h.service.UpdateRole(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка обновления роли")
		return
	}
	response.OK(c, role)
}

// DeleteRole godoc
// DELETE /api/v1/roles/:roleId
func (h *Handler) DeleteRole(c *gin.Context) {
	id, ok := parseID(c, "roleId")
	if !ok {
		return
	}

	if err := h.service.DeleteRole(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка удаления роли")
		return
	}
	response.OK(c, gin.H{"message": "Роль удалена"})
}

// ListRolePermissions godoc
// GET /api/v1/roles/:roleId/permissions
func (h *Handler) ListRolePermissions(c *gin.Context) {
	id, ok := parseID(c, "roleId")
	if !ok {
		return
	}

	perms, err := h.service.ListRolePermissions(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка получения прав роли")
		return
	}
	response.OK(c, perms)
}

// --- Role assignment handlers ---

// AssignRole godoc
// POST /api/v1/roles/:roleId/users/:userId
func (h *Handler) AssignRole(c *gin.Context) {
	roleID, ok := parseID(c, "roleId")
	if !ok {
		return
	}
	userID, ok := parseID(c, "userId")
	if !ok {
		return
	}

	if err := h.service.AssignRole(c.Request.Context(), roleID, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка назначения роли")
		return
	}
	response.OK(c, gin.H{"message": "Роль назначена"})
}

// RemoveRole godoc
// DELETE /api/v1/roles/:roleId/users/:userId
func (h *Handler) RemoveRole(c *gin.Context) {
	roleID, ok := parseID(c, "roleId")
	if !ok {
		return
	}
	userID, ok := parseID(c, "userId")
	if !ok {
		return
	}

	if err := h.service.RemoveRole(c.Request.Context(), roleID, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка снятия роли")
		return
	}
	response.OK(c, gin.H{"message": "Роль снята"})
}
