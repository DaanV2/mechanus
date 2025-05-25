package config

import (
	"time"

	"github.com/DaanV2/mechanus/server/pkg/generics"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type BaseFlag interface {
	Name() string
	Description() string
	Type() string
	AddToSet(set *pflag.FlagSet)
}

type Flag[T any] interface {
	BaseFlag
	Value() T
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

func (in *infoFlag[T]) Type() string {
	return generics.NameOf[T]()
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

func String(name, def, usage string) Flag[string] {
	flags.String(name, def, usage)

	return newFlag(name, usage, viper.GetString)
}

func Int(name string, def int, usage string) Flag[int] {
	flags.Int(name, def, usage)

	return newFlag(name, usage, viper.GetInt)
}

func UInt16(name string, def int, usage string) Flag[uint16] {
	flags.Int(name, def, usage)

	return newFlag(name, usage, viper.GetUint16)
}

func Duration(name string, def time.Duration, usage string) Flag[time.Duration] {
	flags.Duration(name, def, usage)

	return newFlag(name, usage, viper.GetDuration)
}
