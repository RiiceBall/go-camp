package ioc

import (
	grpc2 "webook/comment/grpc"
	"webook/pkg/grpcx"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCxServer(comment *grpc2.CommentServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	comment.Register(server)
	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
