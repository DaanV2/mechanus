package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/internal/routes"
	"github.com/DaanV2/mechanus/server/pkg/application"
)

func NewRouter(comps *application.ComponentManager) *http.ServeMux {
	router := http.NewServeMux()

	routes.RegisterHealthChecks(router, comps)
	routes.RegisterReadyChecks(router, comps)

	// Files
	router.Handle("/static", http.FileServer(http.Dir(FolderFlag.Value())))

	return router
}
