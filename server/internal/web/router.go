package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/internal/routes"
	"github.com/DaanV2/mechanus/server/pkg/application"
)

func WebRouter(comps *application.ComponentManager, folder string) *http.ServeMux {
	router := http.NewServeMux()

	routes.RegisterHealthChecks(router, comps)
	routes.RegisterReadyChecks(router, comps)

	// Files
	router.Handle("/", http.FileServer(http.Dir(folder)))

	return router
}
