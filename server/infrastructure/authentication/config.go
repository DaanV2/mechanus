package authentication

import (
	"time"

	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	// AuthConfigSet is the configuration set for authentication settings.
	AuthConfigSet            = config.New("auth")
	AccessTokenDurationFlag  = AuthConfigSet.Duration("auth.access-token-duration", time.Hour, "How long an access token is valid for")
	RefreshTokenDurationFlag = AuthConfigSet.Duration("auth.refresh-token-duration", time.Hour*24*7, "How long a refresh token is valid for, a session that isn't refreshed within this time has to log in again")
	KeyRotationFlag          = AuthConfigSet.Duration("auth.key-rotation", time.Hour*24*30, "How often a new jwt signing key is generated, old keys are removed once every token they signed has expired. 0 disables rotation")
	SignupEnabledFlag        = AuthConfigSet.Bool("auth.signup-enabled", false, "Whenever or not anyone can create an account, when disabled only admins can create accounts")
)

// AuthConfig contains the authentication configuration.
type AuthConfig struct {
	JWT           JWTOptions
	KeyRotation   time.Duration
	SignupEnabled bool
}

// GetAuthConfig reads the authentication configuration from the flags.
func GetAuthConfig() AuthConfig {
	return AuthConfig{
		JWT: JWTOptions{
			TokenDuration:        AccessTokenDurationFlag.Value(),
			RefreshTokenDuration: RefreshTokenDurationFlag.Value(),
		},
		KeyRotation:   KeyRotationFlag.Value(),
		SignupEnabled: SignupEnabledFlag.Value(),
	}
}
