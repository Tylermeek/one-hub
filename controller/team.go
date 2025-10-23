package controller

import (
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTeamRequest struct {
	Name   string `json:"name" binding:"required"`
	Status int    `json:"status"`
}

type AllocateTeamQuotaRequest struct {
	Quota      int  `json:"quota"`
	Unlimited  bool `json:"unlimited"`
}

func CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	ownerId := c.GetInt("id")
	team := &model.Team{
		Name:    req.Name,
		OwnerId: ownerId,
		Status:  1, // 默认启用
	}

	if err := team.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 自动将创建者添加为团队成员（管理员角色）
	member := &model.TeamMember{
		TeamId: team.Id,
		UserId: ownerId,
		Role:   1, // 管理员
		Status: 1, // 正常状态
	}
	if err := member.Insert(); err != nil {
		// 如果添加成员失败，删除团队
		team.Delete()
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "创建团队失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "团队创建成功",
		"data":    team,
	})
}

func GetUserTeams(c *gin.Context) {
	var params model.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	userId := c.GetInt("id")
	teams, err := model.GetUserTeams(userId, &params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 为每个团队添加当前用户的角色信息
	if teams != nil && teams.Data != nil {
		for _, team := range *teams.Data {
			// 检查是否为团队所有者
			if model.IsTeamOwner(team.Id, userId) {
				team.CurrentUserRole = 1
				team.IsOwner = true
			} else {
				// 获取团队成员信息
				member, err := model.GetTeamMember(team.Id, userId)
				if err != nil {
					team.CurrentUserRole = 2 // 默认为普通成员
					team.IsOwner = false
				} else {
					team.CurrentUserRole = member.Role
					team.IsOwner = false
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    teams,
	})
}

func GetTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否有权限查看此团队
	if !model.IsTeamOwner(id, userId) && !model.IsTeamMember(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权限查看此团队",
		})
		return
	}

	team, err := model.GetTeamById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 🔐 仅 Owner 可以看到个人余额
	if model.IsTeamOwner(id, userId) {
		owner, err := model.GetUserById(team.OwnerId, false)
		if err == nil {
			team.OwnerBalance = owner.Quota
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    team,
	})
}

func UpdateTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以修改团队信息",
		})
		return
	}

	var req UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	team := &model.Team{
		Id:     id,
		Name:   req.Name,
		Status: req.Status,
	}

	if err := team.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "团队信息更新成功",
	})
}

func DeleteTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以删除团队",
		})
		return
	}

	team, err := model.GetTeamById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	if err := team.Delete(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "团队删除成功",
	})
}

func AllocateTeamQuota(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以分配团队额度",
		})
		return
	}

	var req AllocateTeamQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if !req.Unlimited && req.Quota <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "额度必须大于0",
		})
		return
	}

	if err := model.AllocateQuotaToTeam(userId, id, req.Quota, req.Unlimited); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	message := "团队额度分配成功"
	if req.Unlimited {
		message = "团队已设置为无限额度"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}
