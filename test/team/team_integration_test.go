package team

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"one-hub/controller"
	"one-hub/middleware"
	"one-hub/model"
	"one-hub/router"
)

// 测试服务器设置
func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Team{}, &model.TeamMember{}, &model.Log{})
	require.NoError(t, err)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	r := gin.New()
	r.Use(middleware.Recover())
	r.Use(middleware.Logger())

	// 设置API路由
	apiRouter := r.Group("/api")
	router.SetupAPIRoutes(apiRouter, db)

	return r, db
}

// 创建测试用户并获取token
func createTestUserAndToken(db *gorm.DB, t *testing.T, username, password string, quota int) (int, string) {
	user := &model.User{
		Username: username,
		Password: password,
		Quota:    quota,
		Status:   1,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// 生成简单的测试token（实际应用中应该使用JWT）
	token := fmt.Sprintf("test_token_%d_%d", user.Id, time.Now().Unix())
	return user.Id, token
}

// 模拟认证中间件
func mockAuthMiddleware(userId int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userId)
		c.Next()
	}
}

// TestCreateTeam 测试创建团队
func TestCreateTeam(t *testing.T) {
	server, db := setupTestServer(t)
	
	// 创建测试用户
	userId, token := createTestUserAndToken(db, t, "testuser", "password", 1000000)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(userId))
	
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
	err := db.Create(team).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(userId))
	
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
	err := db.Create(team).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(userId))
	
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
	err := db.Create(team).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(userId))
	
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
	
	// 验证用户额度减少
	var updatedUser model.User
	err = db.First(&updatedUser, userId).Error
	require.NoError(t, err)
	assert.Equal(t, 500000, updatedUser.Quota)
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
	err := db.Create(team).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(ownerId))
	
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
	err := db.Create(team).Error
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
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(ownerId))
	
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
	err := db.Create(team).Error
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
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(ownerId))
	
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
	err := db.Create(team).Error
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
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(ownerId))
	
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
	err := db.Create(team).Error
	require.NoError(t, err)
	
	// 设置认证中间件
	server.Use(mockAuthMiddleware(ownerId))
	
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
	ownerId, ownerToken := createTestUserAndToken(db, t, "owner", "password", 1000000)
	memberId, memberToken := createTestUserAndToken(db, t, "member", "password", 200000)
	otherId, otherToken := createTestUserAndToken(db, t, "other", "password", 200000)
	
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
	err := db.Create(team).Error
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
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 测试普通成员无法分配团队额度
	server.Use(mockAuthMiddleware(memberId))
	
	quotaData := map[string]interface{}{
		"quota":     50000,
		"unlimited": false,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/team/%d/allocate", team.Id), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+memberToken)
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	// 测试非团队成员无法查看团队
	server.Use(mockAuthMiddleware(otherId))
	
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team/%d", team.Id), nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

