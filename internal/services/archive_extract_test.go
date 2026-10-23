package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveExtractionRecovery(t *testing.T) {
	tests := []struct {
		name, format, primary, unrar, unar, password string
		wantSuccess, wantPassword                    bool
		wantCalled                                   string
	}{
		{name: "RAR头部错误使用原生工具", format: "rar", primary: "Headers Error", unrar: "success", wantSuccess: true, wantCalled: "7zz unrar "},
		{name: "头部错误回退到unar", format: "rar", primary: "Headers Error", unar: "success", wantSuccess: true, wantCalled: "7zz unar "},
		{name: "unrar密码保留空格与符号", format: "rar", primary: "Headers Error", unrar: "success", password: "pa ss;$()", wantSuccess: true, wantCalled: "7zz unrar "},
		{name: "unar密码保留空格与符号", format: "rar", primary: "Headers Error", unar: "success", password: "pa ss;$()", wantSuccess: true, wantCalled: "7zz unar "},
		{name: "备用工具之间清理残留", format: "rar", primary: "Unsupported Method", unrar: "Headers Error", unar: "success", wantSuccess: true, wantCalled: "7zz unrar unar "},
		{name: "缺少7zip仍可使用原生RAR工具", format: "rar", unrar: "success", wantSuccess: true, wantCalled: "unrar "},
		{name: "密码错误不回退", format: "rar", primary: "Headers Error: Wrong password?", unrar: "success", wantPassword: true, wantCalled: "7zz "},
		{name: "CRC错误不回退", format: "rar", primary: "CRC Failed", unrar: "success", wantCalled: "7zz "},
		{name: "头部错误伴随数据损坏不回退", format: "rar", primary: "Headers Error\nCRC Failed", unrar: "success", wantCalled: "7zz "},
		{name: "7z头部错误不用RAR工具", format: "7z", primary: "Headers Error", unrar: "success", wantCalled: "7zz "},
		{name: "7z算法错误不用RAR工具", format: "7z", primary: "Unsupported Method", unrar: "success", wantCalled: "7zz "},
		{name: "全部失败保留两级诊断", format: "rar", primary: "Headers Error", unrar: "CRC Failed", unar: "Data Error", wantCalled: "7zz unrar unar "},
		{name: "备用工具识别密码错误", format: "rar", primary: "Headers Error", unrar: "Wrong password", unar: "success", wantPassword: true, wantCalled: "7zz unrar "},
		{name: "没有兼容工具说明依赖", format: "rar", primary: "Headers Error", wantCalled: "7zz "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			bin, out := filepath.Join(root, "bin"), filepath.Join(root, "output")
			if err := os.Mkdir(bin, 0o755); err != nil {
				t.Fatal(err)
			}
			archive := filepath.Join(root, "中文 archive."+tt.format)
			if err := os.WriteFile(archive, []byte("fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			t.Setenv("ARCHIVE_TEST_OUTPUT", out)
			t.Setenv("ARCHIVE_TEST_INPUT", archive)
			t.Setenv("ARCHIVE_TEST_PASSWORD", tt.password)
			t.Setenv("ARCHIVE_TEST_CALLS", filepath.Join(root, "calls"))
			for _, tool := range []struct{ name, result string }{{"7zz", tt.primary}, {"unrar", tt.unrar}, {"unar", tt.unar}} {
				if tool.result == "" {
					continue
				}
				script := `#!/bin/sh
set -eu
printf '%s ' "${0##*/}" >> "$ARCHIVE_TEST_CALLS"
case "${0##*/}" in
unrar)
  password_arg=-p-
  if [ -n "$ARCHIVE_TEST_PASSWORD" ]; then password_arg="-p$ARCHIVE_TEST_PASSWORD"; fi
  [ "$1" = x ] && [ "$2" = -y ] && [ "$3" = -o+ ] && [ "$4" = "$password_arg" ] && [ "$5" = -- ] && [ "$6" = "$ARCHIVE_TEST_INPUT" ] && [ "$7" = "$ARCHIVE_TEST_OUTPUT/" ] || exit 31
  [ ! -f "$ARCHIVE_TEST_OUTPUT/partial" ] || exit 32
  ;;
unar)
  [ "$1" = -quiet ] && [ "$2" = -force-overwrite ] && [ "$3" = -no-directory ] && [ "$4" = -output-directory ] && [ "$5" = "$ARCHIVE_TEST_OUTPUT" ] && [ "$6" = -password ] && [ "$7" = "$ARCHIVE_TEST_PASSWORD" ] && [ "$8" = "$ARCHIVE_TEST_INPUT" ] || exit 33
  [ ! -f "$ARCHIVE_TEST_OUTPUT/partial" ] || exit 34
  ;;
esac
`
				if tool.result == "success" {
					script += "printf complete > \"$ARCHIVE_TEST_OUTPUT/result\"\n"
				} else {
					script += "printf partial > \"$ARCHIVE_TEST_OUTPUT/partial\"\nprintf '" + tool.result + "\\n'\nexit 2\n"
				}
				if err := os.WriteFile(filepath.Join(bin, tool.name), []byte(script), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			_, err := (&ArchiveImportService{}).extractArchive(context.Background(), tt.format, archive, out, tt.password, archiveEncodingAuto)
			if (err == nil) != tt.wantSuccess {
				t.Fatalf("extractArchive() = %v, wantSuccess=%v", err, tt.wantSuccess)
			}
			if isArchivePasswordRelatedError(err) != tt.wantPassword {
				t.Fatalf("extractArchive() = %v, wantPassword=%v", err, tt.wantPassword)
			}
			calls, readErr := os.ReadFile(filepath.Join(root, "calls"))
			if readErr != nil || string(calls) != tt.wantCalled {
				t.Fatalf("calls=%q, err=%v, want %q", calls, readErr, tt.wantCalled)
			}
			if !tt.wantSuccess {
				if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
					t.Fatalf("failed extraction left output: %v", statErr)
				}
			} else if data, readErr := os.ReadFile(filepath.Join(out, "result")); readErr != nil || string(data) != "complete" {
				t.Fatalf("successful extraction result=%q, err=%v", data, readErr)
			}
			if tt.name == "全部失败保留两级诊断" {
				for _, marker := range []string{"Headers Error", "CRC Failed", "Data Error"} {
					if !strings.Contains(err.Error(), marker) {
						t.Fatalf("missing %q in %v", marker, err)
					}
				}
			}
			if tt.name == "没有兼容工具说明依赖" && !strings.Contains(err.Error(), "未安装") {
				t.Fatalf("missing dependency hint: %v", err)
			}
		})
	}
}

func TestArchiveExtractorCancellationPreservesContext(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "7zz")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nprintf 'Headers Error\\n'\nexec /bin/sleep 30\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := runArchiveExtractor(ctx, exe, nil, "sample.rar")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("runArchiveExtractor() = %v, want context.DeadlineExceeded", err)
	}
}

