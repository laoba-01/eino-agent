package svc

import (
	authpb "smart-coding-assistant/app/auth/pb"
	corepb "smart-coding-assistant/app/core/pb"
	"smart-coding-assistant/app/gateway/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	AuthRpc   authpb.AuthServiceClient
	CoreRpc   corepb.CoreServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	authConn := zrpc.MustNewClient(c.AuthRpc)
	coreConn := zrpc.MustNewClient(c.CoreRpc)

	return &ServiceContext{
		Config:  c,
		AuthRpc: authpb.NewAuthServiceClient(authConn.Conn()),
		CoreRpc: corepb.NewCoreServiceClient(coreConn.Conn()),
	}
}
