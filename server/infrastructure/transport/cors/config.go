package cors

import (
	"github.com/DaanV2/mechanus/server/infrastructure/config"
)

var (
	CorsConfig         = config.New("server.cors")
	OriginsFlag        = CorsConfig.StringArray("server.cors.allowed-origins", []string{"*"}, "The origins that are allowed to be used by requesters, if empty will skip this header. Allowed strings are matched via prefix check")
	AllowLocalHostFlag = CorsConfig.Bool("server.cors.allow-localhost", true, "Whenever or not as an origin, localhost are allowed")
)

type CORSConfig struct {
	AllowedOrigins []string
	AllowLocalHost bool
}

func GetCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: OriginsFlag.Value(),
		AllowLocalHost: AllowLocalHostFlag.Value(),
	}
}
