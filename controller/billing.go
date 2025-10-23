package controller

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/model"

	"github.com/gin-gonic/gin"
)

func GetSubscription(c *gin.Context) {
	var remainQuota int
	var usedQuota int
	var err error
	var token *model.Token
	var expiredTime int64

	tokenId := c.GetInt("token_id")
	token, err = model.GetTokenById(tokenId)
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("获取信息失败: %v", err))
		return
	}

	if token.UnlimitedQuota {
		userId := c.GetInt("id")
		
		// 检查用户是否有团队
		teamId, err := model.GetUserTeamForConsumption(userId)
		if err == nil && teamId > 0 {
			// 用户有团队，计算团队+个人额度
			totalAvailable, unlimited, err := model.GetUserTotalAvailableQuota(userId)
			if err == nil {
				if unlimited {
					remainQuota = 999999999 // 表示无限额度
				} else {
					remainQuota = totalAvailable
				}
			}
			
			// 获取用户已用额度
			userData, err := model.GetUserFields(userId, []string{"used_quota"})
			if err == nil {
				usedQuota = userData["used_quota"].(int)
			}
		} else {
			// 用户没有团队，使用原有逻辑
			userData, err := model.GetUserFields(userId, []string{"quota", "used_quota"})
			if err != nil {
				common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("获取用户信息失败: %v", err))
				return
			}
			remainQuota = userData["quota"].(int)
			usedQuota = userData["used_quota"].(int)
		}
	} else {
		expiredTime = token.ExpiredTime
		remainQuota = token.RemainQuota
		usedQuota = token.UsedQuota
	}

	if expiredTime <= 0 {
		expiredTime = 0
	}

	quota := remainQuota + usedQuota
	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}

	subscription := OpenAISubscriptionResponse{
		Object:             "billing_subscription",
		HasPaymentMethod:   true,
		SoftLimitUSD:       amount,
		HardLimitUSD:       amount,
		SystemHardLimitUSD: amount,
		AccessUntil:        expiredTime,
	}
	c.JSON(200, subscription)
}

func GetUsage(c *gin.Context) {
	var quota int
	var err error
	var token *model.Token

	tokenId := c.GetInt("token_id")
	token, err = model.GetTokenById(tokenId)
	if err != nil {
		common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("获取信息失败: %v", err))
		return
	}

	if token.UnlimitedQuota {
		userId := c.GetInt("id")
		
		// 检查用户是否有团队
		teamId, err := model.GetUserTeamForConsumption(userId)
		if err == nil && teamId > 0 {
			// 用户有团队，获取用户已用额度（团队使用量已包含在用户使用量中）
			userData, err := model.GetUserFields(userId, []string{"used_quota"})
			if err == nil {
				quota = userData["used_quota"].(int)
			}
		} else {
			// 用户没有团队，使用原有逻辑
			userData, err := model.GetUserFields(userId, []string{"used_quota"})
			if err != nil {
				common.APIRespondWithError(c, http.StatusOK, fmt.Errorf("获取用户信息失败: %v", err))
				return
			}
			quota = userData["used_quota"].(int)
		}
	} else {
		quota = token.UsedQuota
	}

	amount := float64(quota)
	if config.DisplayInCurrencyEnabled {
		amount /= config.QuotaPerUnit
	}
	usage := OpenAIUsageResponse{
		Object:     "list",
		TotalUsage: amount * 100,
	}
	c.JSON(200, usage)
}
