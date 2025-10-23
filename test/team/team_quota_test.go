package team

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"one-hub/model"
)

// 测试数据库设置
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&model.User{}, &model.Team{}, &model.TeamMember{}, &model.Log{})
	require.NoError(t, err)

	return db
}

// 创建测试用户
func createTestUser(db *gorm.DB, t *testing.T, username string, quota int) *model.User {
	user := &model.User{
		Username: username,
		Quota:    quota,
		Status:   1,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// 创建测试团队
func createTestTeam(db *gorm.DB, t *testing.T, name string, ownerId int, quota int) *model.Team {
	team := &model.Team{
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
	err := db.Create(team).Error
	require.NoError(t, err)
	return team
}

// 创建测试团队成员
func createTestTeamMember(db *gorm.DB, t *testing.T, teamId, userId, role int, maxQuota int) *model.TeamMember {
	member := &model.TeamMember{
		TeamId:     teamId,
		UserId:     userId,
		Role:       role,
		MaxQuota:   maxQuota,
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err := db.Create(member).Error
	require.NoError(t, err)
	return member
}

// TestAllocateQuotaToTeam 测试团队额度分配
func TestAllocateQuotaToTeam(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户（管理员）
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 0)
	
	// 测试分配有限额度
	err := model.AllocateQuotaToTeam(owner.Id, team.Id, 500000, false)
	assert.NoError(t, err)
	
	// 验证管理员额度减少
	var updatedOwner model.User
	db.First(&updatedOwner, owner.Id)
	assert.Equal(t, 500000, updatedOwner.Quota)
	
	// 验证团队额度增加
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 500000, updatedTeam.Quota)
	
	// 测试分配无限额度
	err = model.AllocateQuotaToTeam(owner.Id, team.Id, 0, true)
	assert.NoError(t, err)
	
	// 验证团队设置为无限额度
	db.First(&updatedTeam, team.Id)
	assert.True(t, updatedTeam.UnlimitedQuota)
}

// TestConsumeTeamQuota 测试团队额度消费
func TestConsumeTeamQuota(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0) // 普通成员，无限制
	
	// 测试消费团队额度
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(t, err)
	assert.Equal(t, 100000, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed)
	
	// 验证团队额度减少
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 400000, updatedTeam.Quota)
	assert.Equal(t, 100000, updatedTeam.UsedQuota)
	
	// 验证成员已用额度增加
	var updatedMember model.TeamMember
	db.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&updatedMember)
	assert.Equal(t, 100000, updatedMember.UsedQuota)
}

// TestConsumeTeamQuotaWithMixedSource 测试混合额度消费
func TestConsumeTeamQuotaWithMixedSource(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建测试团队（额度不足）
	team := createTestTeam(db, t, "测试团队", owner.Id, 50000)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 测试消费超过团队额度的请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(t, err)
	assert.Equal(t, 50000, teamQuotaUsed)  // 团队额度全部用完
	assert.Equal(t, 50000, userQuotaUsed) // 个人额度补足差额
	
	// 验证团队额度用完
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 0, updatedTeam.Quota)
	assert.Equal(t, 50000, updatedTeam.UsedQuota)
	
	// 验证成员个人额度减少
	var updatedMember model.User
	db.First(&updatedMember, member.Id)
	assert.Equal(t, 150000, updatedMember.Quota)
}

// TestConsumeTeamQuotaWithMemberLimit 测试成员额度限制
func TestConsumeTeamQuotaWithMemberLimit(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 创建团队成员（有额度限制）
	createTestTeamMember(db, t, team.Id, member.Id, 2, 100000) // 最大额度100000
	
	// 测试消费超过成员限制的请求
	_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 150000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "超出团队成员额度限制")
	
	// 测试消费在成员限制内的请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 80000)
	assert.NoError(t, err)
	assert.Equal(t, 80000, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed)
}

// TestConsumeTeamQuotaInsufficientUserQuota 测试用户个人额度不足
func TestConsumeTeamQuotaInsufficientUserQuota(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 50000) // 个人额度较少
	
	// 创建测试团队（额度不足）
	team := createTestTeam(db, t, "测试团队", owner.Id, 50000)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 测试消费超过团队和个人额度的请求
	_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 150000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户额度不足")
}

// TestConsumeTeamQuotaUnlimitedTeam 测试无限额度团队
func TestConsumeTeamQuotaUnlimitedTeam(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建无限额度团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 0)
	team.UnlimitedQuota = true
	db.Save(team)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 测试消费大额度请求
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 1000000)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed)
	
	// 验证团队额度不变（无限额度）
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 0, updatedTeam.Quota)
	assert.Equal(t, 1000000, updatedTeam.UsedQuota)
}

