package response

import "github.com/gin-gonic/gin"

// JSON writes unified API response payload.
func JSON(c *gin.Context, code int, msg string, data any) {
	c.JSON(200, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// Error writes unified API error payload.
func Error(c *gin.Context, code int, msg string) {
	JSON(c, code, msg, gin.H{})
}
