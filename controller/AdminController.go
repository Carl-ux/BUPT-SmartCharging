package controller

import (
	"BSC/dao"
	"BSC/model"
	service "BSC/service/schedule"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllPileStatus(ctx *gin.Context) {
	//验证是否为admin
	if role, _ := ctx.Get("role"); role != "admin" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	//验证请求方法
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	statusList := dao.GetAllPilesStatus()

	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
		"data":    statusList,
	})
}

func UpdatePileStatus(ctx *gin.Context) {
	//验证是否为admin
	if role, _ := ctx.Get("role"); role != "admin" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	//绑定json参数
	var req struct {
		PileID string `json:"pile_id" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": err.Error(),
		})
		return
	}

	//ID和更新后状态
	pileID, _ := strconv.Atoi(req.PileID[1:])
	status := model.PileStatusMap[req.Status]

	statusBefore, err := dao.GetPileStatusByID(pileID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": err.Error(),
		})
		return
	}
	if err = dao.UpdatePile(pileID, status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": err.Error(),
		})
		return
	}

	//状态不同的处理
	switch statusBefore {
	case model.Running:
		switch status {
		case model.Shutdown, model.Unavailable:
			service.Schd.Brake(pileID)
		default:
			//do nothing
		}
	case model.Shutdown, model.Unavailable:
		switch status {
		case model.Running:
			service.Schd.Recover(pileID)
		default:
			//do nothing
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
	})
}

func QueryReportAPI(ctx *gin.Context) {
	//验证是否为admin
	if role, _ := ctx.Get("role"); role != "admin" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	//验证请求方法
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	report := dao.GetReport()

	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
		"data":    report,
	})
}

func QueryQueueAPI(ctx *gin.Context) {
	//验证是否为admin
	if role, _ := ctx.Get("role"); role != "admin" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    FAIL,
			"message": "Permission denied",
		})
		return
	}

	//验证请求方法
	if ctx.Request.Method != http.MethodGet {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    FAIL,
			"message": "invalid request method",
		})
		return
	}

	queue := service.Schd.Snapshot()
	ctx.JSON(http.StatusOK, gin.H{
		"code":    SUCCESS,
		"message": "success",
		"data":    queue,
	})
}
