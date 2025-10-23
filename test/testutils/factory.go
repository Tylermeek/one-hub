package testutils

import (
	"fmt"
	"math/rand"
	"time"

	"one-api/model"
)

// TestDataFactory 测试数据工厂
type TestDataFactory struct {
	rand *rand.Rand
}

// NewTestDataFactory 创建测试数据工厂
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// UserFactory 用户数据工厂
type UserFactory struct {
	factory *TestDataFactory
}

// NewUserFactory 创建用户工厂
func (f *TestDataFactory) NewUserFactory() *UserFactory {
	return &UserFactory{factory: f}
}

// CreateUser 创建测试用户
func (uf *UserFactory) CreateUser() *model.User {
	return &model.User{
		Username:    uf.generateUsername(),
		Password:    "test_password_123",
		DisplayName: uf.generateDisplayName(),
		Email:       uf.generateEmail(),
		Quota:       uf.factory.rand.Intn(1000000) + 100000, // 100k-1.1M
		Status:      1,
	}
}

// CreateUserWithQuota 创建指定额度的用户
func (uf *UserFactory) CreateUserWithQuota(quota int) *model.User {
	user := uf.CreateUser()
	user.Quota = quota
	return user
}

// CreateUserWithUsername 创建指定用户名的用户
func (uf *UserFactory) CreateUserWithUsername(username string) *model.User {
	user := uf.CreateUser()
	user.Username = username
	return user
}

// generateUsername 生成随机用户名
func (uf *UserFactory) generateUsername() string {
	return fmt.Sprintf("testuser_%d_%d", time.Now().Unix(), uf.factory.rand.Intn(10000))
}

// generateDisplayName 生成随机显示名
func (uf *UserFactory) generateDisplayName() string {
	names := []string{"测试用户", "Test User", "用户", "User", "测试", "Test"}
	return fmt.Sprintf("%s_%d", names[uf.factory.rand.Intn(len(names))], uf.factory.rand.Intn(1000))
}

// generateEmail 生成随机邮箱
func (uf *UserFactory) generateEmail() string {
	return fmt.Sprintf("test_%d_%d@example.com", time.Now().Unix(), uf.factory.rand.Intn(10000))
}

// TeamFactory 团队数据工厂
type TeamFactory struct {
	factory *TestDataFactory
}

// NewTeamFactory 创建团队工厂
func (f *TestDataFactory) NewTeamFactory() *TeamFactory {
	return &TeamFactory{factory: f}
}

// CreateTeam 创建测试团队
func (tf *TeamFactory) CreateTeam(ownerId int) *model.Team {
	return &model.Team{
		Name:           tf.generateTeamName(),
		OwnerId:        ownerId,
		Quota:          tf.factory.rand.Intn(500000) + 50000, // 50k-550k
		UsedQuota:      0,
		UnlimitedQuota: false,
		Status:         1,
		InviteCode:     tf.generateInviteCode(),
		CreatedTime:    time.Now().Unix(),
		UpdatedTime:    time.Now().Unix(),
	}
}

// CreateTeamWithName 创建指定名称的团队
func (tf *TeamFactory) CreateTeamWithName(name string, ownerId int) *model.Team {
	team := tf.CreateTeam(ownerId)
	team.Name = name
	return team
}

// CreateTeamWithQuota 创建指定额度的团队
func (tf *TeamFactory) CreateTeamWithQuota(ownerId int, quota int) *model.Team {
	team := tf.CreateTeam(ownerId)
	team.Quota = quota
	return team
}

// CreateUnlimitedTeam 创建无限额度团队
func (tf *TeamFactory) CreateUnlimitedTeam(ownerId int) *model.Team {
	team := tf.CreateTeam(ownerId)
	team.Quota = 0
	team.UnlimitedQuota = true
	return team
}

// generateTeamName 生成随机团队名
func (tf *TeamFactory) generateTeamName() string {
	prefixes := []string{"测试团队", "Test Team", "开发团队", "Dev Team", "项目组", "Project"}
	return fmt.Sprintf("%s_%d", prefixes[tf.factory.rand.Intn(len(prefixes))], tf.factory.rand.Intn(1000))
}

// generateInviteCode 生成随机邀请码
func (tf *TeamFactory) generateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[tf.factory.rand.Intn(len(charset))]
	}
	return string(code)
}

// TeamMemberFactory 团队成员数据工厂
type TeamMemberFactory struct {
	factory *TestDataFactory
}

// NewTeamMemberFactory 创建团队成员工厂
func (f *TestDataFactory) NewTeamMemberFactory() *TeamMemberFactory {
	return &TeamMemberFactory{factory: f}
}

// CreateTeamMember 创建测试团队成员
func (tmf *TeamMemberFactory) CreateTeamMember(teamId, userId int) *model.TeamMember {
	return &model.TeamMember{
		TeamId:     teamId,
		UserId:     userId,
		Role:       2, // 默认普通成员
		MaxQuota:   0, // 默认无限制
		UsedQuota:  0,
		Status:     1,
		JoinedTime: time.Now().Unix(),
	}
}

