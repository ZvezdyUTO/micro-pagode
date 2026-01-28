package middleware

import (
	"mp/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

type JWTMid struct {
	jwt *jwt.JWT
}

func NewJWTMid(j *jwt.JWT) *JWTMid {
	return &JWTMid{
		jwt: j,
	}
}

func (m *JWTMid) Handler(ctx *gin.Context) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		ctx.JSON(401, gin.H{"msg": "未登录"})
		ctx.Abort()
		return
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.JSON(401, gin.H{"msg": "Token 格式错误"})
		ctx.Abort()
		return
	}

	claims, err := m.jwt.ParseToken(parts[1])
	if err != nil {
		ctx.JSON(401, gin.H{"msg": "Token 无效或已过期"})
		ctx.Abort()
		return
	}

	// 1. 处理 uid
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		ctx.JSON(401, gin.H{"msg": "Token 数据异常"})
		ctx.Abort()
		return
	}
	ctx.Set("uid", int64(uidFloat))

	// 2. 处理 isAdmin
	isAdmin := false
	if v, ok := claims["is_admin"]; ok {
		if b, ok := v.(bool); ok {
			isAdmin = b
		}
	}
	ctx.Set("is_admin", isAdmin)

	// 3. 写入 name
	if name, ok := claims["name"].(string); ok {
		ctx.Set("name", name)
	}

	ctx.Next()
}
