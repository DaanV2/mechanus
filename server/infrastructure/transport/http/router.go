package http

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/infrastructure/health"
	"github.com/charmbracelet/log"
)

func WebRouter(conf ServerConfig, healthChecker health.HealthCheck, readyChecker health.ReadyCheck) *http.ServeMux {
	router := http.NewServeMux()

	health.RegisterHealthChecks(router, healthChecker)
	health.RegisterReadyChecks(router, readyChecker)

	// Files
	log.Debug("serving files from: " + conf.StaticFolder)
	router.Handle("/", http.FileServer(http.Dir(conf.StaticFolder)))

	return router
}
