package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"video-server/internal/utils"
)

func TestTVScreensaverRouteRequiresLoginButNotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{jwtSecret: "screensaver-test-only"}
	router := gin.New()
	api.Register(router)
	// 不带 jti 的测试令牌可独立验证认证/角色路由，无需访问 Redis。
	claims := utils.JWTClaims{
		UserID: uuid.NewString(), Role: "user", TokenType: utils.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(api.jwtSecret))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, token string
		want        int
	}{
		{"未登录", "", 401},
		{"无效令牌", "invalid", 401},
		{"普通用户进入业务处理而非管理员拒绝", token, 1050},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/tv/screensaver/images", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			var envelope struct{ Code int }
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Code != tc.want {
				t.Fatalf("code = %d, want %d; body=%s", envelope.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestTVScreensaverRejectsInvalidCursorAndUnboundedLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{}
	router := gin.New()
	router.GET("/images", api.TVScreensaverImages)
	for _, tc := range []struct {
		query string
		want  int
	}{
		{"?after=not-a-uuid", 1},
		{"?limit=0", 1},
		{"?limit=-1", 1},
		{"?limit=201", 1},
		{"?limit=99999999999999999999999999", 1},
		{"?limit=word", 1},
		{"?limit=1", 1050},
		{"?limit=200&after=11111111-1111-1111-1111-111111111111", 1050},
		{"", 1050},
	} {
		t.Run(tc.query, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest("GET", "/images"+tc.query, nil))
			var envelope struct{ Code int }
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Code != tc.want {
				t.Fatalf("code=%d, want=%d; body=%s", envelope.Code, tc.want, rec.Body.String())
			}
		})
	}
}
