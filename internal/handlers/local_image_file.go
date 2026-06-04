package handlers

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	localImageOpenTimeout       = 2 * time.Second
	localImageOpenLimit         = 8
	localImageRetryAfterSeconds = 5
)

var (
	errLocalImageOpenTimeout = errors.New("local image open timeout")
	errLocalImageOpenBusy    = errors.New("local image open busy")
	errLocalImageIsDir       = errors.New("local image path is directory")

	localImageOpenLimiter = make(chan struct{}, localImageOpenLimit)
)

type localImageOpenFunc func(string) (*os.File, error)

type localImageOpenResult struct {
	file *os.File
	info os.FileInfo
	err  error
}

func openLocalImageFile(ctx context.Context, path string) (*os.File, os.FileInfo, error) {
	return openLocalImageFileWith(ctx, path, localImageOpenTimeout, localImageOpenLimiter, os.Open)
}

func openLocalImageFileWith(ctx context.Context, path string, timeout time.Duration, limiter chan struct{}, opener localImageOpenFunc) (*os.File, os.FileInfo, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil, os.ErrNotExist
	}
	if timeout <= 0 {
		timeout = localImageOpenTimeout
	}
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case limiter <- struct{}{}:
	default:
		return nil, nil, errLocalImageOpenBusy
	}

	openCtx, cancel := context.WithCancel(ctx)
	resultCh := make(chan localImageOpenResult)
	go func() {
		defer func() {
			<-limiter
		}()

		file, err := opener(path)
		if err != nil {
			sendLocalImageOpenResult(openCtx, resultCh, localImageOpenResult{err: err})
			return
		}
		if file == nil {
			sendLocalImageOpenResult(openCtx, resultCh, localImageOpenResult{err: os.ErrNotExist})
			return
		}
		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			sendLocalImageOpenResult(openCtx, resultCh, localImageOpenResult{err: err})
			return
		}
		if info.IsDir() {
			_ = file.Close()
			sendLocalImageOpenResult(openCtx, resultCh, localImageOpenResult{err: errLocalImageIsDir})
			return
		}
		sendLocalImageOpenResult(openCtx, resultCh, localImageOpenResult{
			file: file,
			info: info,
		})
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	defer cancel()

	select {
	case result := <-resultCh:
		if result.err != nil {
			return nil, nil, result.err
		}
		return result.file, result.info, nil
	case <-timer.C:
		return nil, nil, errLocalImageOpenTimeout
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}

func sendLocalImageOpenResult(ctx context.Context, resultCh chan<- localImageOpenResult, result localImageOpenResult) {
	select {
	case resultCh <- result:
	case <-ctx.Done():
		if result.file != nil {
			_ = result.file.Close()
		}
	}
}

func serveOpenedLocalImage(c *gin.Context, path string, file *os.File, info os.FileInfo) {
	modTime := time.Time{}
	if info != nil {
		c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
		modTime = info.ModTime()
	}
	if mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); mimeType != "" {
		c.Header("Content-Type", mimeType)
	}
	c.Header("Cache-Control", "public, max-age=86400")
	http.ServeContent(c.Writer, c.Request, filepath.Base(path), modTime, file)
}

func tryServeLocalImagePath(c *gin.Context, path, notFoundMessage, unavailableMessage string) bool {
	file, info, err := openLocalImageFile(c.Request.Context(), path)
	if err != nil {
		writeLocalImageOpenError(c, err, notFoundMessage, unavailableMessage)
		return false
	}
	defer file.Close()
	serveOpenedLocalImage(c, path, file, info)
	return true
}

func writeLocalImageOpenError(c *gin.Context, err error, notFoundMessage, unavailableMessage string) {
	switch {
	case isLocalImageNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"msg": notFoundMessage})
	case isLocalImageTemporarilyUnavailable(err):
		c.Header("Retry-After", strconv.Itoa(localImageRetryAfterSeconds))
		c.JSON(http.StatusServiceUnavailable, gin.H{"msg": unavailableMessage})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"msg": unavailableMessage})
	}
}

func isLocalImageNotFound(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, errLocalImageIsDir)
}

func isLocalImageTemporarilyUnavailable(err error) bool {
	return errors.Is(err, errLocalImageOpenTimeout) ||
		errors.Is(err, errLocalImageOpenBusy) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}
