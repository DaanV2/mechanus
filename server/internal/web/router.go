package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/internal/routes"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/charmbracelet/log"
)

type WEBServices struct {
	Components *application.ComponentManager
}

func WebRouter(conf ServerConfig, services WEBServices) *http.ServeMux {
	router := http.NewServeMux()

	routes.RegisterHealthChecks(router, services.Components)
	routes.RegisterReadyChecks(router, services.Components)

	// Files
	log.Debug("serving files from: " + conf.StaticFolder)
	router.Handle("/", http.FileServer(http.Dir(conf.StaticFolder)))

	return router
}
