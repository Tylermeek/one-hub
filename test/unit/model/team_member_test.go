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

// TeamMemberModelSuite TeamMember 模型测试套件
type TeamMemberModelSuite struct {
	suite.Suite
	db      *gorm.DB
	factory *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamMemberModelSuite) SetupSuite() {
	suite.db = testutils.SetupTestDB(suite.T(), nil)
	suite.factory = testutils.NewTestDataFactory()
}

// TearDownSuite 测试套件清理
func (suite *TeamMemberModelSuite) TearDownSuite() {
	testutils.CleanupTestDB(suite.T(), suite.db)
}

// SetupTest 每个测试前的设置
func (suite *TeamMemberModelSuite) SetupTest() {
	testutils.ResetTestDB(suite.T(), suite.db)
}

// TestInsert 测试创建团队成员
func (suite *TeamMemberModelSuite) TestInsert() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	
	// 测试插入
	err = teamMember.Insert()
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), teamMember.Id)
	assert.Greater(suite.T(), teamMember.JoinedTime, int64(0))
	
	// 验证数据库中的数据
	var dbMember model.TeamMember
	err = suite.db.First(&dbMember, teamMember.Id).Error
	require.NoError(suite.T(), err)
	testutils.AssertTeamMemberEqual(suite.T(), teamMember, &dbMember)
}

// TestInsertWithZeroTeamId 测试创建团队成员时 TeamId 为 0
func (suite *TeamMemberModelSuite) TestInsertWithZeroTeamId() {
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试团队成员（TeamId 为 0）
	teamMember := memberFactory.CreateTeamMember(0, 1)
	
	// 测试插入应该失败
	err := teamMember.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team_id 为空")
}

// TestInsertWithZeroUserId 测试创建团队成员时 UserId 为 0
func (suite *TeamMemberModelSuite) TestInsertWithZeroUserId() {
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试团队成员（UserId 为 0）
	teamMember := memberFactory.CreateTeamMember(1, 0)
	
	// 测试插入应该失败
	err := teamMember.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user_id 为空")
}

// TestInsertWithDuplicateMember 测试添加重复成员
func (suite *TeamMemberModelSuite) TestInsertWithDuplicateMember() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建第一个团队成员
	teamMember1 := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	
	// 尝试创建重复的团队成员
	teamMember2 := memberFactory.CreateTeamMember(team.Id, member.Id)
	
	// 测试插入应该失败
	err = teamMember2.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户已经是团队成员")
}

// TestInsertWithNonExistentTeam 测试添加成员到不存在的团队
func (suite *TeamMemberModelSuite) TestInsertWithNonExistentTeam() {
	userFactory := suite.factory.NewUserFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员（团队不存在）
	teamMember := memberFactory.CreateTeamMember(99999, member.Id)
	
	// 测试插入应该失败
	err = teamMember.Insert()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队不存在")
}

// TestUpdate 测试更新团队成员
func (suite *TeamMemberModelSuite) TestUpdate() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 更新团队成员信息
	teamMember.Role = 1 // 提升为管理员
	teamMember.MaxQuota = 50000
	teamMember.Status = 2 // 禁用
	
	// 测试更新
	err = teamMember.Update()
	assert.NoError(suite.T(), err)
	
	// 验证数据库中的数据
	var dbMember model.TeamMember
	err = suite.db.First(&dbMember, teamMember.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, dbMember.Role)
	assert.Equal(suite.T(), 50000, dbMember.MaxQuota)
	assert.Equal(suite.T(), 2, dbMember.Status)
}

// TestUpdateWithZeroId 测试更新团队成员时 ID 为 0
func (suite *TeamMemberModelSuite) TestUpdateWithZeroId() {
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试团队成员（ID 为 0）
	teamMember := memberFactory.CreateTeamMember(1, 1)
	teamMember.Id = 0
	
	// 测试更新应该失败
	err := teamMember.Update()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "member id 为空")
}

// TestDelete 测试删除团队成员
func (suite *TeamMemberModelSuite) TestDelete() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	teamMember.UsedQuota = 10000 // 设置已使用额度
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试删除
	err = teamMember.Delete()
	assert.NoError(suite.T(), err)
	
	// 验证团队成员已被软删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.TeamMember{}, "id = ? AND deleted_at IS NULL", teamMember.Id)
}

// TestDeleteWithZeroId 测试删除团队成员时 ID 为 0
func (suite *TeamMemberModelSuite) TestDeleteWithZeroId() {
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试团队成员（ID 为 0）
	teamMember := memberFactory.CreateTeamMember(1, 1)
	teamMember.Id = 0
	
	// 测试删除应该失败
	err := teamMember.Delete()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "member id 为空")
}

