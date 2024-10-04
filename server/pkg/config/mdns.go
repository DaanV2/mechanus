package config

import (
	"runtime"

	"github.com/spf13/pflag"
)

type MDNS struct {
	IPV4 bool `mapstructure:"ipv4.enabled"`
	IPV6 bool `mapstructure:"ipv6.enabled"`
}

const (
	MDNS_IPV4_ENABLED = "mdns.ipv4.enabled"
	MDNS_IPV6_ENABLED = "mdns.ipv6.enabled"
)

func MDNSFlags(flags *pflag.FlagSet) {
	ipv6 := true
	if runtime.GOOS == "windows" {
		ipv6 = false
	}

	flags.Bool("mdns.ipv4.enabled", true, "Whenever or not to use ipv4 for mdns")
	flags.Bool("mdns.ipv6.enabled", ipv6, "Whenever or not to use ipv6 for mdns")
}
