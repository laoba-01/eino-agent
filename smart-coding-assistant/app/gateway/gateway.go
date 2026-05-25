package main

import (
	"flag"
	"fmt"
	"net/http"

	"smart-coding-assistant/app/gateway/internal/config"
	"smart-coding-assistant/app/gateway/internal/handler"
	"smart-coding-assistant/app/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)

	mux := http.NewServeMux()
	handler.RegisterHandlers(mux, ctx)

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("API Gateway 启动，监听 %s...\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(fmt.Sprintf("服务启动失败: %v", err))
	}
}
