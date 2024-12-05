package entity

import (
	"go-template/internal/types/dto"
	"time"
)

// Model 是所有 entity 的基类，后期要将所有的 Base 改成这种形式
type Model struct {
	Id        dto.EntityId `gorm:"primarykey" json:"id"`
	CreatedAt time.Time    `gorm:"autoUpdateTime:milli" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime:milli" json:"updated_at"`
	//DeletedAt gorm.DeletedAt  `gorm:"index"`
}
