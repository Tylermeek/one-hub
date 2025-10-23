package team

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"one-api/controller"
	"one-api/middleware"
	"one-api/model"
	"one-api/test/testutils"
)

// 测试服务器设置
func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	// 设置测试数据库
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	// 设置全局数据库实例
	model.SetDB(db)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	r := gin.New()
	r.Use(middleware.RelayPanicRecover())
	middleware.SetUpLogger(r)

	// 设置 session 中间件（用于测试）
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("test-session", store))

	// 设置API路由
	apiRouter := r.Group("/api")
	setupTestAPIRoutes(apiRouter)

	return r, db
}

// setupTestAPIRoutes 设置测试API路由（不使用session）
func setupTestAPIRoutes(apiRouter *gin.RouterGroup) {
	// 团队相关路由
	teamRoute := apiRouter.Group("/team")
	teamRoute.Use(mockAuthMiddleware()) // 使用模拟认证中间件
	
	{
		teamRoute.POST("/", controller.CreateTeam)
		teamRoute.GET("/list", controller.GetUserTeams)
		teamRoute.GET("/:id", controller.GetTeam)
		teamRoute.PUT("/:id", controller.UpdateTeam)
		teamRoute.DELETE("/:id", controller.DeleteTeam)
		teamRoute.POST("/:id/allocate", controller.AllocateTeamQuota)
		
		// 团队成员相关路由
		teamRoute.GET("/:id/members", controller.GetTeamMembers)
		teamRoute.GET("/search_users", controller.SearchUsers)
		teamRoute.POST("/:id/invite", controller.InviteMember)
		teamRoute.DELETE("/:id/member/:userId", controller.RemoveMember)
		teamRoute.PUT("/:id/member/:userId/quota", controller.UpdateMemberQuota)
		teamRoute.GET("/:id/member/:userId/usage", controller.GetMemberUsage)
	}
	
	// 团队注册路由
	apiRouter.POST("/team/register", controller.RegisterWithInvite)
	
	// 上下文管理路由
	contextRoute := apiRouter.Group("/context")
	contextRoute.Use(mockAuthMiddleware())
	{
		contextRoute.POST("/switch", controller.SwitchContext)
		contextRoute.GET("/current", controller.GetCurrentContext)
	}
}

// 创建测试用户并获取token
func createTestUserAndToken(db *gorm.DB, t *testing.T, username, password string, quota int) (int, string) {
	user := &model.User{
		Username:    username,
		Password:    password,
		Quota:       quota,
		Status:      1,
		AccessToken: fmt.Sprintf("test_token_%s_%d", username, time.Now().UnixNano()),
		AffCode:     fmt.Sprintf("test_aff_%s_%d", username, time.Now().UnixNano()),
	}
	err := user.Insert(0)
	require.NoError(t, err)

	// 生成简单的测试token（实际应用中应该使用JWT）
	token := fmt.Sprintf("test_token_%d_%d", user.Id, time.Now().Unix())
	return user.Id, token
}

// 模拟认证中间件
func mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取用户 ID（测试用）
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "1" // 默认用户 ID
		}
		
		// 解析用户 ID
		var userId int
		if userID == "1" {
			userId = 1
		} else if userID == "2" {
			userId = 2
		} else {
			userId = 1 // 默认
		}
		
		c.Set("id", userId)
		c.Set("username", "testuser")
		c.Set("role", 1)
		c.Set("status", 1)
		c.Next()
	}
}

// TestCreateTeam 测试创建团队
func TestCreateTeam(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, token := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	// 测试数据
	teamData := map[string]interface{}{
		"name": "测试团队",
	}
	
	jsonData, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", "/api/team/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	assert.Contains(t, response["message"], "成功")
	
	// 验证数据库中的数据
	var team model.Team
	err = db.Where("name = ?", "测试团队").First(&team).Error
	require.NoError(t, err)
	assert.Equal(t, "测试团队", team.Name)
	assert.Equal(t, userId, team.OwnerId)
}

