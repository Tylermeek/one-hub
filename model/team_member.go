package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/utils"

	"gorm.io/gorm"
)

type TeamMember struct {
	Id         int    `json:"id"`
	TeamId     int    `json:"team_id" gorm:"not null;index"`
	UserId     int    `json:"user_id" gorm:"not null;index"`
	Role       int    `json:"role" gorm:"type:int;default:2"`        // 角色：1管理员 2普通成员
	MaxQuota   int    `json:"max_quota" gorm:"type:int;default:0"`   // 该成员最大可用额度（0表示无限制）
	UsedQuota  int    `json:"used_quota" gorm:"type:int;default:0"`  // 该成员已使用的团队额度
	Status     int    `json:"status" gorm:"type:int;default:1"`      // 状态：1正常 2禁用
	JoinedTime int64  `json:"joined_time" gorm:"bigint"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 计算字段（不存储到数据库）
	AvailableQuota int  `json:"available_quota" gorm:"-"`  // 计算出的可用额度
	UnlimitedQuota bool `json:"unlimited_quota" gorm:"-"`  // 是否无限额度

	// 关联
	Team *Team `json:"team,omitempty" gorm:"foreignKey:TeamId;references:Id"`
	User *User `json:"user,omitempty" gorm:"foreignKey:UserId;references:Id"`
}

type SearchTeamMemberParams struct {
	TeamMember
	PaginationParams
}

var allowedTeamMemberOrderFields = map[string]bool{
	"id":           true,
	"role":         true,
	"status":       true,
	"joined_time":  true,
	"used_quota":   true,
}

func GetTeamMembersList(teamId int, params *SearchTeamMemberParams) (*DataResult[TeamMember], error) {
	var members []*TeamMember
	db := DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name, email")
	}).Where("team_id = ?", teamId)

	if params.Role != 0 {
		db = db.Where("role = ?", params.Role)
	}
	if params.Status != 0 {
		db = db.Where("status = ?", params.Status)
	}

	result, err := PaginateAndOrder(db, &params.PaginationParams, &members, allowedTeamMemberOrderFields)
	if err != nil {
		return nil, err
	}

	// 为每个成员计算可用额度
	for _, member := range *result.Data {
		availableQuota, unlimited, err := GetMemberAvailableQuota(teamId, member.UserId)
		if err == nil {
			member.AvailableQuota = availableQuota
			member.UnlimitedQuota = unlimited
		}
	}

	return result, nil
}

func GetTeamMember(teamId, userId int) (*TeamMember, error) {
	if teamId == 0 || userId == 0 {
		return nil, errors.New("team_id 或 user_id 为空")
	}
	var member TeamMember
	err := DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name, email")
	}).First(&member, "team_id = ? AND user_id = ?", teamId, userId).Error
	return &member, err
}

// GetMemberAvailableQuota - 计算成员可用额度
// 返回：可用额度，是否无限额度，错误
func GetMemberAvailableQuota(teamId, userId int) (int, bool, error) {
	// 获取成员信息
	member, err := GetTeamMember(teamId, userId)
	if err != nil {
		return 0, false, err
	}

	// 获取团队信息
	team, err := GetTeamById(teamId)
	if err != nil {
		return 0, false, err
	}

	// 如果团队是无限额度
	if team.UnlimitedQuota {
		// 如果成员有额度限制，返回成员剩余额度
		if member.MaxQuota > 0 {
			remaining := member.MaxQuota - member.UsedQuota
			if remaining < 0 {
				remaining = 0
			}
			return remaining, false, nil
		}
		// 成员无限制，团队也无限制
		return 0, true, nil
	}

	// 团队有额度限制
	teamAvailableQuota := team.Quota - team.UsedQuota
	if teamAvailableQuota < 0 {
		teamAvailableQuota = 0
	}

	// 如果成员有额度限制
	if member.MaxQuota > 0 {
		memberRemaining := member.MaxQuota - member.UsedQuota
		if memberRemaining < 0 {
			memberRemaining = 0
		}
		// 取团队可用额度和成员剩余额度的较小值
		if teamAvailableQuota < memberRemaining {
			return teamAvailableQuota, false, nil
		}
		return memberRemaining, false, nil
	}

	// 成员无限制，返回团队可用额度
	return teamAvailableQuota, false, nil
}

func GetUserTeamMembers(userId int, params *PaginationParams) (*DataResult[TeamMember], error) {
	var members []*TeamMember
	db := DB.Preload("Team", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, owner_id")
	}).Preload("Team.Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name")
	}).Where("user_id = ?", userId)

	return PaginateAndOrder(db, params, &members, allowedTeamMemberOrderFields)
}

func (member *TeamMember) Insert() error {
	if member.TeamId == 0 {
		return errors.New("team_id 为空")
	}
	if member.UserId == 0 {
		return errors.New("user_id 为空")
	}

	// 检查用户是否已经是团队成员
	var count int64
	DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", member.TeamId, member.UserId).Count(&count)
	if count > 0 {
		return errors.New("用户已经是团队成员")
	}

	// 检查团队是否存在
	var team Team
	if err := DB.First(&team, "id = ?", member.TeamId).Error; err != nil {
		return errors.New("团队不存在")
	}

	member.JoinedTime = utils.GetTimestamp()
	return DB.Create(member).Error
}

func (member *TeamMember) Update() error {
	if member.Id == 0 {
		return errors.New("member id 为空")
	}

	return DB.Model(member).Select("role", "max_quota", "status").Updates(member).Error
}

func (member *TeamMember) Delete() error {
	if member.Id == 0 {
		return errors.New("member id 为空")
	}

	// 开始事务
	return DB.Transaction(func(tx *gorm.DB) error {
		// 获取成员信息
		var memberInfo TeamMember
		if err := tx.First(&memberInfo, "id = ?", member.Id).Error; err != nil {
			return err
		}

		// 如果成员使用了团队额度，记录日志但不退还
		if memberInfo.UsedQuota > 0 {
			RecordLog(memberInfo.UserId, LogTypeSystem, 
				fmt.Sprintf("离开团队，已使用团队额度 %s", common.LogQuota(memberInfo.UsedQuota)))
		}

		// 删除成员记录
		return tx.Delete(member).Error
	})
}

func DeleteTeamMember(teamId, userId int) error {
	if teamId == 0 || userId == 0 {
		return errors.New("team_id 或 user_id 为空")
	}

	var member TeamMember
	err := DB.Where("team_id = ? AND user_id = ?", teamId, userId).First(&member).Error
	if err != nil {
		return err
	}

	return member.Delete()
}

func IsTeamAdmin(teamId, userId int) bool {
	var count int64
	DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ? AND role = ? AND status = ?", 
		teamId, userId, 1, 1).Count(&count)
	return count > 0
}

func GetTeamMemberCount(teamId int) (int64, error) {
	var count int64
	err := DB.Model(&TeamMember{}).Where("team_id = ? AND status = ?", teamId, 1).Count(&count).Error
	return count, err
}

func UpdateTeamMemberQuota(teamId, userId, maxQuota int) error {
	if teamId == 0 || userId == 0 {
		return errors.New("team_id 或 user_id 为空")
	}

	return DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", teamId, userId).
		Update("max_quota", maxQuota).Error
}

func UpdateTeamMemberUsedQuota(teamId, userId, quota int) error {
	if teamId == 0 || userId == 0 {
		return errors.New("team_id 或 user_id 为空")
	}

	return DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", teamId, userId).
		Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
}
