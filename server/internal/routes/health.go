package routes

import (
	"net/http"

	"github.com/DaanV2/mechanus/server/pkg/application"
)

func RegisterHealthChecks(router *http.ServeMux, component application.HealthCheck) {
	// Health Checks
	hc := HealthChecks(component)
	router.Handle("/health", hc)
	router.Handle("/healthz", hc)
}

func RegisterReadyChecks(router *http.ServeMux, component application.ReadyCheck) {
	rc := ReadyChecks(component)
	router.Handle("/ready", rc)
	router.Handle("/readyz", rc)
}

func HealthChecks(component application.HealthCheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := component.HealthCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ReadyChecks(component application.ReadyCheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := component.ReadyCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
