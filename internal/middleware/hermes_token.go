package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HermesTokenMiddleware(configuredToken string) gin.HandlerFunc {
	expected := strings.TrimSpace(configuredToken)
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		provided, ok := strings.CutPrefix(authorization, "Bearer ")
		provided = strings.TrimSpace(provided)
		valid := ok && expected != "" && len(provided) == len(expected) &&
			subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Hermes 鉴权失败",
				"data": nil,
			})
			return
		}
		c.Next()
	}
}