func TestArchiveHeaderDiagnosticIsConcise(t *testing.T) {
	err := archiveCommandError("extract /private/archive.rar with 7zz", errors.New("exit status 2"), "7-Zip copyright banner\nERRORS: Headers Error\nPath = /private/archive.rar\nType = Rar5\nERRORS: Headers Error\n")
	if !strings.Contains(err.Error(), "Headers Error") || !strings.Contains(err.Error(), "头部") {
		t.Fatalf("missing actionable header diagnostic: %v", err)
	}
	if strings.Contains(err.Error(), "copyright") || strings.Contains(err.Error(), "/private") {
		t.Fatalf("diagnostic contains banner or private path: %v", err)
	}
}

func TestArchiveHeaderDiagnosticIgnoresFilenameMarkers(t *testing.T) {
	for _, tt := range []struct {
		output string
		want   bool
	}{
		{output: "Path = /Headers Error.rar\nCRC Failed", want: false},
		{output: "Path = /CRC Failed.rar\nERRORS: Headers Error", want: true},
	} {
		err := archiveCommandError("extract archive", errors.New("exit status 2"), tt.output)
		if errors.Is(err, ErrArchiveHeaders) != tt.want {
			t.Fatalf("output=%q, err=%v, want header=%v", tt.output, err, tt.want)
		}
	}
}

func TestArchiveExtractionCancelledDoesNotRunFallback(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&ArchiveImportService{}).extractArchive(ctx, "rar", filepath.Join(root, "sample.rar"), filepath.Join(root, "out"), "", archiveEncodingAuto)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("extractArchive() = %v, want context.Canceled", err)
	}
}
