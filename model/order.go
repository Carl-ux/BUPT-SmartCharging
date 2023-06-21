package model

import "time"

// 用户订单
type ChargingOrder struct {
	OrderID          uint64    `json:"order_id" gorm:"column:order_id;primaryKey;autoIncrement"`
	UserID           uint64    `json:"user_id" gorm:"column:user_id;not null"`
	ChargingMode     string    `json:"charging_mode" gorm:"column:charging_mode;not null"`         // T(慢）或 F(快）
	ChargingCapacity float64   `json:"charging_capacity" gorm:"column:charging_capacity;not null"` // 请求充电的电量
	EditTime         time.Time `json:"edit_time" gorm:"column:edit_time;type:timestamp"`           // 最近编辑时间 初始化为订单创建时间
}
