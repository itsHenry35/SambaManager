package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// ResponseOK sends a successful response with data
// Always returns HTTP 200 with code 200 in JSON
func ResponseOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// ResponseCreated sends a successful creation response
// Always returns HTTP 200 with code 201 in JSON
func ResponseCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusCreated,
		Message: "created successfully",
		Data:    data,
	})
}

// ResponseSuccessWithCustomMessage sends a successful response with custom message
func ResponseSuccessWithCustomMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    nil,
	})
}

// ResponseSuccessWithMessageAndData sends a successful response with custom message and data
func ResponseSuccessWithMessageAndData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// ResponsePaginated sends a paginated response
func ResponsePaginated(c *gin.Context, data interface{}, total, page, size int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// ResponseError sends an error response
// Always returns HTTP 200 with error code in JSON
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ResponseBadRequest sends a bad request error (400)
func ResponseBadRequest(c *gin.Context, message string) {
	ResponseError(c, http.StatusBadRequest, message)
}

// ResponseUnauthorized sends an unauthorized error (401)
func ResponseUnauthorized(c *gin.Context, message string) {
	ResponseError(c, http.StatusUnauthorized, message)
}

// ResponseForbidden sends a forbidden error (403)
func ResponseForbidden(c *gin.Context, message string) {
	ResponseError(c, http.StatusForbidden, message)
}

// ResponseNotFound sends a not found error (404)
func ResponseNotFound(c *gin.Context, message string) {
	ResponseError(c, http.StatusNotFound, message)
}

// ResponseInternalServerError sends an internal server error (500)
func ResponseInternalServerError(c *gin.Context, message string) {
	ResponseError(c, http.StatusInternalServerError, message)
}
