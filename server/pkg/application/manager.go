package application

import (
	"context"
	"errors"
	"slices"
	"time"
)

type (
	AfterShutDown interface {
		AfterShutDown(ctx context.Context) error
	}

	BeforeShutdown interface {
		BeforeShutdown(ctx context.Context) error
	}

	HealthCheck interface {
		HealthCheck(ctx context.Context) error
	}

	ReadyCheck interface {
		ReadyCheck(ctx context.Context) error
	}
)

type ComponentManager struct {
	aftershutdown  []AfterShutDown
	beforeshutdown []BeforeShutdown
	healthcheck    []HealthCheck
	readycheck     []ReadyCheck
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{}
}

func (m *ComponentManager) AfterShutDown(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	var err error
	for _, comp := range m.aftershutdown {
		err = errors.Join(err, comp.AfterShutDown(ctx))
	}

	return err
}

func (m *ComponentManager) BeforeShutdown(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	var err error
	for _, comp := range m.beforeshutdown {
		err = errors.Join(err, comp.BeforeShutdown(ctx))
	}

	return err
}

func (m *ComponentManager) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	var err error
	for _, comp := range m.healthcheck {
		err = errors.Join(err, comp.HealthCheck(ctx))
	}

	return err
}

func (m *ComponentManager) ReadyCheck(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*30))
	defer cancel()

	var err error
	for _, comp := range m.readycheck {
		err = errors.Join(err, comp.ReadyCheck(ctx))
	}

	return err
}

func (m *ComponentManager) Add(components ...any) {
	for _, comp := range components {
		m.add(comp)
	}
}

func (m *ComponentManager) add(component any) {
	if v, ok := component.(AfterShutDown); ok {
		m.aftershutdown = append(m.aftershutdown, v)
	}
	if v, ok := component.(BeforeShutdown); ok {
		m.beforeshutdown = append(m.beforeshutdown, v)
	}
	if v, ok := component.(HealthCheck); ok {
		m.healthcheck = append(m.healthcheck, v)
	}
	if v, ok := component.(ReadyCheck); ok {
		m.readycheck = append(m.readycheck, v)
	}
}

func (m *ComponentManager) Remove(components ...any) {
	for _, comp := range components {
		m.remove(comp)
	}
}

func (m *ComponentManager) remove(component any) {
	if v, ok := component.(AfterShutDown); ok {
		m.aftershutdown = slices.DeleteFunc(m.aftershutdown, func(item AfterShutDown) bool {
			return item == v
		})
	}
	if v, ok := component.(BeforeShutdown); ok {
		m.beforeshutdown = slices.DeleteFunc(m.beforeshutdown, func(item BeforeShutdown) bool {
			return item == v
		})
	}
	if v, ok := component.(HealthCheck); ok {
		m.healthcheck = slices.DeleteFunc(m.healthcheck, func(item HealthCheck) bool {
			return item == v
		})
	}
	if v, ok := component.(ReadyCheck); ok {
		m.readycheck = slices.DeleteFunc(m.readycheck, func(item ReadyCheck) bool {
			return item == v
		})
	}
}
