package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/services"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// SystemHandler handles system-related HTTP requests
type SystemHandler struct {
	systemService *services.SystemService
	configService *services.ConfigService
}

// NewSystemHandler creates a new system handler
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		systemService: services.NewSystemService(),
		configService: services.NewConfigService(),
	}
}

// CheckEnvironment performs system environment checks
func (h *SystemHandler) CheckEnvironment(c *gin.Context) {
	result, err := h.systemService.CheckEnvironment()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseOK(c, result)
}

// GetSambaConfig retrieves Samba configuration
func (h *SystemHandler) GetSambaConfig(c *gin.Context) {
	config, err := h.configService.GetSambaConfig()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseOK(c, config)
}

// UpdateSambaConfig updates Samba configuration
func (h *SystemHandler) UpdateSambaConfig(c *gin.Context) {
	var req types.UpdateSambaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	if err := h.configService.UpdateSambaConfig(&req); err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Samba configuration updated successfully")
}

// GetSambaConfigFile retrieves raw smb.conf content
func (h *SystemHandler) GetSambaConfigFile(c *gin.Context) {
	content, err := h.configService.GetSambaConfigFile()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseOK(c, content)
}

// UpdateSambaConfigFile updates raw smb.conf content
func (h *SystemHandler) UpdateSambaConfigFile(c *gin.Context) {
	var req types.UpdateSambaConfigFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	if err := h.configService.UpdateSambaConfigFile(&req); err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseSuccessWithCustomMessage(c, "Samba configuration file updated successfully")
}

// GetSambaStatus retrieves current Samba status
func (h *SystemHandler) GetSambaStatus(c *gin.Context) {
	status, err := h.systemService.GetSambaStatus()
	if err != nil {
		utils.ResponseInternalServerError(c, err.Error())
		return
	}

	utils.ResponseOK(c, status)
}
