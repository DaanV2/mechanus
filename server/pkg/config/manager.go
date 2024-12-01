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

type Config struct {
	name string
	data *xsync.Map[string, BaseFlag]
}

func (c *Config) AddToSet(set *pflag.FlagSet) {
	for _, d := range c.data.Items() {
		d.AddToSet(set)
	}
}

func (c *Config) MustLoad(name string) BaseFlag {
	i, ok := c.data.Load(name)
	if !ok {
		panic("couldn't find " + name + " from " + c.name)
	}

	return i
}

func (c *Config) Bool(name string, def bool, usage string) Flag[bool] {
	f := Bool(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetBool(name string) bool {
	return getValue[bool](c, name)
}

func (c *Config) String(name string, def string, usage string) Flag[string] {
	f := String(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetString(name string) string {
	return getValue[string](c, name)
}

func (c *Config) Int(name string, def int, usage string) Flag[int] {
	f := Int(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetInt(name string) int {
	return getValue[int](c, name)
}

func getValue[T any](c *Config, name string) T {
	f := c.MustLoad(name)

	v, ok := f.(Flag[T])
	if !ok {
		panic("item " + name + " from " + c.name + " is not of type bool but: " + f.Type())
	}

	return v.Value()
}
