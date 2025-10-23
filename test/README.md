# Testify 测试框架使用指南

本项目使用 Testify 搭建了完整的测试框架，支持 Team 相关功能的单元测试、集成测试和并发测试。

## 目录结构

```
test/
├── testutils/                    # 测试工具和辅助函数
│   ├── setup.go                 # 测试环境设置
│   ├── db_adapter.go            # 数据库适配器接口
│   ├── factory.go               # 测试数据工厂
│   ├── assertions.go            # 自定义断言函数
│   ├── mock_helpers.go          # Mock 辅助函数
│   └── mocks/                   # Mock 对象
│       ├── mock_db.go           # Mock 数据库
│       └── mock_model.go        # Mock Model 方法
├── unit/                        # 单元测试
│   ├── model/                   # Model 层测试
│   │   ├── team_test.go         # Team 模型测试
│   │   ├── team_member_test.go  # TeamMember 模型测试
│   │   └── team_quota_test.go   # 团队额度测试
│   └── controller/              # Controller 层测试
│       └── team_controller_test.go
└── integration/                 # 集成测试
    ├── team_api_test.go         # API 端到端测试
    └── team_concurrent_test.go  # 并发测试
```

## 快速开始

### 1. 运行所有测试

```bash
# 使用 Taskfile（推荐）
task test:all

# 或直接使用 go test
go test -v ./test/...
```

### 2. 运行特定类型的测试

```bash
# 运行单元测试
task test:unit

# 运行集成测试
task test:integration

# 运行 Model 层测试
task test:model

# 运行 Controller 层测试
task test:controller
```

### 3. 运行性能测试

```bash
# 运行基准测试
task test:benchmark

# 运行竞态条件检测
task test:race

# 生成覆盖率报告
task test:coverage
```

## 数据库配置

### 环境变量

通过 `TEST_DB` 环境变量控制测试数据库类型：

```bash
# 使用 Mock 数据库（默认）
export TEST_DB=mock
task test:all

# 使用 SQLite 数据库
export TEST_DB=sqlite
task test:all

# 使用 MySQL 数据库（预留）
export TEST_DB=mysql
task test:all
```

### 数据库适配器

框架支持多种数据库适配器：

- **MockDBAdapter**: 使用 Mock 对象，快速执行
- **SQLiteDBAdapter**: 使用内存 SQLite，真实数据库操作
- **MySQLDBAdapter**: 使用真实 MySQL（预留接口）

## 测试工具使用

### 1. 测试数据工厂

```go
// 创建工厂实例
factory := testutils.NewTestDataFactory()

// 创建测试用户
userFactory := factory.NewUserFactory()
user := userFactory.CreateUserWithQuota(1000000)

// 创建测试团队
teamFactory := factory.NewTeamFactory()
team := teamFactory.CreateTeamWithQuota(ownerId, 500000)

// 创建测试团队成员
memberFactory := factory.NewTeamMemberFactory()
member := memberFactory.CreateTeamMember(teamId, userId, role)
```

### 2. 测试场景预设

```go
// 创建测试场景
scenario := factory.NewTestScenario()

// Owner 场景
owner, team, member := scenario.OwnerScenario()

// 成员场景
owner, member, team, ownerMember, memberRecord := scenario.MemberScenario()

// 多团队场景
owner, teams, members := scenario.MultiTeamScenario()

// 无限额度场景
owner, member, team, ownerMember, memberRecord := scenario.UnlimitedQuotaScenario()
```

### 3. 自定义断言

```go
// 断言团队相等
testutils.AssertTeamEqual(t, expected, actual)

// 断言团队成员相等
testutils.AssertTeamMemberEqual(t, expected, actual)

// 断言额度被正确扣除
testutils.AssertQuotaDeducted(t, originalQuota, currentQuota, deductedAmount)

// 断言 API 响应成功
testutils.AssertAPIResponseSuccess(t, recorder)

// 断言权限被拒绝
testutils.AssertPermissionDenied(t, recorder)
```

### 4. Mock 对象使用

```go
// 创建 Mock 对象
mockModel := mocks.NewMockModelInterface()

// 设置期望
mockModel.On("GetTeamById", 1).Return(team, nil)
mockModel.On("CreateTeam", mock.AnythingOfType("*model.Team")).Return(nil)

// 执行测试
// ...

// 验证期望
mockModel.AssertExpectations(t)
```

## 测试套件组织

### 使用 Testify Suite

