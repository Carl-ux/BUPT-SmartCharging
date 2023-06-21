package main

import (
	"BSC/controller"
	"BSC/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) *gin.Engine {
	//middleware
	r.Use(middleware.RecoveryMiddleware(), middleware.CORSMiddleware())
	//auth
	r.POST("account_login", controller.Login)
	r.POST("account_register", controller.Register)
	r.GET("/api/auth/info", middleware.AuthMiddleware(), controller.Info)

	r.GET("get_time", controller.QueryTime)
	r.GET("test_cost", controller.QueryCostTest)

	//user
	r.POST("submit_request", middleware.AuthMiddleware(), controller.SubmitChargingRequest)
	r.POST("edit_request", middleware.AuthMiddleware(), controller.EditChargingRequest)
	r.GET("end_request", middleware.AuthMiddleware(), controller.EndChargingRequest)

	r.GET("query_order", middleware.AuthMiddleware(), controller.QueryRecord)
	r.GET("preview_queue", middleware.AuthMiddleware(), controller.PreviewQueue)

	//admin
	r.GET("query_pile", middleware.AuthMiddleware(), controller.GetAllPileStatus)
	r.POST("update_pile", middleware.AuthMiddleware(), controller.UpdatePileStatus)
	r.GET("query_report", middleware.AuthMiddleware(), controller.QueryReportAPI)
	r.GET("query_queue", middleware.AuthMiddleware(), controller.QueryQueueAPI)
	return r
}
