package health

import (
	"context"
)

type (
	HealthCheck interface {
		HealthCheck(ctx context.Context) error
	}

	ReadyCheck interface {
		ReadyCheck(ctx context.Context) error
	}
)
