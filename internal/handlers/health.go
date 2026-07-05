// Package handlers contains all HTTP handler functions for DevHelp.
package handlers

import (
	"devhelp/internal/config"
	"devhelp/internal/services/dto"
	"devhelp/internal/utils"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles the health-check endpoint.
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// Check godoc
// @Summary     Health check
// @Description Returns the health status of the DevHelp API.
// @Tags        Health
// @Produce     json
// @Success     200 {object} dto.APIResponse{data=dto.HealthResponse}
// @Router      /api/v1/health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	utils.SuccessResponse(c, dto.HealthResponse{
		Status:  "ok",
		Version: h.cfg.App.Version,
		Env:     h.cfg.App.Env,
	})
}
