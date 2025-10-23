package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/redis"
	"one-api/common/utils"

	"gorm.io/gorm"
)

// AllocateQuotaToTeam - 管理员设置团队消费上限
func AllocateQuotaToTeam(ownerId, teamId, quota int, unlimited bool) error {
	if ownerId == 0 || teamId == 0 {
		return errors.New("owner_id 或 team_id 为空")
	}

	// 检查管理员权限
	if !IsTeamOwner(teamId, ownerId) {
		return errors.New("无权限操作此团队")
	}

	// 开始事务
	return DB.Transaction(func(tx *gorm.DB) error {
		// 获取团队信息
		var team Team
		if err := tx.First(&team, "id = ?", teamId).Error; err != nil {
			return errors.New("团队不存在")
		}

		// 获取管理员信息
		var owner User
		if err := tx.First(&owner, "id = ?", ownerId).Error; err != nil {
			return errors.New("管理员不存在")
		}

		// 如果不是无限额度，检查团队额度不能超过管理员个人余额
		if !unlimited && quota > 0 {
			if owner.Quota < quota {
				return errors.New("团队消费上限不能超过您的个人余额")
			}
		}

		// 更新团队额度（替换逻辑，不扣除管理员个人额度）
		updateData := map[string]interface{}{
			"unlimited_quota": unlimited,
			"updated_time":    utils.GetTimestamp(),
		}
		
		if unlimited {
			// 无限额度时，quota 设为 0
			updateData["quota"] = 0
		} else {
			// 有限额度时，直接设置新的额度上限
			updateData["quota"] = quota
		}

		if err := tx.Model(&Team{}).Where("id = ?", teamId).Updates(updateData).Error; err != nil {
			return err
		}

		// 记录日志
		var logContent string
		if unlimited {
			logContent = fmt.Sprintf("设置团队 %s 为无限额度（基于个人余额）", team.Name)
		} else {
			logContent = fmt.Sprintf("设置团队 %s 消费上限为 %s", team.Name, common.LogQuota(quota))
		}
		RecordLog(ownerId, LogTypeManage, logContent)

		// 清除缓存
		if config.RedisEnabled {
			redis.RedisDel(fmt.Sprintf(UserQuotaCacheKey, ownerId))
		}

		return nil
	})
}

