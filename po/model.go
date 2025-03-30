package po

import "time"

// UserInfo 用户信息模型
type UserInfo struct {
	BaseModel
	UserID            int64     `gorm:"column:user_id;not null;unique"`
	Username          string    `gorm:"column:username;not null"`
	Points            int64     `gorm:"column:points;not null;default:0"`
	Experience        int64     `gorm:"column:experience;not null;default:0"`
	Level             int       `gorm:"column:level;not null;default:1"`
	IsSigned          bool      `gorm:"column:is_signed;not null;default:false"`
	ContinuousSignDay int32     `gorm:"column:continuous_sign_day;not null;default:0"`
	TotalSignDay      int32     `gorm:"column:total_sign_day;not null;default:0"`
	LastSignDate      time.Time `gorm:"column:last_sign_date"`
}

// PointRecord 积分记录模型
type PointRecord struct {
	BaseModel
	UserID     string `gorm:"column:user_id;not null;index"`
	Points     int64  `gorm:"column:points;not null"`
	Experience int64  `gorm:"column:experience;not null"`
	Reason     string `gorm:"column:reason;not null"`
}

// LikeRecord 点赞记录模型
type LikeRecord struct {
	BaseModel
	UserID       string `gorm:"column:user_id;not null;index"`
	PostID       string `gorm:"column:post_id;not null;index"`
	TargetUserID string `gorm:"column:target_user_id;not null;index"`
}
