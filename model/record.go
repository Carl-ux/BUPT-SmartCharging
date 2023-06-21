package model

import "time"

// 充电详单
type ChargingRecord struct {
	RecordID      uint      `json:"record_id" gorm:"column:record_id;primaryKey;autoIncrement"`
	UserID        uint      `json:"user_id" gorm:"column:user_id;not null"`
	ChargerID     int       `json:"charger_id" gorm:"column:charger_id;not null"`
	ChargedAmount float64   `json:"charged_amount" gorm:"column:charged_amount;type:decimal(6,2)"` // 充电量
	CreateTime    time.Time `json:"create_time" gorm:"column:create_time;type:timestamp"`
	BeginTime     time.Time `json:"begin_time" gorm:"column:begin_time;type:timestamp"`
	EndTime       time.Time `json:"end_time" gorm:"column:end_time;type:timestamp"`
	ChargedTime   int       `json:"charged_time" gorm:"column:charged_time"` // 充电时长
	ChargingCost  float64   `json:"charging_cost" gorm:"column:charging_cost;type:decimal(5,2)"`
	ServiceCost   float64   `json:"service_cost" gorm:"column:service_cost;type:decimal(5,2)"`
	TotalCost     float64   `json:"total_cost" gorm:"column:total_cost;type:decimal(5,2)"`
	IsBrake       bool      `json:"is_brake" gorm:"column:is_brake;not null;default:false"` // 订单是否非正常完成(充电桩故障或被关闭)
}
