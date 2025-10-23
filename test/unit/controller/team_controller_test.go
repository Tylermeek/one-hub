package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"one-api/controller"
	"one-api/model"
	"one-api/test/testutils"
	"one-api/test/testutils/mocks"
)

// TeamControllerSuite Team Controller 单元测试套件
type TeamControllerSuite struct {
	suite.Suite
	mockModel *mocks.MockModelInterface
	factory   *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamControllerSuite) SetupSuite() {
	suite.factory = testutils.NewTestDataFactory()
}

// SetupTest 每个测试前的设置
func (suite *TeamControllerSuite) SetupTest() {
	suite.mockModel = mocks.NewMockModelInterface()
}

// TearDownTest 每个测试后的清理
func (suite *TeamControllerSuite) TearDownTest() {
	suite.mockModel.AssertExpectations(suite.T())
}

// createTestGinContext 创建测试 Gin 上下文
func (suite *TeamControllerSuite) createTestGinContext(method, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		require.NoError(suite.T(), err)
	}
	
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	
	c.Request = req
	return c, recorder
}

// setUserContext 设置用户上下文
func (suite *TeamControllerSuite) setUserContext(c *gin.Context, userId int) {
	c.Set("id", userId)
}

// TestCreateTeamSuccess 测试创建团队成功
func (suite *TeamControllerSuite) TestCreateTeamSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望
	suite.mockModel.On("GetUserById", owner.Id, false).Return(owner, nil)
	suite.mockModel.On("CreateTeam", mock.AnythingOfType("*model.Team")).Return(nil)
	suite.mockModel.On("CreateTeamMember", mock.AnythingOfType("*model.TeamMember")).Return(nil)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("POST", "/api/team/", map[string]interface{}{
		"name": "测试团队",
	})
	suite.setUserContext(c, owner.Id)
	
	// 执行测试
	controller.CreateTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "成功")
}

// TestCreateTeamInvalidInput 测试创建团队无效输入
func (suite *TeamControllerSuite) TestCreateTeamInvalidInput() {
	// 创建测试上下文（空名称）
	c, recorder := suite.createTestGinContext("POST", "/api/team/", map[string]interface{}{
		"name": "",
	})
	suite.setUserContext(c, 1)
	
	// 执行测试
	controller.CreateTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "参数错误")
}

// TestCreateTeamModelError 测试创建团队模型错误
func (suite *TeamControllerSuite) TestCreateTeamModelError() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	owner := userFactory.CreateUserWithQuota(1000000)
	
	// 设置 Mock 期望（返回错误）
	suite.mockModel.On("GetUserById", owner.Id, false).Return(owner, nil)
	suite.mockModel.On("CreateTeam", mock.AnythingOfType("*model.Team")).Return(assert.AnError)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("POST", "/api/team/", map[string]interface{}{
		"name": "测试团队",
	})
	suite.setUserContext(c, owner.Id)
	
	// 执行测试
	controller.CreateTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "创建团队失败")
}

// TestGetUserTeamsSuccess 测试获取用户团队列表成功
func (suite *TeamControllerSuite) TestGetUserTeamsSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	teams := &model.DataResult[model.Team]{
		Data:  &[]*model.Team{team},
		Total: 1,
		Page:  1,
		Size:  10,
	}
	
	// 设置 Mock 期望
	suite.mockModel.On("GetUserTeams", owner.Id, mock.AnythingOfType("*model.PaginationParams")).Return(teams, nil)
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("GET", "/api/team/list?page=1&size=10", nil)
	suite.setUserContext(c, owner.Id)
	
	// 执行测试
	controller.GetUserTeams(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	
	data := response["data"].(map[string]interface{})
	teamsData := data["data"].([]interface{})
	assert.Len(suite.T(), teamsData, 1)
}

// TestGetUserTeamsModelError 测试获取用户团队列表模型错误
func (suite *TeamControllerSuite) TestGetUserTeamsModelError() {
	// 设置 Mock 期望（返回错误）
	suite.mockModel.On("GetUserTeams", 1, mock.AnythingOfType("*model.PaginationParams")).Return((*model.DataResult[model.Team])(nil), assert.AnError)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("GET", "/api/team/list?page=1&size=10", nil)
	suite.setUserContext(c, 1)
	
	// 执行测试
	controller.GetUserTeams(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "获取团队列表失败")
}

