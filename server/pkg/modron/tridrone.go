package modron

import "github.com/charmbracelet/log"

var (
	_ HealthCheck = &Tridrone{}
	_ ReadyCheck  = &Tridrone{}
)

type Tridrone struct {
	logger *log.Logger

	HealthChecks       *DuoDrone[HealthCheck]
	ReadyChecks        *DuoDrone[ReadyCheck]
	BeforeTerminations *DuoDrone[BeforeTermination]
	AfterTerminations  *DuoDrone[AfterTermination]
}

func NewTridrone(logger *log.Logger) *Tridrone {
	return &Tridrone{
		logger: logger,

		HealthChecks: NewDuoDrone[HealthCheck](),
		ReadyChecks:  NewDuoDrone[ReadyCheck](),
	}
}

func (drone *Tridrone) Add(task any) bool {
	added := false
	if v, ok := task.(HealthCheck); ok {
		drone.HealthChecks.Add(v)
		added = true
	}
	if v, ok := task.(ReadyCheck); ok {
		drone.ReadyChecks.Add(v)
		added = true
	}
	return added
}

func (drone *Tridrone) Remove(task any) bool {
	removed := false
	if v, ok := task.(HealthCheck); ok {
		removed = drone.HealthChecks.Remove(v) || removed
	}
	if v, ok := task.(ReadyCheck); ok {
		removed = drone.ReadyChecks.Remove(v) || removed
	}
	return removed
}

// HealthCheck implements HealthCheck.
func (drone *Tridrone) HealthCheck() error {
	drone.logger.Debug("executing: health checks")
	return drone.HealthChecks.RangeE(func(task HealthCheck) error {
		return task.HealthCheck()
	})
}

// ReadyCheck implements ReadyCheck.
func (drone *Tridrone) ReadyCheck() error {
	drone.logger.Debug("executing: ready checks")
	return drone.ReadyChecks.RangeE(func(task ReadyCheck) error {
		return task.ReadyCheck()
	})
}
