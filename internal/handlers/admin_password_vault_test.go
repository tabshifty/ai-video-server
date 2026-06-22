package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterIncludesAdminPasswordVaultRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/admin/password-vault",
		"POST /api/v1/admin/password-vault",
		"PUT /api/v1/admin/password-vault/:id",
		"DELETE /api/v1/admin/password-vault/:id",
		"GET /api/v1/admin/password-vault/:id/password",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}
