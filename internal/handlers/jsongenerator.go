package handlers

import (
	"devhelp/internal/services/dto"
	"devhelp/internal/services/jsongenerator"
	"devhelp/internal/utils"

	"github.com/gin-gonic/gin"
)

// JSONGeneratorHandler handles JSON-to-code generation requests.
type JSONGeneratorHandler struct {
	svc jsongenerator.Service
}

// NewJSONGeneratorHandler constructs a JSONGeneratorHandler.
func NewJSONGeneratorHandler(svc jsongenerator.Service) *JSONGeneratorHandler {
	return &JSONGeneratorHandler{svc: svc}
}

// Generate godoc
// @Summary     Generate type definitions from JSON
// @Description Parses a JSON payload and returns equivalent Go struct and TypeScript interface definitions.
// @Tags        JSON Generator
// @Accept      json
// @Produce     json
// @Param       request body     dto.JSONGeneratorRequest  true  "JSON generation input"
// @Success     200     {object} dto.APIResponse{data=dto.JSONGeneratorResponse}
// @Failure     400     {object} dto.APIResponse
// @Router      /api/v1/json-generator [post]
func (h *JSONGeneratorHandler) Generate(c *gin.Context) {
	var req dto.JSONGeneratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.svc.Generate(req)
	if err != nil {
		utils.BadRequest(c, "Failed to generate type definitions", err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}
