package handlers

import (
	"devhelp/internal/services/curlconverter"
	"devhelp/internal/services/dto"
	"devhelp/internal/utils"

	"github.com/gin-gonic/gin"
)

// CurlConverterHandler handles cURL conversion requests.
type CurlConverterHandler struct {
	svc curlconverter.Service
}

// NewCurlConverterHandler constructs a CurlConverterHandler.
func NewCurlConverterHandler(svc curlconverter.Service) *CurlConverterHandler {
	return &CurlConverterHandler{svc: svc}
}

// Convert godoc
// @Summary     Convert a cURL command to multiple languages
// @Description Parses a cURL command and returns equivalent code in Go, Python, Java, C, C++, Ruby, JavaScript, and Kotlin.
// @Tags        cURL Converter
// @Accept      json
// @Produce     json
// @Param       request body     dto.CurlConverterRequest  true  "cURL command input"
// @Success     200     {object} dto.APIResponse{data=dto.CurlConverterResponse}
// @Failure     400     {object} dto.APIResponse
// @Failure     500     {object} dto.APIResponse
// @Router      /api/v1/curl-converter [post]
func (h *CurlConverterHandler) Convert(c *gin.Context) {
	var req dto.CurlConverterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.svc.Convert(req)
	if err != nil {
		utils.BadRequest(c, "Failed to convert cURL command", err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}

// Languages godoc
// @Summary     List supported languages
// @Description Returns the list of languages the cURL Converter can generate code for.
// @Tags        cURL Converter
// @Produce     json
// @Success     200 {object} dto.APIResponse
// @Router      /api/v1/curl-converter/languages [get]
func (h *CurlConverterHandler) Languages(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{
		"languages": h.svc.SupportedLanguages(),
	})
}
