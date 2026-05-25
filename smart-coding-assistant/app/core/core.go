package main

import (
	"flag"
	"fmt"

	"smart-coding-assistant/app/core/internal/config"
	"smart-coding-assistant/app/core/internal/server"
	"smart-coding-assistant/app/core/internal/svc"
	"smart-coding-assistant/app/core/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.MCPClient.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterCoreServiceServer(grpcServer, server.NewCoreServer(ctx))
	})
	defer s.Stop()

	fmt.Printf("Core RPC 服务启动，监听 %s...\n", c.ListenOn)
	s.Start()
}
