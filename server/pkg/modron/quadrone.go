package modron

import "github.com/charmbracelet/log"

var (
	_ BeforeStart       = &Quadrone{}
	_ AfterStart        = &Quadrone{}
	_ BeforeTermination = &Quadrone{}
	_ AfterTermination  = &Quadrone{}
)

type Quadrone struct {
	logger *log.Logger

	BeforeStarts       *DuoDrone[BeforeStart]
	AfterStarts        *DuoDrone[AfterStart]
	BeforeTerminations *DuoDrone[BeforeTermination]
	AfterTerminations  *DuoDrone[AfterTermination]
}

func NewQuadrone(logger *log.Logger) *Quadrone {
	return &Quadrone{
		logger: logger,

		BeforeStarts:       NewDuoDrone[BeforeStart](),
		AfterStarts:        NewDuoDrone[AfterStart](),
		BeforeTerminations: NewDuoDrone[BeforeTermination](),
		AfterTerminations:  NewDuoDrone[AfterTermination](),
	}
}

func (drone *Quadrone) Add(task any) bool {
	added := false
	if v, ok := task.(BeforeStart); ok {
		drone.BeforeStarts.Add(v)
		added = true
	}
	if v, ok := task.(AfterStart); ok {
		drone.AfterStarts.Add(v)
		added = true
	}
	if v, ok := task.(BeforeTermination); ok {
		drone.BeforeTerminations.Add(v)
		added = true
	}
	if v, ok := task.(AfterTermination); ok {
		drone.AfterTerminations.Add(v)
		added = true
	}
	return added
}

func (drone *Quadrone) Remove(task any) bool {
	removed := false
	if v, ok := task.(BeforeStart); ok {
		removed = drone.BeforeStarts.Remove(v) || removed
	}
	if v, ok := task.(AfterStart); ok {
		removed = drone.AfterStarts.Remove(v) || removed
	}
	if v, ok := task.(BeforeTermination); ok {
		removed = drone.BeforeTerminations.Remove(v) || removed
	}
	if v, ok := task.(AfterTermination); ok {
		removed = drone.AfterTerminations.Remove(v) || removed
	}
	return removed
}

func (drone *Quadrone) AfterTermination() error {
	drone.logger.Debug("executing: after termination tasks")
	return drone.AfterTerminations.RangeE(func(task AfterTermination) error {
		return task.AfterTermination()
	})
}

// BeforeTermination implements BeforeTermination.
func (drone *Quadrone) BeforeTermination() error {
	drone.logger.Debug("executing: before termination tasks")
	return drone.BeforeTerminations.RangeE(func(task BeforeTermination) error {
		return task.BeforeTermination()
	})
}

// AfterStart implements AfterStart.
func (drone *Quadrone) AfterStart() error {
	drone.logger.Debug("executing: after start tasks")
	return drone.AfterStarts.RangeE(func(task AfterStart) error {
		return task.AfterStart()
	})
}

// BeforeStart implements BeforeStart.
func (drone *Quadrone) BeforeStart() error {
	drone.logger.Debug("executing: before start tasks")
	return drone.BeforeStarts.RangeE(func(task BeforeStart) error {
		return task.BeforeStart()
	})
}