package dao

import (
	"BSC/common"
	"BSC/model"
	"BSC/utils"
	"strconv"
	"time"
)

func GetAllPiles() ([]model.ChargerPile, error) {
	DB := common.GetDB()
	var Piles []model.ChargerPile
	if err := DB.Find(&Piles).Error; err != nil {
		return nil, err
	}
	return Piles, nil
}

func GetAllPilesStatus() []map[string]interface{} {
	pilesStatus := []map[string]interface{}{}
	piles, _ := GetAllPiles()
	for _, pile := range piles {
		pileInfo := map[string]interface{}{
			"pile_id":               "P" + strconv.Itoa(pile.PileId),
			"status":                pile.ChargerStatus.String(),
			"total_usage_times":     pile.CumulativeUsageTimes,
			"total_charging_time":   pile.CumulativeChargingTime,
			"total_charging_amount": pile.CumulativeChargingAmount,
		}
		pilesStatus = append(pilesStatus, pileInfo)
	}
	return pilesStatus
}

func GetPileStatusByID(pileID int) (model.PileStatus, error) {
	DB := common.GetDB()
	var pile model.ChargerPile
	if err := DB.Where("pile_id = ?", pileID).First(&pile).Error; err != nil {
		return model.Unavailable, err
	}
	return pile.ChargerStatus, nil
}

func UpdatePile(pileID int, status model.PileStatus) error {
	DB := common.GetDB()
	var pile model.ChargerPile
	if err := DB.Where("pile_id = ?", pileID).First(&pile).Error; err != nil {
		return err
	}

	pile.ChargerStatus = status
	if err := DB.Save(&pile).Error; err != nil {
		return err
	}
	return nil
}

// 生成报表
func GetReport() []map[string]interface{} {
	var piles []struct {
		PileID                    int
		RegisterTime              time.Time
		CumulativeUsageTimes      int
		CumulativeChargingTime    int
		CumulativeChargingAmount  float64
		CumulativeChargingEarning float64
		CumulativeServiceEarning  float64
	}

	var statusList []map[string]interface{}
	DB := common.GetDB()
	DB.Table("charge_pile").
		Select("pile_id, register_time, " +
			"cumulative_usage_times, " +
			"cumulative_charging_time, " +
			"cumulative_charging_amount, " +
			"SUM(charging_records.charging_cost) AS cumulative_charging_earning, " +
			"SUM(charging_records.service_cost) AS cumulative_service_earning").
		Joins("left join charging_records on charge_pile.pile_id = charging_records.charger_id").
		Group("charge_pile.pile_id").
		Scan(&piles)
	for _, pile := range piles {
		status := map[string]interface{}{
			"pile_id":                "P" + strconv.Itoa(pile.PileID),
			"day":                    int(time.Since(pile.RegisterTime).Hours() / 24),
			"week":                   int(time.Since(pile.RegisterTime).Hours() / 24 / 7),
			"month":                  int(time.Since(pile.RegisterTime).Hours() / 24 / 30),
			"total_usage_times":      pile.CumulativeUsageTimes,
			"total_charging_time":    pile.CumulativeChargingTime,
			"total_charging_amount":  utils.Decimal(pile.CumulativeChargingAmount),
			"total_charging_earning": utils.Decimal(pile.CumulativeChargingEarning),
			"total_service_earning":  utils.Decimal(pile.CumulativeServiceEarning),
			"total_earning": utils.Decimal(pile.CumulativeChargingEarning) +
				utils.Decimal(pile.CumulativeServiceEarning),
		}
		statusList = append(statusList, status)
	}
	return statusList
}
