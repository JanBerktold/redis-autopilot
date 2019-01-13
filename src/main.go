package main // import "github.com/janberktold/redis-autopilot"

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) != 2 {
		log.Info("Invoke as redis-autopilot [path to config file]")
		os.Exit(1)
	}

	logger := log.StandardLogger()

	configurationFilePath := os.Args[1]

	configManager := NewConfigurationManager()
	config, changeSignal, err := configManager.Load(configurationFilePath)
	if err != nil {
		logger.Panicf("Failed to load runtime configuration file: %v", err)
	}

	fmt.Printf("%+v \n", *config)

	redisInstanceProvider, err := NewRedisInstanceProvider(func() string {
		return config.RedisAddress
	}, changeSignal, logger)
	if err != nil {
		logger.Panicf("Failed to create RedisInstanceProvider: %v.", err.Error())
	}

	watcher, err := NewRedisWatcher(redisInstanceProvider, logger, func() time.Duration {
		return config.RedisMonitorInterval
	})
	if err != nil {
		logger.Panicf("Failed to create RedisInstanceProvider: %v.", err.Error())
	}

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	pilot := NewMasterWithSlavesPilot(redisInstanceProvider, nil)

	for {
		select {
		case redisStatus := <-watcher.ChangeChannel():
			log.Infof("Interrupt from redis watcher, new status: %v", redisStatus)
			pilot.Execute()
		case <-time.After(4 * time.Second):
			log.Info("Interrupt from time delay")
			pilot.Execute()
		case signal := <-interruptChannel:
			log.Infof("We were asked to shutdown %v", signal)
			break
		}
	}

	pilot.Shutdown()
}
