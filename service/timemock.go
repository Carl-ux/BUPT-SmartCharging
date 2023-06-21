package service

import "time"

var (
	RealTime        time.Time
	RealTimestamp   int64
	BootTime        time.Time
	BootTimestamp   int64
	fastForwardRate int64 = 60 //快进速率
	tz, _                 = time.LoadLocation("Asia/Shanghai")
)

func init() {

	//验收初始时间为6:00:00
	BootTime = time.Date(2023, 2, 3, 5, 40, 0, 0, tz)
	BootTimestamp = BootTime.Unix()
	RealTime = time.Now().UTC().In(tz)
	RealTimestamp = RealTime.Unix()

	// BootTime = time.Now().UTC().In(tz) //当前时间
	// BootTimestamp = time.Now().Unix()  //时间戳
}

func ResetTime() {
	BootTime = time.Now().UTC().In(tz)
	BootTimestamp = time.Now().Unix()
}

func GetTimestampNow() int64 {
	realTimestamp := time.Now().Unix()
	delta := realTimestamp - RealTimestamp //偏移量
	mockedTimestamp := BootTimestamp + delta*fastForwardRate
	return mockedTimestamp
}

func GetDatetimeNow() time.Time {
	realDatetime := time.Now().UTC().In(tz)
	delta := realDatetime.Sub(RealTime)
	mockedDatetime := BootTime.Add(delta * time.Duration(fastForwardRate))
	return mockedDatetime
}
