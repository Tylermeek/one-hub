package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/limit"
	"one-api/common/utils"
	"one-api/model"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	if !config.PasswordLoginEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了密码登录",
			"success": false,
		})
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无效的参数",
			"success": false,
		})
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"message": "无效的参数",
			"success": false,
		})
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

// setup session & cookies and then return user info
func setupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "无法保存会话信息，请重试",
			"success": false,
		})
		return
	}
	user.LastLoginTime = time.Now().Unix()
  user.LastLoginIp = c.ClientIP()

	user.Update(false)

	cleanUser := model.User{
		Id:          user.Id,
		AvatarUrl:   user.AvatarUrl,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    cleanUser,
	})
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
	})
}

func Register(c *gin.Context) {
	if !config.RegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了新用户注册",
			"success": false,
		})
		return
	}
	if !config.PasswordRegisterEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册",
			"success": false,
		})
		return
	}
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	if config.EmailVerificationEnabled {
		if user.Email == "" || user.VerificationCode == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员开启了邮箱验证，请输入邮箱地址和验证码",
			})
			return
		}
		if !common.VerifyCodeWithKey(user.Email, user.VerificationCode, common.EmailVerificationPurpose) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "验证码错误或已过期",
			})
			return
		}
	}
	affCode := user.AffCode // this code is the inviter's code, not the user's own code
	inviterId, _ := model.GetUserIdByAffCode(affCode)
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.Username,
		InviterId:   inviterId,
	}
	if config.EmailVerificationEnabled {
		cleanUser.Email = user.Email
	}
	if err := cleanUser.Insert(inviterId); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func GetUsersList(c *gin.Context) {
	var params model.GenericParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	users, err := model.GetUsersList(&params)
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	// 为每个用户计算团队数量
	if users.Data != nil {
		for _, user := range *users.Data {
			// 获取团队数量
			teamIds, err := model.GetUserTeamIds(user.Id)
			if err == nil {
				user.TeamCount = len(teamIds)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    users,
	})
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != config.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权获取同级或更高等级用户的信息",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
}

const API_LIMIT_KEY = "api-limiter:%d"

func GetRateRealtime(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	limiter := model.GlobalUserGroupRatio.GetAPILimiter(user.Group)
	key := fmt.Sprintf(API_LIMIT_KEY, id)
	// 获取当前已使用的速率
	rpm, err := limiter.GetCurrentRate(key)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	maxRPM := limit.GetMaxRate(limiter)
	var usageRpmRate float64 = 0
	if maxRPM > 0 {
		usageRpmRate = math.Floor(float64(rpm)/float64(maxRPM)*100*100) / 100
	}

	data := map[string]interface{}{
		"rpm":          rpm,
		"maxRPM":       maxRPM,
		"usageRpmRate": usageRpmRate,
		"tpm":          0,
		"maxTPM":       0,
		"usageTpmRate": 0,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

func GetUserDashboard(c *gin.Context) {
	id := c.GetInt("id")
	
	// 获取上下文信息
	contextType := c.GetString("context_type")
	contextId := c.GetInt("context_id")
	
	// 根据上下文类型确定 teamId
	var teamId int
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !model.IsTeamMember(contextId, id) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队的统计信息",
			})
			return
		}
		teamId = contextId
	} else {
		// 个人空间，teamId = 0
		teamId = 0
	}

	now := time.Now()
	toDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := toDay.Add(-time.Second).Add(time.Hour * 24).Format("2006-01-02")
	startOfDay := toDay.AddDate(0, 0, -7).Format("2006-01-02")

	dashboards, err := model.GetUserModelStatisticsByContext(id, teamId, startOfDay, endOfDay)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取统计信息.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    dashboards,
	})
}

