package handlers

import (
	"devhelp/internal/services/dto"
	"devhelp/internal/services/jwttool"
	"devhelp/internal/utils"

	"github.com/gin-gonic/gin"
)

// JWTHandler handles JWT generation and decoding requests.
type JWTHandler struct {
	svc jwttool.Service
}

// NewJWTHandler constructs a JWTHandler.
func NewJWTHandler(svc jwttool.Service) *JWTHandler {
	return &JWTHandler{svc: svc}
}

// Generate godoc
// @Summary     Generate a JWT
// @Description Signs and returns a new HS256 JWT with the provided claims and secret.
// @Tags        JWT Toolkit
// @Accept      json
// @Produce     json
// @Param       request body     dto.JWTGenerateRequest  true  "JWT generation input"
// @Success     200     {object} dto.APIResponse{data=dto.JWTGenerateResponse}
// @Failure     400     {object} dto.APIResponse
// @Router      /api/v1/jwt/generate [post]
func (h *JWTHandler) Generate(c *gin.Context) {
	var req dto.JWTGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.svc.Generate(req)
	if err != nil {
		utils.BadRequest(c, "Failed to generate JWT", err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}

// Decode godoc
// @Summary     Decode a JWT
// @Description Decodes a JWT and returns header, payload, signature, and time claims. Optionally verifies the signature.
// @Tags        JWT Toolkit
// @Accept      json
// @Produce     json
// @Param       request body     dto.JWTDecodeRequest  true  "JWT decode input"
// @Success     200     {object} dto.APIResponse{data=dto.JWTDecodeResponse}
// @Failure     400     {object} dto.APIResponse
// @Router      /api/v1/jwt/decode [post]
func (h *JWTHandler) Decode(c *gin.Context) {
	var req dto.JWTDecodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.svc.Decode(req)
	if err != nil {
		utils.BadRequest(c, "Failed to decode JWT", err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}
