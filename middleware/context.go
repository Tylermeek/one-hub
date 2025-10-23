package middleware

import (
	"one-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ContextMiddleware 从 Session 读取上下文信息并注入到 gin.Context
func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		session := sessions.Default(c)

		// 优先从 Session 读取上下文
		contextType := session.Get("context_type")
		contextId := session.Get("context_id")

		// 如果 Session 中没有，默认为用户空间
		if contextType == nil || contextId == nil {
			contextType = "user"
			contextId = userId
			// 初始化 Session
			session.Set("context_type", contextType)
			session.Set("context_id", contextId)
			session.Save()
		}

		// 二次验证：确保用户仍有权限访问该空间
		if contextType == "team" {
			teamId := contextId.(int)
			if !model.IsTeamOwner(teamId, userId) && !model.IsTeamMember(teamId, userId) {
				// 权限丢失，重置为用户空间
				contextType = "user"
				contextId = userId
				session.Set("context_type", contextType)
				session.Set("context_id", contextId)
				session.Save()
			}
		}

		// 注入到 gin.Context
		c.Set("context_type", contextType.(string))
		c.Set("context_id", contextId.(int))

		c.Next()
	}
}
