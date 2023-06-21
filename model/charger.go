package model

type PileStatus int

const (
	Running PileStatus = iota
	Shutdown
	Unavailable
)

var PileStatusMap = map[string]PileStatus{
	"RUNNING":     Running,
	"SHUTDOWN":    Shutdown,
	"UNAVAILABLE": Unavailable,
}

// 充电桩
type ChargerPile struct {
	PileId                   int        `gorm:"column:pile_id;primaryKey;autoIncrement"`
	ChargerStatus            PileStatus `gorm:"type:int;default:0"`
	Type                     string     `gorm:"type:varchar(20)"`             //T或F
	CumulativeUsageTimes     uint       `gorm:"type:int;default:0"`           //使用次数
	CumulativeChargingTime   int        `gorm:"type:int;default:0"`           //累计充电时间
	CumulativeChargingAmount float64    `gorm:"type:varchar(20);default:0.0"` //充电总量
}

// TableName 指定 Charger 模型对应的数据库表名为 "charger_pile"
func (c *ChargerPile) TableName() string {
	return "charge_pile"
}

func (p PileStatus) String() string {
	switch p {
	case Running:
		return "RUNNING"
	case Shutdown:
		return "SHUTDOWN"
	case Unavailable:
		return "UNAVAILABLE"
	default:
		return "UNKNOWN"
	}
}
