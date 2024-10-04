package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	flags = pflag.NewFlagSet("global", pflag.ContinueOnError)
)

type Flag[T any] interface {
	Value() T
	Name() string
	Description() string
	AddToSet(set *pflag.FlagSet)
}

type infoFlag[T any] struct {
	name        string
	description string
	f           *pflag.Flag
	viperFn     func(name string) T
}

func (in *infoFlag[T]) Name() string {
	return in.name
}

func (in *infoFlag[T]) Description() string {
	return in.description
}

func (in *infoFlag[T]) Value() T {
	return in.viperFn(in.name)
}

func (in *infoFlag[T]) AddToSet(set *pflag.FlagSet) {
	set.AddFlag(in.f)
}

func (in *infoFlag[T]) setup() *infoFlag[T] {
	BindFlag(in.f)

	return in
}

func newFlag[T any](name, description string, viperFn func(string) T) *infoFlag[T] {
	f := &infoFlag[T]{
		name:        name,
		description: description,
		f:           flags.Lookup(name),
		viperFn:     viperFn,
	}

	return f.setup()
}

func Bool(name string, def bool, usage string) Flag[bool] {
	flags.Bool(name, def, usage)
	return newFlag(name, usage, viper.GetBool)
}

func String(name string, def string, usage string) Flag[string] {
	flags.String(name, def, usage)
	return newFlag(name, usage, viper.GetString)
}
