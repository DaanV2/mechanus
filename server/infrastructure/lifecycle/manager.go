package lifecycle

import (
	"context"
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/health"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
)

type (
	AfterInitialize interface {
		// AfterInitialize
		AfterInitialize(ctx context.Context) error
	}

	AfterShutDown interface {
		// BeforeShutdown is called after a shutdown is inititated
		AfterShutDown(ctx context.Context) error
	}

	BeforeShutdown interface {
		// BeforeShutdown is called before a shutdown is inititated
		BeforeShutdown(ctx context.Context) error
	}

	ShutdownCleanup interface {
		// ShutdownCleanup is called as the very last call for anything to close
		ShutdownCleanup(ctx context.Context) error
	}
)

type Manager struct {
	afterInitialize *xsync.Slice[AfterInitialize]
	aftershutdown   *xsync.Slice[AfterShutDown]
	beforeshutdown  *xsync.Slice[BeforeShutdown]
	healthcheck     *xsync.Slice[health.HealthCheck]
	readycheck      *xsync.Slice[health.ReadyCheck]
	shutdownCleanup *xsync.Slice[ShutdownCleanup]
}

func NewManager() *Manager {
	return &Manager{
		afterInitialize: xsync.NewSlice[AfterInitialize](),
		aftershutdown:   xsync.NewSlice[AfterShutDown](),
		beforeshutdown:  xsync.NewSlice[BeforeShutdown](),
		healthcheck:     xsync.NewSlice[health.HealthCheck](),
		readycheck:      xsync.NewSlice[health.ReadyCheck](),
		shutdownCleanup: xsync.NewSlice[ShutdownCleanup](),
	}
}

func Register[T any](c *Manager, v T) T {
	c.Add(v)

	return v
}

func (m *Manager) AfterInitialize(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.afterInitialize.RangeErrorCollect(func(item AfterInitialize) error {
		return item.AfterInitialize(ctx)
	})
}

func (m *Manager) AfterShutDown(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.aftershutdown.RangeErrorCollect(func(item AfterShutDown) error {
		return item.AfterShutDown(ctx)
	})
}

func (m *Manager) BeforeShutdown(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.beforeshutdown.RangeErrorCollect(func(item BeforeShutdown) error {
		return item.BeforeShutdown(ctx)
	})
}

func (m *Manager) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.healthcheck.RangeErrorCollect(func(item health.HealthCheck) error {
		return item.HealthCheck(ctx)
	})
}

func (m *Manager) ReadyCheck(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.readycheck.RangeErrorCollect(func(item health.ReadyCheck) error {
		return item.ReadyCheck(ctx)
	})
}

func (m *Manager) ShutdownCleanup(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	return m.shutdownCleanup.RangeErrorCollect(func(item ShutdownCleanup) error {
		return item.ShutdownCleanup(ctx)
	})
}

func (m *Manager) Add(components ...any) {
	for _, comp := range components {
		m.add(comp)
	}
}

func (m *Manager) Remove(components ...any) {
	for _, comp := range components {
		m.remove(comp)
	}
}

func (m *Manager) add(component any) {
	if v, ok := component.(AfterShutDown); ok {
		m.aftershutdown.Append(v)
	}
	if v, ok := component.(AfterInitialize); ok {
		m.afterInitialize.Append(v)
	}
	if v, ok := component.(BeforeShutdown); ok {
		m.beforeshutdown.Append(v)
	}
	if v, ok := component.(health.HealthCheck); ok {
		m.healthcheck.Append(v)
	}
	if v, ok := component.(health.ReadyCheck); ok {
		m.readycheck.Append(v)
	}
	if v, ok := component.(ShutdownCleanup); ok {
		m.shutdownCleanup.Append(v)
	}
}

func (m *Manager) remove(component any) {
	if v, ok := component.(AfterShutDown); ok {
		m.aftershutdown.Remove(func(item AfterShutDown) bool {
			return item == v
		})
	}
	if v, ok := component.(AfterInitialize); ok {
		m.afterInitialize.Remove(func(item AfterInitialize) bool {
			return item == v
		})
	}
	if v, ok := component.(BeforeShutdown); ok {
		m.beforeshutdown.Remove(func(item BeforeShutdown) bool {
			return item == v
		})
	}
	if v, ok := component.(health.HealthCheck); ok {
		m.healthcheck.Remove(func(item health.HealthCheck) bool {
			return item == v
		})
	}
	if v, ok := component.(health.ReadyCheck); ok {
		m.readycheck.Remove(func(item health.ReadyCheck) bool {
			return item == v
		})
	}
	if v, ok := component.(ShutdownCleanup); ok {
		m.shutdownCleanup.Remove(func(item ShutdownCleanup) bool {
			return item == v
		})
	}
}
