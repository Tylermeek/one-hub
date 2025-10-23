package relay_util

import (
	"context"
	"errors"
	"math"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/model"
	"one-api/types"
	"time"

	"github.com/gin-gonic/gin"
)

type Quota struct {
	modelName        string
	promptTokens     int
	price            model.Price
	groupName        string
	isBackupGroup    bool // 新增字段记录是否使用备用分组
	backupGroupName  string
	groupRatio       float64
	inputRatio       float64
	outputRatio      float64
	preConsumedQuota int
	cacheQuota       int
	userId           int
	channelId        int
	tokenId          int
	HandelStatus     bool

	// 团队相关字段
	teamId        int    // 团队ID（0表示非团队消费）
	teamQuotaUsed int    // 预消费使用的团队额度
	userQuotaUsed int    // 预消费使用的用户个人额度
	contextType   string // 上下文类型："user" 或 "team"
	contextId     int    // 上下文ID

	startTime         time.Time
	firstResponseTime time.Time
	extraBillingData  map[string]ExtraBillingData
}

func NewQuota(c *gin.Context, modelName string, promptTokens int) *Quota {
	isBackupGroup := c.GetBool("is_backupGroup")

	quota := &Quota{
		modelName:     modelName,
		promptTokens:  promptTokens,
		userId:        c.GetInt("id"),
		channelId:     c.GetInt("channel_id"),
		tokenId:       c.GetInt("token_id"),
		HandelStatus:  false,
		isBackupGroup: isBackupGroup, // 记录是否使用备用分组
		// 从请求头获取上下文信息
		contextType:   c.GetString("context_type"),
		contextId:     c.GetInt("context_id"),
	}

	quota.price = *model.PricingInstance.GetPrice(quota.modelName)
	quota.groupName = c.GetString("token_group")
	quota.backupGroupName = c.GetString("token_backup_group")
	quota.groupRatio = c.GetFloat64("group_ratio") // 这里的倍率已经在 common.go 中正确设置了
	quota.inputRatio = quota.price.GetInput() * quota.groupRatio
	quota.outputRatio = quota.price.GetOutput() * quota.groupRatio

	return quota

}

func (q *Quota) PreQuotaConsumption() *types.OpenAIErrorWithStatusCode {
	// 计算预消费额度
	if q.price.Type == model.TimesPriceType {
		q.preConsumedQuota = int(1000 * q.inputRatio)
	} else if q.price.Input != 0 || q.price.Output != 0 {
		q.preConsumedQuota = int(float64(q.promptTokens)*q.inputRatio) + config.PreConsumedQuota
	}

	if q.preConsumedQuota == 0 {
		return nil
	}

	// 根据上下文类型进行额度消费
	if q.contextType == "team" && q.contextId > 0 {
		return q.consumeTeamContextQuota()
	} else {
		return q.consumeUserContextQuota()
	}
}

func (q *Quota) consumeTeamContextQuota() *types.OpenAIErrorWithStatusCode {
	// 检查是否为 Owner
	if model.IsTeamOwner(q.contextId, q.userId) {
		return q.consumeOwnerQuota()
	} else {
		return q.consumeMemberQuota()
	}
}

