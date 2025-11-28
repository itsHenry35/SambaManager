package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/queue"
	"github.com/itsHenry35/SambaManager/services"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// ShareHandler handles share-related HTTP requests
type ShareHandler struct {
	service *services.SambaService
	queue   *queue.Queue
}

// NewShareHandler creates a new share handler
func NewShareHandler(service *services.SambaService, q *queue.Queue) *ShareHandler {
	return &ShareHandler{
		service: service,
		queue:   q,
	}
}

// CreateShare creates a new Samba share
func (h *ShareHandler) CreateShare(c *gin.Context) {
	var req types.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// Convert to internal share model
	share := &types.Share{
		Name:       req.Name,
		Owner:      req.Owner,
		SharedWith: req.SharedWith,
		ReadOnly:   req.ReadOnly,
		Comment:    req.Comment,
		SubPath:    req.SubPath,
	}

	// Submit to queue for processing
	var shareId string
	err := h.queue.SubmitSync(func() error {
		id, err := h.service.CreateShare(share)
		shareId = id
		return err
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseCreated(c, gin.H{
		"id": shareId,
	})
}

// UpdateShare updates an existing Samba share
func (h *ShareHandler) UpdateShare(c *gin.Context) {
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

	// Convert to internal share model
	share := &types.Share{
		SharedWith: req.SharedWith,
		ReadOnly:   req.ReadOnly,
		Comment:    req.Comment,
		SubPath:    req.SubPath,
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.UpdateShare(shareId, share)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithMessageAndData(c, "Share updated successfully", gin.H{
		"id": shareId,
	})
}

// DeleteShare deletes a Samba share
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	shareId := c.Param("shareId")
	if shareId == "" {
		utils.ResponseBadRequest(c, "Share ID is required")
		return
	}

	// Submit to queue for processing
	err := h.queue.SubmitSync(func() error {
		return h.service.DeleteShare(shareId)
	})

	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Share deleted successfully")
}

// ListShares retrieves Samba shares with pagination and search
func (h *ShareHandler) ListShares(c *gin.Context) {
	var query types.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}
	query = query.GetDefaults()

	shares, err := h.service.ListShares()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	// Filter by search query (search in owner, share ID, and comment)
	filteredShares := shares
	if query.Search != "" {
		filteredShares = []types.ShareResponse{}
		searchLower := strings.ToLower(query.Search)
		for _, share := range shares {
			if strings.Contains(strings.ToLower(share.Owner), searchLower) ||
				strings.Contains(strings.ToLower(share.ID), searchLower) ||
				strings.Contains(strings.ToLower(share.Comment), searchLower) {
				filteredShares = append(filteredShares, share)
			}
		}
	}

	// Calculate pagination
	total := len(filteredShares)
	offset := query.GetOffset()
	end := offset + query.PageSize
	if end > total {
		end = total
	}

	// Get paginated slice
	paginatedShares := []types.ShareResponse{}
	if offset < total {
		paginatedShares = filteredShares[offset:end]
	}

	utils.ResponsePaginated(c, paginatedShares, total, query.Page, query.PageSize)
}
