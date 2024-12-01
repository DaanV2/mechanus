package modron

import (
	"context"

	"github.com/DaanV2/mechanus/server/internal/logging"
)

// Primus is the highest ranked godlike modron that manages and command all in their serivce
type Primus struct {
	appCtx context.Context

	*Tridrone
	*Quadrone
}

func NewPrimus(ctx context.Context) *Primus {
	logger := logging.From(ctx)

	return &Primus{
		appCtx:   ctx,
		Tridrone: NewTridrone(logger.WithPrefix("state")),
		Quadrone: NewQuadrone(logger.WithPrefix("events")),
	}
}

func (p *Primus) Context() context.Context {
	return p.appCtx
}

func (p *Primus) Done() <-chan struct{} {
	return p.appCtx.Done()
}

// Add checks if the given task matches any interface: [HealthCheck], [ReadyCheck], [BeforeStart], [AfterStart], [BeforeTermination], [AfterTermination].
// If they do, its add to the specific collection managing those interfaces
func (p *Primus) Add(task any) bool {
	return p.Tridrone.Add(task) || p.Quadrone.Add(task)
}

// Remove works the same as [Primus.Add] but with the intention of removing the item from the collection
func (p *Primus) Remove(task any) bool {
	return p.Tridrone.Remove(task) || p.Quadrone.Remove(task)
}