func (q *Quota) consumeOwnerQuota() *types.OpenAIErrorWithStatusCode {
	// Owner 直接扣除个人额度
	userQuota, err := model.CacheGetUserQuota(q.userId)
	if err != nil {
		return common.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota < q.preConsumedQuota {
		return common.ErrorWrapper(errors.New("owner quota is not enough"), "insufficient_owner_quota", http.StatusPaymentRequired)
	}

	err = model.CacheDecreaseUserQuota(q.userId, q.preConsumedQuota)
	if err != nil {
		return common.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}

	// 同时累计团队已用额度
	err = model.IncreaseTeamUsedQuota(q.contextId, q.preConsumedQuota)
	if err != nil {
		// 如果团队额度累计失败，回滚用户额度
		model.CacheIncreaseUserQuota(q.userId, q.preConsumedQuota)
		return common.ErrorWrapper(err, "increase_team_used_quota_failed", http.StatusInternalServerError)
	}

	// 标记为团队消费
	q.teamId = q.contextId
	q.userQuotaUsed = q.preConsumedQuota
	q.teamQuotaUsed = q.preConsumedQuota

	return nil
}

func (q *Quota) consumeMemberQuota() *types.OpenAIErrorWithStatusCode {
	// 获取团队信息
	team, err := model.GetTeamById(q.contextId)
	if err != nil {
		return common.ErrorWrapper(err, "get_team_failed", http.StatusInternalServerError)
	}

	if team.Status != 1 {
		return common.ErrorWrapper(errors.New("team is disabled"), "team_disabled", http.StatusForbidden)
	}

	// 获取 Owner 信息用于校验
	owner, err := model.GetUserById(team.OwnerId, false)
	if err != nil {
		return common.ErrorWrapper(err, "get_owner_failed", http.StatusInternalServerError)
	}

	// 校验 Owner 实际余额（所有团队都需要）
	ownerAvailableQuota := owner.Quota - owner.UsedQuota
	if ownerAvailableQuota < q.preConsumedQuota {
		return common.ErrorWrapper(errors.New("owner wallet insufficient"), "insufficient_owner_wallet", http.StatusPaymentRequired)
	}

	if team.UnlimitedQuota {
		// 无限额度团队：仅校验 Owner 余额，从 Owner 扣除
		ownerQuota, err := model.CacheGetUserQuota(team.OwnerId)
		if err != nil {
			return common.ErrorWrapper(err, "get_owner_quota_failed", http.StatusInternalServerError)
		}

		if ownerQuota < q.preConsumedQuota {
			return common.ErrorWrapper(errors.New("team owner quota is not enough"), "insufficient_team_quota", http.StatusPaymentRequired)
		}

		err = model.CacheDecreaseUserQuota(team.OwnerId, q.preConsumedQuota)
		if err != nil {
			return common.ErrorWrapper(err, "decrease_owner_quota_failed", http.StatusInternalServerError)
		}

		// 同时累计团队已用额度
		err = model.IncreaseTeamUsedQuota(q.contextId, q.preConsumedQuota)
		if err != nil {
			// 如果团队额度累计失败，回滚 Owner 额度
			model.CacheIncreaseUserQuota(team.OwnerId, q.preConsumedQuota)
			return common.ErrorWrapper(err, "increase_team_used_quota_failed", http.StatusInternalServerError)
		}

		q.teamId = q.contextId
		q.teamQuotaUsed = q.preConsumedQuota
		q.userQuotaUsed = 0
	} else {
		// 有限额度团队：双重校验
		// 第一层：团队上限校验
		availableTeamQuota := team.Quota - team.UsedQuota
		if availableTeamQuota < q.preConsumedQuota {
			return common.ErrorWrapper(errors.New("team quota limit exceeded"), "team_quota_limit_exceeded", http.StatusPaymentRequired)
		}

		// 第二层：Owner 实际余额校验（已在上面完成）

		// 扣除团队额度
		err = model.IncreaseTeamUsedQuota(q.contextId, q.preConsumedQuota)
		if err != nil {
			return common.ErrorWrapper(err, "increase_team_used_quota_failed", http.StatusInternalServerError)
		}

		q.teamId = q.contextId
		q.teamQuotaUsed = q.preConsumedQuota
		q.userQuotaUsed = 0
	}

	return nil
}

func (q *Quota) consumeUserContextQuota() *types.OpenAIErrorWithStatusCode {
	// 个人上下文，保持原有逻辑
	userQuota, err := model.CacheGetUserQuota(q.userId)
	if err != nil {
		return common.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota < q.preConsumedQuota {
		return common.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusPaymentRequired)
	}

	err = model.CacheDecreaseUserQuota(q.userId, q.preConsumedQuota)
	if err != nil {
		return common.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}

	if userQuota > 100*q.preConsumedQuota {
		q.preConsumedQuota = 0
	}

	if q.preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(q.tokenId, q.preConsumedQuota)
		if err != nil {
			return common.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
		q.HandelStatus = true
	}

	return nil
}

// 更新用户实时配额
func (q *Quota) UpdateUserRealtimeQuota(usage *types.UsageEvent, nowUsage *types.UsageEvent) error {
	usage.Merge(nowUsage)

	// 不开启Redis，则不更新实时配额
	if !config.RedisEnabled {
		return nil
	}

	promptTokens, completionTokens := q.getComputeTokensByUsageEvent(nowUsage)
	increaseQuota := q.GetTotalQuota(promptTokens, completionTokens, nil)

	cacheQuota, err := model.CacheIncreaseUserRealtimeQuota(q.userId, increaseQuota)
	if err != nil {
		return errors.New("error update user realtime quota cache: " + err.Error())
	}

	q.cacheQuota += increaseQuota
	userQuota, err := model.CacheGetUserQuota(q.userId)
	if err != nil {
		return errors.New("error get user quota cache: " + err.Error())
	}

	if cacheQuota >= int64(userQuota) {
		return errors.New("user quota is not enough")
	}

	return nil
}

func (q *Quota) completedQuotaConsumption(usage *types.Usage, tokenName string, isStream bool, sourceIp string, ctx context.Context) error {
	defer func() {
		if q.cacheQuota > 0 {
			model.CacheDecreaseUserRealtimeQuota(q.userId, q.cacheQuota)
		}
	}()

	quota := q.GetTotalQuotaByUsage(usage)

	if quota > 0 {
		quotaDelta := quota - q.preConsumedQuota
		
		if q.teamId > 0 {
			// 处理团队消费的额度差异
			if quotaDelta != 0 {
				if model.IsTeamOwner(q.teamId, q.userId) {
					// Owner: 调整个人额度和团队已用额度
					if quotaDelta > 0 {
						model.DecreaseUserQuota(q.userId, quotaDelta)
						model.IncreaseTeamUsedQuota(q.teamId, quotaDelta)
					} else {
						model.IncreaseUserQuota(q.userId, -quotaDelta)
						model.DecreaseTeamUsedQuota(q.teamId, -quotaDelta)
					}
					model.CacheUpdateUserQuota(q.userId)
				} else {
					// 成员: 调整团队/Owner 额度和团队已用额度
					team, _ := model.GetTeamById(q.teamId)
					if team != nil && team.UnlimitedQuota {
						// 无限额度团队，调整 Owner 个人额度和团队已用额度
						if quotaDelta > 0 {
							model.DecreaseUserQuota(team.OwnerId, quotaDelta)
							model.IncreaseTeamUsedQuota(q.teamId, quotaDelta)
						} else {
							model.IncreaseUserQuota(team.OwnerId, -quotaDelta)
							model.DecreaseTeamUsedQuota(q.teamId, -quotaDelta)
						}
						model.CacheUpdateUserQuota(team.OwnerId)
					} else {
						// 有限额度团队，调整团队已用额度
						if quotaDelta > 0 {
							model.IncreaseTeamUsedQuota(q.teamId, quotaDelta)
						} else {
							model.DecreaseTeamUsedQuota(q.teamId, -quotaDelta)
						}
					}
				}
			}
			
			// 记录团队消费日志
			model.RecordConsumeLogWithTeam(
				ctx,
				q.userId,
				q.channelId,
				usage.PromptTokens,
				usage.CompletionTokens,
				q.modelName,
				tokenName,
				quota,
				"",
				q.getRequestTime(),
				isStream,
				q.GetLogMeta(usage),
				sourceIp,
				q.teamId,
			)
		} else {
			// 个人消费，使用原有逻辑
			err := model.PostConsumeTokenQuota(q.tokenId, quotaDelta)
			if err != nil {
				return errors.New("error consuming token remain quota: " + err.Error())
			}
			err = model.CacheUpdateUserQuota(q.userId)
			if err != nil {
				return errors.New("error consuming token remain quota: " + err.Error())
			}
			
			model.RecordConsumeLog(
				ctx,
				q.userId,
				q.channelId,
				usage.PromptTokens,
				usage.CompletionTokens,
				q.modelName,
				tokenName,
				quota,
				"",
				q.getRequestTime(),
				isStream,
				q.GetLogMeta(usage),
				sourceIp,
			)
		}
		
		model.UpdateChannelUsedQuota(q.channelId, quota)
	}

	model.UpdateUserUsedQuotaAndRequestCount(q.userId, quota)

	return nil
}

func (q *Quota) Undo(c *gin.Context) {
	if q.preConsumedQuota == 0 {
		return
	}

	go func(ctx context.Context) {
		if q.teamId > 0 {
			// 团队上下文：统一对 team.used_quota 做反向冲正
			if model.IsTeamOwner(q.teamId, q.userId) {
				// Owner: 退还个人额度和团队已用额度
				model.IncreaseUserQuota(q.userId, q.preConsumedQuota)
				model.DecreaseTeamUsedQuota(q.teamId, q.preConsumedQuota)
				model.CacheUpdateUserQuota(q.userId)
			} else {
				// 成员: 退还团队/Owner 额度和团队已用额度
				team, _ := model.GetTeamById(q.teamId)
				if team != nil && team.UnlimitedQuota {
					model.IncreaseUserQuota(team.OwnerId, q.preConsumedQuota)
					model.DecreaseTeamUsedQuota(q.teamId, q.preConsumedQuota)
					model.CacheUpdateUserQuota(team.OwnerId)
				} else {
					model.DecreaseTeamUsedQuota(q.teamId, q.preConsumedQuota)
				}
			}
		} else if q.HandelStatus {
			// 个人上下文，原有逻辑
			tokenId := c.GetInt("token_id")
			err := model.PostConsumeTokenQuota(tokenId, -q.preConsumedQuota)
			if err != nil {
				logger.LogError(ctx, "error return pre-consumed quota: "+err.Error())
			}
		}
	}(c.Request.Context())
}

func (q *Quota) Consume(c *gin.Context, usage *types.Usage, isStream bool) {
	tokenName := c.GetString("token_name")
	q.startTime = c.GetTime("requestStartTime")
	// 如果没有报错，则消费配额
	go func(ctx context.Context) {
		err := q.completedQuotaConsumption(usage, tokenName, isStream, c.ClientIP(), ctx)
		if err != nil {
			logger.LogError(ctx, err.Error())
		}
	}(c.Request.Context())
}

func (q *Quota) GetInputRatio() float64 {
	return q.inputRatio
}

func (q *Quota) GetLogMeta(usage *types.Usage) map[string]any {
	meta := map[string]any{
		"group_name":        q.groupName,
		"backup_group_name": q.backupGroupName,
		"is_backup_group":   q.isBackupGroup, // 添加是否使用备用分组的标识
		"price_type":        q.price.Type,
		"group_ratio":       q.groupRatio,
		"input_ratio":       q.price.GetInput(),
		"output_ratio":      q.price.GetOutput(),
	}

	firstResponseTime := q.GetFirstResponseTime()
	if firstResponseTime > 0 {
		meta["first_response"] = firstResponseTime
	}

	if usage != nil {
		extraTokens := usage.GetExtraTokens()

		for key, value := range extraTokens {
			meta[key] = value
			extraRatio := q.price.GetExtraRatio(key)
			meta[key+"_ratio"] = extraRatio
		}
	}

	if q.extraBillingData != nil {
		meta["extra_billing"] = q.extraBillingData
	}

	return meta
}

func (q *Quota) getRequestTime() int {
	return int(time.Since(q.startTime).Milliseconds())
}

// 通过 token 数获取消费配额
func (q *Quota) GetTotalQuota(promptTokens, completionTokens int, extraBilling map[string]types.ExtraBilling) (quota int) {
	if q.price.Type == model.TimesPriceType {
		quota = int(1000 * q.inputRatio)
	} else {
		quota = int(math.Ceil((float64(promptTokens) * q.inputRatio) + (float64(completionTokens) * q.outputRatio)))
	}

	q.GetExtraBillingData(extraBilling)
	extraBillingQuota := 0
	if q.extraBillingData != nil {
		for _, value := range q.extraBillingData {
			extraBillingQuota += int(math.Ceil(
				float64(value.Price)*float64(config.QuotaPerUnit),
			)) * value.CallCount
		}
	}

	if extraBillingQuota > 0 {
		quota += int(math.Ceil(
			float64(extraBillingQuota) * q.groupRatio,
		))
	}

	if q.inputRatio != 0 && quota <= 0 {
		quota = 1
	}
	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}

	return quota
}