// TestGetTeamMember 测试获取团队成员
func (suite *TeamMemberModelSuite) TestGetTeamMember() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取团队成员
	retrievedMember, err := model.GetTeamMember(team.Id, member.Id)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrievedMember)
	testutils.AssertTeamMemberEqual(suite.T(), teamMember, retrievedMember)
	
	// 验证预加载的 User 信息
	assert.NotNil(suite.T(), retrievedMember.User)
	assert.Equal(suite.T(), member.Username, retrievedMember.User.Username)
}

// TestGetTeamMemberWithZeroIds 测试获取团队成员时参数为 0
func (suite *TeamMemberModelSuite) TestGetTeamMemberWithZeroIds() {
	// 测试 TeamId 为 0
	member, err := model.GetTeamMember(0, 1)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), member)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
	
	// 测试 UserId 为 0
	member, err = model.GetTeamMember(1, 0)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), member)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
}

// TestGetTeamMemberNotFound 测试获取不存在的团队成员
func (suite *TeamMemberModelSuite) TestGetTeamMemberNotFound() {
	// 测试获取不存在的团队成员
	member, err := model.GetTeamMember(99999, 99999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), member)
}

// TestGetTeamMembersList 测试获取团队成员列表
func (suite *TeamMemberModelSuite) TestGetTeamMembersList() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member1 := userFactory.CreateUserWithQuota(200000)
	member2 := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member1).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member2).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember1 := memberFactory.CreateTeamMember(team.Id, member1.Id)
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	
	teamMember2 := memberFactory.CreateTeamMember(team.Id, member2.Id)
	err = teamMember2.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取团队成员列表
	params := &model.SearchTeamMemberParams{
		PaginationParams: model.PaginationParams{Page: 1, Size: 10},
	}
	members, err := model.GetTeamMembersList(team.Id, params)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), members)
	require.NotNil(suite.T(), members.Data)
	assert.Len(suite.T(), *members.Data, 2)
	
	// 验证返回的成员信息包含可用额度计算
	for _, member := range *members.Data {
		assert.GreaterOrEqual(suite.T(), member.AvailableQuota, 0)
	}
}

// TestGetTeamMembersListWithRoleFilter 测试根据角色过滤团队成员列表
func (suite *TeamMemberModelSuite) TestGetTeamMembersListWithRoleFilter() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member1 := userFactory.CreateUserWithQuota(200000)
	member2 := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member1).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member2).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员（不同角色）
	teamMember1 := memberFactory.CreateAdminMember(team.Id, member1.Id) // 管理员
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	
	teamMember2 := memberFactory.CreateTeamMember(team.Id, member2.Id) // 普通成员
	err = teamMember2.Insert()
	require.NoError(suite.T(), err)
	
	// 测试根据角色过滤（只获取管理员）
	params := &model.SearchTeamMemberParams{
		TeamMember: model.TeamMember{Role: 1}, // 管理员
		PaginationParams: model.PaginationParams{Page: 1, Size: 10},
	}
	members, err := model.GetTeamMembersList(team.Id, params)
	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), members)
	require.NotNil(suite.T(), members.Data)
	assert.Len(suite.T(), *members.Data, 1)
	assert.Equal(suite.T(), 1, (*members.Data)[0].Role)
}

// TestGetMemberAvailableQuota 测试获取成员可用额度
func (suite *TeamMemberModelSuite) TestGetMemberAvailableQuota() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建有限额度团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员（无额度限制）
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取成员可用额度
	availableQuota, unlimited, err := model.GetMemberAvailableQuota(team.Id, member.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100000, availableQuota) // 团队可用额度
	assert.False(suite.T(), unlimited)
}

// TestGetMemberAvailableQuotaWithMemberLimit 测试获取有额度限制的成员可用额度
func (suite *TeamMemberModelSuite) TestGetMemberAvailableQuotaWithMemberLimit() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建有限额度团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员（有额度限制）
	teamMember := memberFactory.CreateMemberWithQuotaLimit(team.Id, member.Id, 50000)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取成员可用额度
	availableQuota, unlimited, err := model.GetMemberAvailableQuota(team.Id, member.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50000, availableQuota) // 成员额度限制
	assert.False(suite.T(), unlimited)
}

