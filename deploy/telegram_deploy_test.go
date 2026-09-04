package deploy_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Build       composeBuild      `yaml:"build"`
	Command     []string          `yaml:"command"`
	Environment map[string]string `yaml:"environment"`
	Image       string            `yaml:"image"`
	User        string            `yaml:"user"`
	Volumes     []string          `yaml:"volumes"`
}

type composeBuild struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

func TestTelegramComposeDeploymentContract(t *testing.T) {
	content := readProjectFile(t, "deploy/docker-compose.telegram.yml")
	var compose composeFile
	if err := yaml.Unmarshal(content, &compose); err != nil {
		t.Fatalf("解析 Telegram Compose: %v", err)
	}

	serviceNames := []string{"api", "worker", "telegram-ingestor"}
	sharedMounts := map[string]string{}
	for _, name := range serviceNames {
		service, ok := compose.Services[name]
		if !ok {
			t.Fatalf("Compose 缺少 %s 服务", name)
		}
		if service.Build.Context != "." || service.Build.Dockerfile != "deploy/video-server.Dockerfile" {
			t.Errorf("%s 未复用统一镜像构建配置：%+v", name, service.Build)
		}
		if service.Image != "ai-video-server:telegram" {
			t.Errorf("%s 的镜像 = %q，期望 ai-video-server:telegram", name, service.Image)
		}
		for _, env := range []string{"STORAGE_ROOT=/data/storage", "UPLOAD_TEMP_DIR=/data/tmp/uploads"} {
			parts := strings.SplitN(env, "=", 2)
			if got := service.Environment[parts[0]]; got != parts[1] {
				t.Errorf("%s 的 %s = %q，期望 %q", name, parts[0], got, parts[1])
			}
		}
		for key, want := range map[string]string{
			"HTTP_ADDR":    ":8080",
			"POSTGRES_DSN": "postgres://video:video@postgres:5432/video_server?sslmode=disable",
			"REDIS_ADDR":   "redis:6379",
		} {
			if got := service.Environment[key]; got != want {
				t.Errorf("%s 的 %s = %q，期望 %q", name, key, got, want)
			}
		}
		for _, target := range []string{"/data/storage", "/data/tmp/uploads"} {
			mount := findMount(service.Volumes, target)
			if mount == "" {
				t.Errorf("%s 未挂载 %s", name, target)
				continue
			}
			if previous, ok := sharedMounts[target]; ok && previous != mount {
				t.Errorf("%s 的 %s 挂载 %q 与其它服务 %q 不一致", name, target, mount, previous)
			} else {
				sharedMounts[target] = mount
			}
		}
	}
	if got := compose.Services["api"].Environment["APP_MODE"]; got != "server" {
		t.Errorf("api 的 APP_MODE = %q，期望 server", got)
	}
	if got := compose.Services["worker"].Environment["APP_MODE"]; got != "worker" {
		t.Errorf("worker 的 APP_MODE = %q，期望 worker", got)
	}

	for _, name := range []string{"api", "worker"} {
		if mount := findMount(compose.Services[name].Volumes, "/data/telegram-session"); mount != "" {
			t.Errorf("%s 不应挂载 Telegram session：%s", name, mount)
		}
	}
	ingestor := compose.Services["telegram-ingestor"]
	if mount := findMount(ingestor.Volumes, "/data/telegram-session"); mount == "" {
		t.Error("telegram-ingestor 未挂载 /data/telegram-session")
	}
	if user := strings.TrimSpace(ingestor.User); user == "" || user == "root" || strings.HasPrefix(user, "0:") {
		t.Errorf("telegram-ingestor 必须配置非 root 用户，当前为 %q", ingestor.User)
	}
	if command := strings.Join(ingestor.Command, " "); !strings.Contains(command, "telegram-ingestor") || !strings.Contains(command, "-mode run") {
		t.Errorf("telegram-ingestor command 必须显式包含 -mode run，当前为 %q", command)
	}
}