// TestGetUserTeams 测试获取用户团队列表
func TestGetUserTeams(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, token := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	req, _ := http.NewRequest("GET", "/api/team/list?page=1&size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	teams := data["data"].([]interface{})
	assert.Len(t, teams, 1)
	
	teamData := teams[0].(map[string]interface{})
	assert.Equal(t, "测试团队", teamData["name"])
}

// TestGetTeamDetail 测试获取团队详情
func TestGetTeamDetail(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, token := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/team/%d", team.Id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	teamData := response["data"].(map[string]interface{})
	assert.Equal(t, "测试团队", teamData["name"])
	assert.Equal(t, float64(100000), teamData["quota"])
}

// TestAllocateTeamQuota 测试分配团队额度
func TestAllocateTeamQuota(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, token := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userId,
		Quota:       0,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	// 测试分配有限额度
	quotaData := map[string]interface{}{
		"quota":     500000,
		"unlimited": false,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/team/%d/allocate", team.Id), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	// 验证数据库中的团队额度
	var updatedTeam model.Team
	err = db.First(&updatedTeam, team.Id).Error
	require.NoError(t, err)
	assert.Equal(t, 500000, updatedTeam.Quota)
	
	// 验证用户额度不变（新逻辑：不扣除管理员个人额度）
	var updatedUser model.User
	err = db.First(&updatedUser, userId).Error
	require.NoError(t, err)
	assert.Equal(t, 1000000, updatedUser.Quota)
}

// TestInviteMember 测试邀请成员
func TestInviteMember(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, _ := createTestUserAndToken(db, t, "member", "password", 200000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	// 测试邀请成员
	inviteData := map[string]interface{}{
		"user_id": memberId,
	}
	
	jsonData, _ := json.Marshal(inviteData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/team/%d/invite", team.Id), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	// 验证数据库中的团队成员
	var member model.TeamMember
	err = db.Where("team_id = ? AND user_id = ?", team.Id, memberId).First(&member).Error
	require.NoError(t, err)
	assert.Equal(t, team.Id, member.TeamId)
	assert.Equal(t, memberId, member.UserId)
	assert.Equal(t, 2, member.Role) // 普通成员
}

// TestGetTeamMembers 测试获取团队成员列表
func TestGetTeamMembers(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, _ := createTestUserAndToken(db, t, "member", "password", 200000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 创建团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     memberId,
		Role:       2,
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/team/%d/members?page=1&size=10", team.Id), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	members := data["data"].([]interface{})
	assert.Len(t, members, 1)
	
	memberData := members[0].(map[string]interface{})
	assert.Equal(t, float64(memberId), memberData["user_id"])
	assert.Equal(t, float64(2), memberData["role"])
}

// TestUpdateMemberQuota 测试更新成员额度限制
func TestUpdateMemberQuota(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, _ := createTestUserAndToken(db, t, "member", "password", 200000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 创建团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     memberId,
		Role:       2,
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	// 测试更新成员额度限制
	quotaData := map[string]interface{}{
		"max_quota": 50000,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/team/%d/member/%d/quota", team.Id, memberId), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	// 验证数据库中的成员额度限制
	var updatedMember model.TeamMember
	err = db.Where("team_id = ? AND user_id = ?", team.Id, memberId).First(&updatedMember).Error
	require.NoError(t, err)
	assert.Equal(t, 50000, updatedMember.MaxQuota)
}

// TestRemoveMember 测试移除成员
func TestRemoveMember(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, _ := createTestUserAndToken(db, t, "member", "password", 200000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 创建团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     memberId,
		Role:       2,
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/team/%d/member/%d", team.Id, memberId), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	// 验证数据库中的成员已被删除
	var deletedMember model.TeamMember
	err = db.Where("team_id = ? AND user_id = ?", team.Id, memberId).First(&deletedMember).Error
	assert.Error(t, err) // 应该找不到记录
}

// TestDeleteTeam 测试删除团队
func TestDeleteTeam(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/team/%d", team.Id), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.True(t, response["success"].(bool))
	
	// 验证数据库中的团队已被删除
	var deletedTeam model.Team
	err = db.First(&deletedTeam, team.Id).Error
	assert.Error(t, err) // 应该找不到记录
}

// TestPermissionControl 测试权限控制
func TestPermissionControl(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	ownerId, _ := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, _ := createTestUserAndToken(db, t, "member", "password", 200000)
	_, _ = createTestUserAndToken(db, t, "other", "password", 200000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     ownerId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 创建团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     memberId,
		Role:       2,
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 测试普通成员无法分配团队额度
	server.Use(mockAuthMiddleware())
	
	quotaData := map[string]interface{}{
		"quota":     50000,
		"unlimited": false,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/team/%d/allocate", team.Id), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)  // 修改为 StatusOK，因为现在返回 JSON 格式的错误
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "只有团队管理员可以分配团队额度")
	
	// 测试非团队成员无法查看团队
	server.Use(mockAuthMiddleware())
	
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team/%d", team.Id), nil)
	
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)  // 修改为 StatusOK，因为现在返回 JSON 格式的错误
	
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "无权限查看此团队")
}

// TestContextSwitching 测试空间切换功能
func TestContextSwitching(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, _ := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 添加用户为团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     userId,
		Role:       1, // 管理员
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	t.Run("测试切换到团队空间", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "空间切换成功", response["message"].(string))
	})
	
	t.Run("测试切换到用户空间", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"type": "user",
			"id":   userId,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "空间切换成功", response["message"].(string))
	})
	
	t.Run("测试越权访问团队空间", func(t *testing.T) {
		// 创建另一个用户
		_, _ = createTestUserAndToken(db, t, "otheruser", "password", 100000)
		
		// 使用另一个用户的身份尝试切换
		server.Use(mockAuthMiddleware())
		
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "无权限访问该团队空间")
	})
}

// TestTokenContextBinding 测试Token空间绑定
func TestTokenContextBinding(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, _ := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userId,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: time.Now().Unix(),
		UpdatedTime: time.Now().Unix(),
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 添加用户为团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     userId,
		Role:       1, // 管理员
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert()
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware())
	
	t.Run("测试在团队空间创建Token", func(t *testing.T) {
		// 先切换到团队空间
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// 在团队空间创建Token
		tokenData := map[string]interface{}{
			"name":            "团队Token",
			"expired_time":    -1,
			"remain_quota":    10000,
			"unlimited_quota": false,
		}
		jsonData, _ = json.Marshal(tokenData)
		
		req, _ = http.NewRequest("POST", "/api/token/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		
		// 验证Token已绑定到团队空间
		var token model.Token
		err = db.Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
			userId, "team", team.Id).First(&token).Error
		require.NoError(t, err)
		assert.Equal(t, "team", token.OwnerType)
		assert.Equal(t, team.Id, token.OwnerId)
	})
	
	t.Run("测试在个人空间创建Token", func(t *testing.T) {
		// 切换到个人空间
		reqBody := map[string]interface{}{
			"type": "user",
			"id":   userId,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		// 在个人空间创建Token
		tokenData := map[string]interface{}{
			"name":            "个人Token",
			"expired_time":    -1,
			"remain_quota":    5000,
			"unlimited_quota": false,
		}
		jsonData, _ = json.Marshal(tokenData)
		
		req, _ = http.NewRequest("POST", "/api/token/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		
		// 验证Token已绑定到个人空间
		var token model.Token
		err = db.Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
			userId, "user", userId).First(&token).Error
		require.NoError(t, err)
		assert.Equal(t, "user", token.OwnerType)
		assert.Equal(t, userId, token.OwnerId)
	})
}

