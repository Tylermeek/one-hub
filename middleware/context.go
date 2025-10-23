package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ContextMiddleware 解析请求头中的上下文信息并注入到 gin.Context
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取上下文信息
		contextType := c.GetHeader("X-Context-Type") // "user" or "team"
		contextIdStr := c.GetHeader("X-Context-Id")  // id

		// 如果没有传递上下文信息，默认为用户上下文
		if contextType == "" {
			contextType = "user"
			// 从用户认证中间件获取用户ID
			if userId, exists := c.Get("id"); exists {
				contextIdStr = strconv.Itoa(userId.(int))
			}
		}

		// 解析上下文ID
		var contextId int
		if contextIdStr != "" {
			if id, err := strconv.Atoi(contextIdStr); err == nil {
				contextId = id
			}
		}

		// 将上下文信息注入到 gin.Context
		c.Set("context_type", contextType)
		c.Set("context_id", contextId)

		c.Next()
	}
}
