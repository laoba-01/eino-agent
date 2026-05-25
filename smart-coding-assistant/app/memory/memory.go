package main

import (
	"flag"
	"fmt"

	"smart-coding-assistant/app/memory/internal/config"
	"smart-coding-assistant/app/memory/internal/server"
	"smart-coding-assistant/app/memory/internal/svc"
	"smart-coding-assistant/app/memory/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/memory.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterMemoryServiceServer(grpcServer, server.NewMemoryServer(ctx))
	})
	defer s.Stop()

	fmt.Printf("Memory RPC 服务启动，监听 %s...\n", c.ListenOn)
	s.Start()
}
