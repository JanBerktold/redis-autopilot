package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// RuntimeConfiguration stores all configuration parameters the user
// is able to pass in. Unless otherwise noted, we support reloading
// these at runtime by watching the configuration file for changes.
type RuntimeConfiguration struct {
	// Base URL for the consul cluster to use. Defaults to 127.0.0.1,
	// which supports the local agent installation model recommended
	// by Hashicorp. Depending on the docker networking mode, this
	// will have to be updated to point at the IP of the host.
	ConsulURL string
	// The address to reach our redis instance at. Defaults to
	// 127.0.0.1:6379 to support the default deployment mode
	// of redis and consul running in the same container.
	RedisAddress             string
	RedisMonitorInterval     time.Duration
	PilotExecuteTimeInterval time.Duration
}

// The ConfigurationManager handles loading, parsing and updating
// the RuntimeConfiguration based off a configuration file which is
// provided by the user.
type ConfigurationManager interface {
	Load(path string) (config *RuntimeConfiguration, changeSignal chan struct{}, err error)
}

type configurationManager struct {
	fs afero.Fs
}

// NewConfigurationManager creates a ConfigurationManager instance
// for real-world use. Should be bypassed for testing purposes.
func NewConfigurationManager() ConfigurationManager {
	return &configurationManager{
		fs: afero.NewOsFs(),
	}
}

func (c *configurationManager) Load(path string) (config *RuntimeConfiguration, changeSignal chan struct{}, err error) {
	v := viper.New()
	v.SetFs(c.fs)

	v.SetConfigType("yaml")

	v.SetConfigFile(path)

	c.setViperDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, nil, errors.Wrap(err, "read configuration")
	}

	config = &RuntimeConfiguration{}
	if err := v.Unmarshal(config); err != nil {
		return nil, nil, errors.Wrap(err, "unmarshal configuration")
	}

	changeSignal = make(chan struct{})

	v.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		changeSignal <- struct{}{}
	})

	return
}

func (c *configurationManager) setViperDefaults(v *viper.Viper) {
	v.SetDefault("ConsulURL", "127.0.0.1")
	v.SetDefault("RedisAddress", "127.0.0.1:6379")
}
