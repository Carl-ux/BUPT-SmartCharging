package common

import (
	"BSC/model"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// 进行jwt鉴权
// jwt加密解密的key
var jwtKey = []byte("BUPT")

// jwt 声明的结构体
type Claims struct {
	UserId uint
	jwt.StandardClaims
}

// 生成token(加密字符串)
func ReleaseToken(user model.User) (string, error) {
	//过期时间
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserId: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "carl-learn",
			Subject:   "user token",
		},
	}
	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 生成加密字符串
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 解析token 获取claims
func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, nil, err
	}
	return token, claims, nil
}