// TestGetMemberAvailableQuotaUnlimitedTeam 测试获取无限额度团队成员的可用额度
func (suite *TeamMemberModelSuite) TestGetMemberAvailableQuotaUnlimitedTeam() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建无限额度团队
	team := teamFactory.CreateUnlimitedTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员（无额度限制）
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取成员可用额度
	availableQuota, unlimited, err := model.GetMemberAvailableQuota(team.Id, member.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, availableQuota) // 无限额度
	assert.True(suite.T(), unlimited)
}

// TestGetMemberAvailableQuotaNotFound 测试获取不存在成员的可用额度
func (suite *TeamMemberModelSuite) TestGetMemberAvailableQuotaNotFound() {
	// 测试获取不存在成员的可用额度
	availableQuota, unlimited, err := model.GetMemberAvailableQuota(99999, 99999)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), 0, availableQuota)
	assert.False(suite.T(), unlimited)
}

// TestUpdateTeamMemberQuota 测试更新成员额度限制
func (suite *TeamMemberModelSuite) TestUpdateTeamMemberQuota() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试更新成员额度限制
	err = model.UpdateTeamMemberQuota(team.Id, member.Id, 50000)
	assert.NoError(suite.T(), err)
	
	// 验证数据库中的数据
	var dbMember model.TeamMember
	err = suite.db.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&dbMember).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50000, dbMember.MaxQuota)
}

// TestUpdateTeamMemberQuotaWithZeroIds 测试更新成员额度限制时参数为 0
func (suite *TeamMemberModelSuite) TestUpdateTeamMemberQuotaWithZeroIds() {
	// 测试 TeamId 为 0
	err := model.UpdateTeamMemberQuota(0, 1, 50000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
	
	// 测试 UserId 为 0
	err = model.UpdateTeamMemberQuota(1, 0, 50000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
}

// TestDeleteTeamMember 测试删除团队成员
func (suite *TeamMemberModelSuite) TestDeleteTeamMember() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	teamMember.UsedQuota = 10000 // 设置已使用额度
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试删除团队成员
	err = model.DeleteTeamMember(team.Id, member.Id)
	assert.NoError(suite.T(), err)
	
	// 验证团队成员已被软删除
	testutils.AssertDatabaseRecordNotExists(suite.T(), suite.db, &model.TeamMember{}, "team_id = ? AND user_id = ? AND deleted_at IS NULL", team.Id, member.Id)
}

// TestDeleteTeamMemberWithZeroIds 测试删除团队成员时参数为 0
func (suite *TeamMemberModelSuite) TestDeleteTeamMemberWithZeroIds() {
	// 测试 TeamId 为 0
	err := model.DeleteTeamMember(0, 1)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
	
	// 测试 UserId 为 0
	err = model.DeleteTeamMember(1, 0)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "team_id 或 user_id 为空")
}

// TestIsTeamAdmin 测试检查是否为团队管理员
func (suite *TeamMemberModelSuite) TestIsTeamAdmin() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	admin := userFactory.CreateUserWithQuota(200000)
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(admin).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建管理员成员
	adminMember := memberFactory.CreateAdminMember(team.Id, admin.Id)
	err = adminMember.Insert()
	require.NoError(suite.T(), err)
	
	// 创建普通成员
	normalMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = normalMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试检查管理员
	assert.True(suite.T(), model.IsTeamAdmin(team.Id, admin.Id))
	assert.False(suite.T(), model.IsTeamAdmin(team.Id, member.Id))
	assert.False(suite.T(), model.IsTeamAdmin(team.Id, owner.Id)) // Owner 不是管理员成员
	assert.False(suite.T(), model.IsTeamAdmin(99999, admin.Id))
}

// TestGetTeamMemberCount 测试获取团队成员数量
func (suite *TeamMemberModelSuite) TestGetTeamMemberCount() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	member1 := userFactory.CreateUserWithQuota(200000)
	member2 := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member1).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member2).Error
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试团队成员
	teamMember1 := memberFactory.CreateTeamMember(team.Id, member1.Id)
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	
	teamMember2 := memberFactory.CreateTeamMember(team.Id, member2.Id)
	err = teamMember2.Insert()
	require.NoError(suite.T(), err)
	
	// 测试获取团队成员数量
	count, err := model.GetTeamMemberCount(team.Id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), count)
	
	// 测试获取不存在团队的成员数量
	count, err = model.GetTeamMemberCount(99999)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

// TestTeamMemberModelSuite 运行测试套件
func TestTeamMemberModelSuite(t *testing.T) {
	suite.Run(t, new(TeamMemberModelSuite))
}
