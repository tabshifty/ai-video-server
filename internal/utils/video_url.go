package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func VideoPlayURL(videoID uuid.UUID) string {
	return fmt.Sprintf("/api/v1/videos/%s/source", videoID.String())
}
