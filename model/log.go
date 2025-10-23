package model

import (
	"context"
	"errors"
	"fmt"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/common/utils"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Log struct {
	Id               int                                `json:"id"`
	UserId           int                                `json:"user_id" gorm:"index"`
	TeamId           int                                `json:"team_id" gorm:"default:0;index"` // 所属团队ID（0表示非团队消费）
	CreatedAt        int64                              `json:"created_at" gorm:"bigint;index:idx_created_at_type"`
	Type             int                                `json:"type" gorm:"index:idx_created_at_type"`
	Content          string                             `json:"content"`
	Username         string                             `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	TokenName        string                             `json:"token_name" gorm:"index;default:''"`
	ModelName        string                             `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota            int                                `json:"quota" gorm:"default:0"`
	PromptTokens     int                                `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int                                `json:"completion_tokens" gorm:"default:0"`
	ChannelId        int                                `json:"channel_id" gorm:"index"`
	RequestTime      int                                `json:"request_time" gorm:"default:0"`
	IsStream         bool                               `json:"is_stream" gorm:"default:false"`
	SourceIp         string                             `json:"source_ip" gorm:"default:''"`
	Metadata         datatypes.JSONType[map[string]any] `json:"metadata" gorm:"type:json"`

	Channel *Channel `json:"channel" gorm:"foreignKey:Id;references:ChannelId"`
}

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
)

func RecordQuotaLog(userId int, logType int, quota int, ip string, content string) {
	if logType == LogTypeConsume && !config.LogConsumeEnabled {
		return
	}
	username, _ := CacheGetUsername(userId)
	log := &Log{
		UserId:    userId,
		Username:  username,
		Quota:     quota,
		CreatedAt: utils.GetTimestamp(),
		Type:      logType,
		SourceIp:  ip,
		Content:   content,
	}
	err := DB.Create(log).Error
	if err != nil {
		logger.SysError("failed to record log: " + err.Error())
	}
}

func RecordLog(userId int, logType int, content string) {
	if logType == LogTypeConsume && !config.LogConsumeEnabled {
		return
	}
	username, _ := CacheGetUsername(userId)

	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: utils.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	err := DB.Create(log).Error
	if err != nil {
		logger.SysError("failed to record log: " + err.Error())
	}
}

func RecordConsumeLog(
	ctx context.Context,
	userId int,
	channelId int,
	promptTokens int,
	completionTokens int,
	modelName string,
	tokenName string,
	quota int,
	content string,
	requestTime int,
	isStream bool,
	metadata map[string]any,
	sourceIp string) {
	RecordConsumeLogWithTeam(ctx, userId, channelId, promptTokens, completionTokens, modelName, tokenName, quota, content, requestTime, isStream, metadata, sourceIp, 0)
}

func RecordConsumeLogWithTeam(
	ctx context.Context,
	userId int,
	channelId int,
	promptTokens int,
	completionTokens int,
	modelName string,
	tokenName string,
	quota int,
	content string,
	requestTime int,
	isStream bool,
	metadata map[string]any,
	sourceIp string,
	teamId int,
	) {
	logger.LogInfo(ctx, fmt.Sprintf("record consume log: userId=%d, channelId=%d, promptTokens=%d, completionTokens=%d, modelName=%s, tokenName=%s, quota=%d, content=%s ,sourceIp=%s", userId, channelId, promptTokens, completionTokens, modelName, tokenName, quota, content, sourceIp))
	if !config.LogConsumeEnabled {
		return
	}

	username, _ := CacheGetUsername(userId)

	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        utils.GetTimestamp(),
		Type:             LogTypeConsume,
		Content:          content,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TokenName:        tokenName,
		ModelName:        modelName,
		Quota:            quota,
		ChannelId:        channelId,
		RequestTime:      requestTime,
		IsStream:         isStream,
		SourceIp:         sourceIp,
		TeamId:           teamId,

	}

	if metadata != nil {
		log.Metadata = datatypes.NewJSONType(metadata)
	}

	err := DB.Create(log).Error
	if err != nil {
		logger.LogError(ctx, "failed to record log: "+err.Error())
	}
}

type LogsListParams struct {
	PaginationParams
	LogType        int    `form:"log_type"`
	StartTimestamp int64  `form:"start_timestamp"`
	EndTimestamp   int64  `form:"end_timestamp"`
	ModelName      string `form:"model_name"`
	Username       string `form:"username"`
	TokenName      string `form:"token_name"`
	ChannelId      int    `form:"channel_id"`
	SourceIp       string `form:"source_ip"`
	// 团队相关字段
	UserId         int    `form:"user_id"`
	TeamId         int    `form:"team_id"`
}

var allowedLogsOrderFields = map[string]bool{
	"created_at": true,
	"channel_id": true,
	"user_id":    true,
	"token_name": true,
	"model_name": true,
	"type":       true,
	"source_ip":  true,
}

func GetLogsList(params *LogsListParams) (*DataResult[Log], error) {
	var tx *gorm.DB
	var logs []*Log

	tx = DB.Preload("Channel", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	})

	if params.LogType != LogTypeUnknown {
		tx = tx.Where("type = ?", params.LogType)
	}
	if params.ModelName != "" {
		tx = tx.Where("model_name = ?", params.ModelName)
	}
	if params.Username != "" {
		tx = tx.Where("username = ?", params.Username)
	}
	if params.TokenName != "" {
		tx = tx.Where("token_name = ?", params.TokenName)
	}
	if params.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", params.EndTimestamp)
	}
	if params.ChannelId != 0 {
		tx = tx.Where("channel_id = ?", params.ChannelId)
	}
	if params.SourceIp != "" {
		tx = tx.Where("source_ip = ?", params.SourceIp)
	}

	return PaginateAndOrder[Log](tx, &params.PaginationParams, &logs, allowedLogsOrderFields)
}

func GetUserLogsList(userId int, params *LogsListParams) (*DataResult[Log], error) {
	return GetUserLogsListWithTeamFilter(userId, -1, params)
}

// GetUserLogsListWithTeamFilter 按团队筛选查询用户日志列表
// teamId: -1=全部, 0=个人空间, >0=特定团队
func GetUserLogsListWithTeamFilter(userId int, teamId int, params *LogsListParams) (*DataResult[Log], error) {
	var logs []*Log

	tx := DB.Where("user_id = ?", userId).Omit("id")

	// 根据 teamId 添加筛选条件
	if teamId >= 0 {
		tx = tx.Where("team_id = ?", teamId)
	}
	// teamId = -1 时不添加 team_id 筛选，显示所有空间的数据

	if params.LogType != LogTypeUnknown {
		tx = tx.Where("type = ?", params.LogType)
	}
	if params.ModelName != "" {
		tx = tx.Where("model_name = ?", params.ModelName)
	}
	if params.TokenName != "" {
		tx = tx.Where("token_name = ?", params.TokenName)
	}
	if params.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", params.EndTimestamp)
	}

	return PaginateAndOrder[Log](tx, &params.PaginationParams, &logs, allowedLogsOrderFields)
}

// GetLogsByContext 按上下文查询 Log 列表
func GetLogsByContext(contextType string, contextId int, userId int, params *LogsListParams) (*DataResult[Log], error) {
	var logs []*Log

	var tx *gorm.DB
	
	if contextType == "team" {
		// 检查用户是否为团队成员
		if !IsTeamMember(contextId, userId) {
			return nil, errors.New("无权限访问该团队的日志")
		}
		// 查询团队日志
		tx = DB.Where("team_id = ?", contextId).Omit("id")
	} else {
		// 个人上下文：查询个人日志
		tx = DB.Where("user_id = ? AND team_id = 0", userId).Omit("id")
	}

	if params.LogType != LogTypeUnknown {
		tx = tx.Where("type = ?", params.LogType)
	}
	if params.ModelName != "" {
		tx = tx.Where("model_name = ?", params.ModelName)
	}
	if params.TokenName != "" {
		tx = tx.Where("token_name = ?", params.TokenName)
	}
	if params.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", params.StartTimestamp)
	}
	if params.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", params.EndTimestamp)
	}

	return PaginateAndOrder[Log](tx, &params.PaginationParams, &logs, allowedLogsOrderFields)
}

// InsertWithContext 创建 Log 并绑定上下文
func (log *Log) InsertWithContext(contextType string, contextId int) error {
	if contextType == "team" {
		log.TeamId = contextId
	} else {
		log.TeamId = 0 // 个人日志
	}
	return DB.Create(log).Error
}

func SearchAllLogs(keyword string) (logs []*Log, err error) {
	err = DB.Where("type = ? or content LIKE ?", keyword, keyword+"%").Order("id desc").Limit(config.MaxRecentItems).Find(&logs).Error
	return logs, err
}

func SearchUserLogs(userId int, keyword string) (logs []*Log, err error) {
	err = DB.Where("user_id = ? and type = ?", userId, keyword).Order("id desc").Limit(config.MaxRecentItems).Omit("id").Find(&logs).Error
	return logs, err
}

func SumUsedQuota(startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) (quota int) {
	tx := DB.Table("logs").Select(assembleSumSelectStr("quota"))
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&quota)
	return quota
}

func DeleteOldLog(targetTimestamp int64) (int64, error) {
	result := DB.Where("type = ? AND created_at < ?", LogTypeConsume, targetTimestamp).Delete(&Log{})
	return result.RowsAffected, result.Error
}

type LogStatistic struct {
	Date             string `gorm:"column:date"`
	RequestCount     int64  `gorm:"column:request_count"`
	Quota            int64  `gorm:"column:quota"`
	PromptTokens     int64  `gorm:"column:prompt_tokens"`
	CompletionTokens int64  `gorm:"column:completion_tokens"`
	RequestTime      int64  `gorm:"column:request_time"`
}

type LogStatisticGroupModel struct {
	LogStatistic
	ModelName string `gorm:"column:model_name"`
}

type LogStatisticGroupChannel struct {
	LogStatistic
	Channel string `gorm:"column:channel"`
}
