package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"one-api/model"
)

// AssertTeamEqual 断言团队相等
func AssertTeamEqual(t *testing.T, expected, actual *model.Team, msgAndArgs ...interface{}) {
	require.NotNil(t, expected, "Expected team should not be nil")
	require.NotNil(t, actual, "Actual team should not be nil")
	
	assert.Equal(t, expected.Name, actual.Name, "Team name should be equal")
	assert.Equal(t, expected.OwnerId, actual.OwnerId, "Team owner ID should be equal")
	assert.Equal(t, expected.Quota, actual.Quota, "Team quota should be equal")
	assert.Equal(t, expected.UsedQuota, actual.UsedQuota, "Team used quota should be equal")
	assert.Equal(t, expected.UnlimitedQuota, actual.UnlimitedQuota, "Team unlimited quota should be equal")
	assert.Equal(t, expected.Status, actual.Status, "Team status should be equal")
	assert.NotEmpty(t, actual.InviteCode, "Team invite code should not be empty")
	assert.Greater(t, actual.CreatedTime, int64(0), "Team created time should be greater than 0")
	assert.GreaterOrEqual(t, actual.UpdatedTime, actual.CreatedTime, "Team updated time should be greater than or equal to created time")
}

// AssertTeamMemberEqual 断言团队成员相等
func AssertTeamMemberEqual(t *testing.T, expected, actual *model.TeamMember, msgAndArgs ...interface{}) {
	require.NotNil(t, expected, "Expected team member should not be nil")
	require.NotNil(t, actual, "Actual team member should not be nil")
	
	assert.Equal(t, expected.TeamId, actual.TeamId, "Team member team ID should be equal")
	assert.Equal(t, expected.UserId, actual.UserId, "Team member user ID should be equal")
	assert.Equal(t, expected.Role, actual.Role, "Team member role should be equal")
	assert.Equal(t, expected.MaxQuota, actual.MaxQuota, "Team member max quota should be equal")
	assert.Equal(t, expected.UsedQuota, actual.UsedQuota, "Team member used quota should be equal")
	assert.Equal(t, expected.Status, actual.Status, "Team member status should be equal")
	assert.Greater(t, actual.JoinedTime, int64(0), "Team member joined time should be greater than 0")
}

// AssertUserEqual 断言用户相等
func AssertUserEqual(t *testing.T, expected, actual *model.User, msgAndArgs ...interface{}) {
	require.NotNil(t, expected, "Expected user should not be nil")
	require.NotNil(t, actual, "Actual user should not be nil")
	
	assert.Equal(t, expected.Username, actual.Username, "User username should be equal")
	assert.Equal(t, expected.DisplayName, actual.DisplayName, "User display name should be equal")
	assert.Equal(t, expected.Email, actual.Email, "User email should be equal")
	assert.Equal(t, expected.Quota, actual.Quota, "User quota should be equal")
	assert.Equal(t, expected.Status, actual.Status, "User status should be equal")
}

// AssertQuotaDeducted 断言额度被正确扣除
func AssertQuotaDeducted(t *testing.T, originalQuota, currentQuota, deductedAmount int, msgAndArgs ...interface{}) {
	expectedQuota := originalQuota - deductedAmount
	assert.Equal(t, expectedQuota, currentQuota, 
		"Quota should be deducted correctly. Original: %d, Current: %d, Deducted: %d, Expected: %d",
		originalQuota, currentQuota, deductedAmount, expectedQuota)
}

// AssertQuotaAdded 断言额度被正确增加
func AssertQuotaAdded(t *testing.T, originalQuota, currentQuota, addedAmount int, msgAndArgs ...interface{}) {
	expectedQuota := originalQuota + addedAmount
	assert.Equal(t, expectedQuota, currentQuota,
		"Quota should be added correctly. Original: %d, Current: %d, Added: %d, Expected: %d",
		originalQuota, currentQuota, addedAmount, expectedQuota)
}

// AssertAPIResponseSuccess 断言 API 响应成功
func AssertAPIResponseSuccess(t *testing.T, recorder *httptest.ResponseRecorder, msgAndArgs ...interface{}) {
	assert.Equal(t, http.StatusOK, recorder.Code, "HTTP status should be 200 OK")
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	success, exists := response["success"]
	require.True(t, exists, "Response should contain 'success' field")
	assert.True(t, success.(bool), "Response success should be true")
}

// AssertAPIResponseError 断言 API 响应错误
func AssertAPIResponseError(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, msgAndArgs ...interface{}) {
	assert.Equal(t, expectedStatus, recorder.Code, "HTTP status should match expected error status")
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	success, exists := response["success"]
	require.True(t, exists, "Response should contain 'success' field")
	assert.False(t, success.(bool), "Response success should be false")
	
	message, exists := response["message"]
	require.True(t, exists, "Response should contain 'message' field")
	assert.NotEmpty(t, message, "Error message should not be empty")
}

