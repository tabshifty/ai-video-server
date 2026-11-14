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

	"video-server/internal/models"
	"video-server/internal/repository"
	"video-server/internal/utils"
)

// TestUploadAttachPostgres 覆盖秒传命中后的合集追加接口：鉴权、幂等、类型冲突与待删除视频。
func TestUploadAttachPostgres(t *testing.T) {
	dsn := os.Getenv("UPLOAD_ATTACH_TEST_DSN")
	if dsn == "" {
		t.Skip("设置 UPLOAD_ATTACH_TEST_DSN 运行隔离数据库接口测试")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host != "127.0.0.1" && cfg.ConnConfig.Host != "localhost" {
		t.Fatal("只允许本机隔离测试数据库")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	admin, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	name := "upload_attach_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
		_, dropErr := admin.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)")
		admin.Close()
		if dropErr != nil {
			t.Error(dropErr)
		}
	})
	paths, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("缺少迁移", err)
	}
	for _, path := range paths {
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, execErr := pool.Exec(ctx, strings.TrimPrefix(string(raw), "\ufeff")); execErr != nil {
			t.Fatalf("迁移 %s: %v", path, execErr)
		}
	}

	repo := repository.NewVideoRepository(pool)
	userID := uuid.New()
	suffix := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO users(id,username,email,password_hash) VALUES($1,$2,$3,'unused')`, userID, "attach-"+suffix, "attach-"+suffix+"@localhost"); err != nil {
		t.Fatal(err)
	}
	hash := strings.Repeat("a", 64)
	videoID := uuid.New()
	if err := repo.CreateVideo(ctx, models.Video{ID: videoID, UserID: &userID, Title: "已存在短视频", Type: "short", Status: "ready"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.InsertFileHash(ctx, hash, videoID, 1024); err != nil {
		t.Fatal(err)
	}
	collectionID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO collections(id,name,normalized_name) VALUES($1,$2,$2)`, collectionID, "追加合集-"+collectionID.String()); err != nil {
		t.Fatal(err)
	}

	api := &API{repo: repo, jwtSecret: "upload-attach-test"}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api.Register(router)
	tokenFor := func(role string) string {
		t.Helper()
		claims := utils.JWTClaims{UserID: userID.String(), Role: role, TokenType: utils.TokenTypeAccess, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
		token, signErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(api.jwtSecret))
		if signErr != nil {
			t.Fatal(signErr)
		}
		return token
	}
	adminToken := tokenFor("admin")

	type attachResponse struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	attach := func(token, body string) attachResponse {
		t.Helper()
		req := httptest.NewRequest("POST", "/api/v1/upload/attach", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var result attachResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatalf("解析响应失败: %s %v", rec.Body.String(), err)
		}
		return result
	}
	payload := `{"hash":"` + hash + `","file_size":1024,"type":"short","collection_ids":["` + collectionID.String() + `"]}`

	nonAdmin := attach(tokenFor("user"), payload)
	if nonAdmin.Code != 403 {
		t.Fatalf("非管理员应被拒绝, got %d %+v", nonAdmin.Code, nonAdmin.Data)
	}

	if result := attach(adminToken, payload); result.Code != 0 {
		t.Fatalf("首次追加失败: %d %+v", result.Code, result.Data)
	} else if result.Data["added_collections"] != float64(1) {
		t.Fatalf("首次追加计数 = %v, want 1", result.Data["added_collections"])
	}
	var linked int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM video_collections WHERE video_id=$1 AND collection_id=$2`, videoID, collectionID).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 1 {
		t.Fatalf("video_collections 实际行数 = %d, want 1", linked)
	}
	if result := attach(adminToken, payload); result.Code != 0 {
		t.Fatalf("重复追加失败: %d %+v", result.Code, result.Data)
	} else if result.Data["added_collections"] != float64(0) {
		t.Fatalf("重复追加计数 = %v, want 0", result.Data["added_collections"])
	}

	unknown := attach(adminToken, `{"hash":"`+strings.Repeat("b", 64)+`","file_size":1024,"type":"short","collection_ids":["`+collectionID.String()+`"]}`)
	if unknown.Code != 2008 {
		t.Fatalf("未知哈希返回 %d, want 2008", unknown.Code)
	}
	// 入口层先拦非短视频携带合集，因此类型冲突用例必须不带合集选择。
	conflict := attach(adminToken, `{"hash":"`+hash+`","file_size":1024,"type":"movie"}`)
	if conflict.Code != 409 {
		t.Fatalf("类型冲突返回 %d, want 409", conflict.Code)
	}
	nonShortCollections := attach(adminToken, strings.Replace(payload, `"type":"short"`, `"type":"movie"`, 1))
	if nonShortCollections.Code == 0 {
		t.Fatalf("非短视频携带合集应被入口层拒绝: %+v", nonShortCollections.Data)
	}

	if _, err := pool.Exec(ctx, `UPDATE videos SET status='pending_delete' WHERE id=$1`, videoID); err != nil {
		t.Fatal(err)
	}
	// 本次没有选择任何合集时也必须报告待删除，而不是回「合集无变化」。
	pending := attach(adminToken, `{"hash":"`+hash+`","file_size":1024,"type":"short"}`)
	if pending.Code != 0 {
		t.Fatalf("待删除视频返回错误: %d %+v", pending.Code, pending.Data)
	}
	if pending.Data["pending_delete"] != true || pending.Data["added_collections"] != float64(0) {
		t.Fatalf("待删除视频不应追加关系: %+v", pending.Data)
	}
}
