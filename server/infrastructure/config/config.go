package config

import (
	"errors"
	"iter"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xsync"
	"github.com/spf13/pflag"
)

type Config struct {
	name       string
	data       *xsync.Map[string, BaseFlag]
	validateFn func(*Config) error
}

func (c *Config) AddToSet(set *pflag.FlagSet) {
	for _, d := range c.data.Items() {
		d.AddToSet(set)
	}
}

func (c *Config) Load(name string) (BaseFlag, error) {
	i, ok := c.data.Load(name)
	if !ok {
		return nil, errors.New("couldn't find " + name + " from " + c.name)
	}

	return i, nil
}

func (c *Config) MustLoad(name string) BaseFlag {
	item, err := c.Load(name)
	if err != nil {
		panic(err)
	}

	return item
}

func (c *Config) All() iter.Seq2[string, BaseFlag] {
	return c.data.Items()
}

func (c *Config) Bool(name string, def bool, usage string) Flag[bool] {
	f := Bool(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetBool(name string) bool {
	return getValue[bool](c, name)
}

func (c *Config) String(name, def, usage string) Flag[string] {
	f := String(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetString(name string) string {
	return getValue[string](c, name)
}

func (c *Config) StringArray(name string, def []string, usage string) Flag[[]string] {
	if def == nil {
		def = []string{}
	}

	f := Strings(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetStringArray(name string) []string {
	return getValue[[]string](c, name)
}

func (c *Config) Int(name string, def int, usage string) Flag[int] {
	f := Int(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetInt(name string) int {
	return getValue[int](c, name)
}

func (c *Config) UInt16(name string, def int, usage string) Flag[uint16] {
	f := UInt16(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetUInt16(name string) uint16 {
	return getValue[uint16](c, name)
}

func (c *Config) Duration(name string, def time.Duration, usage string) Flag[time.Duration] {
	f := Duration(name, def, usage)
	c.data.Store(name, f)

	return f
}

func (c *Config) GetDuration(name string) time.Duration {
	return getValue[time.Duration](c, name)
}

// WithValidate couples the given function as the function used to validate this config object.
// If nill, no checks will be made
func (c *Config) WithValidate(validatefn func(*Config) error) *Config {
	c.validateFn = validatefn

	return c
}

func (c *Config) Validate() error {
	if c.validateFn == nil {
		return nil
	}

	return c.validateFn(c)
}

func getValue[T any](c *Config, name string) T {
	f := c.MustLoad(name)

	v, ok := f.(Flag[T])
	if !ok {
		panic("item " + name + " from " + c.name + " is not of type bool but: " + f.Type())
	}

	return v.Value()
}
