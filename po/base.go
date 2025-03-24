package po

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int64     `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`

	// 包含这个类型的字段就是软删除
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
