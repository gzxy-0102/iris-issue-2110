package app

import (
	"2110/app/middleware"
	"2110/app/model"
	cfg "2110/config"
	"2110/pkg/database/dynamicSource"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/iris-contrib/middleware/throttler"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/monitor"
	"github.com/kataras/iris/v12/middleware/requestid"
	log "github.com/sirupsen/logrus"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

type App struct {
	*iris.Application
	config  cfg.Configuration
	closers []func()
	orm     *gorm.DB
	cache   *redis.Client
	limiter throttler.RateLimiter
}

func NewApp(config cfg.Configuration) *App {
	app := iris.New()
	app.SetName(config.ServerName)
	app.Configure(iris.WithConfiguration(config.Iris), iris.WithLowercaseRouting)
	srv := &App{
		Application: app,
		config:      config,
	}
	if err := srv.prepare(); err != nil {
		srv.Logger().Fatal(err)
		return nil
	}

	return srv
}

func (app *App) prepare() error {
	var err error
	if app.Logger().Level == golog.DebugLevel {
		app.registerDebugFeatures()
	}
	//	初始化日志及设置日志格式
	app.bootstrapLog()
	//	检测已注册数据源活跃状态
	app.registerDynamicSourceState()
	err = app.registerOrm()
	if app.config.Redis.Enable {
		err = app.registerCache()
	}
	err = app.registerLimiter()
	err = app.registerMiddlewares()
	app.buildRouter()
	return err
}

func (app *App) bootstrapLog() {
	//	初始化log 提供给 json api 等使用
	// 为true时显示文件及行号等信息
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return "", path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
		},
		DisableColors: false,
	})
	file, err := os.OpenFile("./database-api.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	var writer io.Writer
	if err != nil {
		writer = os.Stdout
	} else {
		writer = io.MultiWriter(os.Stdout, file)
	}
	log.SetOutput(writer)
	log.SetLevel(log.DebugLevel)
}

func (app *App) registerDynamicSourceState() {
	go func() {
		timer := time.NewTicker(time.Minute * 1)
		for {
			select {
			case <-timer.C:
				dynamicSource.RDB.Range(func(key, value any) bool {
					manager, ok := value.(*dynamicSource.RDBManager)
					log.Infof("当前数据源连接信息：%s 上次使用时间：%v", manager.DSN, manager.LastTime)
					if ok {
						//	超过60分钟 关闭连接并删除
						since := time.Since(manager.LastTime).Minutes()
						if since > 60 {
							log.Warningf("数据源：%s 超过60分钟无活动 正在关闭", manager.DSN)
							err := dynamicSource.UNRegisterSource(model.Source{
								Base: model.Base{
									ID: key.(uint64),
								},
								Device: manager.Device,
							})
							if err != nil {
								log.Warningf("数据源：%s 关闭错误：%v", manager.DSN, err)
							}
						}
					}
					return true
				})
				break
			}
		}
	}()
}

func (app *App) registerDebugFeatures() {
	// TODO 自定义调试部分
}

func (app *App) registerOrm() error {
	db, err := gorm.Open(mysql.Open(app.config.Database.DSN))
	if err == nil {
		app.orm = db
		app.AddCloser(func() {
			logger := app.Logger()
			logger.Info("正在关闭数据库...")
			sqlDb, err := db.DB()
			if sqlDb != nil {
				err = sqlDb.Close()
			}
			if err != nil {
				logger.Errorf("数据库关闭错误：%+v", err)
				return
			}
			logger.Info("数据库关闭成功")
		})
	}
	return err
}

