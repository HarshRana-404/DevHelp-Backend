package handlers

import (
	"net/http"

	"devhelp/internal/services/dto"
	"devhelp/internal/services/httpinspector"
	"devhelp/internal/utils"

	"github.com/gin-gonic/gin"
)

// HTTPInspectorHandler handles HTTP inspection requests.
type HTTPInspectorHandler struct {
	svc httpinspector.Service
}

// NewHTTPInspectorHandler constructs an HTTPInspectorHandler with the given service.
func NewHTTPInspectorHandler(svc httpinspector.Service) *HTTPInspectorHandler {
	return &HTTPInspectorHandler{svc: svc}
}

// Inspect godoc
// @Summary     Inspect an HTTP request
// @Description Parses a raw HTTP request or cURL command and returns a detailed structural breakdown.
// @Tags        HTTP Inspector
// @Accept      json
// @Produce     json
// @Param       request body     dto.HTTPInspectorRequest  true  "Raw HTTP or cURL input"
// @Success     200     {object} dto.APIResponse{data=dto.HTTPInspectorResponse}
// @Failure     400     {object} dto.APIResponse
// @Failure     500     {object} dto.APIResponse
// @Router      /api/v1/http-inspector [post]
func (h *HTTPInspectorHandler) Inspect(c *gin.Context) {
	var req dto.HTTPInspectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if req.RawHTTP == "" && req.CurlCommand == "" {
		utils.BadRequest(c, "Either raw_http or curl_command must be provided", "")
		return
	}

	result, err := h.svc.Inspect(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PARSE_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	utils.SuccessResponse(c, result)
}
