package testutils

import (
	"testing"

	"gorm.io/gorm"
)

// DBAdapter 数据库适配器接口
type DBAdapter interface {
	// GetDB 获取数据库实例
	GetDB() *gorm.DB
	
	// Setup 设置测试环境
	Setup(t *testing.T) error
	
	// Teardown 清理测试环境
	Teardown() error
	
	// Reset 重置数据库状态
	Reset() error
	
	// IsMock 是否为 Mock 数据库
	IsMock() bool
}

// MockDBAdapter Mock 数据库适配器
type MockDBAdapter struct {
	db *gorm.DB
}

// NewMockDBAdapter 创建 Mock 数据库适配器
func NewMockDBAdapter() *MockDBAdapter {
	return &MockDBAdapter{}
}

// GetDB 获取 Mock 数据库实例
func (m *MockDBAdapter) GetDB() *gorm.DB {
	return m.db
}

// Setup 设置 Mock 数据库
func (m *MockDBAdapter) Setup(t *testing.T) error {
	// TODO: 实现真正的 Mock 数据库
	// 目前暂时使用 SQLite 作为占位符
	config := &TestConfig{
		DBType:      "sqlite",
		LogLevel:    logger.Silent,
		AutoMigrate: true,
	}
	
	db := SetupTestDB(t, config)
	m.db = db
	return nil
}

// Teardown 清理 Mock 数据库
func (m *MockDBAdapter) Teardown() error {
	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err == nil {
			sqlDB.Close()
		}
		m.db = nil
	}
	return nil
}

// Reset 重置 Mock 数据库
func (m *MockDBAdapter) Reset() error {
	if m.db == nil {
		return nil
	}
	
	// 清空所有表
	tables := []interface{}{
		&model.TeamMember{},
		&model.Team{},
		&model.Log{},
		&model.User{},
	}

	for _, table := range tables {
		if err := m.db.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// IsMock 返回 true，表示这是 Mock 数据库
func (m *MockDBAdapter) IsMock() bool {
	return true
}

// SQLiteDBAdapter SQLite 数据库适配器
type SQLiteDBAdapter struct {
	db *gorm.DB
}

// NewSQLiteDBAdapter 创建 SQLite 数据库适配器
func NewSQLiteDBAdapter() *SQLiteDBAdapter {
	return &SQLiteDBAdapter{}
}

// GetDB 获取 SQLite 数据库实例
func (s *SQLiteDBAdapter) GetDB() *gorm.DB {
	return s.db
}

// Setup 设置 SQLite 数据库
func (s *SQLiteDBAdapter) Setup(t *testing.T) error {
	config := &TestConfig{
		DBType:      "sqlite",
		LogLevel:    logger.Silent,
		AutoMigrate: true,
	}
	
	db := SetupTestDB(t, config)
	s.db = db
	return nil
}

// Teardown 清理 SQLite 数据库
func (s *SQLiteDBAdapter) Teardown() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			sqlDB.Close()
		}
		s.db = nil
	}
	return nil
}

// Reset 重置 SQLite 数据库
func (s *SQLiteDBAdapter) Reset() error {
	if s.db == nil {
		return nil
	}
	
	// 清空所有表
	tables := []interface{}{
		&model.TeamMember{},
		&model.Team{},
		&model.Log{},
		&model.User{},
	}

	for _, table := range tables {
		if err := s.db.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// IsMock 返回 false，表示这是真实数据库
func (s *SQLiteDBAdapter) IsMock() bool {
	return false
}

// MySQLDBAdapter MySQL 数据库适配器（预留）
type MySQLDBAdapter struct {
	db *gorm.DB
}

// NewMySQLDBAdapter 创建 MySQL 数据库适配器
func NewMySQLDBAdapter() *MySQLDBAdapter {
	return &MySQLDBAdapter{}
}

// GetDB 获取 MySQL 数据库实例
func (m *MySQLDBAdapter) GetDB() *gorm.DB {
	return m.db
}

// Setup 设置 MySQL 数据库
func (m *MySQLDBAdapter) Setup(t *testing.T) error {
	// TODO: 实现 MySQL 连接
	t.Skip("MySQL test database not implemented yet")
	return nil
}

// Teardown 清理 MySQL 数据库
func (m *MySQLDBAdapter) Teardown() error {
	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err == nil {
			sqlDB.Close()
		}
		m.db = nil
	}
	return nil
}

// Reset 重置 MySQL 数据库
func (m *MySQLDBAdapter) Reset() error {
	// TODO: 实现 MySQL 重置逻辑
	return nil
}

// IsMock 返回 false，表示这是真实数据库
func (m *MySQLDBAdapter) IsMock() bool {
	return false
}

// CreateDBAdapter 根据配置创建数据库适配器
func CreateDBAdapter(dbType string) DBAdapter {
	switch dbType {
	case "sqlite":
		return NewSQLiteDBAdapter()
	case "mysql":
		return NewMySQLDBAdapter()
	case "mock":
		fallthrough
	default:
		return NewMockDBAdapter()
	}
}

// GetDBAdapterFromEnv 从环境变量创建数据库适配器
func GetDBAdapterFromEnv() DBAdapter {
	dbType := os.Getenv("TEST_DB")
	if dbType == "" {
		dbType = "mock"
	}
	return CreateDBAdapter(dbType)
}