// TestConcurrentTeamQuotaConsumption 测试并发团队额度消费
func TestConcurrentTeamQuotaConsumption(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member1 := createTestUser(db, t, "member1", 200000)
	member2 := createTestUser(db, t, "member2", 200000)
	member3 := createTestUser(db, t, "member3", 200000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member1.Id, 2, 0)
	createTestTeamMember(db, t, team.Id, member2.Id, 2, 0)
	createTestTeamMember(db, t, team.Id, member3.Id, 2, 0)
	
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
	assert.GreaterOrEqual(t, successCount, 2)
	
	// 验证最终团队额度
	var finalTeam model.Team
	db.First(&finalTeam, team.Id)
	assert.LessOrEqual(t, finalTeam.UsedQuota, 500000)
}

// TestConcurrentQuotaAllocation 测试并发额度分配
func TestConcurrentQuotaAllocation(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户（大额度）
	owner := createTestUser(db, t, "owner", 10000000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 0)
	
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
	assert.Equal(t, 5, successCount)
	
	// 验证最终团队额度
	var finalTeam model.Team
	db.First(&finalTeam, team.Id)
	assert.Equal(t, 500000, finalTeam.Quota)
	
	// 验证管理员剩余额度
	var finalOwner model.User
	db.First(&finalOwner, owner.Id)
	assert.Equal(t, 9500000, finalOwner.Quota)
}

// TestTeamQuotaEdgeCases 测试边界情况
func TestTeamQuotaEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 100000)
	
	// 创建团队成员
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 测试消费0额度
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed)
	
	// 测试消费负数额度（应该被拒绝）
	_, _, err = model.ConsumeTeamQuota(member.Id, team.Id, -100)
	assert.Error(t, err)
	
	// 测试不存在的团队成员
	_, _, err = model.ConsumeTeamQuota(999, team.Id, 1000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不是团队成员")
	
	// 测试不存在的团队
	_, _, err = model.ConsumeTeamQuota(member.Id, 999, 1000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "团队不存在")
}

// BenchmarkTeamQuotaConsumption 性能测试
func BenchmarkTeamQuotaConsumption(b *testing.B) {
	db := setupTestDB(&testing.T{})
	
	// 创建测试数据
	owner := createTestUser(db, &testing.T{}, "owner", 100000000)
	member := createTestUser(db, &testing.T{}, "member", 100000000)
	team := createTestTeam(db, &testing.T{}, "测试团队", owner.Id, 50000000)
	createTestTeamMember(db, &testing.T{}, team.Id, member.Id, 2, 0)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _, _ = model.ConsumeTeamQuota(member.Id, team.Id, 1000)
	}
}

// TestGetEffectiveQuotaForContext 测试统一额度查询接口
func TestGetEffectiveQuotaForContext(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建有限额度团队
	team1 := createTestTeam(db, t, "有限额度团队", owner.Id, 500000)
	createTestTeamMember(db, t, team1.Id, member.Id, 2, 0)
	
	// 创建无限额度团队
	team2 := createTestTeam(db, t, "无限额度团队", owner.Id, 0)
	team2.UnlimitedQuota = true
	db.Save(team2)
	createTestTeamMember(db, t, team2.Id, member.Id, 2, 0)
	
	// 测试场景 A: Owner 在个人上下文
	quota, unlimited, err := model.GetEffectiveQuotaForContext(owner.Id, "user", owner.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, quota)
	assert.False(t, unlimited)
	
	// 测试场景 B: Owner 在团队上下文（有限额度）
	quota, unlimited, err = model.GetEffectiveQuotaForContext(owner.Id, "team", team1.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, quota) // Owner 返回个人余额
	assert.False(t, unlimited)
	
	// 测试场景 C: Owner 在团队上下文（无限额度）
	quota, unlimited, err = model.GetEffectiveQuotaForContext(owner.Id, "team", team2.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, quota) // Owner 返回个人余额
	assert.False(t, unlimited)
	
	// 测试场景 D: 成员在团队上下文（有限额度）
	quota, unlimited, err = model.GetEffectiveQuotaForContext(member.Id, "team", team1.Id)
	assert.NoError(t, err)
	assert.Equal(t, 500000, quota) // 返回团队可用额度
	assert.False(t, unlimited)
	
	// 测试场景 E: 成员在团队上下文（无限额度）
	quota, unlimited, err = model.GetEffectiveQuotaForContext(member.Id, "team", team2.Id)
	assert.NoError(t, err)
	assert.Equal(t, 1000000, quota) // 返回 Owner 个人余额
	assert.False(t, unlimited)
	
	// 测试场景 F: 成员在个人上下文
	quota, unlimited, err = model.GetEffectiveQuotaForContext(member.Id, "user", member.Id)
	assert.NoError(t, err)
	assert.Equal(t, 200000, quota)
	assert.False(t, unlimited)
}

