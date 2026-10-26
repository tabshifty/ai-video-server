package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"video-server/internal/services"
)

type workspaceHandlerStub struct {
	archiveImportService
	err       error
	version   int64
	guardErr  error
	guardKind string
	released  bool
}

func (s *workspaceHandlerStub) GetArchiveWorkspace(context.Context, uuid.UUID) (services.ArchiveWorkspaceResponse, error) {
	return services.ArchiveWorkspaceResponse{Version: 7}, s.err
}
func (s *workspaceHandlerStub) SaveArchiveWorkspace(_ context.Context, _ uuid.UUID, in services.ArchiveWorkspaceSave) (services.ArchiveWorkspaceResponse, error) {
	s.version = in.Version
	return services.ArchiveWorkspaceResponse{Version: in.Version + 1}, s.err
}
func (s *workspaceHandlerStub) SubmitArchiveWorkspace(_ context.Context, _ uuid.UUID, in services.ArchiveWorkspaceSubmit) (services.ArchiveWorkspaceResponse, error) {
	s.version = in.Version
	return services.ArchiveWorkspaceResponse{}, s.err
}
func (s *workspaceHandlerStub) GenerateArchiveWorkspaceGroups(_ context.Context, _ uuid.UUID, v int64) (services.ArchiveWorkspaceResponse, error) {
	s.version = v
	return services.ArchiveWorkspaceResponse{Version: v + 1}, s.err
}
func (s *workspaceHandlerStub) AcquireLegacyWorkspaceGuard(_ context.Context, kind string, _ uuid.UUID) (func(), error) {
	s.guardKind = kind
	return func() { s.released = true }, s.guardErr
}

func TestArchiveWorkspaceHandlerConflictAndVersion(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		err                      error
		want                     int
	}{
		{name: "加载工作区", method: "GET", path: "/workspace", want: 0},
		{name: "保存版本", method: "PUT", path: "/workspace", body: `{"version":7,"draft":{"directories":[],"groups":[],"files":{}}}`, want: 0},
		{name: "冲突保留版本", method: "PUT", path: "/workspace", body: `{"version":6}`, err: services.ErrArchiveWorkspaceConflict, want: 409},
		{name: "后台提交失败", method: "POST", path: "/submit", body: `{"version":7,"file_ids":[]}`, err: errors.New("入库队列不可用"), want: 1071},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &workspaceHandlerStub{err: tc.err}
			a := &API{archiveImportSvc: stub}
			r := gin.New()
			r.GET("/:id/workspace", a.AdminArchiveWorkspace)
			r.PUT("/:id/workspace", a.AdminSaveArchiveWorkspace)
			r.POST("/:id/submit", a.AdminSubmitArchiveWorkspace)
			req := httptest.NewRequest(tc.method, "/"+uuid.NewString()+tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			var got struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Code != tc.want {
				t.Fatalf("响应 %s", w.Body.String())
			}
			if tc.method == "PUT" && tc.want == 0 && stub.version != 7 {
				t.Fatal("未传递版本")
			}
		})
	}
}

func TestArchiveWorkspaceLegacyGuard(t *testing.T) {
	t.Run("初始化后旧接口拒绝", func(t *testing.T) {
		stub := &workspaceHandlerStub{guardErr: errors.New("请在工作台内编辑")}
		a := &API{archiveImportSvc: stub}
		r := gin.New()
		called := false
		r.PUT("/:id", a.legacyArchiveWorkspaceGuard("file"), func(c *gin.Context) { called = true })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("PUT", "/"+uuid.NewString(), nil))
		if called || !strings.Contains(w.Body.String(), `"code":409`) {
			t.Fatalf("旧接口绕过保护: %s", w.Body.String())
		}
	})
	t.Run("批量请求体可复读并释放锁", func(t *testing.T) {
		stub := &workspaceHandlerStub{}
		a := &API{archiveImportSvc: stub}
		r := gin.New()
		seen := false
		r.PUT("/batch", a.legacyArchiveWorkspaceGuard("targets"), func(c *gin.Context) {
			var req adminArchiveImportBatchUpdateRequest
			if err := c.ShouldBindBodyWithJSON(&req); err != nil {
				t.Fatal(err)
			}
			seen = len(req.Targets) == 1
		})
		req := httptest.NewRequest("PUT", "/batch", strings.NewReader(`{"targets":[{"id":"`+uuid.NewString()+`"}]}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
		if !seen || stub.guardKind != "file" || !stub.released {
			t.Fatalf("批量保护未正确处理: %+v", stub)
		}
	})
}
