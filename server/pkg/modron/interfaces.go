package modron

import (
	"context"
)

type (
	DoneNotifier interface {
		Done() <-chan struct{}
	}

	ContextCarrier interface {
		Context() context.Context
	}
)
