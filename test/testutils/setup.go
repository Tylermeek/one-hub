package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"one-api/common/logger"
	"one-api/model"
)

// TestConfig 测试配置
type TestConfig struct {
	DBType     string // mock, sqlite, mysql
	LogLevel   gormLogger.LogLevel
	AutoMigrate bool
}

// DefaultTestConfig 默认测试配置
func DefaultTestConfig() *TestConfig {
	dbType := os.Getenv("TEST_DB")
	if dbType == "" {
		dbType = "mock" // 默认使用 Mock
	}

	return &TestConfig{
		DBType:      dbType,
		LogLevel:    gormLogger.Silent, // 测试时静默日志
		AutoMigrate: true,
	}
}

// SetupTestDB 设置测试数据库
func SetupTestDB(t *testing.T, config *TestConfig) *gorm.DB {
	if config == nil {
		config = DefaultTestConfig()
	}

	// 初始化 logger（如果还没有初始化）
	if logger.Logger == nil {
		logger.SetupLogger()
	}

	var db *gorm.DB
	var err error

	switch config.DBType {
	case "sqlite":
		db, err = setupSQLiteDB(t, config)
	case "mysql":
		db, err = setupMySQLDB(t, config)
	case "mock":
		fallthrough
	default:
		db, err = setupSQLiteDB(t, config) // Mock 暂时使用 SQLite，后续会替换
	}

	require.NoError(t, err, "Failed to setup test database")

	if config.AutoMigrate {
		err = db.AutoMigrate(
			&model.User{},
			&model.Team{},
			&model.TeamMember{},
			&model.Token{},
			&model.Log{},
		)
		require.NoError(t, err, "Failed to auto migrate test database")
	}

	return db
}

// setupSQLiteDB 设置 SQLite 测试数据库
func setupSQLiteDB(t *testing.T, config *TestConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormLogger.Default.LogMode(config.LogLevel),
	})
	if err != nil {
		return nil, err
	}

	// 设置全局 DB 实例（用于 model 包）
	model.SetDB(db)
	
	return db, nil
}

// setupMySQLDB 设置 MySQL 测试数据库（预留）
func setupMySQLDB(t *testing.T, config *TestConfig) (*gorm.DB, error) {
	// TODO: 实现 MySQL 测试数据库连接
	// 这里预留接口，后续可以扩展
	t.Skip("MySQL test database not implemented yet")
	return nil, nil
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// ResetTestDB 重置测试数据库（清空所有表）
func ResetTestDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// 按依赖关系顺序删除数据
	tables := []interface{}{
		&model.TeamMember{},
		&model.Token{},
		&model.Team{},
		&model.Log{},
		&model.User{},
	}

	for _, table := range tables {
		err := db.Unscoped().Where("1 = 1").Delete(table).Error
		if err != nil {
			t.Logf("Warning: Failed to reset table %T: %v", table, err)
		}
	}
}

// SetupTestEnvironment 设置完整的测试环境
func SetupTestEnvironment(t *testing.T, config *TestConfig) *gorm.DB {
	db := SetupTestDB(t, config)
	
	// 设置 Gin 测试模式
	// gin.SetMode(gin.TestMode)
	
	return db
}

// TeardownTestEnvironment 清理测试环境
func TeardownTestEnvironment(t *testing.T, db *gorm.DB) {
	CleanupTestDB(t, db)
}
