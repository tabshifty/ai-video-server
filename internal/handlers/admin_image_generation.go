package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

const (
	imageGenerationMaxImages          = 4
	imageGenerationMaxReferenceImages = 4
	imageGenerationMaxReferenceBytes  = 10 * 1024 * 1024
	imageGenerationMaxPayloadBytes    = 40 * 1024 * 1024
	imageGenerationMaxRequestBytes    = 56 * 1024 * 1024
)

var allowedImageGenerationMIMEs = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
}

type ImageGenerationConfig struct {
	APIURL  string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func (cfg ImageGenerationConfig) normalized() ImageGenerationConfig {
	cfg.APIURL = strings.TrimSuffix(strings.TrimSpace(cfg.APIURL), "/")
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	if cfg.Model == "" {
		cfg.Model = "gpt-image-2"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 180 * time.Second
	}
	return cfg
}

func (cfg ImageGenerationConfig) enabled() bool {
	cfg = cfg.normalized()
	return cfg.APIURL != "" && cfg.APIKey != ""
}

type adminImageGenerationRequest struct {
	Prompt            string                          `json:"prompt"`
	Size              string                          `json:"size"`
	Quality           string                          `json:"quality"`
	OutputFormat      string                          `json:"output_format"`
	OutputCompression *int                            `json:"output_compression"`
	N                 int                             `json:"n"`
	ReferenceImages   []adminImageGenerationInputFile `json:"reference_images"`
	Mask              *adminImageGenerationMaskInput  `json:"mask"`
}

type adminImageGenerationInputFile struct {
	Name    string `json:"name"`
	MIME    string `json:"mime"`
	DataURL string `json:"data_url"`
}

type adminImageGenerationMaskInput struct {
	Name        string `json:"name"`
	MIME        string `json:"mime"`
	DataURL     string `json:"data_url"`
	TargetIndex int    `json:"target_index"`
}

type adminImageGenerationResult struct {
	DataURL       string `json:"data_url"`
	MIME          string `json:"mime"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type upstreamImageResponse struct {
	Data []upstreamImageItem `json:"data"`
}

type upstreamImageItem struct {
	B64JSON       string `json:"b64_json"`
	URL           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}

func (a *API) AdminImageGenerationStatus(c *gin.Context) {
	cfg := a.imageGenerationConfig.normalized()
	ok(c, gin.H{
		"enabled":             cfg.enabled(),
		"model":               cfg.Model,
		"base_url_configured": cfg.APIURL != "",
		"api_key_configured":  cfg.APIKey != "",
		"timeout_seconds":     int(cfg.Timeout / time.Second),
		"limits": gin.H{
			"max_images":                  imageGenerationMaxImages,
			"max_reference_images":        imageGenerationMaxReferenceImages,
			"max_reference_image_bytes":   imageGenerationMaxReferenceBytes,
			"max_reference_payload_bytes": imageGenerationMaxPayloadBytes,
			"max_request_body_bytes":      imageGenerationMaxRequestBytes,
			"allowed_mime_types":          []string{"image/png", "image/jpeg", "image/webp"},
		},
	})
}

func (a *API) AdminImageGenerate(c *gin.Context) {
	cfg := a.imageGenerationConfig.normalized()
	if !cfg.enabled() {
		response.Error(c, 2401, "后端未配置图像生成服务")
		return
	}

	var req adminImageGenerationRequest
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, imageGenerationMaxRequestBytes)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 2402, "图像生成请求体过大或格式无效")
		return
	}
	normalizedReq, references, mask, err := normalizeAdminImageGenerationRequest(req)
	if err != nil {
		response.Error(c, 2402, err.Error())
		return
	}

	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Timeout)
	defer cancel()
	results, err := callImageGenerationUpstream(ctx, cfg, normalizedReq, references, mask)
	if err != nil {
		requestID := fmt.Sprintf("img-%d", startedAt.UnixNano())
		if a.logger != nil {
			a.logger.Error("admin image generation failed", "request_id", requestID, "error", err)
		}
		response.Error(c, 2403, fmt.Sprintf("图像生成失败（请求 ID：%s）", requestID))
		return
	}

	ok(c, gin.H{
		"items":      results,
		"model":      cfg.Model,
		"created_at": time.Now().Format(time.RFC3339),
	})
}

func normalizeAdminImageGenerationRequest(req adminImageGenerationRequest) (adminImageGenerationRequest, []decodedImageGenerationFile, *normalizedImageGenerationMask, error) {
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		return req, nil, nil, errors.New("提示词不能为空")
	}
	req.Size = normalizeImageGenerationChoice(req.Size, "auto", map[string]struct{}{
		"auto": {}, "1024x1024": {}, "1024x1536": {}, "1536x1024": {},
	})
	req.Quality = normalizeImageGenerationChoice(req.Quality, "auto", map[string]struct{}{
		"auto": {}, "low": {}, "medium": {}, "high": {},
	})
	req.OutputFormat = normalizeImageGenerationChoice(req.OutputFormat, "png", map[string]struct{}{
		"png": {}, "jpeg": {}, "webp": {},
	})
	if req.N <= 0 {
		req.N = 1
	}
	if req.N > imageGenerationMaxImages {
		return req, nil, nil, fmt.Errorf("单次最多生成 %d 张图片", imageGenerationMaxImages)
	}
	if len(req.ReferenceImages) > imageGenerationMaxReferenceImages {
		return req, nil, nil, fmt.Errorf("参考图最多 %d 张", imageGenerationMaxReferenceImages)
	}

	references := make([]decodedImageGenerationFile, 0, len(req.ReferenceImages))
	totalBytes := 0
	for i, item := range req.ReferenceImages {
		decoded, err := decodeImageGenerationReference(item, i)
		if err != nil {
			return req, nil, nil, err
		}
		if len(decoded.Data) > imageGenerationMaxReferenceBytes {
			return req, nil, nil, fmt.Errorf("单张参考图最多 10 MiB")
		}
		totalBytes += len(decoded.Data)
		if totalBytes > imageGenerationMaxPayloadBytes {
			return req, nil, nil, fmt.Errorf("参考图总大小最多 40 MiB")
		}
		references = append(references, decoded)
	}
	if req.OutputFormat == "png" {
		req.OutputCompression = nil
	}

	var mask *normalizedImageGenerationMask
	if req.Mask != nil {
		if len(references) == 0 {
			return req, nil, nil, errors.New("局部蒙版至少需要一张参考图")
		}
		if req.Mask.TargetIndex < 0 || req.Mask.TargetIndex >= len(references) {
			return req, nil, nil, errors.New("蒙版目标参考图不存在")
		}
		if references[req.Mask.TargetIndex].MIME != "image/png" {
			return req, nil, nil, errors.New("局部蒙版的目标参考图必须为 PNG")
		}
		decodedMask, err := decodeImageGenerationMask(*req.Mask)
		if err != nil {
			return req, nil, nil, err
		}
		if len(decodedMask.Data) > imageGenerationMaxReferenceBytes {
			return req, nil, nil, fmt.Errorf("局部蒙版最多 10 MiB")
		}
		totalBytes += len(decodedMask.Data)
		if totalBytes > imageGenerationMaxPayloadBytes {
			return req, nil, nil, fmt.Errorf("参考图与蒙版总大小最多 40 MiB")
		}
		targetWidth, targetHeight := probeImageGenerationDimensions(references[req.Mask.TargetIndex].Data)
		maskWidth, maskHeight := probeImageGenerationDimensions(decodedMask.Data)
		if targetWidth <= 0 || targetHeight <= 0 || maskWidth <= 0 || maskHeight <= 0 {
			return req, nil, nil, errors.New("无法解析局部蒙版或目标参考图尺寸")
		}
		if targetWidth != maskWidth || targetHeight != maskHeight {
			return req, nil, nil, errors.New("局部蒙版尺寸与目标参考图不一致")
		}
		mask = &normalizedImageGenerationMask{
			File:        decodedMask,
			TargetIndex: req.Mask.TargetIndex,
		}
	}

	return req, references, mask, nil
}

func normalizeImageGenerationChoice(value, fallback string, allowed map[string]struct{}) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowed[normalized]; ok {
		return normalized
	}
	return fallback
}

type decodedImageGenerationFile struct {
	Name string
	MIME string
	Data []byte
}

type normalizedImageGenerationMask struct {
	File        decodedImageGenerationFile
	TargetIndex int
}

func decodeImageGenerationReference(item adminImageGenerationInputFile, index int) (decodedImageGenerationFile, error) {
	mime := strings.ToLower(strings.TrimSpace(item.MIME))
	dataURL := strings.TrimSpace(item.DataURL)
	if dataURL == "" {
		return decodedImageGenerationFile{}, fmt.Errorf("第 %d 张参考图为空", index+1)
	}
	if strings.HasPrefix(dataURL, "data:") {
		parsedMIME, payload, ok := strings.Cut(strings.TrimPrefix(dataURL, "data:"), ",")
		if !ok {
			return decodedImageGenerationFile{}, fmt.Errorf("第 %d 张参考图 data URL 无效", index+1)
		}
		if semi := strings.Index(parsedMIME, ";"); semi >= 0 {
			parsedMIME = parsedMIME[:semi]
		}
		if mime == "" {
			mime = strings.ToLower(strings.TrimSpace(parsedMIME))
		}
		dataURL = payload
	}
	if _, ok := allowedImageGenerationMIMEs[mime]; !ok {
		return decodedImageGenerationFile{}, fmt.Errorf("第 %d 张参考图类型不支持", index+1)
	}
	data, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		return decodedImageGenerationFile{}, fmt.Errorf("第 %d 张参考图 data URL 无效", index+1)
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = fmt.Sprintf("reference-%d.%s", index+1, imageGenerationExtForMIME(mime))
	}
	return decodedImageGenerationFile{Name: name, MIME: mime, Data: data}, nil
}

func decodeImageGenerationMask(mask adminImageGenerationMaskInput) (decodedImageGenerationFile, error) {
	mime := strings.ToLower(strings.TrimSpace(mask.MIME))
	dataURL := strings.TrimSpace(mask.DataURL)
	if dataURL == "" {
		return decodedImageGenerationFile{}, errors.New("局部蒙版不能为空")
	}
	if strings.HasPrefix(dataURL, "data:") {
		parsedMIME, payload, ok := strings.Cut(strings.TrimPrefix(dataURL, "data:"), ",")
		if !ok {
			return decodedImageGenerationFile{}, errors.New("局部蒙版 data URL 无效")
		}
		if semi := strings.Index(parsedMIME, ";"); semi >= 0 {
			parsedMIME = parsedMIME[:semi]
		}
		if mime == "" {
			mime = strings.ToLower(strings.TrimSpace(parsedMIME))
		}
		dataURL = payload
	}
	if mime != "image/png" {
		return decodedImageGenerationFile{}, errors.New("局部蒙版文件必须为 PNG")
	}
	data, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		return decodedImageGenerationFile{}, errors.New("局部蒙版 data URL 无效")
	}
	name := strings.TrimSpace(mask.Name)
	if name == "" {
		name = "mask.png"
	}
	return decodedImageGenerationFile{Name: name, MIME: mime, Data: data}, nil
}

func callImageGenerationUpstream(ctx context.Context, cfg ImageGenerationConfig, req adminImageGenerationRequest, references []decodedImageGenerationFile, mask *normalizedImageGenerationMask) ([]adminImageGenerationResult, error) {
	if len(references) > 0 {
		return callImageGenerationEdit(ctx, cfg, req, references, mask)
	}
	return callImageGenerationCreate(ctx, cfg, req)
}

func callImageGenerationCreate(ctx context.Context, cfg ImageGenerationConfig, req adminImageGenerationRequest) ([]adminImageGenerationResult, error) {
	payload := map[string]any{
		"model":         cfg.Model,
		"prompt":        req.Prompt,
		"size":          req.Size,
		"quality":       req.Quality,
		"output_format": req.OutputFormat,
		"moderation":    "auto",
		"n":             req.N,
	}
	if shouldRequestImageGenerationB64JSON(cfg.Model) {
		payload["response_format"] = "b64_json"
	}
	if req.OutputFormat != "png" && req.OutputCompression != nil {
		payload["output_compression"] = *req.OutputCompression
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIURL+"/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return doImageGenerationRequest(httpReq, cfg, req.OutputFormat)
}

func callImageGenerationEdit(ctx context.Context, cfg ImageGenerationConfig, req adminImageGenerationRequest, references []decodedImageGenerationFile, mask *normalizedImageGenerationMask) ([]adminImageGenerationResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"model":         cfg.Model,
		"prompt":        req.Prompt,
		"size":          req.Size,
		"quality":       req.Quality,
		"output_format": req.OutputFormat,
		"moderation":    "auto",
		"n":             fmt.Sprintf("%d", req.N),
	}
	if shouldRequestImageGenerationB64JSON(cfg.Model) {
		fields["response_format"] = "b64_json"
	}
	if req.OutputFormat != "png" && req.OutputCompression != nil {
		fields["output_compression"] = fmt.Sprintf("%d", *req.OutputCompression)
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	executionReferences := references
	if mask != nil {
		executionReferences = reorderImageGenerationReferencesForMask(references, mask.TargetIndex)
	}
	for _, reference := range executionReferences {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, escapeMultipartFilename(reference.Name)))
		header.Set("Content-Type", reference.MIME)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(reference.Data); err != nil {
			return nil, err
		}
	}
	if mask != nil {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="mask"; filename="%s"`, escapeMultipartFilename(mask.File.Name)))
		header.Set("Content-Type", mask.File.MIME)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(mask.File.Data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIURL+"/images/edits", &body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	return doImageGenerationRequest(httpReq, cfg, req.OutputFormat)
}

