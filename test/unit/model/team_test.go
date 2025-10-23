package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"one-api/model"
	"one-api/test/testutils"
)

// TeamModelSuite Team 模型测试套件
type TeamModelSuite struct {
	suite.Suite
	db     *gorm.DB
	factory *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamModelSuite) SetupSuite() {
	suite.db = testutils.SetupTestDB(suite.T(), nil)
	suite.factory = testutils.NewTestDataFactory()
}

// TearDownSuite 测试套件清理
func (suite *TeamModelSuite) TearDownSuite() {
	testutils.CleanupTestDB(suite.T(), suite.db)
}

// SetupTest 每个测试前的设置
func (suite *TeamModelSuite) SetupTest() {
	testutils.ResetTestDB(suite.T(), suite.db)
}

// TestInsert 测试创建团队
func (suite *TeamModelSuite) TestInsert() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeam(owner.Id)
	
	// 测试插入
	err = team.Insert()
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), team.Id)
	assert.NotEmpty(suite.T(), team.InviteCode)
	assert.Greater(suite.T(), team.CreatedTime, int64(0))
	assert.Equal(suite.T(), team.CreatedTime, team.UpdatedTime)
	
	// 验证数据库中的数据
	var dbTeam model.Team
	err = suite.db.First(&dbTeam, team.Id).Error
	require.NoError(suite.T(), err)
	testutils.AssertTeamEqual(suite.T(), team, &dbTeam)
}

// TestInsertWithEmptyName 测试创建团队时名称为空
func (suite *TeamModelSuite) TestInsertWithEmptyName() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队（名称为空）
	team := teamFactory.CreateTeam(owner.Id)
	team.Name = ""
	
	// 测试插入应该失败
	err = team.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队名称不能为空")
}

// TestInsertWithZeroOwnerId 测试创建团队时 OwnerId 为 0
func (suite *TeamModelSuite) TestInsertWithZeroOwnerId() {
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试团队（OwnerId 为 0）
	team := teamFactory.CreateTeam(0)
	
	// 测试插入应该失败
	err := team.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队管理员不能为空")
}

// TestInsertWithDuplicateInviteCode 测试邀请码重复处理
func (suite *TeamModelSuite) TestInsertWithDuplicateInviteCode() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner1 := userFactory.CreateUserWithQuota(1000000)
	owner2 := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner1).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(owner2).Error
	require.NoError(suite.T(), err)
	
	// 创建第一个团队
	team1 := teamFactory.CreateTeam(owner1.Id)
	err = team1.Insert()
	require.NoError(suite.T(), err)
	
	// 创建第二个团队，手动设置相同的邀请码
	team2 := teamFactory.CreateTeam(owner2.Id)
	team2.InviteCode = team1.InviteCode
	
	// 测试插入应该成功（系统会自动生成新的邀请码）
	err = team2.Insert()
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), team1.InviteCode, team2.InviteCode)
}

// TestUpdate 测试更新团队
func (suite *TeamModelSuite) TestUpdate() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 更新团队信息
	originalUpdatedTime := team.UpdatedTime
	team.Name = "更新后的团队名称"
	team.Status = 2 // 禁用
	
	// 测试更新
	err = team.Update()
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), team.UpdatedTime, originalUpdatedTime)
	
	// 验证数据库中的数据
	var dbTeam model.Team
	err = suite.db.First(&dbTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "更新后的团队名称", dbTeam.Name)
	assert.Equal(suite.T(), 2, dbTeam.Status)
}

// TestUpdateWithEmptyName 测试更新团队时名称为空
func (suite *TeamModelSuite) TestUpdateWithEmptyName() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 更新团队名称为空
	team.Name = ""
	
	// 测试更新应该失败
	err = team.Update()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队名称不能为空")
}

// TestUpdateWithZeroId 测试更新团队时 ID 为 0
func (suite *TeamModelSuite) TestUpdateWithZeroId() {
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试团队（ID 为 0）
	team := teamFactory.CreateTeam(1)
	team.Id = 0
	
	// 测试更新应该失败
	err := team.Update()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team id 为空")
}