// TestTeamQuotaOperations 测试团队额度操作函数
func TestTeamQuotaOperations(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户和团队
	owner := createTestUser(db, t, "owner", 1000000)
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 测试增加团队已用额度
	err := model.IncreaseTeamUsedQuota(team.Id, 100000)
	assert.NoError(t, err)
	
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 100000, updatedTeam.UsedQuota)
	
	// 测试减少团队已用额度
	err = model.DecreaseTeamUsedQuota(team.Id, 50000)
	assert.NoError(t, err)
	
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 50000, updatedTeam.UsedQuota)
	
	// 测试错误参数
	err = model.IncreaseTeamUsedQuota(0, 100000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数错误")
	
	err = model.IncreaseTeamUsedQuota(team.Id, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数错误")
	
	err = model.DecreaseTeamUsedQuota(0, 100000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数错误")
	
	err = model.DecreaseTeamUsedQuota(team.Id, -100000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "参数错误")
}

// TestOwnerUsingTeamToken 测试 Owner 使用团队 Token 的场景
func TestOwnerUsingTeamToken(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建无限额度团队
	team := createTestTeam(db, t, "无限额度团队", owner.Id, 0)
	team.UnlimitedQuota = true
	db.Save(team)
	
	// 模拟 Owner 使用团队 Token 消费
	// 这里应该直接扣除 Owner 的个人额度，而不是团队额度
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(owner.Id, team.Id, 100000)
	assert.NoError(t, err)
	
	// 验证：Owner 使用团队 Token 时，应该扣除个人额度
	// 注意：这里需要根据实际的 ConsumeTeamQuota 实现来调整
	// 如果 ConsumeTeamQuota 没有区分 Owner 和成员，我们需要在调用层面处理
	
	// 验证团队已用额度增加（用于统计）
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 100000, updatedTeam.UsedQuota)
	
	// 验证 Owner 个人额度减少
	var updatedOwner model.User
	db.First(&updatedOwner, owner.Id)
	assert.Equal(t, 900000, updatedOwner.Quota)
}

// TestMemberUsingTeamToken 测试成员使用团队 Token 的场景
func TestMemberUsingTeamToken(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建有限额度团队
	team := createTestTeam(db, t, "有限额度团队", owner.Id, 500000)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 模拟成员使用团队 Token 消费
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(t, err)
	assert.Equal(t, 100000, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed) // 成员个人额度不应该被扣除
	
	// 验证团队额度减少
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 400000, updatedTeam.Quota)
	assert.Equal(t, 100000, updatedTeam.UsedQuota)
	
	// 验证成员个人额度不变
	var updatedMember model.User
	db.First(&updatedMember, member.Id)
	assert.Equal(t, 200000, updatedMember.Quota)
	
	// 验证 Owner 个人额度不变
	var updatedOwner model.User
	db.First(&updatedOwner, owner.Id)
	assert.Equal(t, 1000000, updatedOwner.Quota)
}

// TestMemberUsingUnlimitedTeamToken 测试成员使用无限额度团队 Token 的场景
func TestMemberUsingUnlimitedTeamToken(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建无限额度团队
	team := createTestTeam(db, t, "无限额度团队", owner.Id, 0)
	team.UnlimitedQuota = true
	db.Save(team)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 模拟成员使用无限额度团队 Token 消费
	teamQuotaUsed, userQuotaUsed, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.NoError(t, err)
	assert.Equal(t, 100000, teamQuotaUsed)
	assert.Equal(t, 0, userQuotaUsed) // 成员个人额度不应该被扣除
	
	// 验证团队已用额度增加（用于统计）
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 100000, updatedTeam.UsedQuota)
	assert.Equal(t, 0, updatedTeam.Quota) // 团队额度保持为0（无限额度）
	
	// 验证成员个人额度不变
	var updatedMember model.User
	db.First(&updatedMember, member.Id)
	assert.Equal(t, 200000, updatedMember.Quota)
	
	// 验证 Owner 个人额度不变（这里需要根据实际实现调整）
	// 如果无限额度团队从 Owner 余额扣除，那么 Owner 额度应该减少
	var updatedOwner model.User
	db.First(&updatedOwner, owner.Id)
	// 根据实际实现，这里可能是 900000 或 1000000
	// assert.Equal(t, 900000, updatedOwner.Quota) // 如果从 Owner 扣除
	assert.Equal(t, 1000000, updatedOwner.Quota) // 如果团队额度独立
}

// TestTeamQuotaExhausted 测试团队额度耗尽的情况
func TestTeamQuotaExhausted(t *testing.T) {
	db := setupTestDB(t)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	member := createTestUser(db, t, "member", 200000)
	
	// 创建有限额度团队（额度较少）
	team := createTestTeam(db, t, "有限额度团队", owner.Id, 50000)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 测试消费超过团队额度的请求
	_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 100000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team quota is not enough")
	
	// 验证团队额度没有变化
	var updatedTeam model.Team
	db.First(&updatedTeam, team.Id)
	assert.Equal(t, 50000, updatedTeam.Quota)
	assert.Equal(t, 0, updatedTeam.UsedQuota)
	
	// 验证成员个人额度没有变化
	var updatedMember model.User
	db.First(&updatedMember, member.Id)
	assert.Equal(t, 200000, updatedMember.Quota)
}