func reorderImageGenerationReferencesForMask(references []decodedImageGenerationFile, targetIndex int) []decodedImageGenerationFile {
	if targetIndex <= 0 || targetIndex >= len(references) {
		return append([]decodedImageGenerationFile(nil), references...)
	}
	reordered := append([]decodedImageGenerationFile(nil), references...)
	target := reordered[targetIndex]
	copy(reordered[1:targetIndex+1], reordered[0:targetIndex])
	reordered[0] = target
	return reordered
}

func shouldRequestImageGenerationB64JSON(model string) bool {
	return !strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "gpt-image-")
}

func escapeMultipartFilename(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(value)
}

func doImageGenerationRequest(req *http.Request, cfg ImageGenerationConfig, outputFormat string) ([]adminImageGenerationResult, error) {
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream image api status=%d error=%s", resp.StatusCode, extractUpstreamImageError(body))
	}
	var payload upstreamImageResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode upstream image response: %w", err)
	}
	if len(payload.Data) == 0 {
		return nil, errors.New("upstream image api returned no image")
	}
	return normalizeImageGenerationResults(req.Context(), payload.Data, outputFormat, cfg.Timeout)
}

func extractUpstreamImageError(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return strings.TrimSpace(string(body))
	}
	if errValue, ok := payload["error"].(map[string]any); ok {
		if msg, ok := errValue["message"].(string); ok {
			return msg
		}
	}
	if msg, ok := payload["error"].(string); ok {
		return msg
	}
	if msg, ok := payload["message"].(string); ok {
		return msg
	}
	return strings.TrimSpace(string(body))
}

