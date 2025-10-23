package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"one-api/model"
	"one-api/test/testutils"
)

// TeamQuotaSuite 团队额度测试套件
type TeamQuotaSuite struct {
	suite.Suite
	db      *gorm.DB
	factory *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamQuotaSuite) SetupSuite() {
	suite.db = testutils.SetupTestDB(suite.T(), nil)
	suite.factory = testutils.NewTestDataFactory()
}

// TearDownSuite 测试套件清理
func (suite *TeamQuotaSuite) TearDownSuite() {
	testutils.CleanupTestDB(suite.T(), suite.db)
}

// SetupTest 每个测试前的设置
func (suite *TeamQuotaSuite) SetupTest() {
	testutils.ResetTestDB(suite.T(), suite.db)
}

// TestAllocateQuotaToTeam 测试分配团队额度
func (suite *TeamQuotaSuite) TestAllocateQuotaToTeam() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户（管理员）
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试分配有限额度
	err = model.AllocateQuotaToTeam(owner.Id, team.Id, 500000, false)
	assert.NoError(suite.T(), err)
	
	// 验证管理员额度减少
	var updatedOwner model.User
	err = suite.db.First(&updatedOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	testutils.AssertQuotaDeducted(suite.T(), owner.Quota, updatedOwner.Quota, 500000)
	
	// 验证团队额度增加
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500000, updatedTeam.Quota)
	assert.False(suite.T(), updatedTeam.UnlimitedQuota)
}

// TestAllocateQuotaToTeamUnlimited 测试分配无限额度
func (suite *TeamQuotaSuite) TestAllocateQuotaToTeamUnlimited() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户（管理员）
	owner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试分配无限额度
	err = model.AllocateQuotaToTeam(owner.Id, team.Id, 0, true)
	assert.NoError(suite.T(), err)
	
	// 验证管理员额度不变
	var updatedOwner model.User
	err = suite.db.First(&updatedOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), owner.Quota, updatedOwner.Quota)
	
	// 验证团队设置为无限额度
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.True(suite.T(), updatedTeam.UnlimitedQuota)
}

// TestAllocateQuotaToTeamInsufficientBalance 测试管理员额度不足
func (suite *TeamQuotaSuite) TestAllocateQuotaToTeamInsufficientBalance() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户（管理员额度不足）
	owner := userFactory.CreateUserWithQuota(100000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试分配超过管理员额度的团队额度
	err = model.AllocateQuotaToTeam(owner.Id, team.Id, 500000, false)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户额度不足")
	
	// 验证管理员额度不变
	var updatedOwner model.User
	err = suite.db.First(&updatedOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), owner.Quota, updatedOwner.Quota)
	
	// 验证团队额度不变
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), team.Quota, updatedTeam.Quota)
}

// TestAllocateQuotaToTeamNotOwner 测试非管理员分配团队额度
func (suite *TeamQuotaSuite) TestAllocateQuotaToTeamNotOwner() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	nonOwner := userFactory.CreateUserWithQuota(1000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(nonOwner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试非管理员分配团队额度
	err = model.AllocateQuotaToTeam(nonOwner.Id, team.Id, 500000, false)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户不是团队管理员")
}

// TestConsumeTeamQuota 测试消费团队额度
func (suite *TeamQuotaSuite) TestConsumeTeamQuota() {
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
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费团队额度
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100000, teamQuotaUsed)
	assert.Equal(suite.T(), 0, userQuotaUsed)
	
	// 验证团队额度减少
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 400000, updatedTeam.Quota)
	assert.Equal(suite.T(), 100000, updatedTeam.UsedQuota)
	
	// 验证成员已用额度增加
	var updatedMember model.TeamMember
	err = suite.db.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&updatedMember).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100000, updatedMember.UsedQuota)
	
	// 验证成员个人额度不变
	var updatedUser model.User
	err = suite.db.First(&updatedUser, member.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), member.Quota, updatedUser.Quota)
}

