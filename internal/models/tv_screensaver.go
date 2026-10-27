package models

import (
	"time"

	"github.com/google/uuid"
)

// TVScreensaverImage 是普通客户端可见的屏保图片，不暴露文件路径或管理员元数据。
type TVScreensaverImage struct {
	ID        uuid.UUID `json:"id"`
	ViewURL   string    `json:"view_url"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TVScreensaverImages 按图片身份游标读取全图库，空游标表示本轮读取完成。
type TVScreensaverImages struct {
	Items      []TVScreensaverImage `json:"items"`
	NextCursor string               `json:"next_cursor"`
}
