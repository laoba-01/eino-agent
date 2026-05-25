package main

import (
	"flag"
	"fmt"

	"smart-coding-assistant/app/auth/internal/config"
	"smart-coding-assistant/app/auth/internal/server"
	"smart-coding-assistant/app/auth/internal/svc"
	"smart-coding-assistant/app/auth/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/auth.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterAuthServiceServer(grpcServer, server.NewAuthServer(ctx))
	})
	defer s.Stop()

	fmt.Printf("Auth RPC 服务启动，监听 %s...\n", c.ListenOn)
	s.Start()
}
