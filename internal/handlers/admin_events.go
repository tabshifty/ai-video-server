package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"video-server/internal/response"
)

// AdminEventsStream streams server-sent events for admin dashboard/task updates.
// Endpoint path keeps /events/ws for compatibility with planned frontend route naming.
func (a *API) AdminEventsStream(c *gin.Context) {
	if a.redis == nil {
		response.Error(c, 1030, "redis unavailable")
		return
	}
	pubsub := a.redis.Subscribe(c.Request.Context(), "admin:events")
	defer pubsub.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, 1031, "streaming unsupported")
		return
	}

	_, _ = fmt.Fprintf(c.Writer, "event: ready\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	ch := pubsub.Channel()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			_, _ = fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", msg.Payload)
			flusher.Flush()
		}
	}
}
