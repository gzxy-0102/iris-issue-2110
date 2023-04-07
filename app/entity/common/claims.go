package common

import "github.com/golang-jwt/jwt/v4"

// LoginClaims 登录Token中Payload实体
type LoginClaims struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
