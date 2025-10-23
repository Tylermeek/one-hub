package controller

import (
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchUsersRequest struct {
	Keyword string `form:"keyword" binding:"required"`
}

type InviteMemberRequest struct {
	UserId int `json:"user_id" binding:"required"`
}

type UpdateMemberQuotaRequest struct {
	MaxQuota int `json:"max_quota" binding:"min=0"`
}

type RegisterWithInviteRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	InviteCode string `json:"invite_code" binding:"required"`
}

func SearchUsers(c *gin.Context) {
	var req SearchUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	// 搜索用户（排除当前用户）
	userId := c.GetInt("id")
	var params model.GenericParams
	params.Keyword = req.Keyword
	params.Page = 1
	params.Size = 20

	users, err := model.GetUsersList(&params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 过滤掉当前用户
	var filteredUsers []*model.User
	for _, user := range *users.Data {
		if user.Id != userId {
			filteredUsers = append(filteredUsers, user)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    filteredUsers,
	})
}

func GetTeamMembers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否有权限查看团队成员
	if !model.IsTeamOwner(id, userId) && !model.IsTeamMember(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权限查看此团队成员",
		})
		return
	}

	// 获取当前用户在该团队中的角色信息
	var currentUserRole int
	var isOwner bool
	
	if model.IsTeamOwner(id, userId) {
		isOwner = true
		currentUserRole = 1 // Owner 视为管理员角色
	} else {
		// 获取团队成员信息
		member, err := model.GetTeamMember(id, userId)
		if err != nil {
			currentUserRole = 2 // 默认为普通成员
		} else {
			currentUserRole = member.Role
		}
	}

	var params model.SearchTeamMemberParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	// 如果是普通成员，只返回自己的信息
	if currentUserRole == 2 {
		member, err := model.GetTeamMember(id, userId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "获取成员信息失败",
			})
			return
		}
		
		// 为成员计算可用额度
		availableQuota, unlimited, err := model.GetMemberAvailableQuota(id, userId)
		if err == nil {
			member.AvailableQuota = availableQuota
			member.UnlimitedQuota = unlimited
		}
		
		// 构建单个成员的响应
		responseData := map[string]interface{}{
			"data": []*model.TeamMember{member},
			"total": 1,
			"page": 1,
			"size": 1,
		}
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data":    responseData,
			"current_user_role": currentUserRole,
			"is_owner": isOwner,
		})
		return
	}

	// 管理员和所有者可以看到所有成员
	members, err := model.GetTeamMembersList(id, &params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    members,
		"current_user_role": currentUserRole,
		"is_owner": isOwner,
	})
}

func InviteMember(c *gin.Context) {
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
			"message": "只有团队管理员可以邀请成员",
		})
		return
	}

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查被邀请用户是否存在
	invitedUser, err := model.GetUserById(req.UserId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	// 检查用户是否已经是团队成员
	if model.IsTeamMember(id, req.UserId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已经是团队成员",
		})
		return
	}

	// 添加成员
	member := &model.TeamMember{
		TeamId: id,
		UserId: req.UserId,
		Role:   2, // 普通成员
		Status: 1, // 正常状态
	}

	if err := member.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 记录日志
	model.RecordLog(userId, model.LogTypeManage, 
		"邀请用户 "+invitedUser.Username+" 加入团队")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成员邀请成功",
	})
}

func RemoveMember(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	memberUserId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以移除成员",
		})
		return
	}

	// 不能移除自己
	if memberUserId == userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不能移除自己",
		})
		return
	}

	// 获取被移除用户信息
	memberUser, err := model.GetUserById(memberUserId, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if err := model.DeleteTeamMember(id, memberUserId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 记录日志
	model.RecordLog(userId, model.LogTypeManage, 
		"移除团队成员 "+memberUser.Username)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成员移除成功",
	})
}

func UpdateMemberQuota(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	memberUserId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以设置成员额度",
		})
		return
	}

	var req UpdateMemberQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := model.UpdateTeamMemberQuota(id, memberUserId, req.MaxQuota); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 获取用户信息用于日志
	memberUser, _ := model.GetUserById(memberUserId, false)
	username := "未知用户"
	if memberUser != nil {
		username = memberUser.Username
	}

	quotaText := "无限制"
	if req.MaxQuota > 0 {
		quotaText = common.LogQuota(req.MaxQuota)
	}

	model.RecordLog(userId, model.LogTypeManage, 
		"设置团队成员 "+username+" 最大额度为 "+quotaText)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成员额度设置成功",
	})
}

func GetMemberUsage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队ID格式错误",
		})
		return
	}

	memberUserId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户ID格式错误",
		})
		return
	}

	userId := c.GetInt("id")
	
	// 检查用户是否是团队管理员
	if !model.IsTeamOwner(id, userId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "只有团队管理员可以查看成员使用记录",
		})
		return
	}

	var params model.LogsListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	// 只查询该成员在团队中的使用记录
	params.UserId = memberUserId
	params.TeamId = id

	logs, err := model.GetUserLogsList(memberUserId, &params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}

func RegisterWithInvite(c *gin.Context) {
	var req RegisterWithInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证邀请码
	team, err := model.GetTeamByInviteCode(req.InviteCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邀请码无效",
		})
		return
	}

	// 检查团队状态
	if team.Status != 1 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "团队已禁用",
		})
		return
	}

	// 创建用户
	user := &model.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.Username,
		Email:       req.Email,
	}

	if err := user.Insert(0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户注册失败: " + err.Error(),
		})
		return
	}

	// 添加用户到团队
	member := &model.TeamMember{
		TeamId: team.Id,
		UserId: user.Id,
		Role:   2, // 普通成员
		Status: 1, // 正常状态
	}

	if err := member.Insert(); err != nil {
		// 如果加入团队失败，删除用户
		user.Delete()
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "加入团队失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	model.RecordLog(user.Id, model.LogTypeSystem, 
		"通过邀请码注册并加入团队 "+team.Name)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注册成功并已加入团队",
	})
}
