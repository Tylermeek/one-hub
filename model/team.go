package model

import (
	"errors"
	"fmt"
	"one-api/common"
	"one-api/common/utils"

	"gorm.io/gorm"
)

type Team struct {
	Id             int    `json:"id"`
	Name           string `json:"name" gorm:"type:varchar(100);not null"`
	OwnerId        int    `json:"owner_id" gorm:"not null;index"`
	Quota          int    `json:"quota" gorm:"type:int;default:0"`           // 团队剩余额度
	UsedQuota      int    `json:"used_quota" gorm:"type:int;default:0"`      // 团队已使用额度
	UnlimitedQuota bool   `json:"unlimited_quota" gorm:"default:false"`      // 是否无限额度
	Status         int    `json:"status" gorm:"type:int;default:1"`          // 状态：1启用 2禁用
	InviteCode     string `json:"invite_code" gorm:"type:varchar(32);uniqueIndex"` // 邀请码
	CreatedTime    int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime    int64  `json:"updated_time" gorm:"bigint"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Owner   *User         `json:"owner,omitempty" gorm:"foreignKey:OwnerId;references:Id"`
	Members []*TeamMember `json:"members,omitempty" gorm:"foreignKey:TeamId;references:Id"`

	// 计算字段（不存储到数据库）
	CurrentUserRole int  `json:"current_user_role" gorm:"-"`  // 当前用户在该团队中的角色
	IsOwner         bool `json:"is_owner" gorm:"-"`           // 当前用户是否为团队所有者
	OwnerBalance    int  `json:"owner_balance" gorm:"-"`      // Owner个人余额（用于无限额度团队显示）
	EffectiveQuota  int  `json:"effective_quota" gorm:"-"`    // 有效额度（无限团队=owner.quota，有限团队=team.quota）
	AvailableQuota  int  `json:"available_quota" gorm:"-"`    // 可用余额（effective_quota - used_quota）
}

type SearchTeamParams struct {
	Team
	PaginationParams
}

var allowedTeamOrderFields = map[string]bool{
	"id":           true,
	"name":         true,
	"status":       true,
	"created_time": true,
	"updated_time": true,
}

func GetTeamsList(params *SearchTeamParams) (*DataResult[Team], error) {
	var teams []*Team
	db := DB.Preload("Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name")
	})

	if params.Name != "" {
		db = db.Where("name LIKE ?", params.Name+"%")
	}
	if params.OwnerId != 0 {
		db = db.Where("owner_id = ?", params.OwnerId)
	}
	if params.Status != 0 {
		db = db.Where("status = ?", params.Status)
	}

	return PaginateAndOrder(db, &params.PaginationParams, &teams, allowedTeamOrderFields)
}

func GetTeamById(id int) (*Team, error) {
	if id == 0 {
		return nil, errors.New("team id 为空")
	}
	var team Team
	err := DB.Preload("Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name")
	}).First(&team, "id = ?", id).Error
	return &team, err
}

func GetUserTeams(userId int, params *PaginationParams) (*DataResult[Team], error) {
	var teams []*Team
	
	// 查询用户作为管理员或成员的团队
	db := DB.Preload("Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username, display_name")
	}).Where("owner_id = ? OR id IN (SELECT team_id FROM team_members WHERE user_id = ? AND deleted_at IS NULL)", userId, userId)

	return PaginateAndOrder(db, params, &teams, allowedTeamOrderFields)
}

func GetTeamByInviteCode(inviteCode string) (*Team, error) {
	if inviteCode == "" {
		return nil, errors.New("邀请码为空")
	}
	var team Team
	err := DB.First(&team, "invite_code = ?", inviteCode).Error
	return &team, err
}

func (team *Team) Insert() error {
	if team.Name == "" {
		return errors.New("团队名称不能为空")
	}
	if team.OwnerId == 0 {
		return errors.New("团队管理员不能为空")
	}

	// 生成邀请码
	team.InviteCode = utils.GetRandomString(8)
	team.CreatedTime = utils.GetTimestamp()
	team.UpdatedTime = team.CreatedTime

	// 检查邀请码是否重复
	for {
		var count int64
		DB.Model(&Team{}).Where("invite_code = ?", team.InviteCode).Count(&count)
		if count == 0 {
			break
		}
		team.InviteCode = utils.GetRandomString(8)
	}

	return DB.Create(team).Error
}

func (team *Team) Update() error {
	if team.Id == 0 {
		return errors.New("team id 为空")
	}
	if team.Name == "" {
		return errors.New("团队名称不能为空")
	}

	team.UpdatedTime = utils.GetTimestamp()
	return DB.Model(team).Select("name", "status", "updated_time").Updates(team).Error
}

func (team *Team) Delete() error {
	if team.Id == 0 {
		return errors.New("team id 为空")
	}

	// 开始事务
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除团队成员
		if err := tx.Where("team_id = ?", team.Id).Delete(&TeamMember{}).Error; err != nil {
			return err
		}

		// 2. 如果团队有剩余额度，退还给管理员
		if team.Quota > 0 {
			if err := tx.Model(&User{}).Where("id = ?", team.OwnerId).Update("quota", gorm.Expr("quota + ?", team.Quota)).Error; err != nil {
				return err
			}
			// 记录日志
			RecordLog(team.OwnerId, LogTypeSystem, fmt.Sprintf("团队 %s 删除，退还额度 %s", team.Name, common.LogQuota(team.Quota)))
		}

		// 3. 删除团队
		return tx.Delete(team).Error
	})
}

func IsTeamOwner(teamId, userId int) bool {
	var count int64
	DB.Model(&Team{}).Where("id = ? AND owner_id = ?", teamId, userId).Count(&count)
	return count > 0
}

func IsTeamMember(teamId, userId int) bool {
	var count int64
	DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ? AND status = ?", teamId, userId, 1).Count(&count)
	return count > 0
}

func GetUserTeamIds(userId int) ([]int, error) {
	var teamIds []int
	err := DB.Model(&TeamMember{}).Where("user_id = ? AND status = ?", userId, 1).Pluck("team_id", &teamIds).Error
	return teamIds, err
}

func GetUserPrimaryTeam(userId int) (*Team, error) {
	// 获取用户加入的第一个团队作为主要团队
	var teamMember TeamMember
	err := DB.Where("user_id = ? AND status = ?", userId, 1).Order("joined_time ASC").First(&teamMember).Error
	if err != nil {
		return nil, err
	}
	return GetTeamById(teamMember.TeamId)
}
