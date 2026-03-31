package rbac

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
