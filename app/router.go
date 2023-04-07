package app

import (
	"2110/app/controller"
	"2110/app/middleware"
)

func (app *App) buildRouter() {
	//	新建一个跟路由 并将连接池等信息注入到路由
	r := app.Party("/")
	{
		r.RegisterDependency(app.orm)
		r.RegisterDependency(app.config)
		if app.config.Redis.Enable {
			r.RegisterDependency(app.cache)
		}
		r.PartyConfigure("/json-api", new(controller.JSONApi)).
			UseRouter(
				middleware.JsonApiInit(app.orm),
				app.limiter.RateLimit,
			)

	}

}
