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
		return config.RedisUrl
	}, changeSignal, logger)
	if err != nil {
		logger.Panicf("Failed to create RedisInstanceProvider: %v\n.", err.Error())
	}

	watcher, _ := NewRedisWatcher(redisInstanceProvider, logger, func() time.Duration {
		return config.RedisMonitorInterval
	})

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	pilot := NewMasterWithSlavesPilot(redisInstanceProvider, nil)

	for {
		select {
		case redisStatus := <-watcher.ChangeChannel():
			fmt.Printf("redis status %v\n", redisStatus)
			pilot.Execute()
		case <-time.After(4 * time.Second):
			fmt.Println("Hello")
			pilot.Execute()
		case signal := <-interruptChannel:
			fmt.Printf("We were asked to shutdown %v\n", signal)
			break
		}
	}

	pilot.Shutdown()
}
