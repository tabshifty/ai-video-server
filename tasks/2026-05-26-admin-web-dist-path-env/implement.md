# Implement：admin-web dist 路径环境变量

- 日期：2026-05-26
- 关联 PRD：`./prd.md`
- 关联 ADR：`docs/adr/0005-home-deployment-architecture.md`（[[家用部署机绝对路径契约]] 的代码侧支撑）

## 1. 总体方案

在 `Config` 增加 `AdminWebDistPath` 字段，沿用现有 `getEnv("ADMIN_WEB_DIST_PATH", "admin-web/dist")` 默认值模式。`RegisterRoutes`（或其入口 `New(api *API, ...)`）从 `Config` 接收路径字符串；如果不便从入口注入，退而求其次在 router 注册处直接 `os.Getenv`。**最小改动，不重构**。

## 2. 实施顺序（强制：TDD 红→绿→重构）

```
1. 读 router.go 现状，确认 RegisterRoutes 入口是否能拿到 Config（决定 D3 主路径或退路）
2. 红灯：先写 internal/handlers 的两个单测（env 未设 / env 设了），运行确认 FAIL
3. 红灯：写 internal/config 单测（如走主路径），运行确认 FAIL
4. 绿灯：改 Config 加字段 + getEnv；改 router.go 读字段；让 3 → 2 顺序变绿
5. 跑 go build ./... 通过
6. 跑 go test ./internal/config ./internal/handlers 通过
7. .env.example 加一行，置空值
8. plan.md 追加进度
```

## 3. 文件结构

### 3.1 改动文件

| 路径 | 改动 |
|---|---|
| `internal/config/config.go` | `Config` 增 `AdminWebDistPath string`，`Load` 中读 `getEnv("ADMIN_WEB_DIST_PATH", filepath.Join("admin-web", "dist"))` |
| `internal/handlers/router.go` | 第 221 行改为读 `cfg.AdminWebDistPath`（或 `os.Getenv` 退路）；保留 `os.Stat ... IsDir()` 兜底逻辑不变 |
| `.env.example` | 新增一行 `ADMIN_WEB_DIST_PATH=` 注释说明"留空 = 使用默认 admin-web/dist；家用部署机请填绝对路径" |

### 3.2 新增文件

| 路径 | 用途 |
|---|---|
| `internal/handlers/router_admin_dist_test.go` | 覆盖 env 未设 + dist 不存在 → /admin 未注册；env 设到临时 dir → /admin 返回该 dir 内容 |
| `internal/config/config_admin_dist_test.go`（可选）| 如走 D3 主路径，单测 `Load()` 读取 `ADMIN_WEB_DIST_PATH` 行为 |

## 4. 代码草图

### 4.1 `internal/config/config.go`

```go
// 在 Config 结构里加（按字段聚合度插在 PosterStoragePath 附近）：
AdminWebDistPath string

// 在 Load() 里加：
AdminWebDistPath: getEnv("ADMIN_WEB_DIST_PATH", filepath.Join("admin-web", "dist")),
```

注意：`filepath` 需要 import；如果文件中已 import 就不重复。

### 4.2 `internal/handlers/router.go`

定位第 221 行，从：

```go
adminDist := filepath.Join("admin-web", "dist")
```

改为（假设 `cfg *config.Config` 在该函数作用域可达；若不可达，看 4.3 退路）：

```go
adminDist := cfg.AdminWebDistPath
```

如果 `RegisterRoutes` 的签名上没有 `cfg`，要么把它加入签名（涉及 `main.go` 调用点改动），要么走退路：

```go
adminDist := os.Getenv("ADMIN_WEB_DIST_PATH")
if adminDist == "" {
    adminDist = filepath.Join("admin-web", "dist")
}
```

**实施时先看 router.go 现有签名再定**——能走主路径就走主路径，签名扩散到 main.go 即可接受。

### 4.3 测试草图

```go
// internal/handlers/router_admin_dist_test.go
package handlers

import (
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestAdminDistFromEnvOverridesDefault(t *testing.T) {
    dir := t.TempDir()
    indexPath := filepath.Join(dir, "index.html")
    if err := os.WriteFile(indexPath, []byte("env-admin"), 0o644); err != nil {
        t.Fatal(err)
    }
    if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
        t.Fatal(err)
    }
    t.Setenv("ADMIN_WEB_DIST_PATH", dir)

    r := gin.New()
    // 调用真实的 admin static 挂载入口（实施时确认函数名）
    mountAdminStatic(r) // 或者通过 RegisterRoutes 注入 Config 后调用

    req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("want 200, got %d", w.Code)
    }
    if got := w.Body.String(); got != "env-admin" {
        t.Fatalf("body mismatch: %q", got)
    }
}

func TestAdminDistMissingDirSkipsMount(t *testing.T) {
    t.Setenv("ADMIN_WEB_DIST_PATH", filepath.Join(t.TempDir(), "nonexistent"))

    r := gin.New()
    mountAdminStatic(r)

    req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code == http.StatusOK {
        t.Fatalf("/admin/ should not be mounted when dist missing, got %d", w.Code)
    }
}
```

> 如果 admin 静态挂载逻辑没有独立可调用入口，实施时要么提取一个内部辅助函数 `mountAdminStatic`，要么测试改为构造完整 `RegisterRoutes`（成本高）。**推荐提取**——这是 PRD §3 "N2 不重构" 的合理例外：单纯把 21 行 inline 代码挪进同文件的私有 helper，单测可达且语义零变化。

## 5. 已知陷阱

- **`filepath.Join` 对空字符串的行为**：`filepath.Join("admin-web", "dist")` 返回 `admin-web/dist`，但 `filepath.Join("")` 返回 `""`。如果使用者把 env 显式设为空字符串（不是 unset），代码应当走默认 —— `getEnv` 的实现 (`if v := os.Getenv(key); v != ""`) 已经处理了这种情况，无需额外判断。
- **`t.Setenv` 自动清理**：Go 1.17+ 的 `t.Setenv` 会在测试结束后还原；不要用 `os.Setenv` 直接设。
- **`gin.New()` vs `gin.Default()`**：测试中用 `gin.New()` 避免默认中间件污染 stdout。
- **Static route 与 NoRoute 的冲突**：router.go:229 有 `NoRoute` 兜底，单测中如果只挂 admin 不挂别的，访问 `/admin/` 走的是 `GET /admin/` 显式路由，不走 NoRoute；测试用例应该按显式路由覆盖，不依赖 NoRoute fallback。

## 6. 退出准则

- [ ] 单测两条全绿（env 设 / 未设）
- [ ] `go build ./...` BUILD SUCCESSFUL
- [ ] `go test ./internal/config ./internal/handlers -count=1` 全绿
- [ ] 手动场景 2 条（PRD §6.2）跑通
- [ ] `.env.example` 已加新键
- [ ] `plan.md` 追加本次进度
- [ ] 进入 `./review.md`