// AssertAPIResponseData 断言 API 响应包含数据
func AssertAPIResponseData(t *testing.T, recorder *httptest.ResponseRecorder, msgAndArgs ...interface{}) {
	AssertAPIResponseSuccess(t, recorder)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	data, exists := response["data"]
	require.True(t, exists, "Response should contain 'data' field")
	assert.NotNil(t, data, "Response data should not be nil")
}

// AssertTeamInResponse 断言响应中包含团队数据
func AssertTeamInResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedTeam *model.Team, msgAndArgs ...interface{}) {
	AssertAPIResponseData(t, recorder)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	data := response["data"]
	require.NotNil(t, data, "Response data should not be nil")
	
	// 如果是单个团队对象
	if teamData, ok := data.(map[string]interface{}); ok {
		assert.Equal(t, expectedTeam.Name, teamData["name"], "Team name in response should match")
		assert.Equal(t, float64(expectedTeam.OwnerId), teamData["owner_id"], "Team owner ID in response should match")
		assert.Equal(t, float64(expectedTeam.Quota), teamData["quota"], "Team quota in response should match")
		return
	}
	
	// 如果是团队列表
	if teamList, ok := data.(map[string]interface{}); ok {
		if teams, exists := teamList["data"]; exists {
			if teamArray, ok := teams.([]interface{}); ok {
				require.Greater(t, len(teamArray), 0, "Team list should not be empty")
				
				firstTeam := teamArray[0].(map[string]interface{})
				assert.Equal(t, expectedTeam.Name, firstTeam["name"], "First team name in response should match")
				assert.Equal(t, float64(expectedTeam.OwnerId), firstTeam["owner_id"], "First team owner ID in response should match")
			}
		}
	}
}

// AssertTeamMemberInResponse 断言响应中包含团队成员数据
func AssertTeamMemberInResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedMember *model.TeamMember, msgAndArgs ...interface{}) {
	AssertAPIResponseData(t, recorder)
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	data := response["data"]
	require.NotNil(t, data, "Response data should not be nil")
	
	// 如果是单个成员对象
	if memberData, ok := data.(map[string]interface{}); ok {
		assert.Equal(t, float64(expectedMember.TeamId), memberData["team_id"], "Member team ID in response should match")
		assert.Equal(t, float64(expectedMember.UserId), memberData["user_id"], "Member user ID in response should match")
		assert.Equal(t, float64(expectedMember.Role), memberData["role"], "Member role in response should match")
		return
	}
	
	// 如果是成员列表
	if memberList, ok := data.(map[string]interface{}); ok {
		if members, exists := memberList["data"]; exists {
			if memberArray, ok := members.([]interface{}); ok {
				require.Greater(t, len(memberArray), 0, "Member list should not be empty")
				
				firstMember := memberArray[0].(map[string]interface{})
				assert.Equal(t, float64(expectedMember.TeamId), firstMember["team_id"], "First member team ID in response should match")
				assert.Equal(t, float64(expectedMember.UserId), firstMember["user_id"], "First member user ID in response should match")
			}
		}
	}
}

// AssertPermissionDenied 断言权限被拒绝
func AssertPermissionDenied(t *testing.T, recorder *httptest.ResponseRecorder, msgAndArgs ...interface{}) {
	AssertAPIResponseError(t, recorder, http.StatusOK) // 注意：项目使用 200 状态码返回错误
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	message := response["message"].(string)
	assert.Contains(t, message, "权限", "Error message should contain permission-related text")
}

// AssertValidationError 断言验证错误
func AssertValidationError(t *testing.T, recorder *httptest.ResponseRecorder, msgAndArgs ...interface{}) {
	AssertAPIResponseError(t, recorder, http.StatusOK) // 注意：项目使用 200 状态码返回错误
	
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")
	
	message := response["message"].(string)
	assert.Contains(t, message, "参数错误", "Error message should contain parameter error text")
}

// AssertDatabaseRecordExists 断言数据库记录存在
func AssertDatabaseRecordExists(t *testing.T, db *gorm.DB, table interface{}, condition string, args ...interface{}) {
	var count int64
	err := db.Model(table).Where(condition, args...).Count(&count).Error
	require.NoError(t, err, "Should be able to query database")
	assert.Greater(t, count, int64(0), "Database record should exist")
}

// AssertDatabaseRecordNotExists 断言数据库记录不存在
func AssertDatabaseRecordNotExists(t *testing.T, db *gorm.DB, table interface{}, condition string, args ...interface{}) {
	var count int64
	err := db.Model(table).Where(condition, args...).Count(&count).Error
	require.NoError(t, err, "Should be able to query database")
	assert.Equal(t, int64(0), count, "Database record should not exist")
}

// AssertDatabaseRecordCount 断言数据库记录数量
func AssertDatabaseRecordCount(t *testing.T, db *gorm.DB, table interface{}, expectedCount int64, condition string, args ...interface{}) {
	var count int64
	err := db.Model(table).Where(condition, args...).Count(&count).Error
	require.NoError(t, err, "Should be able to query database")
	assert.Equal(t, expectedCount, count, "Database record count should match expected")
}
