package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB Mock 数据库接口
type MockDB struct {
	mock.Mock
}

// MockGormDB Mock GORM 数据库
type MockGormDB struct {
	mock.Mock
}

// NewMockDB 创建 Mock 数据库
func NewMockDB() *MockDB {
	return &MockDB{}
}

// NewMockGormDB 创建 Mock GORM 数据库
func NewMockGormDB() *MockGormDB {
	return &MockGormDB{}
}

// Create Mock Create 操作
func (m *MockGormDB) Create(value interface{}) *MockGormDB {
	args := m.Called(value)
	return args.Get(0).(*MockGormDB)
}

// First Mock First 操作
func (m *MockGormDB) First(dest interface{}, conds ...interface{}) *MockGormDB {
	args := m.Called(dest, conds)
	return args.Get(0).(*MockGormDB)
}

// Where Mock Where 操作
func (m *MockGormDB) Where(query interface{}, args ...interface{}) *MockGormDB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockGormDB)
}

// Update Mock Update 操作
func (m *MockGormDB) Update(column string, value interface{}) *MockGormDB {
	args := m.Called(column, value)
	return args.Get(0).(*MockGormDB)
}

// Updates Mock Updates 操作
func (m *MockGormDB) Updates(values interface{}) *MockGormDB {
	args := m.Called(values)
	return args.Get(0).(*MockGormDB)
}

// Delete Mock Delete 操作
func (m *MockGormDB) Delete(value interface{}, conds ...interface{}) *MockGormDB {
	args := m.Called(value, conds)
	return args.Get(0).(*MockGormDB)
}

// Count Mock Count 操作
func (m *MockGormDB) Count(count *int64) *MockGormDB {
	args := m.Called(count)
	return args.Get(0).(*MockGormDB)
}

// Preload Mock Preload 操作
func (m *MockGormDB) Preload(query string, args ...interface{}) *MockGormDB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockGormDB)
}

// Select Mock Select 操作
func (m *MockGormDB) Select(query interface{}, args ...interface{}) *MockGormDB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockGormDB)
}

// Order Mock Order 操作
func (m *MockGormDB) Order(value interface{}) *MockGormDB {
	args := m.Called(value)
	return args.Get(0).(*MockGormDB)
}

// Limit Mock Limit 操作
func (m *MockGormDB) Limit(limit int) *MockGormDB {
	args := m.Called(limit)
	return args.Get(0).(*MockGormDB)
}

// Offset Mock Offset 操作
func (m *MockGormDB) Offset(offset int) *MockGormDB {
	args := m.Called(offset)
	return args.Get(0).(*MockGormDB)
}

// Transaction Mock Transaction 操作
func (m *MockGormDB) Transaction(fc func(tx *gorm.DB) error, opts ...*gorm.Session) error {
	args := m.Called(fc, opts)
	return args.Error(0)
}

// Model Mock Model 操作
func (m *MockGormDB) Model(value interface{}) *MockGormDB {
	args := m.Called(value)
	return args.Get(0).(*MockGormDB)
}

// Unscoped Mock Unscoped 操作
func (m *MockGormDB) Unscoped() *MockGormDB {
	args := m.Called()
	return args.Get(0).(*MockGormDB)
}

// Pluck Mock Pluck 操作
func (m *MockGormDB) Pluck(column string, dest interface{}) *MockGormDB {
	args := m.Called(column, dest)
	return args.Get(0).(*MockGormDB)
}

