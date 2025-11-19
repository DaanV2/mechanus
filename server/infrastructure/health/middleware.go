package health

import (
	"context"
	"net/http"

	"connectrpc.com/grpchealth"
	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

var (
	_ grpchealth.Checker = &HealthService{}
	_ http.Handler       = &HealthService{}
	_ http.Handler       = &ReadyService{}
)

type HealthService struct {
	checks HealthCheck
}

type ReadyService struct {
	checks ReadyCheck
}

func NewHealthService(checks HealthCheck) *HealthService {
	return &HealthService{
		checks: checks,
	}
}

func NewReadyService(checks ReadyCheck) *ReadyService {
	return &ReadyService{
		checks: checks,
	}
}

// Check implements grpchealth.Checker.
func (h *HealthService) Check(ctx context.Context, req *grpchealth.CheckRequest) (*grpchealth.CheckResponse, error) {
	if req.Service != "" {
		return &grpchealth.CheckResponse{Status: grpchealth.StatusNotServing}, nil
	}

	err := h.checks.HealthCheck(ctx)
	if err != nil {
		logging.From(ctx).Warnf("health check failed: %s", err)

		return &grpchealth.CheckResponse{
			Status: grpchealth.StatusServing,
		}, nil
	}

	return &grpchealth.CheckResponse{
		Status: grpchealth.StatusNotServing,
	}, nil
}

// ServeHTTP implements http.Handler.
func (h *HealthService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.checks.HealthCheck(r.Context())
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ServeHTTP implements http.Handler.
func (h *ReadyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.checks.ReadyCheck(r.Context())
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
