package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"one-api/model"
)

// HTTPTestHelper HTTP 测试辅助函数
type HTTPTestHelper struct{}

// NewHTTPTestHelper 创建 HTTP 测试辅助函数
func NewHTTPTestHelper() *HTTPTestHelper {
	return &HTTPTestHelper{}
}

// CreateJSONRequest 创建 JSON 请求
func (h *HTTPTestHelper) CreateJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	var jsonBody []byte
	var err error
	
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// CreateFormRequest 创建表单请求
func (h *HTTPTestHelper) CreateFormRequest(method, url string, formData map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	
	// 添加查询参数
	if len(formData) > 0 {
		q := req.URL.Query()
		for key, value := range formData {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
	
	return req, nil
}

// AddAuthHeader 添加认证头
func (h *HTTPTestHelper) AddAuthHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

// AddUserContext 添加用户上下文到 Gin 请求
func (h *HTTPTestHelper) AddUserContext(c *gin.Context, userId int) {
	c.Set("id", userId)
}

// MockAuthMiddleware 模拟认证中间件
func (h *HTTPTestHelper) MockAuthMiddleware(userId int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("id", userId)
		c.Next()
	}
}

// MockPermissionMiddleware 模拟权限中间件
func (h *HTTPTestHelper) MockPermissionMiddleware(hasPermission bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasPermission {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "权限不足",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TestContextHelper 测试上下文辅助函数
type TestContextHelper struct{}

// NewTestContextHelper 创建测试上下文辅助函数
func NewTestContextHelper() *TestContextHelper {
	return &TestContextHelper{}
}

// CreateTestGinContext 创建测试 Gin 上下文
func (h *TestContextHelper) CreateTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	return c, recorder
}

// SetJSONRequest 设置 JSON 请求到 Gin 上下文
func (h *TestContextHelper) SetJSONRequest(c *gin.Context, method, url string, body interface{}) error {
	var jsonBody []byte
	var err error
	
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return nil
}

// SetFormRequest 设置表单请求到 Gin 上下文
func (h *TestContextHelper) SetFormRequest(c *gin.Context, method, url string, formData map[string]string) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	
	// 添加查询参数
	if len(formData) > 0 {
		q := req.URL.Query()
		for key, value := range formData {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}
	
	c.Request = req
	return nil
}

// SetUserContext 设置用户上下文
func (h *TestContextHelper) SetUserContext(c *gin.Context, userId int) {
	c.Set("id", userId)
}

// SetTeamContext 设置团队上下文
func (h *TestContextHelper) SetTeamContext(c *gin.Context, teamId int) {
	c.Set("team_id", teamId)
}

// MockSetupHelper Mock 设置辅助函数
type MockSetupHelper struct{}

// NewMockSetupHelper 创建 Mock 设置辅助函数
func NewMockSetupHelper() *MockSetupHelper {
	return &MockSetupHelper{}
}

// SetupMockTeamSuccess 设置团队操作成功的 Mock
func (h *MockSetupHelper) SetupMockTeamSuccess(mockDB interface{}, team *model.Team) {
	// 这里可以根据具体的 Mock 类型进行设置
	// 例如：mockDB.(*mocks.MockDBInterface).On("CreateTeam", team).Return(nil)
}

// SetupMockTeamFailure 设置团队操作失败的 Mock
func (h *MockSetupHelper) SetupMockTeamFailure(mockDB interface{}, team *model.Team, errMsg string) {
	// 这里可以根据具体的 Mock 类型进行设置
}

// SetupMockUserSuccess 设置用户操作成功的 Mock
func (h *MockSetupHelper) SetupMockUserSuccess(mockDB interface{}, user *model.User) {
	// 这里可以根据具体的 Mock 类型进行设置
}

// SetupMockPermissionSuccess 设置权限检查成功的 Mock
func (h *MockSetupHelper) SetupMockPermissionSuccess(mockDB interface{}) {
	// 这里可以根据具体的 Mock 类型进行设置
}

// SetupMockPermissionFailure 设置权限检查失败的 Mock
func (h *MockSetupHelper) SetupMockPermissionFailure(mockDB interface{}) {
	// 这里可以根据具体的 Mock 类型进行设置
}

// TestDataHelper 测试数据辅助函数
type TestDataHelper struct{}

// NewTestDataHelper 创建测试数据辅助函数
func NewTestDataHelper() *TestDataHelper {
	return &TestDataHelper{}
}

// CreateTestUser 创建测试用户
func (h *TestDataHelper) CreateTestUser(username string, quota int) *model.User {
	return &model.User{
		Username:    username,
		Password:    "test_password",
		DisplayName: username,
		Email:       username + "@example.com",
		Quota:       quota,
		Status:      1,
	}
}

// CreateTestTeam 创建测试团队
func (h *TestDataHelper) CreateTestTeam(name string, ownerId int, quota int) *model.Team {
	return &model.Team{
		Name:           name,
		OwnerId:        ownerId,
		Quota:          quota,
		UsedQuota:      0,
		UnlimitedQuota: false,
		Status:         1,
		InviteCode:     "TEST123",
		CreatedTime:    time.Now().Unix(),
		UpdatedTime:    time.Now().Unix(),
	}
}

// CreateTestTeamMember 创建测试团队成员
func (h *TestDataHelper) CreateTestTeamMember(teamId, userId, role int) *model.TeamMember {
	return &model.TeamMember{
		TeamId:     teamId,
		UserId:     userId,
		Role:       role,
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
}

// TestScenarioHelper 测试场景辅助函数
type TestScenarioHelper struct{}

// NewTestScenarioHelper 创建测试场景辅助函数
func NewTestScenarioHelper() *TestScenarioHelper {
	return &TestScenarioHelper{}
}

// SetupOwnerScenario 设置 Owner 场景
func (h *TestScenarioHelper) SetupOwnerScenario() (*model.User, *model.Team, *model.TeamMember) {
	owner := &model.User{
		Id:          1,
		Username:    "owner",
		DisplayName: "Owner",
		Email:       "owner@example.com",
		Quota:       1000000,
		Status:      1,
	}
	
	team := &model.Team{
		Id:            1,
		Name:          "测试团队",
		OwnerId:       owner.Id,
		Quota:         500000,
		UsedQuota:     0,
		UnlimitedQuota: false,
		Status:        1,
		InviteCode:    "TEST123",
		CreatedTime:   time.Now().Unix(),
		UpdatedTime:   time.Now().Unix(),
	}
	
	member := &model.TeamMember{
		Id:         1,
		TeamId:     team.Id,
		UserId:     owner.Id,
		Role:       1, // 管理员
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	
	return owner, team, member
}

// SetupMemberScenario 设置成员场景
func (h *TestScenarioHelper) SetupMemberScenario() (*model.User, *model.User, *model.Team, *model.TeamMember, *model.TeamMember) {
	owner := &model.User{
		Id:          1,
		Username:    "owner",
		DisplayName: "Owner",
		Email:       "owner@example.com",
		Quota:       1000000,
		Status:      1,
	}
	
	member := &model.User{
		Id:          2,
		Username:    "member",
		DisplayName: "Member",
		Email:       "member@example.com",
		Quota:       200000,
		Status:      1,
	}
	
	team := &model.Team{
		Id:            1,
		Name:          "测试团队",
		OwnerId:       owner.Id,
		Quota:         500000,
		UsedQuota:     0,
		UnlimitedQuota: false,
		Status:        1,
		InviteCode:    "TEST123",
		CreatedTime:   time.Now().Unix(),
		UpdatedTime:   time.Now().Unix(),
	}
	
	ownerMember := &model.TeamMember{
		Id:         1,
		TeamId:     team.Id,
		UserId:     owner.Id,
		Role:       1, // 管理员
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	
	memberRecord := &model.TeamMember{
		Id:         2,
		TeamId:     team.Id,
		UserId:     member.Id,
		Role:       2, // 普通成员
		MaxQuota:   0,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	
	return owner, member, team, ownerMember, memberRecord
}

// TestCleanupHelper 测试清理辅助函数
type TestCleanupHelper struct{}

// NewTestCleanupHelper 创建测试清理辅助函数
func NewTestCleanupHelper() *TestCleanupHelper {
	return &TestCleanupHelper{}
}

// CleanupTestData 清理测试数据
func (h *TestCleanupHelper) CleanupTestData(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}
	
	// 按依赖关系顺序删除数据
	tables := []interface{}{
		&model.TeamMember{},
		&model.Team{},
		&model.Log{},
		&model.User{},
	}
	
	for _, table := range tables {
		err := db.Unscoped().Where("1 = 1").Delete(table).Error
		if err != nil {
			t.Logf("Warning: Failed to cleanup table %T: %v", table, err)
		}
	}
}

// ResetMockExpectations 重置 Mock 期望
func (h *TestCleanupHelper) ResetMockExpectations(mocks ...interface{}) {
	for _, mock := range mocks {
		if mockObj, ok := mock.(mock.TestingT); ok {
			// 重置 Mock 对象的期望
			// 这里需要根据具体的 Mock 类型来实现
		}
	}
}
