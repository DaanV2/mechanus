package modron

type (
	HealthCheck interface {
		HealthCheck() error
	}

	ReadyCheck interface {
		ReadyCheck() error
	}

	BeforeStart interface {
		BeforeStart() error
	}

	AfterStart interface {
		AfterStart() error
	}

	BeforeTermination interface {
		BeforeTermination() error
	}

	AfterTermination interface {
		AfterTermination() error
	}
)
