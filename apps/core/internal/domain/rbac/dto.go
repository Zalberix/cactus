package rbac

import (
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/zalberix/cactus/apps/core/storage/db"
)

type UserRoleRef struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type UserResponse struct {
	ID                      int32            `json:"id"`
	OrganizationID          pgtype.Int4      `json:"organization_id"`
	LastName                string           `json:"last_name"`
	FirstName               string           `json:"first_name"`
	Patronymic              pgtype.Text      `json:"patronymic"`
	Email                   string           `json:"email"`
	ResetPasswordAfterLogin pgtype.Bool      `json:"reset_password_after_login"`
	CreatedAt               pgtype.Timestamp `json:"created_at"`
	UpdatedAt               pgtype.Timestamp `json:"updated_at"`
	Roles                   []UserRoleRef    `json:"roles"`
}

func toUserResponse(u db.User) UserResponse {
	return UserResponse{
		ID:                      u.ID,
		OrganizationID:          u.OrganizationID,
		LastName:                u.LastName,
		FirstName:               u.FirstName,
		Patronymic:              u.Patronymic,
		Email:                   u.Email,
		ResetPasswordAfterLogin: u.ResetPasswordAfterLogin,
		CreatedAt:               u.CreatedAt,
		UpdatedAt:               u.UpdatedAt,
		Roles:                   []UserRoleRef{},
	}
}

func toUserResponses(users []db.User) []UserResponse {
	out := make([]UserResponse, len(users))
	for i, u := range users {
		out[i] = toUserResponse(u)
	}
	return out
}

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
	Code string `json:"code" binding:"required,min=2,max=100"`
}

type UpdateOrgRequest struct {
	Name string `json:"name" binding:"required,min=2,max=255"`
	Code string `json:"code" binding:"required,min=2,max=100"`
}

type CreateUserRequest struct {
	LastName   string `json:"last_name" binding:"required,max=255"`
	FirstName  string `json:"first_name" binding:"required,max=255"`
	Patronymic string `json:"patronymic" binding:"max=255"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	LastName   string `json:"last_name" binding:"required,max=255"`
	FirstName  string `json:"first_name" binding:"required,max=255"`
	Patronymic string `json:"patronymic" binding:"max=255"`
	Email      string `json:"email" binding:"required,email"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=255"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=255"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type PaginationQuery struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page" binding:"max=100"`
}

func (p *PaginationQuery) Defaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PerPage == 0 {
		p.PerPage = 20
	}
}

func (p *PaginationQuery) Offset() int64 {
	return int64((p.Page - 1) * p.PerPage)
}

func (p *PaginationQuery) Limit() int64 {
	return int64(p.PerPage)
}
