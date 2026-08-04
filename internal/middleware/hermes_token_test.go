package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHermesTokenMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configured     string
		authorization  string
		wantStatus     int
		wantNextCalled bool
	}{
		{name: "valid token", configured: "secret-token", authorization: "Bearer secret-token", wantStatus: http.StatusNoContent, wantNextCalled: true},
		{name: "wrong token", configured: "secret-token", authorization: "Bearer wrong", wantStatus: http.StatusUnauthorized},
		{name: "missing bearer", configured: "secret-token", wantStatus: http.StatusUnauthorized},
		{name: "empty server config", authorization: "Bearer secret-token", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			nextCalled := false
			router.GET("/test", HermesTokenMiddleware(tt.configured), func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status=%d, want=%d body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if nextCalled != tt.wantNextCalled {
				t.Fatalf("nextCalled=%v, want=%v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantStatus == http.StatusUnauthorized {
				var body struct {
					Code int `json:"code"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if body.Code != http.StatusUnauthorized {
					t.Fatalf("code=%d, want=%d", body.Code, http.StatusUnauthorized)
				}
			}
		})
	}
}
