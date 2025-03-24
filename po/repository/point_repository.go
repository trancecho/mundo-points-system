package repository

import (
	"context"
	"errors"
	"github.com/trancecho/mundo-points-system/po"
	"gorm.io/gorm"
)

type PointRepositoryImpl struct {
	db *gorm.DB
}

// NewPointRepository 创建积分仓库实例
func NewPointRepository(db *gorm.DB) *PointRepositoryImpl {
	return &PointRepositoryImpl{
		db: db,
	}
}

// AddPointsAndExperience 添加积分和经验值
func (r *PointRepositoryImpl) AddPointsAndExperience(ctx context.Context, userID string, points int64, experience int64, reason string) error {
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 记录积分变更
	pointRecord := &po.PointRecord{
		UserID:     userID,
		Points:     points,
		Experience: experience,
		Reason:     reason,
	}
	if err := tx.Create(pointRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新用户积分和经验
	result := tx.Model(&po.UserInfo{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"points":     gorm.Expr("points + ?", points),
			"experience": gorm.Expr("experience + ?", experience),
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

// RecordLike 记录点赞信息
func (r *PointRepositoryImpl) RecordLike(ctx context.Context, userID string, postID string, targetUserID string) error {
	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 检查是否已经点赞
	var count int64
	if err := tx.Model(&po.LikeRecord{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error; err != nil {
		tx.Rollback()
		return err
	}

	if count > 0 {
		tx.Rollback()
		return errors.New("已经点过赞")
	}

	// 记录点赞
	likeRecord := &po.LikeRecord{
		UserID:       userID,
		PostID:       postID,
		TargetUserID: targetUserID,
	}
	if err := tx.Create(likeRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 给被点赞者加积分
	result := tx.Model(&po.UserInfo{}).
		Where("user_id = ?", targetUserID).
		Update("points", gorm.Expr("points + ?", 1)) // 点赞默认加1分

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return tx.Commit().Error
}
