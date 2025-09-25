package health

import (
	"context"
	"net/http"
)

type (
	HealthCheck interface {
		HealthCheck(ctx context.Context) error
	}

	ReadyCheck interface {
		ReadyCheck(ctx context.Context) error
	}
)

func RegisterHealthChecks(router *http.ServeMux, component HealthCheck) {
	// Health Checks
	hc := HealthChecks(component)
	router.Handle("/health", hc)
	router.Handle("/healthz", hc)
}

func RegisterReadyChecks(router *http.ServeMux, component ReadyCheck) {
	rc := ReadyChecks(component)
	router.Handle("/ready", rc)
	router.Handle("/readyz", rc)
}

func HealthChecks(component HealthCheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := component.HealthCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func ReadyChecks(component ReadyCheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := component.ReadyCheck(r.Context())
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
