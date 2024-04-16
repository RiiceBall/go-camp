package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"webook/ioc"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initViper()
	// initViperRemote()
	initLogger()
	initPrometheus()
	tpCancel := ioc.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server := app.server

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world")
	})

	server.Run(":8080")
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// 监听 8081 端口，你也可以做成可配置的
		http.ListenAndServe(":8081", nil)
	}()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViper() {
	// 加载命令行参数 config 的值，如果没有则使用 config/dev.yaml
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	// 解析命令行参数，这一步之后 cflie 中才会有值
	pflag.Parse()
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("watch", viper.GetString("test.key"))
	})
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}

// func initViperRemote() {
// 	err := viper.AddRemoteProvider("etcd3",
// 		"http://127.0.0.1:12379", "/webook")
// 	if err != nil {
// 		panic(err)
// 	}
// 	viper.SetConfigType("yaml")
// 	err = viper.ReadRemoteConfig()
// 	if err != nil {
// 		panic(err)
// 	}
// 	// 最好不用
// 	// go func() {
// 	// 	for {
// 	// 		err = viper.WatchRemoteConfig()
// 	// 		if err != nil {
// 	// 			panic(err)
// 	// 		}
// 	// 		log.Println("watch", viper.GetString("test.key"))
// 	// 		time.Sleep(time.Second * 3)
// 	// 	}
// 	// }()
// }
