package mqtt

import (
	"errors"
	"fmt"
	"net/url"
	"slices"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	// MQTTConfigSet is the configuration set for the MQTT integration.
	MQTTConfigSet = config.New("integrations.mqtt")
	BrokerFlag    = MQTTConfigSet.String("integrations.mqtt.broker", "", "Address of the MQTT broker, e.g. tcp://192.168.1.10:1883. Empty disables the MQTT integration")
	ClientIDFlag  = MQTTConfigSet.String("integrations.mqtt.client-id", "mechanus", "Client ID on the MQTT broker, has to be unique per broker")
	UsernameFlag  = MQTTConfigSet.String("integrations.mqtt.username", "", "Username for the MQTT broker")
	PasswordFlag  = MQTTConfigSet.String("integrations.mqtt.password", "", "Password for the MQTT broker")
)

// ErrInvalidBroker is returned when the broker address cannot be used.
var ErrInvalidBroker = errors.New("invalid mqtt broker address, expected something like tcp://192.168.1.10:1883")

// Config holds the settings of the MQTT integration.
type Config struct {
	Broker   string // Address of the broker, empty disables the integration
	ClientID string
	Username string
	Password string
}

// GetConfig retrieves the settings of the MQTT integration.
func GetConfig() Config {
	return Config{
		Broker:   BrokerFlag.Value(),
		ClientID: ClientIDFlag.Value(),
		Username: UsernameFlag.Value(),
		Password: PasswordFlag.Value(),
	}
}

// Enabled returns true when a broker is configured.
func (c *Config) Enabled() bool {
	return c.Broker != ""
}

// Validate checks if the broker address can be used. The address itself is left out of the error
// since it could hold a password.
func (c *Config) Validate() error {
	u, err := url.Parse(c.Broker)
	if err != nil {
		return ErrInvalidBroker
	}
	if !slices.Contains([]string{"tcp", "mqtt", "ssl", "tls", "mqtts", "ws", "wss"}, u.Scheme) || u.Host == "" {
		return fmt.Errorf("%w, got scheme %q", ErrInvalidBroker, u.Scheme)
	}

	return nil
}
