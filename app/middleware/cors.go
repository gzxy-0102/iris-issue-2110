package middleware

import (
	"github.com/kataras/iris/v12"
	"net/http"
)

func Cors() iris.Handler {
	return func(ctx iris.Context) {
		method := ctx.Method()
		ctx.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		ctx.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "false")

		if method == "OPTIONS" {
			ctx.StopWithStatus(http.StatusNoContent)
		}
		ctx.Next()
	}
}
