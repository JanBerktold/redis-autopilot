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

	configurationFilePath := os.Args[1]

	configManager := NewConfigurationManager()
	config, changeSignal, err := configManager.Load(configurationFilePath)
	if err != nil {
		log.Panicf("Failed to load runtime configuration file: %v", err)
	}

	fmt.Printf("%+v \n", *config)

	redisInstanceProvider := NewRedisInstanceProvider(func() string {
		return config.RedisUrl
	}, changeSignal)

	watcher, _ := NewRedisWatcher(redisInstanceProvider)

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	pilot := &masterWithSlavesPilot{}

	for {
		select {
		case redisStatus := <-watcher.ChangeChannel():
			fmt.Printf("redis status %v\n", redisStatus)
			pilot.Execute()
		case <-time.After(5 * time.Second):
			fmt.Println("Hello")
			pilot.Execute()
		case signal := <-interruptChannel:
			fmt.Printf("We were asked to shutdown %v\n", signal)
			break
		}
	}
}

type Pilot interface {
	Execute() error
	Shutdown()
}

type masterWithSlavesPilot struct {
	redisInstanceGetter func() RedisInstance
}

func (pilot *masterWithSlavesPilot) Execute() error {
	instance := pilot.redisInstanceGetter()

	redisStatus := instance.Status()
	if redisStatus != RedisStatusReady {
		// make sure we're not in consul
	}

	return nil
}