// 获取计算的 token 数
func (q *Quota) getComputeTokensByUsage(usage *types.Usage) (promptTokens, completionTokens int) {
	promptTokens = usage.PromptTokens
	completionTokens = usage.CompletionTokens

	extraTokens := usage.GetExtraTokens()

	for key, value := range extraTokens {
		extraRatio := q.price.GetExtraRatio(key)
		if model.GetExtraPriceIsPrompt(key) {
			promptTokens += model.GetIncreaseTokens(value, extraRatio)
		} else {
			completionTokens += model.GetIncreaseTokens(value, extraRatio)
		}
	}

	return
}

func (q *Quota) getComputeTokensByUsageEvent(usage *types.UsageEvent) (promptTokens, completionTokens int) {
	promptTokens = usage.InputTokens
	completionTokens = usage.OutputTokens
	extraTokens := usage.GetExtraTokens()

	for key, value := range extraTokens {
		extraRatio := q.price.GetExtraRatio(key)
		if model.GetExtraPriceIsPrompt(key) {
			promptTokens += model.GetIncreaseTokens(value, extraRatio)
		} else {
			completionTokens += model.GetIncreaseTokens(value, extraRatio)
		}
	}

	return
}

// 通过 usage 获取消费配额
func (q *Quota) GetTotalQuotaByUsage(usage *types.Usage) (quota int) {
	promptTokens, completionTokens := q.getComputeTokensByUsage(usage)
	return q.GetTotalQuota(promptTokens, completionTokens, usage.ExtraBilling)
}

func (q *Quota) GetFirstResponseTime() int64 {
	// 先判断 firstResponseTime 是否为0
	if q.firstResponseTime.IsZero() {
		return 0
	}

	return q.firstResponseTime.Sub(q.startTime).Milliseconds()
}

func (q *Quota) SetFirstResponseTime(firstResponseTime time.Time) {
	q.firstResponseTime = firstResponseTime
}

type ExtraBillingData struct {
	Type      string  `json:"type"`
	CallCount int     `json:"call_count"`
	Price     float64 `json:"price"`
}

func (q *Quota) GetExtraBillingData(extraBilling map[string]types.ExtraBilling) {
	if extraBilling == nil {
		return
	}

	extraBillingData := make(map[string]ExtraBillingData)
	for serviceType, value := range extraBilling {
		extraBillingData[serviceType] = ExtraBillingData{
			Type:      value.Type,
			CallCount: value.CallCount,
			Price:     getDefaultExtraServicePrice(serviceType, q.modelName, value.Type),
		}

	}

	if len(extraBillingData) == 0 {
		return
	}

	q.extraBillingData = extraBillingData
}
