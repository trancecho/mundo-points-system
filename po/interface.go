package po

import "context"

// UserRepository 用户仓库接口
type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
	UpdateSignStatus(ctx context.Context, userID string, isSigned bool, continuousDay int32, totalDay int32) error
	UpdateLevelByExperience(ctx context.Context, userID string) error
	UpdateActivityScore(ctx context.Context, userID string, deltaScore int64) error
}

// PointRepository 积分仓库接口
type PointRepository interface {
	AddPointsAndExperience(ctx context.Context, userID string, points int64, experience int64, reason string) error
	RecordLike(ctx context.Context, userID string, postID string, targetUserID string) error
}

// StatisticsRepository 统计仓库接口
type StatisticsRepository interface {
	GetLevelDistribution(ctx context.Context) (map[int]int64, error)
	GetAveragePoints(ctx context.Context) (float32, error)
	GetMonthlyPointsUsed(ctx context.Context) (int64, error)
}
