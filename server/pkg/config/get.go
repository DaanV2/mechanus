package config

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

func Get[T any]() T {
	var config T

	if err := viper.Unmarshal(&config); err != nil {
		log.With("error", err).Fatalf("fatal reading config: '%T'", config)
	}

	return config
}

func GetSub[T any](key string) T {
	var config T
	
	if err := viper.UnmarshalKey(key, &config); err != nil {
		log.With("error", err).Fatalf("fatal reading config: '%T' on key %s", config, key)
	}

	return config
}