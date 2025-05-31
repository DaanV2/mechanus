package config

import (
	"sync"

	xsync "github.com/DaanV2/mechanus/server/pkg/extensions/sync"
	"github.com/spf13/pflag"
)

var (
	flags   = pflag.NewFlagSet("global", pflag.ContinueOnError)
	manager = &ConfigManager{}
)

type ConfigManager struct {
	configs sync.Map
}

func New(name string) *Config {
	c := &Config{
		name: name,
		data: xsync.NewMap[string, BaseFlag](),
	}

	manager.configs.Store(name, c)

	return c
}

func Get(name string) *Config {
	item, ok := manager.configs.Load(name)
	if !ok {
		panic("no such config object exists: " + name)
	}

	c, ok := item.(*Config)
	if !ok {
		panic("config item couldn't be converted to type: " + name)
	}

	return c
}

