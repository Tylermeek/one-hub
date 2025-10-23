package team

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"one-api/model"
	"one-api/test/testutils"
)

// TestConcurrentTeamOperations 测试并发团队操作
func TestConcurrentTeamOperations(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 10000000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 0)
	
	// 并发操作测试
	var wg sync.WaitGroup
	results := make(chan error, 10)
	
	// 启动10个并发操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// 分配额度
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 10000, false)
			results <- err
		}(i)
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
	
	// 所有操作都应该成功
	assert.Equal(t, 10, successCount)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err := db.First(&finalTeam, team.Id).Error
	require.NoError(t, err)
	assert.Equal(t, 100000, finalTeam.Quota)
}

// TestConcurrentMemberOperations 测试并发成员操作
func TestConcurrentMemberOperations(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建多个测试用户
	var members []*model.User
	for i := 0; i < 5; i++ {
		member := createTestUser(db, t, fmt.Sprintf("member%d", i), 200000)
		members = append(members, member)
	}
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 并发邀请成员
	var wg sync.WaitGroup
	results := make(chan error, 5)
	
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(member *model.User) {
			defer wg.Done()
			
			// 创建团队成员记录
			memberRecord := &model.TeamMember{
				TeamId:     team.Id,
				UserId:     member.Id,
				Role:       2,
				MaxQuota:   0,
				UsedQuota:  0,
				Status:     1,
				JoinedTime: time.Now().Unix(),
			}
			err := memberRecord.Insert()
			results <- err
		}(members[i])
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
	
	// 所有操作都应该成功
	assert.Equal(t, 5, successCount)
	
	// 验证团队成员数量
	var memberCount int64
	err := db.Model(&model.TeamMember{}).Where("team_id = ?", team.Id).Count(&memberCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(5), memberCount)
}

// TestConcurrentQuotaConsumption 测试并发额度消费
func TestConcurrentQuotaConsumption(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建多个测试用户
	var members []*model.User
	for i := 0; i < 10; i++ {
		member := createTestUser(db, t, fmt.Sprintf("member%d", i), 100000)
		members = append(members, member)
	}
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 500000)
	
	// 创建团队成员
	for _, member := range members {
		createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	}
	
	// 并发消费额度
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	// 每个成员消费2次
	for _, member := range members {
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(memberId int) {
				defer wg.Done()
				
				_, _, err := model.ConsumeTeamQuota(memberId, team.Id, 10000)
				results <- err
			}(member.Id)
		}
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
	
	// 应该有一些成功的消费（团队额度500000，每次消费10000，最多50次）
	assert.GreaterOrEqual(t, successCount, 30)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err := db.First(&finalTeam, team.Id).Error
	require.NoError(t, err)
	assert.LessOrEqual(t, finalTeam.UsedQuota, 500000)
}

// TestConcurrentMixedOperations 测试混合并发操作
func TestConcurrentMixedOperations(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 10000000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 0)
	
	// 创建测试成员
	member := createTestUser(db, t, "member", 1000000)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 并发混合操作
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	// 分配额度操作
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 10000, false)
			results <- err
		}()
	}
	
	// 消费额度操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 1000)
			results <- err
		}()
	}
	
	// 查询操作
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var team model.Team
			err := db.First(&team, team.Id).Error
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
	
	// 大部分操作应该成功
	assert.GreaterOrEqual(t, successCount, 15)
}

// TestConcurrentWithContext 测试带上下文的并发操作
func TestConcurrentWithContext(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建测试团队
	team := createTestTeam(db, t, "测试团队", owner.Id, 100000)
	
	// 创建测试成员
	member := createTestUser(db, t, "member", 100000)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 并发操作
	var wg sync.WaitGroup
	results := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				results <- ctx.Err()
				return
			default:
			}
			
			// 执行消费操作
			_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 1000)
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
	
	// 应该有成功的操作
	assert.Greater(t, successCount, 0)
}

// TestConcurrentDeadlockPrevention 测试死锁预防
func TestConcurrentDeadlockPrevention(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 1000000)
	
	// 创建两个测试团队
	team1 := createTestTeam(db, t, "团队1", owner.Id, 100000)
	team2 := createTestTeam(db, t, "团队2", owner.Id, 100000)
	
	// 创建测试成员
	member := createTestUser(db, t, "member", 100000)
	createTestTeamMember(db, t, team1.Id, member.Id, 2, 0)
	createTestTeamMember(db, t, team2.Id, member.Id, 2, 0)
	
	// 并发操作不同团队（可能导致死锁）
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	// 操作团队1
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := model.ConsumeTeamQuota(member.Id, team1.Id, 1000)
			results <- err
		}()
	}
	
	// 操作团队2
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := model.ConsumeTeamQuota(member.Id, team2.Id, 1000)
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
	
	// 应该有成功的操作
	assert.Greater(t, successCount, 0)
}

// TestConcurrentErrorHandling 测试并发错误处理
func TestConcurrentErrorHandling(t *testing.T) {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	db := testutils.SetupTestDB(t, config)
	
	// 创建测试用户
	owner := createTestUser(db, t, "owner", 100000)
	
	// 创建测试团队（额度较少）
	team := createTestTeam(db, t, "测试团队", owner.Id, 50000)
	
	// 创建测试成员
	member := createTestUser(db, t, "member", 100000)
	createTestTeamMember(db, t, team.Id, member.Id, 2, 0)
	
	// 并发消费超过团队额度的请求
	var wg sync.WaitGroup
	results := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := model.ConsumeTeamQuota(member.Id, team.Id, 10000)
			results <- err
		}()
	}
	
	wg.Wait()
	close(results)
	
	// 检查结果
	successCount := 0
	errorCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}
	
	// 应该有一些成功和一些失败
	assert.Greater(t, successCount, 0)
	assert.Greater(t, errorCount, 0)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err := db.First(&finalTeam, team.Id).Error
	require.NoError(t, err)
	assert.LessOrEqual(t, finalTeam.UsedQuota, 50000)
}

// BenchmarkConcurrentQuotaConsumption 并发额度消费性能测试
func BenchmarkConcurrentQuotaConsumption(b *testing.B) {
	db := testutils.SetupTestDB(&testing.T{}, nil)
	
	// 创建测试数据
	owner := createTestUser(db, &testing.T{}, "owner", 100000000)
	member := createTestUser(db, &testing.T{}, "member", 100000000)
	team := createTestTeam(db, &testing.T{}, "测试团队", owner.Id, 50000000)
	createTestTeamMember(db, &testing.T{}, team.Id, member.Id, 2, 0)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = model.ConsumeTeamQuota(member.Id, team.Id, 1000)
		}
	})
}

// BenchmarkConcurrentQuotaAllocation 并发额度分配性能测试
func BenchmarkConcurrentQuotaAllocation(b *testing.B) {
	db := testutils.SetupTestDB(&testing.T{}, nil)
	
	// 创建测试数据
	owner := createTestUser(db, &testing.T{}, "owner", 100000000)
	team := createTestTeam(db, &testing.T{}, "测试团队", owner.Id, 0)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = model.AllocateQuotaToTeam(owner.Id, team.Id, 1000, false)
		}
	})
}

