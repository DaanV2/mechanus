package web

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/application"
)

func NewRouter(comps *application.ComponentManager) *http.ServeMux {
	router := http.NewServeMux()

	// Health Checks
	hc := HealthChecks(comps)
	rc := ReadyChecks(comps)
	router.Handle("/health", hc)
	router.Handle("/healthz", hc)
	router.Handle("/ready", rc)
	router.Handle("/readyz", rc)

	// Files
	router.Handle("/", http.FileServer(http.Dir(FolderFlag.Value())))

	return router
}

func HealthChecks(comps *application.ComponentManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := comps.HealthCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ReadyChecks(comps *application.ComponentManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := comps.ReadyCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
