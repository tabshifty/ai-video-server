package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"video-server/internal/repository"
	"video-server/internal/services"
	"video-server/internal/utils"
)

func TestCollectionCleanupPostgresUploadSessionProtection(t *testing.T) {
	dsn := os.Getenv("COLLECTION_CLEANUP_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 COLLECTION_CLEANUP_TEST_DSN 运行隔离数据库接口测试")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host != "127.0.0.1" && cfg.ConnConfig.Host != "localhost" {
		t.Fatal("只允许本机隔离测试数据库")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	name := "cleanup_handlers_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, err := admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		admin.Close()
		if err != nil {
			t.Error(err)
		}
	})
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("缺少迁移", err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, strings.TrimPrefix(string(raw), "\ufeff")); err != nil {
			t.Fatalf("迁移 %s: %v", path, err)
		}
	}
	user, vc, ic := uuid.New(), uuid.New(), uuid.New()
	for table, id := range map[string]uuid.UUID{"collections": vc, "collections_images": ic} {
		if _, err = pool.Exec(ctx, "INSERT INTO "+table+"(id,name,normalized_name) VALUES($1,$2,$2)", id, id.String()); err != nil {
			t.Fatal(err)
		}
	}
	api := &API{repo: repository.NewVideoRepository(pool), chunkUpload: services.NewChunkUploadService(t.TempDir()), jwtSecret: "upload-protection-test"}
	claims := utils.JWTClaims{UserID: user.String(), Role: "admin", TokenType: utils.TokenTypeAccess, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(api.jwtSecret))
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.Register(router)
	request := func(method, path, body string) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var result struct {
			Code int
			Msg  string
			Data map[string]any
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Code != 0 {
			t.Fatalf("%s: %d %s", path, result.Code, result.Msg)
		}
		return result.Data
	}
	previews := map[string]string{}
	for _, path := range []string{"/admin/collections", "/admin/image-collections"} {
		p := request("POST", path+"/cleanup-empty/preview", "")
		if p["candidate_count"] != float64(1) {
			t.Fatalf("初次预览 %+v", p)
		}
		previews[path] = p["preview_token"].(string)
	}
	payload := `{"filename":"demo.mp4","file_size":6,"chunk_size":3,"total_chunks":2,"hash":"` + strings.Repeat("a", 64) + `","type":"short","collection_ids":["` + vc.String() + `"],"image_collection_id":"` + ic.String() + `"}`
	// 清理持有文件引用锁时，新上传必须被拦截，不能绕过扫描创建会话。
	release, err := api.repo.LockCollectionReferenceFiles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	reqBusy := httptest.NewRequest("POST", "/api/v1/upload/init", strings.NewReader(payload))
	reqBusy.Header.Set("Content-Type", "application/json")
	reqBusy.Header.Set("Authorization", "Bearer "+token)
	recBusy := httptest.NewRecorder()
	router.ServeHTTP(recBusy, reqBusy)
	release()
	var busy struct{ Code int }
	if err := json.Unmarshal(recBusy.Body.Bytes(), &busy); err != nil || busy.Code != 2001 {
		t.Fatalf("并发上传绕过清理锁: %s %v", recBusy.Body.String(), err)
	}
	session := request("POST", "/upload/init", payload)["upload_session_id"].(string)
	for path, preview := range previews {
		p := request("POST", path+"/cleanup-empty/preview", "")
		if p["candidate_count"] != float64(0) {
			t.Fatalf("预览遗漏上传引用 %+v", p)
		}
		result := request("POST", path+"/cleanup-empty", `{"preview_token":"`+preview+`"}`)
		if result["deleted_count"] != float64(0) || result["skipped_count"] != float64(1) {
			t.Fatalf("清理遗漏上传引用 %+v", result)
		}
	}
	request("DELETE", "/upload/session/"+session, "")
	for path, preview := range previews {
		result := request("POST", path+"/cleanup-empty", `{"preview_token":"`+preview+`"}`)
		if result["deleted_count"] != float64(1) {
			t.Fatalf("取消上传后无法清理 %+v", result)
		}
	}
	// 已清理目标不能被旧表单重新写入新上传会话。
	req := httptest.NewRequest("POST", "/api/v1/upload/init", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	var result struct{ Code int }
	if err = json.Unmarshal(rec.Body.Bytes(), &result); err != nil || result.Code != 1 {
		t.Fatalf("失效引用被接受 %s %v", rec.Body.String(), err)
	}
}
