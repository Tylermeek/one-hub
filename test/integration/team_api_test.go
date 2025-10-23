package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"one-api/controller"
	"one-api/middleware"
	"one-api/model"
	"one-api/test/testutils"
)

// TeamAPISuite Team API 集成测试套件
type TeamAPISuite struct {
	suite.Suite
	db      *gorm.DB
	router  *gin.Engine
	factory *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamAPISuite) SetupSuite() {
	suite.db = testutils.SetupTestDB(suite.T(), nil)
	suite.factory = testutils.NewTestDataFactory()
	
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)
	
	// 创建路由
	suite.router = gin.New()
	suite.router.Use(middleware.Recover())
	suite.router.Use(middleware.Logger())
	
	// 设置 API 路由
	apiRouter := suite.router.Group("/api")
	suite.setupAPIRoutes(apiRouter)
}

// TearDownSuite 测试套件清理
func (suite *TeamAPISuite) TearDownSuite() {
	testutils.CleanupTestDB(suite.T(), suite.db)
}

// SetupTest 每个测试前的设置
func (suite *TeamAPISuite) SetupTest() {
	testutils.ResetTestDB(suite.T(), suite.db)
}

// setupAPIRoutes 设置 API 路由
func (suite *TeamAPISuite) setupAPIRoutes(apiRouter *gin.RouterGroup) {
	// 团队相关路由
	teamRoute := apiRouter.Group("/team")
	teamRoute.Use(suite.mockAuthMiddleware()) // 使用模拟认证中间件
	
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
}

// mockAuthMiddleware 模拟认证中间件
func (suite *TeamAPISuite) mockAuthMiddleware() gin.HandlerFunc {
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
		c.Next()
	}
}

// TestCreateTeamAPI 测试创建团队 API
func (suite *TeamAPISuite) TestCreateTeamAPI() {
	// 创建测试用户
	userFactory := suite.factory.NewUserFactory()
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 准备请求数据
	teamData := map[string]interface{}{
		"name": "测试团队",
	}
	
	jsonData, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", "/api/team/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的数据
	testutils.AssertDatabaseRecordExists(suite.T(), suite.db, &model.Team{}, "name = ?", "测试团队")
	testutils.AssertDatabaseRecordExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = (SELECT id FROM teams WHERE name = ?) AND user_id = ?", "测试团队", owner.Id)
}

// TestCreateTeamAPIWithInvalidInput 测试创建团队 API 无效输入
func (suite *TeamAPISuite) TestCreateTeamAPIWithInvalidInput() {
	// 准备无效请求数据
	teamData := map[string]interface{}{
		"name": "", // 空名称
	}
	
	jsonData, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", "/api/team/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseError(suite.T(), w, http.StatusOK)
	testutils.AssertValidationError(suite.T(), w)
}

// TestGetUserTeamsAPI 测试获取用户团队列表 API
func (suite *TeamAPISuite) TestGetUserTeamsAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求
	req, _ := http.NewRequest("GET", "/api/team/list?page=1&size=10", nil)
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	testutils.AssertTeamInResponse(suite.T(), w, team)
}

// TestGetTeamDetailAPI 测试获取团队详情 API
func (suite *TeamAPISuite) TestGetTeamDetailAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求
	req, _ := http.NewRequest("GET", "/api/team/1", nil)
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	testutils.AssertTeamInResponse(suite.T(), w, team)
}

// TestGetTeamDetailAPIPermissionDenied 测试获取团队详情 API 权限拒绝
func (suite *TeamAPISuite) TestGetTeamDetailAPIPermissionDenied() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	nonMember := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(nonMember).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求（非团队成员访问）
	req, _ := http.NewRequest("GET", "/api/team/1", nil)
	req.Header.Set("X-User-ID", "2") // 非团队成员
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseError(suite.T(), w, http.StatusOK)
	testutils.AssertPermissionDenied(suite.T(), w)
}

