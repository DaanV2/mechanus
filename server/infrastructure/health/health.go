package health

import (
	"context"
)

type (
	// HealthCheck is an interface for types that can perform health checks.
	HealthCheck interface {
		HealthCheck(ctx context.Context) error
	}

	// ReadyCheck is an interface for types that can perform readiness checks.
	ReadyCheck interface {
		ReadyCheck(ctx context.Context) error
	}
)