func normalizeImageGenerationResults(ctx context.Context, items []upstreamImageItem, outputFormat string, timeout time.Duration) ([]adminImageGenerationResult, error) {
	results := make([]adminImageGenerationResult, 0, len(items))
	fallbackMIME := imageGenerationMIMEForFormat(outputFormat)
	for _, item := range items {
		dataURL := ""
		mime := fallbackMIME
		var raw []byte
		var err error
		if strings.TrimSpace(item.B64JSON) != "" {
			raw, err = base64.StdEncoding.DecodeString(strings.TrimSpace(item.B64JSON))
			if err != nil {
				return nil, fmt.Errorf("decode upstream b64_json: %w", err)
			}
			dataURL = "data:" + mime + ";base64," + strings.TrimSpace(item.B64JSON)
		} else if strings.HasPrefix(strings.TrimSpace(item.URL), "data:") {
			dataURL = strings.TrimSpace(item.URL)
			parsedMIME, parsedRaw, err := parseImageGenerationDataURL(dataURL)
			if err != nil {
				return nil, err
			}
			mime = parsedMIME
			raw = parsedRaw
		} else if isHTTPImageGenerationURL(item.URL) {
			downloadedMIME, downloadedRaw, err := downloadImageGenerationURL(ctx, strings.TrimSpace(item.URL), timeout)
			if err != nil {
				return nil, err
			}
			mime = downloadedMIME
			raw = downloadedRaw
			dataURL = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw)
		}
		if dataURL == "" {
			continue
		}
		width, height := probeImageGenerationDimensions(raw)
		results = append(results, adminImageGenerationResult{
			DataURL:       dataURL,
			MIME:          mime,
			Width:         width,
			Height:        height,
			RevisedPrompt: strings.TrimSpace(item.RevisedPrompt),
		})
	}
	if len(results) == 0 {
		return nil, errors.New("upstream image api returned no usable image")
	}
	return results, nil
}

