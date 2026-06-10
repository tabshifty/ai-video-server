package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRegisterIncludesAdminImageGenerationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &API{}
	api.Register(router)

	routes := map[string]struct{}{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/admin/image-generation/status",
		"POST /api/v1/admin/image-generation/generate",
	} {
		if _, ok := routes[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestAdminImageGenerationStatusIsRedacted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/image-generation/status", nil)
	api := &API{imageGenerationConfig: ImageGenerationConfig{
		APIURL:  "https://image.example/v1",
		APIKey:  "secret-token",
		Model:   "gpt-image-test",
		Timeout: 181 * time.Second,
	}}

	api.AdminImageGenerationStatus(ctx)

	body := rec.Body.String()
	for _, want := range []string{
		`"enabled":true`,
		`"model":"gpt-image-test"`,
		`"base_url_configured":true`,
		`"api_key_configured":true`,
		`"timeout_seconds":181`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected response to contain %s, got %s", want, body)
		}
	}
	for _, forbidden := range []string{"secret-token", "image.example"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("status response leaked %q: %s", forbidden, body)
		}
	}
}

func TestNormalizeAdminImageGenerationRequestRejectsTooManyImages(t *testing.T) {
	_, _, err := normalizeAdminImageGenerationRequest(adminImageGenerationRequest{
		Prompt:       "生成海报",
		OutputFormat: "png",
		N:            imageGenerationMaxImages + 1,
	})
	if err == nil || !strings.Contains(err.Error(), "单次最多生成") {
		t.Fatalf("expected max images error, got %v", err)
	}
}

func TestAdminImageGenerateCallsGenerationEndpointAndReturnsDataURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	inputPNG := testPNGBytes(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/generations" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer image-token" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if _, exists := payload["response_format"]; exists {
			t.Fatalf("gpt-image models should not send response_format, got=%v", payload["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + base64.StdEncoding.EncodeToString(inputPNG) + `","revised_prompt":"改写提示词"}]}`))
	}))
	defer upstream.Close()

	body := bytes.NewBufferString(`{"prompt":"生成一张海报","size":"1024x1024","quality":"auto","output_format":"png","n":1}`)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/image-generation/generate", body)
	ctx.Request.Header.Set("Content-Type", "application/json")
	api := &API{imageGenerationConfig: ImageGenerationConfig{
		APIURL:  upstream.URL,
		APIKey:  "image-token",
		Model:   "gpt-image-2",
		Timeout: 2 * time.Second,
	}}

	api.AdminImageGenerate(ctx)

	responseBody := rec.Body.String()
	for _, want := range []string{
		`"code":0`,
		`"data:image/png;base64,`,
		`"mime":"image/png"`,
		`"width":2`,
		`"height":1`,
		`"revised_prompt":"改写提示词"`,
	} {
		if !strings.Contains(responseBody, want) {
			t.Fatalf("expected response to contain %s, got %s", want, responseBody)
		}
	}
}

func TestAdminImageGenerateCallsEditEndpointForReferenceImages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	inputPNG := testPNGBytes(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/images/edits" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(12 * 1024 * 1024); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if r.FormValue("model") != "gpt-image-test" {
			t.Fatalf("unexpected model: %s", r.FormValue("model"))
		}
		if files := r.MultipartForm.File["image"]; len(files) != 1 {
			t.Fatalf("expected one reference image, got %d", len(files))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + base64.StdEncoding.EncodeToString(inputPNG) + `"}]}`))
	}))
	defer upstream.Close()

	refDataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(inputPNG)
	payload := `{"prompt":"改成夜景","output_format":"png","reference_images":[{"name":"ref.png","mime":"image/png","data_url":"` + refDataURL + `"}]}`
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/image-generation/generate", bytes.NewBufferString(payload))
	ctx.Request.Header.Set("Content-Type", "application/json")
	api := &API{imageGenerationConfig: ImageGenerationConfig{
		APIURL:  upstream.URL,
		APIKey:  "image-token",
		Model:   "gpt-image-test",
		Timeout: 2 * time.Second,
	}}

	api.AdminImageGenerate(ctx)

	if !strings.Contains(rec.Body.String(), `"code":0`) || !strings.Contains(rec.Body.String(), `"data:image/png;base64,`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func TestAdminImageGenerateRequestsB64JSONForCompatibleModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	inputPNG := testPNGBytes(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["response_format"] != "b64_json" {
			t.Fatalf("expected compatible model to request b64_json, got=%v", payload["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + base64.StdEncoding.EncodeToString(inputPNG) + `"}]}`))
	}))
	defer upstream.Close()

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/image-generation/generate", bytes.NewBufferString(`{"prompt":"生成海报"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	api := &API{imageGenerationConfig: ImageGenerationConfig{
		APIURL:  upstream.URL,
		APIKey:  "image-token",
		Model:   "compatible-image-model",
		Timeout: 2 * time.Second,
	}}

	api.AdminImageGenerate(ctx)

	if !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func testPNGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{B: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}
