package middleware

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http/httputil"
	"runtime"
	"runtime/debug"
	"strings"
)

func Recover() iris.Handler {
	return func(ctx *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() { // handled by other middleware.
					return
				}

				var callers []string
				for i := 1; ; i++ {
					_, file, line, got := runtime.Caller(i)
					if !got {
						break
					}

					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}

				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
				logMessage += fmt.Sprint(getRequestLogs(ctx))
				logMessage += fmt.Sprintf("%s\n", err)
				logMessage += fmt.Sprintf("%s\n", strings.Join(callers, "\n"))
				ctx.Application().Logger().Warn(logMessage)

				// get the list of registered handlers and the
				// handler which panic derived from.
				handlers := ctx.Handlers()
				handlersFileLines := make([]string, 0, len(handlers))
				currentHandlerIndex := ctx.HandlerIndex(-1)
				currentHandlerFileLine := "???"
				for i, h := range ctx.Handlers() {
					file, line := context.HandlerFileLine(h)
					fileline := fmt.Sprintf("%s:%d", file, line)
					handlersFileLines = append(handlersFileLines, fileline)
					if i == currentHandlerIndex {
						currentHandlerFileLine = fileline
					}
				}
				exce := &context.ErrPanicRecovery{
					Cause:              err,
					Callers:            callers,
					Stack:              debug.Stack(),
					RegisteredHandlers: handlersFileLines,
					CurrentHandler:     currentHandlerFileLine,
				}
				err = ctx.JSON(iris.Map{
					"code": 200,
					"msg":  "服务器错误",
					"data": exce,
				})
				if err != nil {
					ctx.StopWithPlainError(500, exce)
				}
				ctx.StopWithStatus(200)
			}
		}()
		ctx.Next()
	}
}

func getRequestLogs(ctx *context.Context) string {
	rawReq, _ := httputil.DumpRequest(ctx.Request(), false)
	return string(rawReq)
}