func GetUserDashboardRecentLogs(c *gin.Context) {
	id := c.GetInt("id")
	
	// 获取上下文信息
	contextType := c.GetString("context_type")
	contextId := c.GetInt("context_id")
	
	// 根据上下文类型确定 teamId
	var teamId int
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !model.IsTeamMember(contextId, id) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队的日志信息",
			})
			return
		}
		teamId = contextId
	} else {
		// 个人空间，teamId = 0
		teamId = 0
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	timeRange := c.DefaultQuery("time_range", "7d")

	// 计算时间范围
	now := time.Now()
	var startTimestamp int64
	switch timeRange {
	case "30d":
		startTimestamp = now.AddDate(0, 0, -30).Unix()
	case "90d":
		startTimestamp = now.AddDate(0, 0, -90).Unix()
	default: // 7d
		startTimestamp = now.AddDate(0, 0, -7).Unix()
	}

	// 构建查询参数
	var params model.LogsListParams
	params.UserId = id
	params.TeamId = teamId
	params.StartTimestamp = startTimestamp
	params.EndTimestamp = now.Unix()
	params.Page = page
	params.Size = pageSize

	// 获取日志数据
	logs, err := model.GetUserLogsListWithTeamFilter(id, teamId, &params)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法获取日志信息",
		})
		return
	}

	// 计算总数 - 使用单独的查询来获取总数，避免size限制
	totalParams := params
	totalParams.Page = 1
	totalParams.Size = 1 // 只需要获取总数，不需要实际数据
	totalLogs, err := model.GetUserLogsListWithTeamFilter(id, teamId, &totalParams)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("无法获取日志总数: %v (userId=%d, teamId=%d)", err, id, teamId),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": logs.Data,
			"total": totalLogs.TotalCount,
		},
	})
}

func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.AccessToken = utils.GetUUID()

	if model.DB.Where("access_token = ?", user.AccessToken).First(user).RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请重试，系统生成的 UUID 竟然重复了！",
		})
		return
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AccessToken,
	})
}

func GetAffCode(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if user.AffCode == "" {
		user.AffCode = utils.GetRandomString(4)
		if err := user.Update(false); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user.AffCode,
	})
}

func GetSelf(c *gin.Context) {
	id := c.GetInt("id")
	user, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 获取团队数量
	teamIds, err := model.GetUserTeamIds(id)
	if err == nil {
		user.TeamCount = len(teamIds)
	} else {
		user.TeamCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    user,
	})
}

// GetContextQuota 获取当前上下文的额度信息
func GetContextQuota(c *gin.Context) {
	userId := c.GetInt("id")
	contextType := c.GetString("context_type")
	contextId := c.GetInt("context_id")
	
	var quota, usedQuota int
	var unlimited bool
	var contextName string
	
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !model.IsTeamMember(contextId, userId) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队的额度信息",
			})
			return
		}
		
		team, err := model.GetTeamById(contextId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		
		contextName = team.Name
		
		if team.UnlimitedQuota {
			// 无限团队：基准额度 = owner.quota，used_quota = team.used_quota
			owner, err := model.GetUserById(team.OwnerId, false)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
			quota = owner.Quota
			usedQuota = team.UsedQuota
			unlimited = true // 实际不是无限，而是基于owner余额
		} else {
			// 有限团队：基准额度 = team.quota，used_quota = team.used_quota
			quota = team.Quota
			usedQuota = team.UsedQuota
		}
	} else {
		// 个人空间：quota = user.quota，used_quota = user.used_quota
		contextName = "个人空间"
		user, err := model.GetUserById(userId, false)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		quota = user.Quota
		usedQuota = user.UsedQuota
	}
	
	available := quota - usedQuota
	if available < 0 {
		available = 0
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"context_type": contextType,
			"context_id":   contextId,
			"context_name": contextName,
			"quota":        quota,
			"used_quota":   usedQuota,
			"available":    available,
			"unlimited":    unlimited,
		},
	})
}