func parseImageGenerationDataURL(dataURL string) (string, []byte, error) {
	header, payload, ok := strings.Cut(strings.TrimPrefix(dataURL, "data:"), ",")
	if !ok {
		return "", nil, errors.New("invalid data URL")
	}
	mime := header
	if semi := strings.Index(mime, ";"); semi >= 0 {
		mime = mime[:semi]
	}
	mime = strings.ToLower(strings.TrimSpace(mime))
	if _, ok := allowedImageGenerationMIMEs[mime]; !ok {
		return "", nil, fmt.Errorf("unsupported image MIME: %s", mime)
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", nil, err
	}
	return mime, raw, nil
}

func isHTTPImageGenerationURL(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(normalized, "http://") || strings.HasPrefix(normalized, "https://")
}

func downloadImageGenerationURL(ctx context.Context, url string, timeout time.Duration) (string, []byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("download upstream image status=%d", resp.StatusCode)
	}
	mime := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	if _, ok := allowedImageGenerationMIMEs[mime]; !ok {
		return "", nil, fmt.Errorf("downloaded image MIME is unsupported: %s", mime)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, imageGenerationMaxReferenceBytes*imageGenerationMaxImages))
	if err != nil {
		return "", nil, err
	}
	return mime, raw, nil
}

func probeImageGenerationDimensions(raw []byte) (int, int) {
	if len(raw) == 0 {
		return 0, 0
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func imageGenerationMIMEForFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func imageGenerationExtForMIME(mime string) string {
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}