func (app *App) registerCache() error {
	if app.config.Redis.Addr != "" {
		app.cache = redis.NewClient(&redis.Options{
			Addr:               app.config.Redis.Addr,
			Username:           app.config.Redis.Username,
			Password:           app.config.Redis.Password,
			DB:                 app.config.Redis.Database,
			MaxRetries:         app.config.Redis.MaxRetries,
			MinRetryBackoff:    app.config.Redis.MinRetryBackoff,
			MaxRetryBackoff:    app.config.Redis.MaxRetryBackoff,
			DialTimeout:        app.config.Redis.DialTimeout,
			ReadTimeout:        app.config.Redis.ReadTimeout,
			WriteTimeout:       app.config.Redis.WriteTimeout,
			PoolSize:           app.config.Redis.PoolSize,
			MinIdleConns:       app.config.Redis.MinIdleConns,
			MaxConnAge:         app.config.Redis.MaxConnAge,
			PoolTimeout:        app.config.Redis.PoolTimeout,
			IdleTimeout:        app.config.Redis.IdleTimeout,
			IdleCheckFrequency: app.config.Redis.IdleCheckFrequency,
		})
		app.AddCloser(func() {
			logger := app.Logger()
			logger.Info("正在关闭Redis....")
			err := app.cache.Close()
			if err != nil {
				logger.Errorf("Redis关闭错误：%+v", err)
				return
			}
			logger.Info("Redis关闭成功")
		})
	}
	return nil
}

func (app *App) registerLimiter() error {
	store, err := memstore.NewCtx(65536)
	if err != nil {
		return err
	}
	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(app.config.Limiter.PerMin),
		MaxBurst: 5,
	}
	rateLimiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
	if err != nil {
		return err
	}

	app.limiter = throttler.RateLimiter{
		RateLimiter: rateLimiter,
		VaryBy:      &throttled.VaryBy{Path: true},
		Error: func(ctx iris.Context, err error) {
			_ = ctx.JSON(iris.Map{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
		},
	}
	return nil
}

func (app *App) registerMiddlewares() error {
	app.UseRouter(logger.New())
	app.UseRouter(requestid.New())
	//	请求日志
	if app.config.RequestLog != "" {
		app.registerAccessLogger()
	}
	//	跨域
	app.UseRouter(middleware.Cors())
	//	错误处理
	app.UseRouter(middleware.Recover())
	//	传输压缩
	if app.config.EnableCompression {
		app.Use(iris.Compression)
	}
	if app.config.Monitor.Enable && app.config.Monitor.Path != "" {
		m := monitor.New(monitor.Options{
			RefreshInterval:     30 * time.Second,
			ViewRefreshInterval: 30 * time.Second,
			ViewTitle:           app.config.ServerName,
		})
		app.AddCloser(m.Stop)
		if app.config.Monitor.Auth.Enable && app.config.Monitor.Auth.Username != "" && app.config.Monitor.Auth.Password != "" {
			auth := basicauth.Default(map[string]string{
				app.config.Monitor.Auth.Username: app.config.Monitor.Auth.Password,
			})
			app.Get(app.config.Monitor.Path, auth, m.View)
		} else {
			app.Get(app.config.Monitor.Path, m.View)
		}
		app.Post(app.config.Monitor.Path, m.Stats)
	}
	return nil
}

func (app *App) registerAccessLogger() {
	ac := accesslog.FileUnbuffered(app.config.RequestLog)

	ac.Delim = '|'
	ac.TimeFormat = "2006-01-02 15:04:05"
	ac.Async = true
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = true
	ac.BytesSent = true
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = true
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler

	ac.SetFormatter(&accesslog.JSON{
		Indent:    "  ",
		HumanTime: true,
	})

	app.UseRouter(ac.Handler)
}

func (app *App) Start() error {

	app.ConfigureHost(func(su *host.Supervisor) {
		su.Server.ReadTimeout = 5 * time.Minute
		su.Server.WriteTimeout = 5 * time.Minute
		su.Server.IdleTimeout = 10 * time.Minute
		su.Server.ReadHeaderTimeout = 2 * time.Minute
	})

	addr := fmt.Sprintf("%s:%d", app.config.Host, app.config.Port)
	return app.Listen(addr)
}

func (app *App) AddCloser(closers ...func()) {
	for _, closer := range closers {
		if closer == nil {
			continue
		}
		iris.RegisterOnInterrupt(closer)
	}
	app.closers = append(app.closers, closers...)
}

func (app *App) Close() error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	err := app.Shutdown(ctx)
	cancelCtx()
	for _, closer := range app.closers {
		if closer == nil {
			continue
		}
		closer()
	}
	return err
}