func TestTelegramRuntimeImageContainsMediaTools(t *testing.T) {
	dockerfile := string(readProjectFile(t, "deploy/video-server.Dockerfile"))
	for _, required := range []string{
		"FROM golang:1.22",
		"ca-certificates",
		"ffmpeg",
		"ffprobe -version",
		"go build",
		"./cmd/telegram-ingestor",
	} {
		if !strings.Contains(dockerfile, required) {
			t.Errorf("Dockerfile 缺少 %q", required)
		}
	}
	if strings.Contains(dockerfile, "COPY . .") {
		t.Error("Dockerfile 不应把可能含 .env 或 session 的整个工作区复制进构建层")
	}
	dockerignore := string(readProjectFile(t, ".dockerignore"))
	for _, excluded := range []string{".env", "storage/", "tmp/", "telegram-session/"} {
		if !hasDockerignorePattern(dockerignore, excluded) {
			t.Errorf(".dockerignore 未排除 %s", excluded)
		}
	}
}

func TestTelegramEnvironmentExampleIsComplete(t *testing.T) {
	env := string(readProjectFile(t, ".env.example"))
	for _, name := range []string{
		"TELEGRAM_API_ID",
		"TELEGRAM_API_HASH",
		"TELEGRAM_PHONE",
		"TELEGRAM_SESSION_PATH",
		"TELEGRAM_IMPORT_USER_ID",
		"TELEGRAM_REALTIME_QUEUE",
		"TELEGRAM_BACKFILL_QUEUE",
		"TELEGRAM_CONTROL_QUEUE",
		"TELEGRAM_DOWNLOAD_CONCURRENCY",
		"TELEGRAM_MAX_ACTIVE_TASKS",
		"TELEGRAM_CONTROL_TOKEN",
		"TELEGRAM_SESSION_HOST_PATH",
		"VIDEO_STORAGE_HOST_PATH",
		"VIDEO_UPLOAD_TEMP_HOST_PATH",
	} {
		if !hasEnvAssignment(env, name) {
			t.Errorf(".env.example 缺少 %s", name)
		}
	}
}

func TestTelegramRunbookCoversRequiredOperationsAndSafety(t *testing.T) {
	runbook := string(readProjectFile(t, "docs/telegram-video-ingestion.md"))
	for _, required := range []string{
		"migrate-apply.sh",
		"/admin/telegram",
		"手机号授权",
		"二维码重新授权",
		"预解析",
		"确认添加",
		"恢复失败来源",
		"重新回填",
		"0039_telegram_management.up.sql",
		"TELEGRAM_CONTROL_TOKEN",
		"0038_telegram_video_ingestion.down.sql",
		"0039_telegram_management.down.sql",
		"FloodWait",
		"个人账号",
		"私密邀请",
		"session",
		"版权",
	} {
		if !strings.Contains(runbook, required) {
			t.Errorf("Telegram 运维手册缺少 %q", required)
		}
	}
	for _, removed := range []string{
		"-mode login",
		"/admin/telegram/sources",
		"/progress",
		"/pause",
		"/resume",
		"/backfill",
		"processing_status = 'failed'",
		"telegram_media SET",
		"FLUSHDB",
	} {
		if strings.Contains(runbook, removed) {
			t.Errorf("Telegram 运维手册不应再包含旧手工操作 %q", removed)
		}
	}

	gitignore := string(readProjectFile(t, ".gitignore"))
	for _, excluded := range []string{"/tmp/", "/telegram-session/"} {
		if !hasDockerignorePattern(gitignore, excluded) {
			t.Errorf(".gitignore 未排除 %s", excluded)
		}
	}
}

func readProjectFile(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("读取 %s: %v", name, err)
	}
	return content
}

func findMount(volumes []string, target string) string {
	for _, volume := range volumes {
		parts := strings.Split(volume, ":")
		if len(parts) >= 2 && parts[len(parts)-1] == target {
			return volume
		}
	}
	return ""
}

func hasEnvAssignment(content, name string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), name+"=") {
			return true
		}
	}
	return false
}

func hasDockerignorePattern(content, want string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == want {
			return true
		}
	}
	return false
}
