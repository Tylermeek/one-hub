package controller

import (
	"net/http"
	"one-api/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// POST /api/context/switch
func SwitchContext(c *gin.Context) {
	userId := c.GetInt("id")

	var req struct {
		Type string `json:"type" binding:"required,oneof=user team"`
		Id   int    `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 权限验证
	if req.Type == "team" {
		if !model.IsTeamOwner(req.Id, userId) && !model.IsTeamMember(req.Id, userId) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队空间",
			})
			return
		}
	} else if req.Type == "user" {
		// 用户空间只能是自己
		req.Id = userId
	}

	// 更新 Session
	session := sessions.Default(c)
	session.Set("context_type", req.Type)
	session.Set("context_id", req.Id)
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "空间切换成功",
		"data": gin.H{
			"type": req.Type,
			"id":   req.Id,
		},
	})
}

// GET /api/context/current
func GetCurrentContext(c *gin.Context) {
	userId := c.GetInt("id")
	session := sessions.Default(c)

	contextType := session.Get("context_type")
	contextId := session.Get("context_id")

	// 如果 Session 中没有上下文，默认为用户空间
	if contextType == nil {
		contextType = "user"
		contextId = userId
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"type": contextType,
			"id":   contextId,
		},
	})
}

