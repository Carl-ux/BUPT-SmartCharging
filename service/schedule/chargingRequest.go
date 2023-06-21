package service

import (
	"BSC/common"
	"BSC/model"
	"BSC/service"
	"BSC/utils"
	"fmt"
	"strconv"
	"time"
)

// 常量
const (
	MaxRecycleID        = 200  //最大分配ID
	WaitingAreaCapacity = 6    //等待区容量
	ChargingQueueLen    = 2    //充电桩队列长度
	TricklePilePower    = 7.0  //慢充功率
	FastChargePilePower = 30.0 //快充攻率
)

// 充电桩类型 分为普通充电桩和快充桩
type PileType int

const (
	TricklePile    PileType = iota //T代表慢充
	FastChargePile                 //F代表慢充
)

// 充电请求状态类型
type StatusType int

const (
	NotCharging       StatusType = iota //不在充电
	WaitingStage1                       //正在等待区
	WaitingStage2                       //正在排队队列
	Charging                            //正在充电
	ChangeModeRequeue                   //切换模式重新排队
	FailRequeue                         //充电失败重排
)

// 调度模式
type SchedulingMode int

const (
	Normal      SchedulingMode = iota //普通调度
	Priority                          //优先级调度
	TimeOrdered                       //时间顺序调度
	Recovery                          //故障恢复
)

// 默认调度模式
const DefalutScheduleMode = Priority

// 请求状态
type RequestStatus struct {
	Status   StatusType
	Position int
	PileID   int
}

// 充电请求结构体
type ChargingRequest struct {
	CreateTime       time.Time //创建时间
	RequestID        string    //请求ID
	RequestType      PileType  //请求类型
	Username         string    //用户名
	Amount           float64   //充电量
	BatteryCapacity  float64   //电池容量
	IsInWaitingQueue bool      //是否在等待区
	IsExecuting      bool      //是否在执行
	BeginTime        time.Time //开始时间
	IsRemoved        bool      //是否被移除
	PileID           int       //对应充电桩号
	RequeueFlag      bool      //重排标志
	FailFlag         bool      //失败标志
}

func CreateOrder(requestType PileType, pileID int, userName string, amount float64, beginTime time.Time, endTime time.Time, brakeflag bool) {
	/*
			    request_type (PileType): 充电模式
		        pild_id (int): 充电桩编号
		        username (str): 用户名
		        amount (Decimal): 用量
		        battery_capacity (Decimal): 电池容量
		        begin_time (datetime): 开始时间
		        end_time (datetime): 结束时间
	*/

	totalcost, chargingcost, servicecost := service.CalcCost(beginTime, endTime, amount)
	DB := common.GetDB()
	var loginUser model.User
	DB.Where("name = ?", userName).First(&loginUser).Limit(1)
	record := &model.ChargingRecord{
		UserID:        loginUser.ID,
		ChargerID:     pileID,
		ChargedAmount: amount,
		CreateTime:    service.GetDatetimeNow(),
		BeginTime:     beginTime,
		EndTime:       endTime,
		ChargedTime:   int(endTime.Sub(beginTime).Seconds()),
		ChargingCost:  chargingcost,
		ServiceCost:   servicecost,
		TotalCost:     totalcost,
		IsBrake:       brakeflag,
	}
	DB.Create(record)
	//将record写入CSV文件导出
	utils.WriteRecordToCSV(userName, record)
	fmt.Println("充电详单创建成功！")
	var pile model.ChargerPile
	DB.Where("pile_id = ?", pileID).First(&pile).Limit(1)
	pile.CumulativeUsageTimes += 1
	pile.CumulativeChargingAmount += amount
	pile.CumulativeChargingTime += record.ChargedTime
	DB.Save(&pile)
}

func GetLastRecord(userID int) map[string]interface{} {
	DB := common.GetDB()
	var order model.ChargingRecord
	result := DB.Where("user_id = ?", userID).Order("create_time DESC").Limit(1).Find(&order)
	if result.Error != nil {
		panic(result.Error)
	}
	orderInfo := map[string]interface{}{
		"order_id":       strconv.Itoa(int(order.RecordID)),
		"create_time":    order.CreateTime.Format("2006-01-02T15:04:05.000Z"),
		"charged_amount": order.ChargedAmount,
		"charged_time":   order.ChargedTime,
		"begin_time":     order.BeginTime.Format("2006-01-02T15:04:05.000Z"),
		"end_time":       order.EndTime.Format("2006-01-02T15:04:05.000Z"),
		"charging_cost":  order.ChargingCost,
		"service_cost":   order.ServiceCost,
		"total_cost":     order.TotalCost,
		"pile_id":        "C0" + strconv.Itoa(order.ChargerID),
	}
	return orderInfo
}

func GetAllOrders(userID int) []map[string]interface{} {
	orderList := []map[string]interface{}{}
	DB := common.GetDB()
	var orders []model.ChargingRecord
	result := DB.Where("user_id = ?", userID).Find(&orders)
	if result.Error != nil {
		panic(result.Error)
	}
	for _, order := range orders {
		orderInfo := map[string]interface{}{
			"order_id":       strconv.Itoa(int(order.RecordID)),
			"create_time":    order.CreateTime.Format("2006-01-02T15:04:05.000Z"),
			"charged_amount": order.ChargedAmount,
			"charged_time":   order.ChargedTime,
			"begin_time":     order.BeginTime.Format("2006-01-02T15:04:05.000Z"),
			"end_time":       order.EndTime.Format("2006-01-02T15:04:05.000Z"),
			"charging_cost":  order.ChargingCost,
			"service_cost":   order.ServiceCost,
			"total_cost":     order.TotalCost,
			"pile_id":        "C0" + strconv.Itoa(order.ChargerID),
		}
		orderList = append(orderList, orderInfo)
	}
	return orderList
}

func (s StatusType) String() string {
	switch s {
	case NotCharging:
		return "NOTCHARGING"
	case WaitingStage1:
		return "WAITINGSTAGE1"
	case WaitingStage2:
		return "WAITINGSTAGE2"
	case Charging:
		return "CHARGING"
	case ChangeModeRequeue:
		return "ChANGEMODEREQUEUE"
	case FailRequeue:
		return "FAILREQUEUE"
	default:
		return ""
	}
}