// Error Mock Error 操作
func (m *MockGormDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

// RowsAffected Mock RowsAffected 操作
func (m *MockGormDB) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// AutoMigrate Mock AutoMigrate 操作
func (m *MockGormDB) AutoMigrate(dst ...interface{}) error {
	args := m.Called(dst)
	return args.Error(0)
}

// MockDBInterface 数据库操作接口
type MockDBInterface struct {
	mock.Mock
}

// NewMockDBInterface 创建 Mock 数据库接口
func NewMockDBInterface() *MockDBInterface {
	return &MockDBInterface{}
}

// CreateTeam Mock 创建团队
func (m *MockDBInterface) CreateTeam(team interface{}) error {
	args := m.Called(team)
	return args.Error(0)
}

// GetTeamById Mock 根据 ID 获取团队
func (m *MockDBInterface) GetTeamById(id int) (interface{}, error) {
	args := m.Called(id)
	return args.Get(0), args.Error(1)
}

// UpdateTeam Mock 更新团队
func (m *MockDBInterface) UpdateTeam(team interface{}) error {
	args := m.Called(team)
	return args.Error(0)
}

// DeleteTeam Mock 删除团队
func (m *MockDBInterface) DeleteTeam(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetUserTeams Mock 获取用户团队列表
func (m *MockDBInterface) GetUserTeams(userId int, params interface{}) (interface{}, error) {
	args := m.Called(userId, params)
	return args.Get(0), args.Error(1)
}

// CreateTeamMember Mock 创建团队成员
func (m *MockDBInterface) CreateTeamMember(member interface{}) error {
	args := m.Called(member)
	return args.Error(0)
}

// GetTeamMember Mock 获取团队成员
func (m *MockDBInterface) GetTeamMember(teamId, userId int) (interface{}, error) {
	args := m.Called(teamId, userId)
	return args.Get(0), args.Error(1)
}

// UpdateTeamMember Mock 更新团队成员
func (m *MockDBInterface) UpdateTeamMember(member interface{}) error {
	args := m.Called(member)
	return args.Error(0)
}

// DeleteTeamMember Mock 删除团队成员
func (m *MockDBInterface) DeleteTeamMember(teamId, userId int) error {
	args := m.Called(teamId, userId)
	return args.Error(0)
}

// GetTeamMembersList Mock 获取团队成员列表
func (m *MockDBInterface) GetTeamMembersList(teamId int, params interface{}) (interface{}, error) {
	args := m.Called(teamId, params)
	return args.Get(0), args.Error(1)
}

// IsTeamOwner Mock 检查是否为团队所有者
func (m *MockDBInterface) IsTeamOwner(teamId, userId int) bool {
	args := m.Called(teamId, userId)
	return args.Bool(0)
}

// IsTeamMember Mock 检查是否为团队成员
func (m *MockDBInterface) IsTeamMember(teamId, userId int) bool {
	args := m.Called(teamId, userId)
	return args.Bool(0)
}

// AllocateQuotaToTeam Mock 分配团队额度
func (m *MockDBInterface) AllocateQuotaToTeam(ownerId, teamId, quota int, unlimited bool) error {
	args := m.Called(ownerId, teamId, quota, unlimited)
	return args.Error(0)
}

// ConsumeTeamQuota Mock 消费团队额度
func (m *MockDBInterface) ConsumeTeamQuota(userId, teamId, quota int) (int, int, error) {
	args := m.Called(userId, teamId, quota)
	return args.Int(0), args.Int(1), args.Error(2)
}

// GetMemberAvailableQuota Mock 获取成员可用额度
func (m *MockDBInterface) GetMemberAvailableQuota(teamId, userId int) (int, bool, error) {
	args := m.Called(teamId, userId)
	return args.Int(0), args.Bool(1), args.Error(2)
}

// UpdateTeamMemberQuota Mock 更新成员额度限制
func (m *MockDBInterface) UpdateTeamMemberQuota(teamId, userId, maxQuota int) error {
	args := m.Called(teamId, userId, maxQuota)
	return args.Error(0)
}

// MockHelper Mock 辅助函数
type MockHelper struct{}

// NewMockHelper 创建 Mock 辅助函数
func NewMockHelper() *MockHelper {
	return &MockHelper{}
}

// SetupMockTeamSuccess 设置团队操作成功的 Mock
func (h *MockHelper) SetupMockTeamSuccess(mockDB *MockDBInterface, team interface{}) {
	mockDB.On("CreateTeam", team).Return(nil)
	mockDB.On("GetTeamById", mock.AnythingOfType("int")).Return(team, nil)
	mockDB.On("UpdateTeam", team).Return(nil)
	mockDB.On("DeleteTeam", mock.AnythingOfType("int")).Return(nil)
}

// SetupMockTeamFailure 设置团队操作失败的 Mock
func (h *MockHelper) SetupMockTeamFailure(mockDB *MockDBInterface, team interface{}, errMsg string) {
	mockDB.On("CreateTeam", team).Return(errors.New(errMsg))
	mockDB.On("GetTeamById", mock.AnythingOfType("int")).Return(nil, errors.New(errMsg))
	mockDB.On("UpdateTeam", team).Return(errors.New(errMsg))
	mockDB.On("DeleteTeam", mock.AnythingOfType("int")).Return(errors.New(errMsg))
}

// SetupMockTeamMemberSuccess 设置团队成员操作成功的 Mock
func (h *MockHelper) SetupMockTeamMemberSuccess(mockDB *MockDBInterface, member interface{}) {
	mockDB.On("CreateTeamMember", member).Return(nil)
	mockDB.On("GetTeamMember", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(member, nil)
	mockDB.On("UpdateTeamMember", member).Return(nil)
	mockDB.On("DeleteTeamMember", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(nil)
}

// SetupMockQuotaSuccess 设置额度操作成功的 Mock
func (h *MockHelper) SetupMockQuotaSuccess(mockDB *MockDBInterface) {
	mockDB.On("AllocateQuotaToTeam", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("bool")).Return(nil)
	mockDB.On("ConsumeTeamQuota", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(1000, 0, nil)
	mockDB.On("GetMemberAvailableQuota", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(50000, false, nil)
	mockDB.On("UpdateTeamMemberQuota", mock.AnythingOfType("int"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(nil)
}

// SetupMockPermissionSuccess 设置权限检查成功的 Mock
func (h *MockHelper) SetupMockPermissionSuccess(mockDB *MockDBInterface) {
	mockDB.On("IsTeamOwner", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true)
	mockDB.On("IsTeamMember", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true)
}

// SetupMockPermissionFailure 设置权限检查失败的 Mock
func (h *MockHelper) SetupMockPermissionFailure(mockDB *MockDBInterface) {
	mockDB.On("IsTeamOwner", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false)
	mockDB.On("IsTeamMember", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false)
}
