package servers

import "context"

type Server interface {
	Listen()
	Shutdown(ctx context.Context)
}