package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminCreateUserRejectsInvalidRole(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewBufferString(`{"username":"admin2","email":"admin2@example.com","password":"secret123","role":"owner"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	api.AdminCreateUser(ctx)

	var resp apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != 1 {
		t.Fatalf("expected code=1, got=%d body=%s", resp.Code, rec.Body.String())
	}
}
