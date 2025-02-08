package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/internal/routes"
	"github.com/DaanV2/mechanus/server/pkg/application"
	"github.com/charmbracelet/log"
)

func WebRouter(comps *application.ComponentManager, folder string) *http.ServeMux {
	router := http.NewServeMux()

	routes.RegisterHealthChecks(router, comps)
	routes.RegisterReadyChecks(router, comps)

	// Files
	log.Debug("serving files from: " + folder)
	router.Handle("/", http.FileServer(http.Dir(folder)))

	return router
}