// TestConsumeTeamQuotaWithMixedSource 测试混合额度消费
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaWithMixedSource() {
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
	
	// 创建测试团队（额度不足）
	team := teamFactory.CreateTeamWithQuota(owner.Id, 50000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费超过团队额度的请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50000, teamQuotaUsed)  // 团队额度全部用完
	assert.Equal(suite.T(), 50000, userQuotaUsed) // 个人额度补足差额
	
	// 验证团队额度用完
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, updatedTeam.Quota)
	assert.Equal(suite.T(), 50000, updatedTeam.UsedQuota)
	
	// 验证成员个人额度减少
	var updatedMember model.User
	err = suite.db.First(&updatedMember, member.Id).Error
	require.NoError(suite.T(), err)
	testutils.AssertQuotaDeducted(suite.T(), member.Quota, updatedMember.Quota, 50000)
}

// TestConsumeTeamQuotaWithMemberLimit 测试成员额度限制
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaWithMemberLimit() {
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
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员（有额度限制）
	teamMember := memberFactory.CreateMemberWithQuotaLimit(team.Id, member.Id, 100000)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费超过成员限制的请求
	_, _, err = model.ConsumeTeamQuota(member.Id, team.Id, 150000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "超出团队成员额度限制")
	
	// 测试消费在成员限制内的请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 80000)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 80000, teamQuotaUsed)
	assert.Equal(suite.T(), 0, userQuotaUsed)
}

// TestConsumeTeamQuotaInsufficientUserQuota 测试用户个人额度不足
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaInsufficientUserQuota() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(50000) // 个人额度较少
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队（额度不足）
	team := teamFactory.CreateTeamWithQuota(owner.Id, 50000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费超过团队和个人额度的请求
	_, _, err = model.ConsumeTeamQuota(member.Id, team.Id, 150000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户额度不足")
}

// TestConsumeTeamQuotaUnlimitedTeam 测试无限额度团队
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaUnlimitedTeam() {
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
	
	// 创建无限额度团队
	team := teamFactory.CreateUnlimitedTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费大额度请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 1000000)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1000000, teamQuotaUsed)
	assert.Equal(suite.T(), 0, userQuotaUsed)
	
	// 验证团队已用额度增加（用于统计）
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1000000, updatedTeam.UsedQuota)
	assert.Equal(suite.T(), 0, updatedTeam.Quota) // 团队额度保持为0（无限额度）
	
	// 验证成员个人额度不变
	var updatedMember model.User
	err = suite.db.First(&updatedMember, member.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), member.Quota, updatedMember.Quota)
}

// TestConsumeTeamQuotaNotMember 测试非团队成员消费团队额度
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaNotMember() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	nonMember := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(nonMember).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试非团队成员消费团队额度
	_, _, err = model.ConsumeTeamQuota(nonMember.Id, team.Id, 100000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户不是团队成员")
}