// TestGetTeamSuccess 测试获取团队详情成功
func (suite *TeamControllerSuite) TestGetTeamSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	suite.mockModel.On("GetTeamById", team.Id).Return(team, nil)
	suite.mockModel.On("GetUserById", owner.Id, false).Return(owner, nil)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("GET", "/api/team/1", nil)
	suite.setUserContext(c, owner.Id)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.GetTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	
	teamData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), team.Name, teamData["name"])
}

// TestGetTeamPermissionDenied 测试获取团队详情权限拒绝
func (suite *TeamControllerSuite) TestGetTeamPermissionDenied() {
	// 设置 Mock 期望（权限检查失败）
	suite.mockModel.On("IsTeamOwner", 1, 1).Return(false)
	suite.mockModel.On("IsTeamMember", 1, 1).Return(false)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("GET", "/api/team/1", nil)
	suite.setUserContext(c, 1)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.GetTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "无权限查看此团队")
}

// TestUpdateTeamSuccess 测试更新团队成功
func (suite *TeamControllerSuite) TestUpdateTeamSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	suite.mockModel.On("UpdateTeam", mock.AnythingOfType("*model.Team")).Return(nil)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("PUT", "/api/team/1", map[string]interface{}{
		"name":   "更新后的团队名称",
		"status": 1,
	})
	suite.setUserContext(c, owner.Id)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.UpdateTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "成功")
}

// TestUpdateTeamPermissionDenied 测试更新团队权限拒绝
func (suite *TeamControllerSuite) TestUpdateTeamPermissionDenied() {
	// 设置 Mock 期望（权限检查失败）
	suite.mockModel.On("IsTeamOwner", 1, 1).Return(false)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("PUT", "/api/team/1", map[string]interface{}{
		"name":   "更新后的团队名称",
		"status": 1,
	})
	suite.setUserContext(c, 1)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.UpdateTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "只有团队管理员可以修改团队信息")
}

// TestDeleteTeamSuccess 测试删除团队成功
func (suite *TeamControllerSuite) TestDeleteTeamSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	suite.mockModel.On("GetTeamById", team.Id).Return(team, nil)
	suite.mockModel.On("DeleteTeam", mock.AnythingOfType("*model.Team")).Return(nil)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("DELETE", "/api/team/1", nil)
	suite.setUserContext(c, owner.Id)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.DeleteTeam(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "成功")
}

// TestAllocateTeamQuotaSuccess 测试分配团队额度成功
func (suite *TeamControllerSuite) TestAllocateTeamQuotaSuccess() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	suite.mockModel.On("AllocateQuotaToTeam", owner.Id, team.Id, 500000, false).Return(nil)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("POST", "/api/team/1/allocate", map[string]interface{}{
		"quota":     500000,
		"unlimited": false,
	})
	suite.setUserContext(c, owner.Id)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.AllocateTeamQuota(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "成功")
}

// TestAllocateTeamQuotaInsufficientBalance 测试分配团队额度余额不足
func (suite *TeamControllerSuite) TestAllocateTeamQuotaInsufficientBalance() {
	// 准备测试数据
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	owner := userFactory.CreateUserWithQuota(100000)
	team := teamFactory.CreateTeam(owner.Id)
	
	// 设置 Mock 期望（返回余额不足错误）
	suite.mockModel.On("IsTeamOwner", team.Id, owner.Id).Return(true)
	suite.mockModel.On("AllocateQuotaToTeam", owner.Id, team.Id, 500000, false).Return(assert.AnError)
	
	// 创建测试上下文
	c, recorder := suite.createTestGinContext("POST", "/api/team/1/allocate", map[string]interface{}{
		"quota":     500000,
		"unlimited": false,
	})
	suite.setUserContext(c, owner.Id)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.AllocateTeamQuota(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "分配团队额度失败")
}

// TestAllocateTeamQuotaInvalidInput 测试分配团队额度无效输入
func (suite *TeamControllerSuite) TestAllocateTeamQuotaInvalidInput() {
	// 创建测试上下文（负数额度）
	c, recorder := suite.createTestGinContext("POST", "/api/team/1/allocate", map[string]interface{}{
		"quota":     -1000,
		"unlimited": false,
	})
	suite.setUserContext(c, 1)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	
	// 执行测试
	controller.AllocateTeamQuota(c)
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "额度必须大于0")
}

// TestTeamControllerSuite 运行测试套件
func TestTeamControllerSuite(t *testing.T) {
	suite.Run(t, new(TeamControllerSuite))
}
