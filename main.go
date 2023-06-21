package main

import (
	"BSC/common"
	"BSC/config"
	service "BSC/service/schedule"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 初始化配置
	config.InitConfig()
	// 初始化数据库
	common.InitDB()
	// 初始化调度
	service.InitSchd()

	r := gin.Default()
	r = InitRouter(r)
	// 监听端口
	port := viper.GetString("server.port")
	if port != "" {
		panic(r.Run(":" + port))
	}
	panic(r.Run())
}