// GetUserContexts 获取用户的所有空间列表（个人 + 团队）
func GetUserContexts(c *gin.Context) {
	userId := c.GetInt("id")
	
	contexts := []gin.H{
		{
			"type": "user",
			"id":   userId,
			"name": "个人空间",
		},
	}
	
	// 获取用户加入的团队列表
	teamIds, err := model.GetUserTeamIds(userId)
	if err == nil {
		for _, teamId := range teamIds {
			team, err := model.GetTeamById(teamId)
			if err == nil {
				contexts = append(contexts, gin.H{
					"type": "team",
					"id":   teamId,
					"name": team.Name,
				})
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    contexts,
	})
}

// GetUserDashboardSummary 获取用户 Dashboard 汇总数据
func GetUserDashboardSummary(c *gin.Context) {
	userId := c.GetInt("id")
	
	// 获取上下文信息
	contextType := c.GetString("context_type")
	contextId := c.GetInt("context_id")
	
	// 根据上下文类型确定 teamId
	var teamId int
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !model.IsTeamMember(contextId, userId) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队的统计信息",
			})
			return
		}
		teamId = contextId
	} else {
		// 个人空间，teamId = 0
		teamId = 0
	}

	now := time.Now()
	
	// 计算时间范围
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -6) // 最近7天
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// 格式化日期
	todayStr := today.Format("2006-01-02")
	weekStartStr := weekStart.Format("2006-01-02")
	monthStartStr := monthStart.Format("2006-01-02")
	
	// 获取今日统计
	todayStats, err := model.GetUserModelStatisticsByContext(userId, teamId, todayStr, todayStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取今日统计失败: " + err.Error(),
		})
		return
	}
	
	// 获取本周统计
	weekStats, err := model.GetUserModelStatisticsByContext(userId, teamId, weekStartStr, todayStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取本周统计失败: " + err.Error(),
		})
		return
	}
	
	// 获取本月统计
	monthStats, err := model.GetUserModelStatisticsByContext(userId, teamId, monthStartStr, todayStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取本月统计失败: " + err.Error(),
		})
		return
	}
	
	// 聚合统计数据
	aggregateStats := func(stats []*model.LogStatisticGroupModel) (int64, int64, int64) {
		var requests, quota, tokens int64
		for _, stat := range stats {
			requests += stat.RequestCount
			quota += stat.Quota
			tokens += stat.PromptTokens + stat.CompletionTokens
		}
		return requests, quota, tokens
	}
	
	todayRequests, todayQuota, todayTokens := aggregateStats(todayStats)
	weekRequests, weekQuota, weekTokens := aggregateStats(weekStats)
	monthRequests, monthQuota, monthTokens := aggregateStats(monthStats)
	
	// 计算热门模型 Top 5
	modelUsage := make(map[string]int64)
	for _, stat := range weekStats {
		modelUsage[stat.ModelName] += stat.RequestCount
	}
	
	// 排序并取前5
	type ModelCount struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
		Quota int64  `json:"quota"`
	}
	
	var topModels []ModelCount
	for modelName, count := range modelUsage {
		var quota int64
		for _, stat := range weekStats {
			if stat.ModelName == modelName {
				quota += stat.Quota
			}
		}
		topModels = append(topModels, ModelCount{
			Name:  modelName,
			Count: count,
			Quota: quota,
		})
	}
	
	// 按请求数排序
	for i := 0; i < len(topModels)-1; i++ {
		for j := i + 1; j < len(topModels); j++ {
			if topModels[i].Count < topModels[j].Count {
				topModels[i], topModels[j] = topModels[j], topModels[i]
			}
		}
	}
	
	// 取前5个
	if len(topModels) > 5 {
		topModels = topModels[:5]
	}
	
	// 生成最近7天趋势数据
	var dailyTrend []gin.H
	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		dayStats, err := model.GetUserModelStatisticsByContext(userId, teamId, dateStr, dateStr)
		if err != nil {
			continue
		}
		
		dayRequests, dayQuota, _ := aggregateStats(dayStats)
		dailyTrend = append(dailyTrend, gin.H{
			"date":     dateStr,
			"requests": dayRequests,
			"quota":    dayQuota,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"today": gin.H{
				"requests": todayRequests,
				"quota":    todayQuota,
				"tokens":   todayTokens,
			},
			"week": gin.H{
				"requests": weekRequests,
				"quota":    weekQuota,
				"tokens":   weekTokens,
			},
			"month": gin.H{
				"requests": monthRequests,
				"quota":    monthQuota,
				"tokens":   monthTokens,
			},
			"top_models":   topModels,
			"daily_trend":  dailyTrend,
		},
	})
}

// GetQuotaAlert 获取额度预警信息
func GetQuotaAlert(c *gin.Context) {
	userId := c.GetInt("id")
	contextType := c.GetString("context_type")
	contextId := c.GetInt("context_id")
	
	var quota, usedQuota int
	
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !model.IsTeamMember(contextId, userId) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权限访问该团队的额度信息",
			})
			return
		}
		
		team, err := model.GetTeamById(contextId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		
		if team.UnlimitedQuota {
			owner, err := model.GetUserById(team.OwnerId, false)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
			quota = owner.Quota
			usedQuota = team.UsedQuota
		} else {
			quota = team.Quota
			usedQuota = team.UsedQuota
		}
	} else {
		user, err := model.GetUserById(userId, false)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		quota = user.Quota
		usedQuota = user.UsedQuota
	}
	
	// 计算使用率
	usageRate := 0.0
	if quota > 0 {
		usageRate = float64(usedQuota) / float64(quota) * 100
	}
	
	// 计算最近7天平均消费
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -6)
	
	weekStartStr := weekStart.Format("2006-01-02")
	todayStr := today.Format("2006-01-02")
	
	var teamId int
	if contextType == "team" {
		teamId = contextId
	} else {
		teamId = 0
	}
	
	weekStats, err := model.GetUserModelStatisticsByContext(userId, teamId, weekStartStr, todayStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取消费统计失败: " + err.Error(),
		})
		return
	}
	
	var totalQuota int64
	for _, stat := range weekStats {
		totalQuota += stat.Quota
	}
	
	dailyAvg := float64(totalQuota) / 7.0
	remainingQuota := quota - usedQuota
	remainingDays := 0.0
	
	if dailyAvg > 0 && remainingQuota > 0 {
		remainingDays = float64(remainingQuota) / dailyAvg
	}
	
	alert := usageRate > 80.0
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"alert":          alert,
			"usage_rate":     usageRate,
			"remaining_days": remainingDays,
			"daily_avg":      dailyAvg,
			"remaining_quota": remainingQuota,
		},
	})
}

