package config

import "github.com/spf13/pflag"

type MDNSConfig struct {
	IPV4 Flag[bool]
	IPV6 Flag[bool]
}

var MDNS = &MDNSConfig{
	IPV4: Bool("mdns.ipv4.enabled", true, "Whenever or not to use ipv4 for mdns"),
	IPV6: Bool("mdns.ipv6.enabled", false, "Whenever or not to use ipv6 for mdns"),
}

func (c *MDNSConfig) AddToSet(set *pflag.FlagSet) {
	c.IPV4.AddToSet(set)
	c.IPV6.AddToSet(set)
}