// TestAllocateTeamQuotaAPI 测试分配团队额度 API
func (suite *TeamAPISuite) TestAllocateTeamQuotaAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求数据
	quotaData := map[string]interface{}{
		"quota":     500000,
		"unlimited": false,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("POST", "/api/team/1/allocate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的团队额度
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500000, updatedTeam.Quota)
}

// TestAllocateTeamQuotaAPIPermissionDenied 测试分配团队额度 API 权限拒绝
func (suite *TeamAPISuite) TestAllocateTeamQuotaAPIPermissionDenied() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 添加普通成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求数据（普通成员尝试分配额度）
	quotaData := map[string]interface{}{
		"quota":     50000,
		"unlimited": false,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("POST", "/api/team/1/allocate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "2") // 普通成员
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseError(suite.T(), w, http.StatusOK)
	testutils.AssertPermissionDenied(suite.T(), w)
}

// TestInviteMemberAPI 测试邀请成员 API
func (suite *TeamAPISuite) TestInviteMemberAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求数据
	inviteData := map[string]interface{}{
		"user_id": member.Id,
	}
	
	jsonData, _ := json.Marshal(inviteData)
	req, _ := http.NewRequest("POST", "/api/team/1/invite", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的团队成员
	testutils.AssertDatabaseRecordExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = ? AND user_id = ?", team.Id, member.Id)
}

// TestGetTeamMembersAPI 测试获取团队成员列表 API
func (suite *TeamAPISuite) TestGetTeamMembersAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 添加团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求
	req, _ := http.NewRequest("GET", "/api/team/1/members?page=1&size=10", nil)
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	testutils.AssertTeamMemberInResponse(suite.T(), w, teamMember)
}

// TestUpdateMemberQuotaAPI 测试更新成员额度限制 API
func (suite *TeamAPISuite) TestUpdateMemberQuotaAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 添加团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求数据
	quotaData := map[string]interface{}{
		"max_quota": 50000,
	}
	
	jsonData, _ := json.Marshal(quotaData)
	req, _ := http.NewRequest("PUT", "/api/team/1/member/2/quota", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的成员额度限制
	var updatedMember model.TeamMember
	err = suite.db.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&updatedMember).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50000, updatedMember.MaxQuota)
}

// TestRemoveMemberAPI 测试移除成员 API
func (suite *TeamAPISuite) TestRemoveMemberAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 添加团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求
	req, _ := http.NewRequest("DELETE", "/api/team/1/member/2", nil)
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的成员已被删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = ? AND user_id = ? AND deleted_at IS NULL", team.Id, member.Id)
}

// TestDeleteTeamAPI 测试删除团队 API
func (suite *TeamAPISuite) TestDeleteTeamAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求
	req, _ := http.NewRequest("DELETE", "/api/team/1", nil)
	req.Header.Set("X-User-ID", "1")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的团队已被删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.Team{}, "id = ?", team.Id)
}

// TestRegisterWithInviteAPI 测试通过邀请码注册 API
func (suite *TeamAPISuite) TestRegisterWithInviteAPI() {
	// 创建测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 准备请求数据
	registerData := map[string]interface{}{
		"username":    "newuser",
		"password":    "password123",
		"email":       "newuser@example.com",
		"invite_code": team.InviteCode,
	}
	
	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/team/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// 验证响应
	testutils.AssertAPIResponseSuccess(suite.T(), w)
	
	// 验证数据库中的新用户和团队成员
	testutils.AssertDatabaseRecordExists(suite.T(), suite.db, &model.User{}, "username = ?", "newuser")
	
	// 验证新用户被添加到团队
	var newUser model.User
	err = suite.db.Where("username = ?", "newuser").First(&newUser).Error
	require.NoError(suite.T(), err)
	testutils.AssertDatabaseRecordExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = ? AND user_id = ?", team.Id, newUser.Id)
}

// TestTeamAPISuite 运行测试套件
func TestTeamAPISuite(t *testing.T) {
	suite.Run(t, new(TeamAPISuite))
}
