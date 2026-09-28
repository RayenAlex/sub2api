package admin

import (
	"errors"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GET /api/v1/admin/settings/codex-model-reasoning
func (h *SettingHandler) GetCodexModelReasoningSettings(c *gin.Context) {
	settings, err := h.settingService.GetCodexModelReasoningSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

// PUT /api/v1/admin/settings/codex-model-reasoning
func (h *SettingHandler) UpdateCodexModelReasoningSettings(c *gin.Context) {
	// The per-model and per-rule limits also bound the stored JSON; cap the
	// incoming body independently so unknown fields cannot be arbitrarily large.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
	var req struct {
		Rules *[]service.CodexModelReasoningRule `json:"rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Rules == nil {
		response.BadRequest(c, "rules must be an array")
		return
	}
	settings := service.CodexModelReasoningSettings{Rules: *req.Rules}
	if err := h.settingService.SetCodexModelReasoningSettings(c.Request.Context(), settings); err != nil {
		if errors.Is(err, service.ErrInvalidCodexModelReasoningSettings) {
			response.BadRequest(c, err.Error())
		} else {
			response.ErrorFrom(c, err)
		}
		return
	}
	// The service trims model slugs before persisting. Echo the canonical form.
	stored, err := h.settingService.GetCodexModelReasoningSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stored)
}
