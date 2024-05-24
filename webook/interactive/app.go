package main

import (
	"webook/internal/events"
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
)

type App struct {
	consumers   []events.Consumer
	server      *grpcx.Server
	adminServer *ginx.Server
}
