package controller

import (
	"BSC/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func QueryTime(ctx *gin.Context) {

	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	timeStr := service.GetDatetimeNow().Format("2006-01-02T15:04:05.000Z")
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"datatime":  timeStr,
			"timestamp": service.GetTimestampNow(),
		},
	})
}

func QueryCostTest(ctx *gin.Context) {
	t1 := service.BootTime
	t2 := service.GetDatetimeNow()
	Value1, Value2, Value3 := service.CalcCost(t1, t2, 100.0)
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"totalcost":    Value1,
			"chargingcost": Value2,
			"servicecost":  Value3,
		},
	})
}
