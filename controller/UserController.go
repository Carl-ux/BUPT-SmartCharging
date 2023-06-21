package controller

import (
	"BSC/model"
	s "BSC/service"
	service "BSC/service/schedule"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	SUCCESS = 0
	FAIL    = -1
)

func SubmitChargingRequest(ctx *gin.Context) {
	//验证是否为user
	if role, _ := ctx.Get("role"); role != "user" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	user, _ := ctx.Get("user")
	username := user.(model.User).Name

	var req struct {
		ChargeMode    string `json:"charge_mode" binding:"required"`
		RequireAmount string `json:"require_amount" binding:"required"`
		BatterySize   string `json:"battery_size" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": err.Error(),
		})
		return
	}

	var requestMode service.PileType
	switch req.ChargeMode {
	case "T":
		requestMode = service.TricklePile
	case "F":
		requestMode = service.FastChargePile
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "Invalid charge mode",
		})
		return
	}

	amount, _ := strconv.ParseFloat(req.RequireAmount, 64)
	batterysize, _ := strconv.ParseFloat(req.BatterySize, 64)
	err := service.Schd.SubmitRequest(requestMode, username, amount, batterysize, false)
	if err != nil {
		message := err.Error()

		ctx.JSON(http.StatusOK, gin.H{
			"code":    FAIL,
			"message": message,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
	})
}

func EndChargingRequest(ctx *gin.Context) {

	if role, _ := ctx.Get("role"); role != "user" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	user, _ := ctx.Get("user")
	username := user.(model.User).Name
	requestID := service.Schd.UsernameToRequest[username]
	service.Schd.EndRequest(requestID, s.GetDatetimeNow())
	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
	})
}

func EditChargingRequest(ctx *gin.Context) {

	if role, _ := ctx.Get("role"); role != "user" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	user, _ := ctx.Get("user")
	username := user.(model.User).Name
	requestID := service.Schd.UsernameToRequest[username]

	var req struct {
		ChargeMode    string `json:"charge_mode" binding:"required"`
		RequireAmount string `json:"require_amount" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": err.Error(),
		})
		return
	}
	amount, _ := strconv.ParseFloat(req.RequireAmount, 64)
	var requestMode service.PileType
	switch req.ChargeMode {
	case "T":
		requestMode = service.TricklePile
	case "F":
		requestMode = service.FastChargePile
	}
	err := service.Schd.UpdateRequest(requestID, amount, requestMode)
	if err != nil {
		message := err.Error()

		ctx.JSON(http.StatusOK, gin.H{
			"code":    FAIL,
			"message": message,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
	})
}

func QueryRecord(ctx *gin.Context) {

	if role, _ := ctx.Get("role"); role != "user" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	user, _ := ctx.Get("user")
	userID := user.(model.User).ID

	record := service.GetAllOrders(int(userID))
	//record := service.GetLastRecord(int(userID))
	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
		"data":    record,
	})
}

// 预览排队情况
func PreviewQueue(ctx *gin.Context) {
	//验证权限
	if role, _ := ctx.Get("role"); role != "user" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	// 验证请求方法
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}
	var err error
	user, _ := ctx.Get("user")
	username := user.(model.User).Name

	var place string
	requestID := ""
	pileID := 0
	position := -1

	if requestID, err = service.Schd.GetRequestIDByUsername(username); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    SUCCESS,
			"message": "success",
			"data": gin.H{
				"cur_state": "NOTCHARGING",
			},
		})
		return
	}

	requestStatus := service.Schd.GetRequestStatus(requestID)
	pileID = requestStatus.PileID
	position = requestStatus.Position

	if pileID == 0 {
		place = "WAITINGPLACE"
	} else {
		place = "C0" + strconv.Itoa(pileID)
	}

	curState := requestStatus.Status.String()
	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
		"data": gin.H{
			//本车排队号码
			"charge_id": requestID,
			//队列长度
			"queue_len": position,
			"cur_state": curState,
			"place":     place,
		},
	})

}
