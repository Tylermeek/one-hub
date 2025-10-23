#!/bin/bash

# 团队功能测试运行脚本

set -e

echo "开始运行团队功能测试..."

# 设置测试环境
export GO_ENV=test
export DB_DSN=":memory:"

# 进入项目根目录
cd "$(dirname "$0")/../.."

echo "运行单元测试..."
go test -v ./test/team/team_quota_test.go -run "TestAllocateQuotaToTeam|TestConsumeTeamQuota|TestConsumeTeamQuotaWithMixedSource|TestConsumeTeamQuotaWithMemberLimit|TestConsumeTeamQuotaInsufficientUserQuota|TestConsumeTeamQuotaUnlimitedTeam|TestTeamQuotaEdgeCases"

echo "运行并发测试..."
go test -v ./test/team/team_quota_test.go -run "TestConcurrentTeamQuotaConsumption|TestConcurrentQuotaAllocation"

echo "运行集成测试..."
go test -v ./test/team/team_integration_test.go -run "TestCreateTeam|TestGetUserTeams|TestGetTeamDetail|TestAllocateTeamQuota|TestInviteMember|TestGetTeamMembers|TestUpdateMemberQuota|TestRemoveMember|TestDeleteTeam|TestPermissionControl"

echo "运行并发压力测试..."
go test -v ./test/team/team_concurrent_test.go -run "TestConcurrentTeamOperations|TestConcurrentMemberOperations|TestConcurrentQuotaConsumption|TestConcurrentMixedOperations|TestConcurrentWithContext|TestConcurrentDeadlockPrevention|TestConcurrentErrorHandling"

echo "运行性能基准测试..."
go test -v ./test/team/team_quota_test.go -bench="BenchmarkTeamQuotaConsumption" -benchmem
go test -v ./test/team/team_concurrent_test.go -bench="BenchmarkConcurrentQuotaConsumption|BenchmarkConcurrentQuotaAllocation" -benchmem

echo "运行所有测试..."
go test -v ./test/team/...

echo "测试完成！"

# 生成测试覆盖率报告
echo "生成测试覆盖率报告..."
go test -v ./test/team/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

echo "覆盖率报告已生成: coverage.html"

