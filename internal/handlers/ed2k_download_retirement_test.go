package handlers

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterExcludesRetiredEd2kDownloadRoutes(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	api := &API{}
	router := gin.New()
	api.Register(router)

	for _, route := range router.Routes() {
		if strings.Contains(route.Path, "/ed2k-download") {
			t.Fatalf("退役后仍注册 ED2K 下载路由：%s %s", route.Method, route.Path)
		}
	}
}