func UpdateUser(c *gin.Context) {
	var updatedUser model.User
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil || updatedUser.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if updatedUser.Password == "" {
		updatedUser.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&updatedUser); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(updatedUser.Id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role && myRole != config.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	if myRole <= updatedUser.Role && myRole != config.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权将其他用户权限等级提升到大于等于自己的权限等级",
		})
		return
	}
	if updatedUser.Password == "$I_LOVE_U" {
		updatedUser.Password = "" // rollback to what it should be
	}
	updatePassword := updatedUser.Password != ""
	if err := updatedUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if originUser.Quota != updatedUser.Quota {
		model.RecordLog(originUser.Id, model.LogTypeManage, fmt.Sprintf("管理员将用户额度从 %s修改为 %s", common.LogQuota(originUser.Quota), common.LogQuota(updatedUser.Quota)))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}

	cleanUser := model.User{
		Id: c.GetInt("id"),
		// Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	if err := cleanUser.Update(updatePassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	originUser, err := model.GetUserById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= originUser.Role {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权删除同权限等级或更高权限等级的用户",
		})
		return
	}
	err = model.DeleteUserById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
		return
	}
}

func CreateUser(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "输入不合法 " + err.Error(),
		})
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无法创建权限大于等于自己的用户",
		})
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if err := cleanUser.Insert(0); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

// ManageUser Only admin user can do this
func ManageUser(c *gin.Context) {
	var req ManageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	user := model.User{
		Username: req.Username,
	}
	// Fill attributes
	model.DB.Where(&user).First(&user)
	if user.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	myRole := c.GetInt("role")
	if myRole <= user.Role && myRole != config.RoleRootUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权更新同权限等级或更高权限等级的用户信息",
		})
		return
	}
	switch req.Action {
	case "disable":
		user.Status = config.UserStatusDisabled
		if user.Role == config.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法禁用超级管理员用户",
			})
			return
		}
	case "enable":
		user.Status = config.UserStatusEnabled
	case "delete":
		if user.Role == config.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法删除超级管理员用户",
			})
			return
		}
		if err := user.Delete(); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	case "promote":
		if myRole != config.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "普通管理员用户无法提升其他用户为管理员",
			})
			return
		}
		if user.Role >= config.RoleAdminUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是管理员",
			})
			return
		}
		user.Role = config.RoleAdminUser
	case "demote":
		if user.Role == config.RoleRootUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无法降级超级管理员用户",
			})
			return
		}
		if user.Role == config.RoleCommonUser {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "该用户已经是普通用户",
			})
			return
		}
		user.Role = config.RoleCommonUser
	}

	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    clearUser,
	})
}

func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码错误或已过期",
		})
		return
	}
	id := c.GetInt("id")
	user := model.User{
		Id: id,
	}
	err := user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.Email = email
	// no need to check if this email already taken, because we have used verification code to check it
	err = user.Update(false)
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
	})
}

type topUpRequest struct {
	Key string `json:"key"`
}

func TopUp(c *gin.Context) {
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	id := c.GetInt("id")
	quota, err := model.Redeem(req.Key, id, c.ClientIP())
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
		"data":    quota,
	})
}

type ChangeUserQuotaRequest struct {
	Quota  int    `json:"quota" form:"quota"`
	Remark string `json:"remark" form:"remark"`
}

func ChangeUserQuota(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	var req ChangeUserQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIRespondWithError(c, http.StatusOK, err)
		return
	}

	if req.Quota == 0 {
		common.APIRespondWithError(c, http.StatusOK, errors.New("不能为0"))
		return
	}

	err = model.ChangeUserQuota(userId, req.Quota, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	remark := fmt.Sprintf("管理员增减用户额度 %s", common.LogQuota(req.Quota))

	if req.Remark != "" {
		remark = fmt.Sprintf("%s, 备注: %s", remark, req.Remark)
	}

	model.RecordQuotaLog(userId, model.LogTypeManage, req.Quota, c.ClientIP(), remark)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
