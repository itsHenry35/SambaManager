package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/api/middlewares"
	"github.com/itsHenry35/SambaManager/queue"
	"github.com/itsHenry35/SambaManager/services"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// UserShareHandler handles user share-related HTTP requests
type UserShareHandler struct {
	service *services.SambaService
	queue   *queue.Queue
}

// NewUserShareHandler creates a new user share handler
func NewUserShareHandler(service *services.SambaService, q *queue.Queue) *UserShareHandler {
	return &UserShareHandler{
		service: service,
		queue:   q,
	}
}

// ListMyShares lists all shares owned by the current user
func (h *UserShareHandler) ListMyShares(c *gin.Context) {
	username, exists := middlewares.GetUsernameFromContext(c)
	if !exists {
		utils.ResponseUnauthorized(c, "User not found in context")
		return
	}

	var shares []types.ShareResponse
	err := h.queue.SubmitSync(func() error {
		result, err := h.service.ListShares()
		if err != nil {
			return err
		}
		shares = result
		return nil
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	// Filter shares owned by current user
	myShares := []types.ShareResponse{}
	for _, share := range shares {
		if share.Owner == username {
			myShares = append(myShares, share)
		}
	}

	utils.ResponseOK(c, myShares)
}

// CreateMyShare creates a new share for the current user
func (h *UserShareHandler) CreateMyShare(c *gin.Context) {
	username, exists := middlewares.GetUsernameFromContext(c)
	if !exists {
		utils.ResponseUnauthorized(c, "User not found in context")
		return
	}

	var req types.CreateMyShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Owner is automatically set to current user (from JWT token)
	var shareId string
	err := h.queue.SubmitSync(func() error {
		result, err := h.service.CreateShare(&types.Share{
			Name:       req.Name,
			Owner:      username, // Use current user as owner
			SharedWith: req.SharedWith,
			ReadOnly:   req.ReadOnly,
			Comment:    req.Comment,
			SubPath:    req.SubPath,
		})
		if err != nil {
			return err
		}
		shareId = result
		return nil
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessageAndData(c, "Share created successfully", shareId)
}

// UpdateMyShare updates a share owned by the current user
func (h *UserShareHandler) UpdateMyShare(c *gin.Context) {
	username, exists := middlewares.GetUsernameFromContext(c)
	if !exists {
		utils.ResponseUnauthorized(c, "User not found in context")
		return
	}

	shareId := c.Param("shareId")
	if shareId == "" {
		utils.ResponseBadRequest(c, "Share ID is required")
		return
	}

	var req types.UpdateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Verify share belongs to current user and update in one queue operation
	err := h.queue.SubmitSync(func() error {
		shares, err := h.service.ListShares()
		if err != nil {
			return err
		}

		var targetShare *types.ShareResponse
		for _, share := range shares {
			if share.ID == shareId {
				targetShare = &share
				break
			}
		}

		if targetShare == nil {
			return utils.NewNotFoundError("Share not found")
		}

		if targetShare.Owner != username {
			return utils.NewForbiddenError("You can only update your own shares")
		}

		return h.service.UpdateShare(shareId, &types.Share{
			SharedWith: req.SharedWith,
			ReadOnly:   req.ReadOnly,
			Comment:    req.Comment,
			SubPath:    req.SubPath,
		})
	})

	if err != nil {
		if notFoundErr, ok := err.(*utils.NotFoundError); ok {
			utils.ResponseNotFound(c, notFoundErr.Error())
			return
		}
		if forbiddenErr, ok := err.(*utils.ForbiddenError); ok {
			utils.ResponseForbidden(c, forbiddenErr.Error())
			return
		}
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Share updated successfully")
}

// DeleteMyShare deletes a share owned by the current user
func (h *UserShareHandler) DeleteMyShare(c *gin.Context) {
	username, exists := middlewares.GetUsernameFromContext(c)
	if !exists {
		utils.ResponseUnauthorized(c, "User not found in context")
		return
	}

	shareId := c.Param("shareId")
	if shareId == "" {
		utils.ResponseBadRequest(c, "Share ID is required")
		return
	}

	// Verify share belongs to current user and delete in one queue operation
	err := h.queue.SubmitSync(func() error {
		shares, err := h.service.ListShares()
		if err != nil {
			return err
		}

		var targetShare *types.ShareResponse
		for _, share := range shares {
			if share.ID == shareId {
				targetShare = &share
				break
			}
		}

		if targetShare == nil {
			return utils.NewNotFoundError("Share not found")
		}

		if targetShare.Owner != username {
			return utils.NewForbiddenError("You can only delete your own shares")
		}

		return h.service.DeleteShare(shareId)
	})

	if err != nil {
		if notFoundErr, ok := err.(*utils.NotFoundError); ok {
			utils.ResponseNotFound(c, notFoundErr.Error())
			return
		}
		if forbiddenErr, ok := err.(*utils.ForbiddenError); ok {
			utils.ResponseForbidden(c, forbiddenErr.Error())
			return
		}
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Share deleted successfully")
}
