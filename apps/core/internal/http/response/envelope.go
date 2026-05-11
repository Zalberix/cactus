package response

import (
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type Meta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

type APIError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Type    string `json:"type,omitempty"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

func OKPaginated(c *gin.Context, data interface{}, total int64, page, perPage int) {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	})
}

func Fail(c *gin.Context, status int, code, message string, details ...ErrorDetail) {
	c.JSON(status, Response{
		Success: false,
		Error:   &APIError{Code: code, Message: message, Details: details},
	})
}

func BadRequest(c *gin.Context, code, message string, details ...ErrorDetail) {
	Fail(c, http.StatusBadRequest, code, message, details...)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, "NOT_FOUND", message)
}

func InternalError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
