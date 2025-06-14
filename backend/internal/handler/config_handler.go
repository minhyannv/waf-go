package handler

import (
	"net/http"

	"waf-go/internal/service"
	"waf-go/internal/utils"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetSystemConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取当前系统配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=service.SystemConfig}
// @Failure 500 {object} utils.Response
// @Router /api/v1/config [get]
func (h *ConfigHandler) GetSystemConfig(c *gin.Context) {
	config, err := h.configService.GetSystemConfig()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取系统配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取系统配置成功", config)
}

// UpdateSystemConfig 更新系统配置
// @Summary 更新系统配置
// @Description 更新系统配置信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param config body service.UpdateConfigRequest true "配置信息"
// @Success 200 {object} utils.Response{data=service.SystemConfig}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/config [put]
func (h *ConfigHandler) UpdateSystemConfig(c *gin.Context) {
	var req service.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	config, err := h.configService.UpdateSystemConfig(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "更新系统配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "更新系统配置成功", config)
}

// ResetSystemConfig 重置系统配置
// @Summary 重置系统配置
// @Description 重置系统配置为默认值
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=service.SystemConfig}
// @Failure 500 {object} utils.Response
// @Router /api/v1/config/reset [post]
func (h *ConfigHandler) ResetSystemConfig(c *gin.Context) {
	config, err := h.configService.ResetSystemConfig()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "重置系统配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "重置系统配置成功", config)
}

// GetConfigStats 获取配置统计信息
// @Summary 获取配置统计信息
// @Description 获取系统配置统计信息
// @Tags 系统配置
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=map[string]interface{}}
// @Failure 500 {object} utils.Response
// @Router /api/v1/config/stats [get]
func (h *ConfigHandler) GetConfigStats(c *gin.Context) {
	stats, err := h.configService.GetConfigStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取配置统计信息失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, "获取配置统计信息成功", stats)
}
