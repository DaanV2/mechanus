package http

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/infrastructure/health"
	"github.com/charmbracelet/log"
)

type WEBServices struct {
	Components *application.ComponentManager
}

func WebRouter(conf ServerConfig, services WEBServices) *http.ServeMux {
	router := http.NewServeMux()

	health.RegisterHealthChecks(router, services.Components)
	health.RegisterReadyChecks(router, services.Components)

	// Files
	log.Debug("serving files from: " + conf.StaticFolder)
	router.Handle("/", http.FileServer(http.Dir(conf.StaticFolder)))

	return router
}
