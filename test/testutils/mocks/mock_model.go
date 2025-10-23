package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
	"one-api/model"
)

// MockTeamModel Mock Team 模型
type MockTeamModel struct {
	mock.Mock
}

// NewMockTeamModel 创建 Mock Team 模型
func NewMockTeamModel() *MockTeamModel {
	return &MockTeamModel{}
}

// Insert Mock Team Insert 方法
func (m *MockTeamModel) Insert() error {
	args := m.Called()
	return args.Error(0)
}

// Update Mock Team Update 方法
func (m *MockTeamModel) Update() error {
	args := m.Called()
	return args.Error(0)
}

// Delete Mock Team Delete 方法
func (m *MockTeamModel) Delete() error {
	args := m.Called()
	return args.Error(0)
}

// MockTeamMemberModel Mock TeamMember 模型
type MockTeamMemberModel struct {
	mock.Mock
}

// NewMockTeamMemberModel 创建 Mock TeamMember 模型
func NewMockTeamMemberModel() *MockTeamMemberModel {
	return &MockTeamMemberModel{}
}

// Insert Mock TeamMember Insert 方法
func (m *MockTeamMemberModel) Insert() error {
	args := m.Called()
	return args.Error(0)
}

// Update Mock TeamMember Update 方法
func (m *MockTeamMemberModel) Update() error {
	args := m.Called()
	return args.Error(0)
}

// Delete Mock TeamMember Delete 方法
func (m *MockTeamMemberModel) Delete() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserModel Mock User 模型
type MockUserModel struct {
	mock.Mock
}

// NewMockUserModel 创建 Mock User 模型
func NewMockUserModel() *MockUserModel {
	return &MockUserModel{}
}

// Insert Mock User Insert 方法
func (m *MockUserModel) Insert(quota int) error {
	args := m.Called(quota)
	return args.Error(0)
}

// Update Mock User Update 方法
func (m *MockUserModel) Update() error {
	args := m.Called()
	return args.Error(0)
}

// Delete Mock User Delete 方法
func (m *MockUserModel) Delete() error {
	args := m.Called()
	return args.Error(0)
}

// MockModelInterface Mock 模型接口
type MockModelInterface struct {
	mock.Mock
}

// NewMockModelInterface 创建 Mock 模型接口
func NewMockModelInterface() *MockModelInterface {
	return &MockModelInterface{}
}

// GetTeamById Mock GetTeamById 函数
func (m *MockModelInterface) GetTeamById(id int) (*model.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Team), args.Error(1)
}

// GetUserTeams Mock GetUserTeams 函数
func (m *MockModelInterface) GetUserTeams(userId int, params *model.PaginationParams) (*model.DataResult[model.Team], error) {
	args := m.Called(userId, params)
	return args.Get(0).(*model.DataResult[model.Team]), args.Error(1)
}

// GetTeamByInviteCode Mock GetTeamByInviteCode 函数
func (m *MockModelInterface) GetTeamByInviteCode(inviteCode string) (*model.Team, error) {
	args := m.Called(inviteCode)
	return args.Get(0).(*model.Team), args.Error(1)
}

// IsTeamOwner Mock IsTeamOwner 函数
func (m *MockModelInterface) IsTeamOwner(teamId, userId int) bool {
	args := m.Called(teamId, userId)
	return args.Bool(0)
}

// IsTeamMember Mock IsTeamMember 函数
func (m *MockModelInterface) IsTeamMember(teamId, userId int) bool {
	args := m.Called(teamId, userId)
	return args.Bool(0)
}

// GetTeamMember Mock GetTeamMember 函数
func (m *MockModelInterface) GetTeamMember(teamId, userId int) (*model.TeamMember, error) {
	args := m.Called(teamId, userId)
	return args.Get(0).(*model.TeamMember), args.Error(1)
}

// GetTeamMembersList Mock GetTeamMembersList 函数
func (m *MockModelInterface) GetTeamMembersList(teamId int, params *model.SearchTeamMemberParams) (*model.DataResult[model.TeamMember], error) {
	args := m.Called(teamId, params)
	return args.Get(0).(*model.DataResult[model.TeamMember]), args.Error(1)
}

// GetMemberAvailableQuota Mock GetMemberAvailableQuota 函数
func (m *MockModelInterface) GetMemberAvailableQuota(teamId, userId int) (int, bool, error) {
	args := m.Called(teamId, userId)
	return args.Int(0), args.Bool(1), args.Error(2)
}

// AllocateQuotaToTeam Mock AllocateQuotaToTeam 函数
func (m *MockModelInterface) AllocateQuotaToTeam(ownerId, teamId, quota int, unlimited bool) error {
	args := m.Called(ownerId, teamId, quota, unlimited)
	return args.Error(0)
}

// ConsumeTeamQuota Mock ConsumeTeamQuota 函数
func (m *MockModelInterface) ConsumeTeamQuota(userId, teamId, quota int) (int, int, error) {
	args := m.Called(userId, teamId, quota)
	return args.Int(0), args.Int(1), args.Error(2)
}

// UpdateTeamMemberQuota Mock UpdateTeamMemberQuota 函数
func (m *MockModelInterface) UpdateTeamMemberQuota(teamId, userId, maxQuota int) error {
	args := m.Called(teamId, userId, maxQuota)
	return args.Error(0)
}