// CreateAdminMember 创建管理员成员
func (tmf *TeamMemberFactory) CreateAdminMember(teamId, userId int) *model.TeamMember {
	member := tmf.CreateTeamMember(teamId, userId)
	member.Role = 1 // 管理员
	return member
}

// CreateMemberWithQuotaLimit 创建有额度限制的成员
func (tmf *TeamMemberFactory) CreateMemberWithQuotaLimit(teamId, userId int, maxQuota int) *model.TeamMember {
	member := tmf.CreateTeamMember(teamId, userId)
	member.MaxQuota = maxQuota
	return member
}

// TestScenario 测试场景预设
type TestScenario struct {
	factory *TestDataFactory
}

// NewTestScenario 创建测试场景
func (f *TestDataFactory) NewTestScenario() *TestScenario {
	return &TestScenario{factory: f}
}

// OwnerScenario 创建 Owner 场景
func (ts *TestScenario) OwnerScenario() (*model.User, *model.Team, *model.TeamMember) {
	userFactory := ts.factory.NewUserFactory()
	teamFactory := ts.factory.NewTeamFactory()
	memberFactory := ts.factory.NewTeamMemberFactory()

	owner := userFactory.CreateUserWithQuota(1000000)
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	member := memberFactory.CreateAdminMember(team.Id, owner.Id)

	return owner, team, member
}

// MemberScenario 创建普通成员场景
func (ts *TestScenario) MemberScenario() (*model.User, *model.User, *model.Team, *model.TeamMember, *model.TeamMember) {
	userFactory := ts.factory.NewUserFactory()
	teamFactory := ts.factory.NewTeamFactory()
	memberFactory := ts.factory.NewTeamMemberFactory()

	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	team := teamFactory.CreateTeamWithQuota(owner.Id, 500000)
	
	ownerMember := memberFactory.CreateAdminMember(team.Id, owner.Id)
	memberRecord := memberFactory.CreateTeamMember(team.Id, member.Id)

	return owner, member, team, ownerMember, memberRecord
}

// MultiTeamScenario 创建多团队场景
func (ts *TestScenario) MultiTeamScenario() (*model.User, []*model.Team, []*model.TeamMember) {
	userFactory := ts.factory.NewUserFactory()
	teamFactory := ts.factory.NewTeamFactory()
	memberFactory := ts.factory.NewTeamMemberFactory()

	owner := userFactory.CreateUserWithQuota(2000000)
	
	var teams []*model.Team
	var members []*model.TeamMember

	// 创建 3 个团队
	for i := 0; i < 3; i++ {
		team := teamFactory.CreateTeamWithQuota(owner.Id, 300000)
		teams = append(teams, team)
		
		member := memberFactory.CreateAdminMember(team.Id, owner.Id)
		members = append(members, member)
	}

	return owner, teams, members
}

// UnlimitedQuotaScenario 创建无限额度场景
func (ts *TestScenario) UnlimitedQuotaScenario() (*model.User, *model.User, *model.Team, *model.TeamMember, *model.TeamMember) {
	userFactory := ts.factory.NewUserFactory()
	teamFactory := ts.factory.NewTeamFactory()
	memberFactory := ts.factory.NewTeamMemberFactory()

	owner := userFactory.CreateUserWithQuota(1000000)
	member := userFactory.CreateUserWithQuota(200000)
	team := teamFactory.CreateUnlimitedTeam(owner.Id)
	
	ownerMember := memberFactory.CreateAdminMember(team.Id, owner.Id)
	memberRecord := memberFactory.CreateTeamMember(team.Id, member.Id)

	return owner, member, team, ownerMember, memberRecord
}

// QuotaExhaustedScenario 创建额度耗尽场景
func (ts *TestScenario) QuotaExhaustedScenario() (*model.User, *model.User, *model.Team, *model.TeamMember, *model.TeamMember) {
	userFactory := ts.factory.NewUserFactory()
	teamFactory := ts.factory.NewTeamFactory()
	memberFactory := ts.factory.NewTeamMemberFactory()

	owner := userFactory.CreateUserWithQuota(100000)
	member := userFactory.CreateUserWithQuota(50000)
	team := teamFactory.CreateTeamWithQuota(owner.Id, 10000) // 很少的团队额度
	
	ownerMember := memberFactory.CreateAdminMember(team.Id, owner.Id)
	memberRecord := memberFactory.CreateMemberWithQuotaLimit(team.Id, member.Id, 5000) // 成员额度限制

	return owner, member, team, ownerMember, memberRecord
}

// 全局工厂实例
var globalFactory = NewTestDataFactory()

// GetFactory 获取全局工厂实例
func GetFactory() *TestDataFactory {
	return globalFactory
}
