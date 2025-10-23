package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"one-api/model"
	"one-api/test/testutils"
)

// TeamConcurrentSuite 团队并发测试套件
type TeamConcurrentSuite struct {
	suite.Suite
	db      *gorm.DB
	factory *testutils.TestDataFactory
}

// SetupSuite 测试套件初始化
func (suite *TeamConcurrentSuite) SetupSuite() {
	config := &testutils.TestConfig{
		DBType:      "sqlite",
		LogLevel:    gormLogger.Silent,
		AutoMigrate: true,
	}
	suite.db = testutils.SetupTestDB(suite.T(), config)
	suite.factory = testutils.NewTestDataFactory()
}

// TearDownSuite 测试套件清理
func (suite *TeamConcurrentSuite) TearDownSuite() {
	testutils.CleanupTestDB(suite.T(), suite.db)
}

// SetupTest 每个测试前的设置
func (suite *TeamConcurrentSuite) SetupTest() {
	testutils.ResetTestDB(suite.T(), suite.db)
}

// TestConcurrentTeamCreation 测试并发创建团队
func (suite *TeamConcurrentSuite) TestConcurrentTeamCreation() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(10000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 并发创建团队
	var wg sync.WaitGroup
	results := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			team := teamFactory.CreateTeamWithName("并发团队"+string(rune(index)), owner.Id)
			err := team.Insert()
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
	
	// 所有团队创建都应该成功
	assert.Equal(suite.T(), 10, successCount)
	
	// 验证数据库中的团队数量
	var teamCount int64
	err = suite.db.Model(&model.Team{}).Count(&teamCount).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(10), teamCount)
}

// TestConcurrentQuotaAllocation 测试并发额度分配
func (suite *TeamConcurrentSuite) TestConcurrentQuotaAllocation() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户（大额度）
	owner := userFactory.CreateUserWithQuota(10000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 并发分配额度
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 10000, false)
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
	assert.Equal(suite.T(), 20, successCount)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err = suite.db.First(&finalTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 200000, finalTeam.Quota)
	
	// 验证管理员剩余额度
	var finalOwner model.User
	err = suite.db.First(&finalOwner, owner.Id).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 9800000, finalOwner.Quota)
}

// TestConcurrentQuotaConsumption 测试并发额度消费
func (suite *TeamConcurrentSuite) TestConcurrentQuotaConsumption() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建多个测试用户
	var members []*model.User
	for i := 0; i < 5; i++ {
		member := userFactory.CreateUserWithUsername("member" + string(rune(i)))
		err := member.Insert(0)
		require.NoError(suite.T(), err)
		members = append(members, member)
	}
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	for _, member := range members {
		teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
		err := teamMember.Insert()
		require.NoError(suite.T(), err)
	}
	
	// 并发消费额度
	var wg sync.WaitGroup
	results := make(chan error, 25)
	
	// 每个成员消费5次
	for _, member := range members {
		for i := 0; i < 5; i++ {
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
	assert.GreaterOrEqual(suite.T(), successCount, 30)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err = suite.db.First(&finalTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), finalTeam.UsedQuota, 500000)
}

// TestConcurrentMemberOperations 测试并发成员操作
func (suite *TeamConcurrentSuite) TestConcurrentMemberOperations() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建多个测试用户
	var members []*model.User
	for i := 0; i < 10; i++ {
		member := userFactory.CreateUserWithUsername("member" + string(rune(i)))
		err := member.Insert(0)
		require.NoError(suite.T(), err)
		members = append(members, member)
	}
	
	// 创建测试团队
	team := teamFactory.CreateTeam(owner.Id)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 并发添加成员
	var wg sync.WaitGroup
	results := make(chan error, 10)
	
	for _, member := range members {
		wg.Add(1)
		go func(member *model.User) {
			defer wg.Done()
			
			teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
			err := teamMember.Insert()
			results <- err
		}(member)
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
	
	// 所有成员添加都应该成功
	assert.Equal(suite.T(), 10, successCount)
	
	// 验证团队成员数量
	var memberCount int64
	err = suite.db.Model(&model.TeamMember{}).Where("team_id = ?", team.Id).Count(&memberCount).Error
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(10), memberCount)
}

