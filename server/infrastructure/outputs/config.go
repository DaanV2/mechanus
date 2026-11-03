// Package outputs holds what the output drivers in its subpackages have in common. Each integration
// keeps its own settings (credentials, bridge / broker addresses) in a config set named
// integrations.<integration>, next to its driver.
package outputs

import (
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	// IntegrationsConfigSet is the configuration set for settings shared by all integrations.
	IntegrationsConfigSet = config.New("integrations")
	ApplyTimeoutFlag      = IntegrationsConfigSet.Duration("integrations.apply-timeout", time.Second*5, "How long a device gets to process an instruction before it counts as failed. 0 means no limit")
)

// Config holds the settings shared by all integrations.
type Config struct {
	// How long a device gets to process an instruction.
	ApplyTimeout time.Duration
}

// GetConfig retrieves the settings shared by all integrations.
func GetConfig() Config {
	return Config{
		ApplyTimeout: ApplyTimeoutFlag.Value(),
	}
}
