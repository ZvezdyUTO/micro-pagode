package jwt

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret []byte
	expire time.Duration
}

func NewJWT(secret, expire string) *JWT {
	d, _ := time.ParseDuration(expire)
	return &JWT{
		secret: []byte(secret),
		expire: d,
	}
}

func (j *JWT) GenerateToken(claims map[string]interface{}) (string, error) { // 生成 Token
	mapClaims := jwt.MapClaims{
		"exp": time.Now().Add(j.expire).Unix(),
		"iat": time.Now().Unix(),
	}

	for k, v := range claims {
		mapClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	return token.SignedString(j.secret)
}

func (j *JWT) ParseToken(tokenStr string) (jwt.MapClaims, error) { // 解析 Token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (j *JWT) GetUID(ctx *gin.Context) (int64, error) {
	uidVal, ok := ctx.Get("uid")
	if !ok {
		return 0, errors.New("未登录")
	}

	uid, ok := uidVal.(int64)
	if !ok {
		return 0, errors.New("用户信息异常")
	}

	return uid, nil
}

func (j *JWT) GetIsAdmin(ctx *gin.Context) bool {
	res, ok := ctx.Get("isAdmin")
	if !ok {
		return false
	}
	return res.(bool)
}