// TestConsumeTeamQuotaNonExistentTeam 测试消费不存在团队的额度
func (suite *TeamQuotaSuite) TestConsumeTeamQuotaNonExistentTeam() {
	userFactory := suite.factory.NewUserFactory()
	
	// 创建测试用户
	member := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(member).Error
	require.NoError(suite.T(), err)
	
	// 测试消费不存在团队的额度
	_, _, err = model.ConsumeTeamQuota(member.Id, 99999, 100000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队不存在")
}

// TestConcurrentTeamQuotaConsumption 测试并发团队额度消费
func (suite *TeamQuotaSuite) TestConcurrentTeamQuotaConsumption() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member1 := userFactory.CreateUserWithQuota(200000)
	member2 := userFactory.CreateUserWithQuota(200000)
	member3 := userFactory.CreateUserWithQuota(200000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member1).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member2).Error
	require.NoError(suite.T(), err)
	err = suite.db.Create(member3).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember1 := memberFactory.CreateTeamMember(team.Id, member1.Id)
	teamMember2 := memberFactory.CreateTeamMember(team.Id, member2.Id)
	teamMember3 := memberFactory.CreateTeamMember(team.Id, member3.Id)
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	err = teamMember2.Insert()
	require.NoError(suite.T(), err)
	err = teamMember3.Insert()
	require.NoError(suite.T(), err)
	
	// 并发消费测试
	var wg sync.WaitGroup
	results := make(chan error, 3)
	
	// 启动3个并发消费
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(memberId int) {
			defer wg.Done()
			_, _, err := model.ConsumeTeamQuota(memberId, team.Id, 100000)
			results <- err
		}([]int{member1.Id, member2.Id, member3.Id}[i])
	}
	
	wg.Wait()
	close(results)
	
	// 检查结果
	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}
	
	// 应该至少有2个成功（团队额度500000，每个消费100000）
	assert.GreaterOrEqual(suite.T(), successCount, 2)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err = suite.db.First(&finalTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), finalTeam.UsedQuota, 500000)
}

// TestConcurrentQuotaAllocation 测试并发额度分配
func (suite *TeamQuotaSuite) TestConcurrentQuotaAllocation() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户（大额度）
	owner := userFactory.CreateUserWithQuota(10000000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 并发分配测试
	var wg sync.WaitGroup
	results := make(chan error, 5)
	
	// 启动5个并发分配
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 100000, false)
			results <- err
		}()
	}
	
	wg.Wait()
	close(results)
	
	// 检查结果
	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}
	
	// 所有分配都应该成功
	assert.Equal(suite.T(), 5, successCount)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err = suite.db.First(&finalTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 500000, finalTeam.Quota)
	
	// 验证管理员剩余额度
	var finalOwner model.User
	err = suite.db.First(&finalOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9500000, finalOwner.Quota)
}

// TestTeamQuotaEdgeCases 测试边界情况
func (suite *TeamQuotaSuite) TestTeamQuotaEdgeCases() {
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
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 测试消费0额度
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 0)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, teamQuotaUsed)
	assert.Equal(suite.T(), 0, userQuotaUsed)
	
	// 测试消费负数额度（应该被拒绝）
	_, _, err = model.ConsumeTeamQuota(member.Id, team.Id, -100)
	assert.Error(suite.T(), err)
	
	// 测试不存在的团队成员
	_, _, err = model.ConsumeTeamQuota(999, team.Id, 1000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "用户不是团队成员")
	
	// 测试不存在的团队
	_, _, err = model.ConsumeTeamQuota(member.Id, 999, 1000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "团队不存在")
}

// TestTeamQuotaOperations 测试团队额度操作函数
func (suite *TeamQuotaSuite) TestTeamQuotaOperations() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户和团队
	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err := suite.db.Create(owner).Error
	require.NoError(suite.T(), err)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 测试增加团队已用额度
	err = model.IncreaseTeamUsedQuota(team.Id, 100000)
	assert.NoError(suite.T(), err)
	
	var updatedTeam model.Team
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100000, updatedTeam.UsedQuota)
	
	// 测试减少团队已用额度
	err = model.DecreaseTeamUsedQuota(team.Id, 50000)
	assert.NoError(suite.T(), err)
	
	err = suite.db.First(&updatedTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 50000, updatedTeam.UsedQuota)
	
	// 测试错误参数
	err = model.IncreaseTeamUsedQuota(0, 100000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "参数错误")
	
	err = model.IncreaseTeamUsedQuota(team.Id, 0)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "参数错误")
	
	err = model.DecreaseTeamUsedQuota(0, 100000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "参数错误")
	
	err = model.DecreaseTeamUsedQuota(team.Id, -100000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "参数错误")
}

// TestTeamQuotaSuite 运行测试套件
func TestTeamQuotaSuite(t *testing.T) {
	suite.Run(t, new(TeamQuotaSuite))
}