// TestConcurrentMixedOperations 测试混合并发操作
func (suite *TeamConcurrentSuite) TestConcurrentMixedOperations() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(10000000)
	member := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	// 并发混合操作
	var wg sync.WaitGroup
	results := make(chan error, 30)
	
	// 分配额度操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 10000, false)
			results <- err
		}()
	}
	
	// 消费额度操作
	for i := 0; i < 15; i++ {
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
			err := suite.db.First(&team, team.Id).Error
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
	assert.GreaterOrEqual(suite.T(), successCount, 25)
}

// TestConcurrentWithContext 测试带上下文的并发操作
func (suite *TeamConcurrentSuite) TestConcurrentWithContext() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(100000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
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
	assert.Greater(suite.T(), successCount, 0)
}

// TestConcurrentDeadlockPrevention 测试死锁预防
func (suite *TeamConcurrentSuite) TestConcurrentDeadlockPrevention() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(100000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建两个测试团队
	team1 := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	team2 := teamFactory.CreateTeamWithQuota(owner.Id, 100000)
	err = team1.Insert()
	require.NoError(suite.T(), err)
	err = team2.Insert()
	require.NoError(suite.T(), err)
	
	// 创建测试成员
	teamMember1 := memberFactory.CreateTeamMember(team1.Id, member.Id)
	teamMember2 := memberFactory.CreateTeamMember(team2.Id, member.Id)
	err = teamMember1.Insert()
	require.NoError(suite.T(), err)
	err = teamMember2.Insert()
	require.NoError(suite.T(), err)
	
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
	assert.Greater(suite.T(), successCount, 0)
}

// TestConcurrentErrorHandling 测试并发错误处理
func (suite *TeamConcurrentSuite) TestConcurrentErrorHandling() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(100000)
	member := userFactory.CreateUserWithQuota(100000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队（额度较少）
	team := teamFactory.CreateTeamWithQuota(owner.Id, 50000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 创建团队成员
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
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
	assert.Greater(suite.T(), successCount, 0)
	assert.Greater(suite.T(), errorCount, 0)
	
	// 验证最终团队额度
	var finalTeam model.Team
	err = suite.db.First(&finalTeam, team.Id).Error
	require.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), finalTeam.UsedQuota, 50000)
}

// TestConcurrentRaceCondition 测试竞态条件
func (suite *TeamConcurrentSuite) TestConcurrentRaceCondition() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	// 并发分配和查询操作（可能产生竞态条件）
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	// 分配操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := model.AllocateQuotaToTeam(owner.Id, team.Id, 10000, false)
			results <- err
		}()
	}
	
	// 查询操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var team model.Team
			err := suite.db.First(&team, team.Id).Error
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
	
	// 所有操作都应该成功
	assert.Equal(suite.T(), 20, successCount)
}

// BenchmarkConcurrentQuotaConsumption 并发额度消费性能测试
func (suite *TeamConcurrentSuite) BenchmarkConcurrentQuotaConsumption(b *testing.B) {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	memberFactory := suite.factory.NewTeamMemberFactory()
	
	// 创建测试数据
	owner := userFactory.CreateUserWithQuota(100000000)
	member := userFactory.CreateUserWithQuota(100000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeamWithQuota(owner.Id, 50000000)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	teamMember := memberFactory.CreateTeamMember(team.Id, member.Id)
	err = teamMember.Insert()
	require.NoError(suite.T(), err)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = model.ConsumeTeamQuota(member.Id, team.Id, 1000)
		}
	})
}

// BenchmarkConcurrentQuotaAllocation 并发额度分配性能测试
func (suite *TeamConcurrentSuite) BenchmarkConcurrentQuotaAllocation(b *testing.B) {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试数据
	owner := userFactory.CreateUserWithQuota(100000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	team := teamFactory.CreateTeamWithQuota(owner.Id, 0)
	err = team.Insert()
	require.NoError(suite.T(), err)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = model.AllocateQuotaToTeam(owner.Id, team.Id, 1000, false)
		}
	})
}