// ConsumeTeamQuota - 团队成员消费额度（优先团队，其次个人）
// 返回：teamQuotaUsed, userQuotaUsed, error
func ConsumeTeamQuota(userId, teamId, totalQuota int) (int, int, error) {
	if userId == 0 || teamId == 0 || totalQuota <= 0 {
		return 0, 0, errors.New("参数错误")
	}

	// 检查用户是否是团队成员
	member, err := GetTeamMember(teamId, userId)
	if err != nil {
		return 0, 0, errors.New("用户不是团队成员")
	}

	// 检查成员状态
	if member.Status != 1 {
		return 0, 0, errors.New("成员状态异常")
	}

	// 检查成员最大额度限制
	if member.MaxQuota > 0 && member.UsedQuota+totalQuota > member.MaxQuota {
		return 0, 0, errors.New("超出团队成员额度限制")
	}

	// 获取团队信息
	team, err := GetTeamById(teamId)
	if err != nil {
		return 0, 0, errors.New("团队不存在")
	}

	// 检查团队状态
	if team.Status != 1 {
		return 0, 0, errors.New("团队状态异常")
	}

	// 计算额度分配
	var teamQuotaUsed, userQuotaUsed int

	if team.UnlimitedQuota {
		// 团队无限额度，全部使用团队额度
		teamQuotaUsed = totalQuota
		userQuotaUsed = 0
	} else {
		// 优先使用团队额度
		availableTeamQuota := team.Quota - team.UsedQuota
		if availableTeamQuota >= totalQuota {
			teamQuotaUsed = totalQuota
			userQuotaUsed = 0
		} else {
			teamQuotaUsed = availableTeamQuota
			userQuotaUsed = totalQuota - availableTeamQuota
		}
	}

	// 检查用户个人额度是否足够补足
	if userQuotaUsed > 0 {
		user, err := GetUserById(userId, false)
		if err != nil {
			return 0, 0, errors.New("用户不存在")
		}
		if user.Quota < userQuotaUsed {
			return 0, 0, errors.New("用户个人额度不足")
		}
	}

	// 事务执行扣除
	err = DB.Transaction(func(tx *gorm.DB) error {
		// 扣除团队额度（如果不是无限额度）
		if !team.UnlimitedQuota && teamQuotaUsed > 0 {
			if err := tx.Model(&Team{}).Where("id = ?", teamId).
				Update("used_quota", gorm.Expr("used_quota + ?", teamQuotaUsed)).Error; err != nil {
				return err
			}
		}

		// 扣除用户个人额度
		if userQuotaUsed > 0 {
			if err := tx.Model(&User{}).Where("id = ?", userId).
				Update("quota", gorm.Expr("quota - ?", userQuotaUsed)).Error; err != nil {
				return err
			}
		}

		// 更新成员已用额度
		if err := tx.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", teamId, userId).
			Update("used_quota", gorm.Expr("used_quota + ?", teamQuotaUsed)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	// 清除缓存
	if config.RedisEnabled {
		redis.RedisDel(fmt.Sprintf(UserQuotaCacheKey, userId))
	}

	return teamQuotaUsed, userQuotaUsed, nil
}

// GetUserTeamForConsumption - 获取用户用于消费的团队
// 返回用户的主要团队ID，如果用户不属于任何团队则返回0
func GetUserTeamForConsumption(userId int) (int, error) {
	if userId == 0 {
		return 0, nil
	}

	// 获取用户所属的团队ID列表
	teamIds, err := GetUserTeamIds(userId)
	if err != nil {
		return 0, err
	}

	if len(teamIds) == 0 {
		return 0, nil
	}

	// 如果用户只属于一个团队，直接返回
	if len(teamIds) == 1 {
		return teamIds[0], nil
	}

	// 多团队场景：优先选择有额度的团队
	// 1. 优先选择无限额度的团队
	for _, teamId := range teamIds {
		team, err := GetTeamById(teamId)
		if err == nil && team.Status == 1 && team.UnlimitedQuota {
			return teamId, nil
		}
	}

	// 2. 选择有剩余额度的团队
	for _, teamId := range teamIds {
		team, err := GetTeamById(teamId)
		if err == nil && team.Status == 1 {
			availableQuota := team.Quota - team.UsedQuota
			if availableQuota > 0 {
				return teamId, nil
			}
		}
	}

	// 3. 如果所有团队都没有额度，返回第一个团队（让系统处理额度不足的情况）
	return teamIds[0], nil
}

// ReclaimTeamQuota - 回收团队额度（减少团队额度时调用）
func ReclaimTeamQuota(ownerId, teamId, quota int) error {
	if ownerId == 0 || teamId == 0 || quota <= 0 {
		return errors.New("参数错误")
	}

	// 检查管理员权限
	if !IsTeamOwner(teamId, ownerId) {
		return errors.New("无权限操作此团队")
	}

	// 开始事务
	return DB.Transaction(func(tx *gorm.DB) error {
		// 获取团队信息
		var team Team
		if err := tx.First(&team, "id = ?", teamId).Error; err != nil {
			return errors.New("团队不存在")
		}

		// 检查团队可用额度
		availableQuota := team.Quota - team.UsedQuota
		if availableQuota < quota {
			return errors.New("团队可用额度不足")
		}

		// 减少团队额度
		if err := tx.Model(&Team{}).Where("id = ?", teamId).
			Update("quota", gorm.Expr("quota - ?", quota)).Error; err != nil {
			return err
		}

		// 退还给管理员个人额度
		if err := tx.Model(&User{}).Where("id = ?", ownerId).
			Update("quota", gorm.Expr("quota + ?", quota)).Error; err != nil {
			return err
		}

		// 记录日志
		RecordLog(ownerId, LogTypeManage, 
			fmt.Sprintf("从团队 %s 回收额度 %s", team.Name, common.LogQuota(quota)))

		// 清除缓存
		if config.RedisEnabled {
			redis.RedisDel(fmt.Sprintf(UserQuotaCacheKey, ownerId))
		}

		return nil
	})
}

// GetTeamQuotaInfo - 获取团队额度信息
func GetTeamQuotaInfo(teamId int) (quota, usedQuota int, unlimited bool, err error) {
	var team Team
	err = DB.Select("quota, used_quota, unlimited_quota").First(&team, "id = ?", teamId).Error
	if err != nil {
		return 0, 0, false, err
	}
	return team.Quota, team.UsedQuota, team.UnlimitedQuota, nil
}

// GetEffectiveQuotaForContext - 获取指定上下文的有效可用额度
// 返回：可用额度，是否无限，错误
func GetEffectiveQuotaForContext(userId int, contextType string, contextId int) (int, bool, error) {
	if contextType == "team" && contextId > 0 {
		team, err := GetTeamById(contextId)
		if err != nil {
			return 0, false, err
		}
		
		if IsTeamOwner(contextId, userId) {
			// Owner: 返回个人余额
			owner, err := GetUserById(userId, false)
			if err != nil {
				return 0, false, err
			}
			return owner.Quota, false, nil
		} else {
			// 成员: 返回团队可用额度
			if team.UnlimitedQuota {
				// 无限额度团队，返回 Owner 个人余额
				owner, err := GetUserById(team.OwnerId, false)
				if err != nil {
					return 0, false, err
				}
				return owner.Quota, false, nil
			} else {
				// 有限额度团队
				available := team.Quota - team.UsedQuota
				if available < 0 {
					available = 0
				}
				return available, false, nil
			}
		}
	} else {
		// 个人上下文或空上下文，返回用户个人额度
		user, err := GetUserById(userId, false)
		if err != nil {
			return 0, false, err
		}
		return user.Quota, false, nil
	}
}

// IncreaseTeamUsedQuota - 增加团队已用额度（用于预消费）
func IncreaseTeamUsedQuota(teamId, quota int) error {
	if teamId == 0 || quota <= 0 {
		return errors.New("参数错误")
	}
	return DB.Model(&Team{}).Where("id = ?", teamId).
		Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
}

// DecreaseTeamUsedQuota - 减少团队已用额度（用于退款）
func DecreaseTeamUsedQuota(teamId, quota int) error {
	if teamId == 0 || quota <= 0 {
		return errors.New("参数错误")
	}
	return DB.Model(&Team{}).Where("id = ?", teamId).
		Update("used_quota", gorm.Expr("used_quota - ?", quota)).Error
}