// DeleteTeamMember Mock DeleteTeamMember 函数
func (m *MockModelInterface) DeleteTeamMember(teamId, userId int) error {
	args := m.Called(teamId, userId)
	return args.Error(0)
}

// GetUserById Mock GetUserById 函数
func (m *MockModelInterface) GetUserById(id int, selectPassword bool) (*model.User, error) {
	args := m.Called(id, selectPassword)
	return args.Get(0).(*model.User), args.Error(1)
}

// GetUsersList Mock GetUsersList 函数
func (m *MockModelInterface) GetUsersList(params *model.GenericParams) (*model.DataResult[model.User], error) {
	args := m.Called(params)
	return args.Get(0).(*model.DataResult[model.User]), args.Error(1)
}

// RecordLog Mock RecordLog 函数
func (m *MockModelInterface) RecordLog(userId int, logType int, message string) {
	m.Called(userId, logType, message)
}

// GetUserLogsList Mock GetUserLogsList 函数
func (m *MockModelInterface) GetUserLogsList(userId int, params *model.LogsListParams) (*model.DataResult[model.Log], error) {
	args := m.Called(userId, params)
	return args.Get(0).(*model.DataResult[model.Log]), args.Error(1)
}

// CreateTeam Mock CreateTeam 函数
func (m *MockModelInterface) CreateTeam(team *model.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

// CreateTeamMember Mock CreateTeamMember 函数
func (m *MockModelInterface) CreateTeamMember(member *model.TeamMember) error {
	args := m.Called(member)
	return args.Error(0)
}

// UpdateTeam Mock UpdateTeam 函数
func (m *MockModelInterface) UpdateTeam(team *model.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

// DeleteTeam Mock DeleteTeam 函数
func (m *MockModelInterface) DeleteTeam(team *model.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

// MockModelHelper Mock 模型辅助函数
type MockModelHelper struct{}

// NewMockModelHelper 创建 Mock 模型辅助函数
func NewMockModelHelper() *MockModelHelper {
	return &MockModelHelper{}
}

// SetupMockTeamModelSuccess 设置团队模型操作成功的 Mock
func (h *MockModelHelper) SetupMockTeamModelSuccess(mockModel *MockModelInterface, team *model.Team) {
	mockModel.On("GetTeamById", team.Id).Return(team, nil)
	mockModel.On("GetTeamByInviteCode", team.InviteCode).Return(team, nil)
	mockModel.On("IsTeamOwner", team.Id, team.OwnerId).Return(true)
	mockModel.On("IsTeamMember", team.Id, mock.AnythingOfType("int")).Return(true)
}

// SetupMockTeamModelFailure 设置团队模型操作失败的 Mock
func (h *MockModelHelper) SetupMockTeamModelFailure(mockModel *MockModelInterface, teamId int, errMsg string) {
	mockModel.On("GetTeamById", teamId).Return((*model.Team)(nil), errors.New(errMsg))
	mockModel.On("IsTeamOwner", teamId, mock.AnythingOfType("int")).Return(false)
	mockModel.On("IsTeamMember", teamId, mock.AnythingOfType("int")).Return(false)
}

// SetupMockTeamMemberModelSuccess 设置团队成员模型操作成功的 Mock
func (h *MockModelHelper) SetupMockTeamMemberModelSuccess(mockModel *MockModelInterface, member *model.TeamMember) {
	mockModel.On("GetTeamMember", member.TeamId, member.UserId).Return(member, nil)
	mockModel.On("GetMemberAvailableQuota", member.TeamId, member.UserId).Return(50000, false, nil)
	mockModel.On("UpdateTeamMemberQuota", member.TeamId, member.UserId, mock.AnythingOfType("int")).Return(nil)
	mockModel.On("DeleteTeamMember", member.TeamId, member.UserId).Return(nil)
}

// SetupMockQuotaModelSuccess 设置额度模型操作成功的 Mock
func (h *MockModelHelper) SetupMockQuotaModelSuccess(mockModel *MockModelInterface) {
	mockModel.On("AllocateQuotaToTeam", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("bool")).Return(nil)
	mockModel.On("ConsumeTeamQuota", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1000, 0, nil)
	mockModel.On("GetMemberAvailableQuota", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(50000, false, nil)
}

// SetupMockUserModelSuccess 设置用户模型操作成功的 Mock
func (h *MockModelHelper) SetupMockUserModelSuccess(mockModel *MockModelInterface, user *model.User) {
	mockModel.On("GetUserById", user.Id, mock.AnythingOfType("bool")).Return(user, nil)
	mockModel.On("GetUsersList", mock.AnythingOfType("*model.GenericParams")).Return(&model.DataResult[model.User]{
		Data: &[]*model.User{user},
		TotalCount: 1,
		Page: 1,
		Size: 10,
	}, nil)
}

// SetupMockLogModelSuccess 设置日志模型操作成功的 Mock
func (h *MockModelHelper) SetupMockLogModelSuccess(mockModel *MockModelInterface) {
	mockModel.On("RecordLog", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("string")).Return()
	mockModel.On("GetUserLogsList", mock.AnythingOfType("int"), mock.AnythingOfType("*model.LogsListParams")).Return(&model.DataResult[model.Log]{
		Data: &[]*model.Log{},
		TotalCount: 0,
		Page: 1,
		Size: 10,
	}, nil)
}
