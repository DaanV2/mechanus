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

func WebRouter(services WEBServices) *http.ServeMux {
	router := http.NewServeMux()

	routes.RegisterHealthChecks(router, services.Components)
	routes.RegisterReadyChecks(router, services.Components)

	// Files
	folder := StaticFolderFlag.Value()
	log.Debug("serving files from: " + folder)
	router.Handle("/", http.FileServer(http.Dir(folder)))

	return router
}
