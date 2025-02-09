//go:build wireinject
// +build wireinject

package components

import (
	"github.com/DaanV2/mechanus/server/internal/grpc"
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/servers"
	"github.com/google/wire"
)

func ServerComponent(folder string) (*servers.Manager, error) {
	wire.Build(
		baseSet,
		servicesSet,
		fileStorage,

		wire.Struct(new(web.WEBServies), "*"),
		wire.Struct(new(grpc.GRPCServices), "*"),
		buildServerComponent,
	)

	return nil, nil
}

func buildServerComponent(wserv web.WEBServies, gserv grpc.GRPCServices) *servers.Manager {
	manager := &servers.Manager{}
	manager.Register(
		web.NewServer(web.WebRouter(wserv)),
		grpc.NewServer(grpc.NewRouter(gserv)),
	)

	return manager
}
