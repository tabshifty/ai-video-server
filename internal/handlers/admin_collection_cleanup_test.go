package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"video-server/internal/utils"
)

func TestCollectionCleanupPreviewBoundaries(t *testing.T) {
	api := &API{jwtSecret: "cleanup-test-only"}
	user, id := uuid.New(), uuid.New()
	raw, err := api.signCollectionCleanup("video", user, []uuid.UUID{id})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, raw, kind string
		user            uuid.UUID
		valid           bool
	}{
		{"原候选", raw, "video", user, true},
		{"跨类型", raw, "image", user, false},
		{"跨管理员", raw, "video", uuid.New(), false},
		{"篡改签名", raw + "x", "video", user, false},
		{"缺失预览", "", "video", user, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := api.parseCollectionCleanup(tc.raw, tc.kind, tc.user)
			if (err == nil) != tc.valid {
				t.Fatalf("ids=%v err=%v", ids, err)
			}
			if tc.valid && (len(ids) != 1 || ids[0] != id) {
				t.Fatalf("候选变化 %v", ids)
			}
		})
	}
	claims := collectionCleanupClaims{Kind: "video", Candidates: []uuid.UUID{id}, RegisteredClaims: jwt.RegisteredClaims{Subject: user.String(), Audience: jwt.ClaimStrings{collectionCleanupAudience}}}
	for _, expired := range []bool{false, true} {
		if expired {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
		}
		raw, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("collection-cleanup:" + api.jwtSecret))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = api.parseCollectionCleanup(raw, "video", user); err == nil {
			t.Fatal("接受无过期时间或已过期凭据")
		}
	}
}

func TestCollectionCleanupRoutesRequireAdminAndPreview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &API{jwtSecret: "cleanup-route-test-only"}
	router := gin.New()
	api.Register(router)
	sign := func(role string) string {
		claims := utils.JWTClaims{UserID: uuid.NewString(), Role: role, TokenType: utils.TokenTypeAccess,
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
		raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(api.jwtSecret))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	for _, path := range []string{"/collections/cleanup-empty", "/image-collections/cleanup-empty"} {
		for _, suffix := range []string{"", "/preview"} {
			for _, tc := range []struct {
				name, token string
				want        int
			}{
				{"未登录", "", 401}, {"普通用户", sign("user"), 403},
			} {
				t.Run(path+suffix+tc.name, func(t *testing.T) {
					req := httptest.NewRequest("POST", "/api/v1/admin"+path+suffix, strings.NewReader(`{}`))
					req.Header.Set("Content-Type", "application/json")
					if tc.token != "" {
						req.Header.Set("Authorization", "Bearer "+tc.token)
					}
					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, req)
					var result struct{ Code int }
					if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
						t.Fatal(err)
					}
					if result.Code != tc.want {
						t.Fatalf("code=%d body=%s", result.Code, rec.Body.String())
					}
				})
			}
		}
		// 管理员也不能绕过预览直接提交候选 ID；必须在访问仓储前拒绝。
		req := httptest.NewRequest("POST", "/api/v1/admin"+path, strings.NewReader(`{"candidates":["`+uuid.NewString()+`"]}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sign("admin"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var result struct{ Code int }
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Code != 1 {
			t.Fatalf("缺少预览仍访问仓储: %s", rec.Body.String())
		}
	}
}
