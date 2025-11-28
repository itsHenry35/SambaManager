package handlers

import (
	"os/exec"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/api/middlewares"
	"github.com/itsHenry35/SambaManager/queue"
	"github.com/itsHenry35/SambaManager/services"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// UserProfileHandler handles user profile-related HTTP requests (for users to manage themselves)
type UserProfileHandler struct {
	service *services.SambaService
	queue   *queue.Queue
}

// NewUserProfileHandler creates a new user profile handler
func NewUserProfileHandler(service *services.SambaService, q *queue.Queue) *UserProfileHandler {
	return &UserProfileHandler{
		service: service,
		queue:   q,
	}
}

// Username validation regex to prevent command injection
var userProfileUsernameValidationRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

// ChangeOwnPassword allows user to change their own password
func (h *UserProfileHandler) ChangeOwnPassword(c *gin.Context) {
	username, exists := middlewares.GetUsernameFromContext(c)
	if !exists {
		utils.ResponseUnauthorized(c, "User not found in context")
		return
	}

	// Validate username to prevent command injection
	if !userProfileUsernameValidationRegex.MatchString(username) {
		utils.ResponseBadRequest(c, "Invalid username format")
		return
	}

	var req types.ChangeOwnPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Verify old password and change in one queue operation
	err := h.queue.SubmitSync(func() error {
		// Verify old password using smbclient
		// SECURITY: username is already validated with regex, password is passed safely
		cmd := exec.Command("smbclient", "-L", "localhost", "-U", username+"%"+req.OldPassword, "-N")
		if err := cmd.Run(); err != nil {
			return utils.NewUnauthorizedError("Invalid old password")
		}

		// Change password
		return h.service.ChangePassword(username, req.NewPassword)
	})

	if err != nil {
		if unauthorizedErr, ok := err.(*utils.UnauthorizedError); ok {
			utils.ResponseUnauthorized(c, unauthorizedErr.Error())
			return
		}
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Password changed successfully")
}
