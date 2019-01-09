package main

import (
	"github.com/pkg/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type RuntimeConfiguration struct {
	ConsulUrl string
	RedisUrl string
}

type ConfigurationManager interface {
	Load(path string) (config *RuntimeConfiguration, changeSignal chan struct{}, err error)
}

type configurationManager struct {
	fs afero.Fs
}

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
	v.SetDefault("ConsulUrl", "127.0.0.1")
	v.SetDefault("RedisUrl", "127.0.0.1:6379")
}