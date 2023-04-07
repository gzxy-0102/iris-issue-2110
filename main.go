package main

import (
	"2110/app"
	"2110/config"
	"flag"
)

var configFile = flag.String("c", "./server.yml", "配置文件所在位置")

func main() {
	serverConfig := new(config.Configuration)
	err := serverConfig.BindFile(*configFile)
	if err != nil {
		panic(err)
	}
	srv := app.NewApp(*serverConfig)
	_ = srv.Start()
}
