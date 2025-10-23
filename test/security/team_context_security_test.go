package security

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// TestTeamContextSecurity 测试团队上下文安全
func TestTeamContextSecurity(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	
	// 创建测试数据库和用户
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	
	// 创建测试用户
	userA := createTestUser(t, db, "userA", "password", 1000000)
	userB := createTestUser(t, db, "userB", "password", 1000000)
	
	// 创建测试团队
	team := &model.Team{
		Name:        "测试团队",
		OwnerId:     userA.Id,
		Quota:       100000,
		UsedQuota:   0,
		Status:      1,
		InviteCode:  "TEST123",
		CreatedTime: 1234567890,
		UpdatedTime: 1234567890,
	}
	err := team.Insert()
	require.NoError(t, err)
	
	// 添加 userA 为团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     userA.Id,
		Role:       1, // 管理员
		Status:     1,
		JoinedTime: 1234567890,
	}
	err = db.Create(member).Error
	require.NoError(t, err)
	
	// 创建测试路由
	router := setupTestRouter(t)
	
	t.Run("测试越权访问团队空间", func(t *testing.T) {
		// userB 尝试切换到不属于自己的团队
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// 模拟 userB 的 session
		session := sessions.Default(gin.New())
		session.Set("id", userB.Id)
		session.Set("username", userB.Username)
		session.Set("role", userB.Role)
		session.Set("status", userB.Status)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 应该返回权限错误
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "无权限访问该团队空间")
	})
	
	t.Run("测试正常切换团队空间", func(t *testing.T) {
		// userA 切换到自己的团队
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// 模拟 userA 的 session
		session := sessions.Default(gin.New())
		session.Set("id", userA.Id)
		session.Set("username", userA.Username)
		session.Set("role", userA.Role)
		session.Set("status", userA.Status)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 应该成功
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Equal(t, "空间切换成功", response["message"].(string))
	})
	
	t.Run("测试Token空间绑定", func(t *testing.T) {
		// 创建团队 Token
		token := &model.Token{
			UserId:         userA.Id,
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
		err := token.Insert()
		require.NoError(t, err)
		
		// 使用团队 Token 访问 API
		req, _ := http.NewRequest("GET", "/api/token/", nil)
		req.Header.Set("Authorization", "Bearer "+token.Key)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 应该成功，且自动进入团队空间
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
	})
	
	t.Run("测试权限失效处理", func(t *testing.T) {
		// 移除 userA 的团队成员身份
		err := db.Where("team_id = ? AND user_id = ?", team.Id, userA.Id).Delete(&model.TeamMember{}).Error
		require.NoError(t, err)
		
		// userA 尝试切换到团队
		reqBody := map[string]interface{}{
			"type": "team",
			"id":   team.Id,
		}
		jsonData, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/context/switch", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		// 模拟 userA 的 session
		session := sessions.Default(gin.New())
		session.Set("id", userA.Id)
		session.Set("username", userA.Username)
		session.Set("role", userA.Role)
		session.Set("status", userA.Status)
		session.Save()
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 应该返回权限错误
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["message"].(string), "无权限访问该团队空间")
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

