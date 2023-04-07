package middleware

import (
	"2110/app/model"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

func JsonApiInit(orm *gorm.DB) iris.Handler {
	return func(ctx iris.Context) {
		mark := ctx.Params().GetString("mark")
		log.Infof("Mark: %v", mark)
		if mark != "" {
			var source model.Source
			result := orm.Where("source_mark = ? ", mark).First(&source)
			if result.Error != nil {
				_ = ctx.JSON(iris.Map{
					"code": http.StatusNotFound,
					"msg":  "数据源未找到",
				})
				ctx.StopWithStatus(http.StatusOK)
				return
			}
			ctx.Values().Set("source", source)
			ctx.Next()
			return
		}
		_ = ctx.JSON(iris.Map{
			"code": http.StatusNotFound,
			"msg":  "数据源未找到",
		})
		ctx.StopWithStatus(http.StatusOK)
		return
	}
}
