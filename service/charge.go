package service

import (
	"BSC/utils"
	"time"
)

var matt = "2006-01-02 15:04:05"

const (
	PEAK     = iota //峰时
	SHOULDER        //平时
	OFFPEAK         //谷时

	// 服务费单价（元/度）
	ServiceCostPerKwh float64 = 0.80
	// 阶梯计价（时间左闭右开）
	// 峰时: 10:00~15:00 & 18:00~21:00
	ChargingCostPerKwhPeak float64 = 1.00
	// 平时: 7:00~10:00 & 15:00~18:00 & 21:00~23:00
	ChargingCostPerKwhShoulder float64 = 0.70
	// 谷时: 23:00~次日7:00
	ChargingCostPerKwhOffpeak float64 = 0.40
)

// 计价区间:
// |  index   |   0    |   1    |    2    |    3    |    4    |    5    |   6    |
// |   type   | BOTTOM | MEDIUM |   TOP   | MEDIUM  |   TOP   | MEDIUM  | BOTTOM |
// | interval | 0 ~ 7  | 7 ~ 10 | 10 ~ 15 | 15 ~ 18 | 18 ~ 21 | 21 ~ 23 | 23 ~ 0 |
// |  length  |  420   |  180   |   300   |   180   |   180   |   120   |   60   |

var (
	BillingIntervalsEnd = []int{7, 10, 15, 18, 21, 23, 0} // 计价区间终点 (hour)
	WhichType           = []int{2, 1, 0, 1, 0, 1, 2}      // 各区间电价类型
	WhichInterval       = []int{
		0, 0, 0, 0, 0, 0, 0, // 0～6
		1, 1, 1,
		2, 2, 2, 2, 2,
		3, 3, 3,
		4, 4, 4,
		5, 5,
		6,
	} // 24小时各自对应的计价区间
)

func CalcCost(beginTime, endTime time.Time, chargedAmount float64) (float64, float64, float64) {
	intervalsTimeCnt := [3]int{0, 0, 0}          // 峰平谷各自的时间长度
	intervalsChargingCost := [3]float64{0, 0, 0} // 峰平谷各自的充电费用
	curTime := beginTime
	curInterval := WhichInterval[curTime.Hour()] //开始时间区间
	endInterval := WhichInterval[endTime.Hour()] //结束时间区间
	//当起始和结束不在同一天的同一区间
	for !(GetTimesubDay(curTime.Format(matt), endTime.Format(matt)) == 0 && curInterval == endInterval) {
		curIntervalType := WhichType[curInterval] //当前区间类型
		//下一区间的起始时间(这一阶段的终点)
		nextIntervalBeginTime := time.Date(
			curTime.Year(),
			curTime.Month(),
			curTime.Day(),
			BillingIntervalsEnd[curInterval],
			0, 0, 0, curTime.Location())
		curInterval = curInterval + 1 //进入下一区间
		//下一区间在最后一个区间之后 代表跨天
		if curInterval%7 == 0 {
			curInterval = 0
			if GetTimesubDay(curTime.Format(matt), endTime.Format(matt)) > 0 {
				nextIntervalBeginTime = time.Date(
					curTime.Year(),
					curTime.Month(),
					curTime.Day()+1,
					nextIntervalBeginTime.Hour(), 0, 0, 0, curTime.Location())
			}
		}
		//得到各类型的时间
		intervalsTimeCnt[curIntervalType] += int(nextIntervalBeginTime.Sub(curTime).Seconds())
		//循环直到同天同区间
		curTime = nextIntervalBeginTime
	}
	//计算同天同区间的时间
	curIntervalType := WhichType[curInterval]
	intervalsTimeCnt[curIntervalType] += int(endTime.Sub(curTime).Seconds())

	//充电总耗时
	totalTime := intervalsTimeCnt[PEAK] + intervalsTimeCnt[SHOULDER] + intervalsTimeCnt[OFFPEAK]
	//计算各类型的充电费用
	//充电费 = 单位电价 * 充电度数
	intervalsChargingCost[PEAK] = ChargingCostPerKwhPeak * chargedAmount * float64(intervalsTimeCnt[PEAK]) / float64(totalTime)
	intervalsChargingCost[SHOULDER] = ChargingCostPerKwhShoulder * chargedAmount * float64(intervalsTimeCnt[SHOULDER]) / float64(totalTime)
	intervalsChargingCost[OFFPEAK] = ChargingCostPerKwhOffpeak * chargedAmount * float64(intervalsTimeCnt[OFFPEAK]) / float64(totalTime)
	//充电总费用
	chargingCost := utils.Decimal(intervalsChargingCost[PEAK] + intervalsChargingCost[SHOULDER] + intervalsChargingCost[OFFPEAK])
	//服务费用
	serviceCost := utils.Decimal(ServiceCostPerKwh * chargedAmount)
	//总费用
	totalCost := utils.Decimal(chargingCost + serviceCost)

	return totalCost, chargingCost, serviceCost
}

// 获取自然天之差
func GetTimesubDay(bef, now string) int {
	var day int
	t1, _ := time.Parse("2006-01-02 15:04:05", bef)
	t2, _ := time.Parse("2006-01-02 15:04:05", now)
	swap := false
	if t1.Unix() > t2.Unix() {
		t1, t2 = t2, t1
		swap = true
	}

	t1_ := t1.Add(time.Duration(t2.Sub(t1).Milliseconds()%86400000) * time.Millisecond)
	day = int(t2.Sub(t1).Hours() / 24)
	// 计算在t1+两个时间的余数之后天数是否有变化
	if t1_.Day() != t1.Day() {
		day += 1
	}

	if swap {
		day = -day
	}
	return day
}
