package repository

import (
	"context"
	"errors"
	"github.com/trancecho/mundo-points-system/po"
	"gorm.io/gorm"
	"time"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

// GetUserByID 通过用户ID获取用户信息
func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, userID string) (*po.UserInfo, error) {
	var user po.UserInfo
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateSignStatus 更新用户签到状态
func (r *UserRepositoryImpl) UpdateSignStatus(ctx context.Context, userID string, isSigned bool, continuousDay int32, totalDay int32) error {
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 更新用户签到状态
	result := tx.Model(&po.UserInfo{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_signed":           isSigned,
			"continuous_sign_day": continuousDay,
			"total_sign_day":      totalDay,
			"last_sign_date":      time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("用户不存在")
	}

	return tx.Commit().Error
}

// UpdateLevelByExperience 根据经验值更新用户等级
func (r *UserRepositoryImpl) UpdateLevelByExperience(ctx context.Context, userID string) error {
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 获取用户当前经验值
	var user po.UserInfo
	result := tx.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 根据经验值计算新等级
	newLevel := calculateLevelByExperience(user.Experience)

	// 如果等级有变化，则更新
	if int(newLevel) != user.Level {
		result = tx.Model(&po.UserInfo{}).
			Where("user_id = ?", userID).
			Update("level", newLevel)

		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}

	return tx.Commit().Error
}

// calculateLevelByExperience 根据经验值计算等级
func calculateLevelByExperience(experience int64) int32 {
	// 简单的等级计算规则，可以根据业务需求调整
	if experience < 100 {
		return 1
	} else if experience < 500 {
		return 2
	} else if experience < 1000 {
		return 3
	} else if experience < 2000 {
		return 4
	} else if experience < 5000 {
		return 5
	} else if experience < 10000 {
		return 6
	} else if experience < 18000 {
		return 7
	} else if experience < 30000 {
		return 8
	} else if experience < 50000 {
		return 9
	} else {
		return 10
	}
}
