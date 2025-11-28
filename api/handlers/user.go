package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/queue"
	"github.com/itsHenry35/SambaManager/services"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	service *services.SambaService
	queue   *queue.Queue
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *services.SambaService, q *queue.Queue) *UserHandler {
	return &UserHandler{
		service: service,
		queue:   q,
	}
}

// CreateUser creates a new system user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Convert to internal user model
	user := &types.User{
		Username: req.Username,
		Password: req.Password,
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.CreateUser(user)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseCreated(c, gin.H{
		"username": user.Username,
	})
}

// DeleteUser deletes a system user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.ResponseBadRequest(c, "Username is required")
		return
	}

	var req types.DeleteUserRequest
	// If no body provided, default to deleting home directory (backward compatible)
	if err := c.ShouldBindJSON(&req); err != nil {
		req.DeleteHomeDir = true
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.DeleteUser(username, req.DeleteHomeDir)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "User deleted successfully")
}

// ListUsers retrieves system users with pagination and search
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query types.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}
	query = query.GetDefaults()

	users, err := h.service.ListUsers()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	// Filter by search query
	filteredUsers := users
	if query.Search != "" {
		filteredUsers = []types.UserResponse{}
		searchLower := strings.ToLower(query.Search)
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Username), searchLower) {
				filteredUsers = append(filteredUsers, user)
			}
		}
	}

	// Calculate pagination
	total := len(filteredUsers)
	offset := query.GetOffset()
	end := offset + query.PageSize
	if end > total {
		end = total
	}

	// Get paginated slice
	paginatedUsers := []types.UserResponse{}
	if offset < total {
		paginatedUsers = filteredUsers[offset:end]
	}

	utils.ResponsePaginated(c, paginatedUsers, total, query.Page, query.PageSize)
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.ResponseBadRequest(c, "Username is required")
		return
	}

	// Basic validation - detailed validation happens in service layer
	if len(username) == 0 || len(username) > 32 {
		utils.ResponseBadRequest(c, "Invalid username length")
		return
	}

	var req types.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.ChangePassword(username, req.Password)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Password changed successfully")
}

// ListOrphanedDirectories lists home directories without corresponding users
func (h *UserHandler) ListOrphanedDirectories(c *gin.Context) {
	orphanedDirs, err := h.service.ListOrphanedDirectories()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseOK(c, orphanedDirs)
}

// DeleteOrphanedDirectory deletes an orphaned home directory
func (h *UserHandler) DeleteOrphanedDirectory(c *gin.Context) {
	dirName := c.Param("dirName")
	if dirName == "" {
		utils.ResponseBadRequest(c, "Directory name is required")
		return
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.DeleteOrphanedDirectory(dirName)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Orphaned directory deleted successfully")
}

// SearchUsers searches for users by username (for autocomplete)
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ResponseOK(c, []types.UserResponse{})
		return
	}

	limit := 10 // Max 10 results for autocomplete
	users, err := h.service.ListUsers()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	// Filter by search query
	searchLower := strings.ToLower(query)
	results := []types.UserResponse{}
	for _, user := range users {
		if strings.Contains(strings.ToLower(user.Username), searchLower) {
			results = append(results, user)
			if len(results) >= limit {
				break
			}
		}
	}

	utils.ResponseOK(c, results)
}
