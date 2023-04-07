package controller

import (
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

type Base struct {
}

func (base *Base) response(ctx iris.Context, code uint, msg string, data any) {
	_ = ctx.JSON(iris.Map{
		"code": code,
		"msg":  msg,
		"data": data,
	})
	ctx.StopWithStatus(200)
}

func (base *Base) SUCCESS(ctx iris.Context, msg string, data any) {
	base.response(ctx, 200, msg, data)
}

func (base *Base) FAIL(ctx iris.Context, msg string, data any) {
	base.response(ctx, 500, msg, data)
}

// getLogger 获取iris日志实例
func (base *Base) getLogger(ctx iris.Context) *golog.Logger {
	return ctx.Application().Logger()
}

// Info 打印Info日志
func (base *Base) Info(ctx iris.Context, v ...any) {
	base.getLogger(ctx).Info(v)
}

// Infof 打印格式化Info日志
func (base *Base) Infof(ctx iris.Context, format string, args ...any) {
	base.getLogger(ctx).Infof(format, args)
}

func (base *Base) Error(ctx iris.Context, v ...any) {
	base.getLogger(ctx).Error(v)
}

func (base *Base) Errorf(ctx iris.Context, format string, args ...any) {
	base.getLogger(ctx).Errorf(format, args)
}

func (base *Base) Warning(ctx iris.Context, v ...any) {
	base.getLogger(ctx).Warn(v)
}

func (base *Base) Warningf(ctx iris.Context, format string, args ...any) {
	base.getLogger(ctx).Warnf(format, args)
}
