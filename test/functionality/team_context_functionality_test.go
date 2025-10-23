package functionality

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"one-api/controller"
	"one-api/middleware"
	"one-api/model"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"one-api/test/testutils"
)

// TestTeamContextFunctionality 测试团队上下文功能
func TestTeamContextFunctionality(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	
	// 创建测试数据库和用户
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// 创建测试用户
	user := createTestUser(t, db, "testuser", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     user.Id,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: 1234567890,
		UpdatedTime: 1234567890,
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 添加用户为团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     user.Id,
		Role:       1, // 管理员
		Status:     1,
		JoinedTime: 1234567890,
	}
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 创建测试路由
	router := setupTestRouter(t)
	
	t.Run("测试空间切换流程", func(t *testing.T) {
		// 1. 登录后默认进入个人空间
		req, _ := http.NewRequest("GET", "/api/context/current", nil)
		
		// 模拟用户 session
		session := sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "user", data["type"])
		assert.Equal(t, float64(user.Id), data["id"])
		
		// 2. 切换到团队空间
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ = http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// 模拟用户 session
		session = sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Save()
		
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "空间切换成功", response["message"].(string))
		
		// 3. 验证当前空间已切换
		req, _ = http.NewRequest("GET", "/api/context/current", nil)
		
		// 模拟用户 session（包含切换后的上下文）
		session = sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("context_type", "team")
		session.Set("context_id", team.Id)
		session.Save()
		
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data = response["data"].(map[string]interface{})
		assert.Equal(t, "team", data["type"])
		assert.Equal(t, float64(team.Id), data["id"])
	})
	
	t.Run("测试Token创建与使用", func(t *testing.T) {
		// 1. 在团队空间创建 Token
		tokenData := map[string]interface{}{
			"name":            "团队Token",
			"expired_time":    -1,
			"remain_quota":    10000,
			"unlimited_quota": false,
		}
		jsonData, _ := json.Marshal(tokenData)
		
		req, _ := http.NewRequest("POST", "/api/token/", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// 模拟用户在团队空间的 session
		session := sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("context_type", "team")
		session.Set("context_id", team.Id)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		
		// 2. 验证 Token 已绑定到团队空间
		var token model.Token
		err = db.Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
			user.Id, "team", team.Id).First(&token).Error
		require.NoError(t, err)
		assert.Equal(t, "team", token.OwnerType)
		assert.Equal(t, team.Id, token.OwnerId)
	})
	
	t.Run("测试数据隔离", func(t *testing.T) {
		// 创建个人 Token
		personalToken := &model.Token{
			UserId:         user.Id,
			OwnerType:      "user",
			OwnerId:        user.Id,
			Name:           "个人Token",
			Status:         1,
			CreatedTime:    1234567890,
			AccessedTime:   1234567890,
			ExpiredTime:    -1,
			RemainQuota:    5000,
			UnlimitedQuota: false,
			UsedQuota:      0,
		}
		err := personalToken.Insert()
		require.NoError(t, err)
		
		// 创建团队 Token
		teamToken := &model.Token{
			UserId:         user.Id,
			OwnerType:      "team",
			OwnerId:        team.Id,
			Name:           "团队Token",
			Status:         1,
			CreatedTime:    1234567890,
			AccessedTime:   1234567890,
			ExpiredTime:    -1,
			RemainQuota:    10000,
			UnlimitedQuota: false,
			UsedQuota:      0,
		}
		err = db.Create(teamToken).Error
		require.NoError(t, err)
		
		// 1. 在个人空间查看 Token 列表
		req, _ := http.NewRequest("GET", "/api/token/", nil)
		
		// 模拟用户在个人空间的 session
		session := sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("context_type", "user")
		session.Set("context_id", user.Id)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		tokens := data["data"].([]interface{})
		
		// 应该只看到个人 Token
		assert.Len(t, tokens, 1)
		tokenData := tokens[0].(map[string]interface{})
		assert.Equal(t, "个人Token", tokenData["name"])
		
		// 2. 在团队空间查看 Token 列表
		req, _ = http.NewRequest("GET", "/api/token/", nil)
		
		// 模拟用户在团队空间的 session
		session = sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("context_type", "team")
		session.Set("context_id", team.Id)
		session.Save()
		
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data = response["data"].(map[string]interface{})
		tokens = data["data"].([]interface{})
		
		// 应该只看到团队 Token
		assert.Len(t, tokens, 1)
		tokenData = tokens[0].(map[string]interface{})
		assert.Equal(t, "团队Token", tokenData["name"])
	})
	
	t.Run("测试权限失效处理", func(t *testing.T) {
		// 移除用户团队成员身份
		err := db.Where("team_id = ? AND user_id = ?", team.Id, user.Id).Delete(&model.TeamMember{}).Error
		require.NoError(t, err)
		
		// 尝试访问团队空间
		req, _ := http.NewRequest("GET", "/api/context/current", nil)
		
		// 模拟用户在团队空间的 session（但权限已失效）
		session := sessions.Default(gin.New())
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("context_type", "team")
		session.Set("context_id", team.Id)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		data := response["data"].(map[string]interface{})
		
		// 应该自动重置为用户空间
		assert.Equal(t, "user", data["type"])
		assert.Equal(t, float64(user.Id), data["id"])
	})
}

// setupTestRouter 设置测试路由
func setupTestRouter(t *testing.T) *gin.Engine {
	router := gin.New()
	
	// 设置 session
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))
	
	// 添加认证中间件
	router.Use(middleware.UserAuth())
	router.Use(middleware.ContextMiddleware())
	
	// 注册路由
	apiRouter := router.Group("/api")
	{
		contextRoute := apiRouter.Group("/context")
		{
			contextRoute.POST("/switch", controller.SwitchContext)
			contextRoute.GET("/current", controller.GetCurrentContext)
		}
		
		tokenRoute := apiRouter.Group("/token")
		{
			tokenRoute.GET("/", controller.GetUserTokensList)
			tokenRoute.POST("/", controller.AddToken)
		}
	}
	
	return router
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	// 设置全局数据库实例
	model.SetDB(db)
	return db
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	testutils.CleanupTestDB(t, db)
}

// createTestUser 创建测试用户
func createTestUser(t *testing.T, db *gorm.DB, username, password string, quota int) *model.User {
	user := &model.User{
		Username:    username,
		Password:    password,
		DisplayName: username,
		Quota:       quota,
		Role:        1,
		Status:      1,
	}
	err := user.Insert(0)
	require.NoError(t, err)
	return user
}

