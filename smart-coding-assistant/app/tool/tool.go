package main

import (
	"flag"
	"fmt"
	"net/http"

	"smart-coding-assistant/app/tool/internal/config"
	"smart-coding-assistant/app/tool/internal/mcp"
	"smart-coding-assistant/app/tool/internal/server"
	"smart-coding-assistant/app/tool/internal/svc"
	"smart-coding-assistant/app/tool/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/tool.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	toolSrv := server.NewToolServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterToolServiceServer(grpcServer, toolSrv)
	})
	defer s.Stop()

	// 启动 MCP HTTP 服务
	mcpHandler, err := mcp.BuildHTTPHandler(toolSrv)
	if err != nil {
		panic(fmt.Sprintf("构建 MCP 处理器失败: %v", err))
	}

	mcpPort := c.MCP.Port
	if mcpPort == "" {
		mcpPort = "8081"
	}

	go func() {
		fmt.Printf("MCP Server 监听端口 %s...\n", mcpPort)
		if err := http.ListenAndServe(":"+mcpPort, mcpHandler); err != nil {
			panic(fmt.Sprintf("MCP 服务启动失败: %v", err))
		}
	}()

	fmt.Printf("Tool RPC 服务启动，监听 %s...\n", c.ListenOn)
	s.Start()
}
