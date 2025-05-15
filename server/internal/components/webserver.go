package components

import (
	"github.com/DaanV2/mechanus/server/internal/web"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/DaanV2/mechanus/server/pkg/servers"
)

func WebServer(staticFolder string) (*servers.Manager, error) {
	manager := &servers.Manager{}

	manager.Register(
		web.NewServer(web.WebRouter(web.WEBServies{
			// TODO: this components need to be injected
			Components: application.NewComponentManager(),
		})),
	)

	// TODO built web servers, grpc, etc etc
	return manager, nil
}
