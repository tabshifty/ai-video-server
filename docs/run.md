# 运行指南

## 一键启动（Linux/macOS）

```bash
bash scripts/dev-up.sh
```

默认本地数据层模式会：

1. 通过 `docker compose` 启动 `postgres` 和 `redis`
2. 等待 Postgres ready
3. 应用所有 `migrations/*.up.sql`
4. 构建并启动本机 Go server / worker
5. 启动 admin 前端（默认 `--frontend dev`）
6. 将日志和 PID 文件写入 `.run/`

## 前端模式

默认模式启动 Vite dev server：

```bash
bash scripts/dev-up.sh
# 等价于：
bash scripts/dev-up.sh --frontend dev
```

dev 模式下 Vite 监听 `0.0.0.0:5173`，可从局域网访问：

```bash
http://<server-ip>:5173
```

如果后端不在 Vite 同一台机器，修改 `admin-web/.env.development` 的代理目标：

```bash
VITE_API_PROXY_TARGET=http://<backend-ip>:8080
```

只构建前端（由 Go 后端 `/admin` 服务）：

```bash
bash scripts/dev-up.sh --frontend build
```

跳过前端启动：

```bash
bash scripts/dev-up.sh --frontend off
```

也可以指定自定义 env 文件：

```bash
ENV_FILE=/absolute/path/to/.env bash scripts/dev-up.sh
```

## 远程支撑层开发模式

当需要在开发机复现家用部署机真实数据问题时，可以本机运行 Go server / worker / admin 前端，但 Postgres、Redis 和翻译服务直连部署机支撑层。

本机 `.env.remote-local` 至少需要包含：

```bash
DEV_DATA_MODE=remote
HTTP_ADDR=:18080
POSTGRES_DSN=postgres://video:video@<部署机 IP>:5432/video_server?sslmode=disable
REDIS_ADDR=<部署机 IP>:6379
TRANSLATION_API_URL=http://<部署机 IP>:8000/v1
```

启动：

```bash
ENV_FILE=.env.remote-local bash scripts/dev-up.sh
```

该模式会跳过本地 `postgres` / `redis` 容器启动，并默认跳过 migration；`TRANSLATION_API_URL` 指向部署机或其它外部地址时，`dev-up.sh` 不启动本地翻译服务，而是检查外部 `/v1/models` 可达。停止时建议使用同一个 env 文件：

```bash
ENV_FILE=.env.remote-local bash scripts/dev-down.sh
```

`dev-up.sh` 会把本次数据模式写入 `.run/dev-data-mode`，即使忘记给 `dev-down.sh` 传 `ENV_FILE`，也会尽量按上次启动模式避免误停本地数据容器。

如果开发机没有本地 `llama-server`，并且这次不需要测试刮削文本翻译，可以关闭开发编排器里的翻译服务前置：

```bash
DEV_TRANSLATION_MODE=off ENV_FILE=.env.remote-local bash scripts/dev-up.sh
```

该开关只影响本次开发进程：`dev-up.sh` 会跳过本地翻译服务启动，并把后端 `TRANSLATION_API_URL` 置空；server / worker 仍会启动，只是内容翻译能力关闭。

如果部署机翻译服务仍只监听 `127.0.0.1:8000`，开发机直接访问 `<部署机 IP>:8000` 会失败。此时需要先在部署机把 `llama-server` 绑定到 LAN 地址，或通过 SSH tunnel 暴露到开发机本地端口，再把 `TRANSLATION_API_URL` 指向对应的 `/v1` 端点。

数据库改动流程：

1. 先使用默认本地数据层模式验证 migration：`bash scripts/dev-up.sh`。
2. 确认上一版本二进制仍能在新 schema 上运行，满足 `CONTEXT.md` 的 migration 前向兼容契约。
3. 提交代码和 migration，由家用部署机部署流程在启动 Go server 前执行 migration。

只有紧急人工修复远程库时，才允许显式放行：

```bash
ALLOW_REMOTE_MIGRATIONS=1 POSTGRES_DSN='postgres://...' bash scripts/migrate-apply.sh
```

没有 `ALLOW_REMOTE_MIGRATIONS=1` 时，`scripts/migrate-apply.sh` 会拒绝对非本机 Postgres 执行 migration。

## 停止全部进程

```bash
bash scripts/dev-down.sh
```

默认本地数据层模式会停止：

1. 本机 `server` / `worker` / `frontend` 进程（来自 `.run/*.pid`）
2. `postgres` 和 `redis` 容器

远程数据层模式只停止本机开发进程，不停止本地 compose 数据层。

## 日志

- server: `.run/server.log`
- worker: `.run/worker.log`
- frontend dev: `.run/frontend.log`

## 备注

- 如果 `.env` 不存在，`dev-up.sh` 会从 `.env.example` 复制。
- 默认本地数据层模式需要 `docker`（带 compose plugin）和 `go`；远程数据层模式不要求本机 Docker。
- `main.go` 加载环境变量时会优先读取 `ENV_FILE`。
- `dev-up.sh` 使用安全模式解析 `.env`（处理 BOM/CRLF，忽略非法行），不会直接 `source`。

## 排查 dev 模式 `/api/v1/*` 返回 `404`

如果浏览器 Network 返回的是 Vite HTML 而不是 JSON，说明请求没有打到 Go 后端：

1. 检查后端是否运行在预期地址（默认 `http://127.0.0.1:8080`）
2. 确认 `admin-web/.env.development` 的 `VITE_API_PROXY_TARGET` 正确
3. 修改 env 后重启 Vite dev server
