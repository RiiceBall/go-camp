package main

import (
	"gitee.com/geekbang/basic-go/webook/internal/events"
	"gitee.com/geekbang/basic-go/webook/pkg/ginx"
	"gitee.com/geekbang/basic-go/webook/pkg/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
}