// TestDelete 测试删除团队
func (suite *TeamModelSuite) TestDelete() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = suite.db.Create(teamMember).Error
	require.NoError(suite.T(), err)
	
	// 记录删除前的用户额度
	var originalOwner model.User
	err = suite.db.First(&originalOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	originalQuota := originalOwner.Quota
	
	// 测试删除
	err = team.Delete()
	assert.NoError(suite.T(), err)
	
	// 验证团队已被删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.Team{}, "id = ?", team.Id)
	
	// 验证团队成员已被删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = ?", team.Id)
	
	// 验证团队额度已退还给管理员
	var updatedOwner model.User
	err = suite.db.First(&updatedOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	testutils.AssertQuotaAdded(suite.T(), originalQuota, updatedOwner.Quota, team.Quota)
}

// TestDeleteWithZeroId 测试删除团队时 ID 为 0
func (suite *TeamModelSuite) TestDeleteWithZeroId() {
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试团队（ID 为 0）
	team := teamFactory.CreateTeam(1)
	team.Id = 0
	
	// 测试删除应该失败
	err := team.Delete()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team id 为空")
}

// TestGetTeamById 测试根据 ID 获取团队
func (suite *TeamModelSuite) TestGetTeamById() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取团队
	retrievedTeam, err := model.GetTeamById(team.Id)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedTeam)
	testutils.AssertTeamEqual(suite.T(), team, retrievedTeam)
	
	// 验证预加载的 Owner 信息
	assert.NotNil(suite.T(), retrievedTeam.Owner)
	assert.Equal(suite.T(), owner.Username, retrievedTeam.Owner.Username)
}

// TestGetTeamByIdWithZeroId 测试根据 ID 获取团队时 ID 为 0
func (suite *TeamModelSuite) TestGetTeamByIdWithZeroId() {
	// 测试获取团队应该失败
	team, err := model.GetTeamById(0)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), team)
	assert.Contains(suite.T(), err.Error(), "team id 为空")
}

// TestGetTeamByIdNotFound 测试获取不存在的团队
func (suite *TeamModelSuite) TestGetTeamByIdNotFound() {
	// 测试获取不存在的团队
	team, err := model.GetTeamById(99999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), team)
}

// TestGetUserTeams 测试获取用户团队列表
func (suite *TeamModelSuite) TestGetUserTeams() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team1 := teamFactory.CreateTeam(owner.Id)
	err = team1.Insert()
	require.NoError(suite.T(), err)
	
	team2 := teamFactory.CreateTeam(owner.Id)
	err = team2.Insert()
	require.NoError(suite.T(), err)
	
	// 将 member 添加到 team1
	teamMember := memberFactory.CreateTeamMember(team1.Id, member.Id)
	err = suite.db.Create(teamMember).Error
	require.NoError(suite.T(), err)
	
	// 测试获取 owner 的团队列表
	params := &model.PaginationParams{Page: 1, Size: 10}
	teams, err := model.GetUserTeams(owner.Id, params)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), teams)
	require.NotNil(suite.T(), teams.Data)
	assert.Len(suite.T(), *teams.Data, 2)
	
	// 测试获取 member 的团队列表
	teams, err = model.GetUserTeams(member.Id, params)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), teams)
	require.NotNil(suite.T(), teams.Data)
	assert.Len(suite.T(), *teams.Data, 1)
	assert.Equal(suite.T(), team1.Id, (*teams.Data)[0].Id)
}

// TestGetTeamByInviteCode 测试根据邀请码获取团队
func (suite *TeamModelSuite) TestGetTeamByInviteCode() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取团队
	retrievedTeam, err := model.GetTeamByInviteCode(team.InviteCode)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedTeam)
	testutils.AssertTeamEqual(suite.T(), team, retrievedTeam)
}

// TestGetTeamByInviteCodeWithEmptyCode 测试根据邀请码获取团队时邀请码为空
func (suite *TeamModelSuite) TestGetTeamByInviteCodeWithEmptyCode() {
	// 测试获取团队应该失败
	team, err := model.GetTeamByInviteCode("")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), team)
	assert.Contains(suite.T(), err.Error(), "邀请码为空")
}

// TestGetTeamByInviteCodeNotFound 测试获取不存在的邀请码团队
func (suite *TeamModelSuite) TestGetTeamByInviteCodeNotFound() {
	// 测试获取不存在的团队
	team, err := model.GetTeamByInviteCode("INVALID")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), team)
}

// TestIsTeamOwner 测试检查是否为团队所有者
func (suite *TeamModelSuite) TestIsTeamOwner() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	nonOwner := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(nonOwner).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试检查所有者
	assert.True(suite.T(), model.IsTeamOwner(team.Id, owner.Id))
	assert.False(suite.T(), model.IsTeamOwner(team.Id, nonOwner.Id))
	assert.False(suite.T(), model.IsTeamOwner(99999, owner.Id))
}

// TestIsTeamMember 测试检查是否为团队成员
func (suite *TeamModelSuite) TestIsTeamMember() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	nonMember := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(nonMember).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = suite.db.Create(teamMember).Error
	require.NoError(suite.T(), err)
	
	// 测试检查成员
	assert.True(suite.T(), model.IsTeamMember(team.Id, member.Id))
	assert.False(suite.T(), model.IsTeamMember(team.Id, nonMember.Id))
	assert.False(suite.T(), model.IsTeamMember(99999, member.Id))
}

// TestTeamModelSuite 运行测试套件
func TestTeamModelSuite(t *testing.T) {
	suite.Run(t, new(TeamModelSuite))
}