// TestConcurrentContextSwitching 测试并发空间切换
func (suite *TeamConcurrentSuite) TestConcurrentContextSwitching() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeam(owner.Id)
	err = suite.db.Create(team).Error
	require.NoError(suite.T(), err)
	
	// 添加用户为团队成员
	member := &model.TeamMember{
		TeamId:     team.Id,
		UserId:     owner.Id,
		Role:       1, // 管理员
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
	err = member.Insert(0)
	require.NoError(suite.T(), err)
	
	// 并发测试空间切换
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 模拟切换到团队空间
			token := &model.Token{
				UserId:         owner.Id,
				Name:           "并发测试Token",
				Key:            "sk-concurrent-token-" + time.Now().Format("20060102150405"),
				RemainQuota:    1000,
				UnlimitedQuota: false,
			}
			
			err := token.InsertWithContext("team", team.Id)
			results <- err
		}()
		
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 模拟切换到用户空间
			token := &model.Token{
				UserId:         owner.Id,
				Name:           "并发测试个人Token",
				Key:            "sk-concurrent-user-token-" + time.Now().Format("20060102150405"),
				RemainQuota:    1000,
				UnlimitedQuota: false,
			}
			
			err := token.InsertWithContext("user", owner.Id)
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
	
	// 所有操作都应该成功
	assert.Equal(suite.T(), 20, successCount)
	
	// 验证Token数量
	var teamTokenCount int64
	suite.db.Model(&model.Token{}).Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
		owner.Id, "team", team.Id).Count(&teamTokenCount)
	assert.Equal(suite.T(), int64(10), teamTokenCount)
	
	var userTokenCount int64
	suite.db.Model(&model.Token{}).Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
		owner.Id, "user", owner.Id).Count(&userTokenCount)
	assert.Equal(suite.T(), int64(10), userTokenCount)
}

// TestConcurrentPermissionValidation 测试并发权限验证
func (suite *TeamConcurrentSuite) TestConcurrentPermissionValidation() {
	userFactory := suite.factory.NewUserFactory()
	teamFactory := suite.factory.NewTeamFactory()
	
	// 创建测试用户
	owner := userFactory.CreateUserWithQuota(1000000)
	err := owner.Insert(0)
	require.NoError(suite.T(), err)
	
	otherUser := userFactory.CreateUserWithQuota(100000)
		err = otherUser.Insert(0)
	require.NoError(suite.T(), err)
	
	// 创建测试团队
	team := teamFactory.CreateTeam(owner.Id)
	err = suite.db.Create(team).Error
	require.NoError(suite.T(), err)
	
	// 并发测试权限验证
	var wg sync.WaitGroup
	results := make(chan error, 20)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 有权限的用户创建Token
			token := &model.Token{
				UserId:         owner.Id,
				Name:           "有权限Token",
				Key:            "sk-authorized-token-" + time.Now().Format("20060102150405"),
				RemainQuota:    1000,
				UnlimitedQuota: false,
			}
			
			err := token.InsertWithContext("team", team.Id)
			results <- err
		}()
		
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// 无权限的用户尝试创建Token
			token := &model.Token{
				UserId:         otherUser.Id,
				Name:           "无权限Token",
				Key:            "sk-unauthorized-token-" + time.Now().Format("20060102150405"),
				RemainQuota:    1000,
				UnlimitedQuota: false,
			}
			
			err := token.InsertWithContext("team", team.Id)
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
	
	// 应该有10个成功（有权限）和10个失败（无权限）
	assert.Equal(suite.T(), 10, successCount)
	assert.Equal(suite.T(), 10, errorCount)
	
	// 验证只有有权限的用户创建了Token
	var tokenCount int64
	suite.db.Model(&model.Token{}).Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
		owner.Id, "team", team.Id).Count(&tokenCount)
	assert.Equal(suite.T(), int64(10), tokenCount)
	
	// 验证无权限的用户没有创建Token
	suite.db.Model(&model.Token{}).Where("user_id = ? AND owner_type = ? AND owner_id = ?", 
		otherUser.Id, "team", team.Id).Count(&tokenCount)
	assert.Equal(suite.T(), int64(0), tokenCount)
}

// TestTeamConcurrentSuite 运行测试套件
func TestTeamConcurrentSuite(t *testing.T) {
	suite.Run(t, new(TeamConcurrentSuite))
}