```go
type TeamModelSuite struct {
    suite.Suite
    db      *gorm.DB
    factory *testutils.TestDataFactory
}

func (suite *TeamModelSuite) SetupSuite() {
    suite.db = testutils.SetupTestDB(suite.T(), nil)
    suite.factory = testutils.NewTestDataFactory()
}

func (suite *TeamModelSuite) SetupTest() {
    testutils.ResetTestDB(suite.T(), suite.db)
}

func TestTeamModelSuite(t *testing.T) {
    suite.Run(t, new(TeamModelSuite))
}
```

### 测试生命周期

1. **SetupSuite**: 测试套件初始化（只执行一次）
2. **SetupTest**: 每个测试前的设置
3. **TestXXX**: 具体测试方法
4. **TearDownTest**: 每个测试后的清理
5. **TearDownSuite**: 测试套件清理（只执行一次）

## 并发测试

### 基本并发测试

```go
func (suite *TeamConcurrentSuite) TestConcurrentQuotaConsumption() {
    var wg sync.WaitGroup
    results := make(chan error, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, _, err := model.ConsumeTeamQuota(memberId, teamId, 1000)
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
    
    assert.GreaterOrEqual(suite.T(), successCount, 5)
}
```

### 性能基准测试

```go
func BenchmarkConcurrentQuotaConsumption(b *testing.B) {
    // 设置测试数据
    // ...
    
    b.ResetTimer()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, _, _ = model.ConsumeTeamQuota(memberId, teamId, 1000)
        }
    })
}
```

## API 集成测试

### HTTP 测试

```go
func (suite *TeamAPISuite) TestCreateTeamAPI() {
    // 准备请求数据
    teamData := map[string]interface{}{
        "name": "测试团队",
    }
    
    jsonData, _ := json.Marshal(teamData)
    req, _ := http.NewRequest("POST", "/api/team/", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-User-ID", "1")
    
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    // 验证响应
    testutils.AssertAPIResponseSuccess(suite.T(), w)
}
```

## 最佳实践

### 1. 测试命名

- 测试方法名应该描述测试场景：`TestCreateTeamWithValidInput`
- 使用 `TestXXX_Success` 和 `TestXXX_Failure` 区分成功和失败场景
- 使用 `TestXXX_EdgeCase` 标记边界测试

### 2. 测试数据

- 使用工厂模式创建测试数据
- 避免硬编码测试数据
- 使用场景预设简化复杂测试设置

### 3. Mock 使用

- 只 Mock 外部依赖，不 Mock 被测试的代码
- 设置明确的期望和验证
- 使用 `mock.AnythingOfType` 处理复杂类型

### 4. 断言

- 使用 `require` 进行关键断言（失败时立即停止）
- 使用 `assert` 进行一般断言（继续执行）
- 使用自定义断言提高可读性

### 5. 测试隔离

- 每个测试应该独立运行
- 使用 `SetupTest` 重置测试环境
- 避免测试间的数据依赖

## 故障排除

### 常见问题

1. **测试失败但代码正常**
   - 检查 Mock 期望设置
   - 验证测试数据是否正确
   - 确认测试环境配置

2. **并发测试不稳定**
   - 检查数据竞争
   - 使用 `-race` 标志检测竞态条件
   - 确保测试数据隔离

3. **Mock 验证失败**
   - 检查方法调用次数
   - 验证参数类型和值
   - 确认 Mock 对象生命周期

### 调试技巧

```bash
# 运行单个测试
go test -v -run TestCreateTeam ./test/unit/model/

# 运行测试并显示详细输出
go test -v -test.v ./test/...

# 运行测试并生成覆盖率报告
go test -v -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out -o coverage.html
```

## 扩展指南

### 添加新的测试类型

1. 在相应目录创建测试文件
2. 使用现有的测试工具和辅助函数
3. 遵循命名和结构约定
4. 添加相应的 Taskfile 任务

### 添加新的 Mock 对象

1. 在 `testutils/mocks/` 目录创建 Mock 文件
2. 实现相应的接口方法
3. 提供辅助函数简化 Mock 设置
4. 更新文档和示例

### 添加新的断言函数

1. 在 `testutils/assertions.go` 添加新函数
2. 使用 `assert` 和 `require` 包
3. 提供清晰的错误消息
4. 添加使用示例

## 贡献指南

1. 遵循现有的代码风格和结构
2. 为新功能添加相应的测试
3. 更新文档和示例
4. 确保所有测试通过
5. 提交前运行完整的测试套件
