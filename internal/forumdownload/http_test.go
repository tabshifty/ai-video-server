package forumdownload

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"video-server/internal/services"
)

func TestIntegrationAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name, token, actor, body string
		enabled                  bool
		want                     int
	}{
		{"ingestion token rejected", "ingestion", "123", "{}", true, 401},
		{"missing actor rejected", "download", "", "{}", true, 403},
		{"other actor rejected", "download", "456", "{}", true, 403},
		{"body cannot spoof actor", "download", "123", `{"actor_id":456}`, true, 403},
		{"malformed body rejected", "download", "123", "{", true, 400},
		{"disabled remains identifiable", "download", "123", "{}", false, 503},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var svc *Service
			if tt.enabled {
				svc = &Service{downloads: &services.CloudDownloadService{}}
			}
			router := gin.New()
			Register(router, svc, "download", 123)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/telegram-forum/operations", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("X-Telegram-User-ID", tt.actor)
			req.Header.Set("Content-Type", "application/json")
			out := httptest.NewRecorder()
			router.ServeHTTP(out, req)
			if out.Code != tt.want {
				t.Fatalf("status=%d body=%s", out.Code, out.Body.String())
			}
		})
	}
}
