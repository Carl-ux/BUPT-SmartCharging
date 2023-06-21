package utils

import (
	"BSC/model"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// 保留两位小数
func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return num
}

func WriteRecordToCSV(name string, r *model.ChargingRecord) {
	path := "./orderCSV/" + name + ".csv"
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// 创建一个CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	//如果没有表头，写入表头
	filestat, _ := file.Stat()
	if filestat.Size() == 0 {
		writer.Write([]string{"record_id", "user_name", "user_id", "charger_id", "charged_amount", "create_time", "begin_time", "end_time", "charged_time", "charging_cost", "service_cost", "total_cost", "is_brake"})
	}
	//遍历写入
	writer.Write(
		[]string{
			strconv.Itoa(int(r.RecordID)),
			name,
			strconv.Itoa(int(r.UserID)),
			strconv.Itoa(int(r.ChargerID)),
			strconv.FormatFloat(r.ChargedAmount, 'f', 2, 64),
			r.CreateTime.Format("2006-01-02 15:04:05"),
			r.BeginTime.Format("2006-01-02 15:04:05"),
			r.EndTime.Format("2006-01-02 15:04:05"),
			strconv.Itoa(int(r.ChargedTime)),
			strconv.FormatFloat(r.ChargingCost, 'f', 2, 64),
			strconv.FormatFloat(r.ServiceCost, 'f', 2, 64),
			strconv.FormatFloat(r.TotalCost, 'f', 2, 64),
			strconv.FormatBool(r.IsBrake)})

	fmt.Printf("%v : 详单导出成功\n", path)
}
