package controller

import (
	"BSC/common"
	"BSC/model"
	"BSC/utils/response"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthUser struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RePassword string `json:"re_password"`
}

func Register(ctx *gin.Context) {
	DB := common.GetDB()

	// // 使用map获取请求参数
	// // var requestMap = make(map[string]string)
	// // json.NewDecoder(ctx.Request.Body).Decode(&requestMap)

	// // 结构体获取参数
	// // var requestUser = model.User{}
	// // json.NewDecoder(ctx.Request.Body).Decode(&requestUser)

	// gin的ShouldBindJSON获取json参数
	var requestUser = AuthUser{}
	if err := ctx.ShouldBindJSON(&requestUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := requestUser.Username
	password := requestUser.Password
	re_password := requestUser.RePassword

	fmt.Println(name, password)

	// 数据验证 1.数据是否为空 2.密码格式是否正确 3.用户名是否已经注册 4.重复密码是否正确
	// gin.H 就是一个map[string]interface{}
	if len(name) == 0 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "用户名不能为空!")
		return
	}

	if len(password) < 8 {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "密码不能少于8位!")
		return
	}

	if password != re_password {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "两次密码不一致!")
		return
	}

	log.Println(name, password)

	// 判断用户名是否存在
	if isUserExist(DB, name) {
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "用户已经存在!")
		return
	}
	// 创建用户
	// 加密用户密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Response(ctx, http.StatusInternalServerError, 500, nil, "密码加密失败!")
		return
	}
	newUser := model.User{
		Name:     name,
		Password: string(hashedPassword),
		Admin:    false,
	}

	DB.Create(&newUser)

	//返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func Login(ctx *gin.Context) {
	DB := common.GetDB()
	// 获取参数

	var requestUser = AuthUser{}
	if err := ctx.ShouldBindJSON(&requestUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := requestUser.Username
	password := requestUser.Password
	// 数据验证

	// 判断用户是否存在
	var loginUser model.User
	DB.Where("name = ?", name).First(&loginUser)
	if loginUser.ID == 0 {
		//用户不存在
		response.Response(ctx, http.StatusUnprocessableEntity, 422, nil, "用户不存在!")
		return
	}
	// 判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(loginUser.Password), []byte(password)); err != nil {
		response.Response(ctx, http.StatusBadRequest, 400, nil, "密码错误!")
		return
	}
	// 密码正确 发放token
	token, err := common.ReleaseToken(loginUser)
	if err != nil {
		response.Response(ctx, http.StatusInternalServerError, 500, nil, "生成token失败!")
		log.Printf("token generate error: %v", err)
		return
	}
	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"token":    token,
			"is_admin": loginUser.Admin},
		"message": "success",
	})

}

func Info(ctx *gin.Context) {
	// 获取上下文中的user 返回interface{}类型
	user, _ := ctx.Get("user")
	// 类型断言
	response.Success(ctx, gin.H{"user": model.ToUserDto(user.(model.User))}, "ok")
}

func isUserExist(db *gorm.DB, name string) bool {
	var user model.User
	if res := db.Where("name = ?", name).First(&user); res.Error != nil {
		return false
	}
	return true
}
