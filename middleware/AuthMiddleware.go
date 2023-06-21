package middleware

import (
	"BSC/common"
	"BSC/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取authorization header
		// 由前端在登录时保存并在每次请求时携带
		tokenString := ctx.GetHeader("Authorization")
		// validate token formate
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "权限不足"})
		}
		// // drop Bearer prefix
		// tokenString = tokenString[7:]

		// get token
		token, claims, err := common.ParseToken(tokenString)
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "权限不足"})
			ctx.Abort()
			return
		}

		// 验证通过后获取claim中的userId
		userId := claims.UserId
		DB := common.GetDB()
		var user model.User
		// 从数据库中查询userId对应的user
		DB.First(&user, userId)
		//  is user exist
		if user.ID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "权限不足"})
			ctx.Abort()
			return
		}
		// 用户存在 将user信息写入上下文
		if user.Admin {
			ctx.Set("role", "admin")
		} else {
			ctx.Set("role", "user")
		}
		ctx.Set("user", user)
		ctx.Next()
	}
}
