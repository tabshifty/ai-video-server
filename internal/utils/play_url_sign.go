package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

func SignVideoSource(secret string, videoID uuid.UUID, expUnix int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d", videoID.String(), expUnix)))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyVideoSourceSign(secret string, videoID uuid.UUID, expUnix int64, sigHex string) bool {
	got := SignVideoSource(secret, videoID, expUnix)
	return hmac.Equal([]byte(got), []byte(sigHex))
}
