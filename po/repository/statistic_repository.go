package repository

import (
	"context"
	"github.com/trancecho/mundo-points-system/po"
	"gorm.io/gorm"
	"time"
)

type StatisticsRepositoryImpl struct {
	db *gorm.DB
}

// NewStatisticsRepository 创建统计仓库实例
func NewStatisticsRepository(db *gorm.DB) *StatisticsRepositoryImpl {
	return &StatisticsRepositoryImpl{
		db: db,
	}
}

// GetLevelDistribution 获取等级分布
func (r *StatisticsRepositoryImpl) GetLevelDistribution(ctx context.Context) (map[int]int64, error) {
	var results []struct {
		Level int
		Count int64
	}

	err := r.db.WithContext(ctx).
		Model(&po.UserInfo{}).
		Select("level, count(*) as count").
		Group("level").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[int]int64)
	for _, result := range results {
		distribution[result.Level] = result.Count
	}

	return distribution, nil
}

// GetAveragePoints 获取平均积分
func (r *StatisticsRepositoryImpl) GetAveragePoints(ctx context.Context) (float32, error) {
	var result struct {
		AvgPoints float32
	}

	err := r.db.WithContext(ctx).
		Model(&po.UserInfo{}).
		Select("AVG(points) as avg_points").
		First(&result).Error

	if err != nil {
		return 0, err
	}

	return result.AvgPoints, nil
}

// GetMonthlyPointsUsed 获取本月使用的积分
func (r *StatisticsRepositoryImpl) GetMonthlyPointsUsed(ctx context.Context) (int64, error) {
	var result struct {
		TotalPoints int64
	}

	// 获取当月第一天和最后一天
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := now.AddDate(0, 1, 0)
	firstDayNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, now.Location())

	// 查询当月消耗的积分总量（负值积分记录之和）
	err := r.db.WithContext(ctx).
		Model(&po.PointRecord{}).
		Select("ABS(SUM(points)) as total_points").
		Where("points < 0 AND created_at >= ? AND created_at < ?", firstDay, firstDayNextMonth).
		First(&result).Error

	if err != nil {
		return 0, err
	}

	return result.TotalPoints, nil
}
