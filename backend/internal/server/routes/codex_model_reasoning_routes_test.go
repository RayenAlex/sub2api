package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCodexModelReasoningRoutesRequireAdminAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlers := &handler.Handlers{Admin: &handler.AdminHandlers{Setting: adminhandler.NewSettingHandler(nil, nil, nil, nil, nil, nil, nil)}}
	adminAuth := servermiddleware.AdminAuthMiddleware(func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			servermiddleware.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required")
			return
		}
		servermiddleware.AbortWithError(c, http.StatusForbidden, "FORBIDDEN", "Admin access required")
	})
	auditLog := servermiddleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() })
	stepUp := servermiddleware.StepUpAuthMiddleware(func(c *gin.Context) { c.Next() })
	RegisterAdminRoutes(router.Group("/api/v1"), handlers, adminAuth, auditLog, stepUp, nil, nil)

	path := "/api/v1/admin/settings/codex-model-reasoning"
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		for _, auth := range []string{"", "Bearer user-token"} {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(method, path, nil)
			if auth != "" {
				request.Header.Set("Authorization", auth)
			}
			router.ServeHTTP(recorder, request)
			expected := http.StatusUnauthorized
			if auth != "" {
				expected = http.StatusForbidden
			}
			require.Equal(t, expected, recorder.Code, "%s with %q", method, auth)
		}
	}
}
