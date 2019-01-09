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
	config, _, err := configManager.Load(configurationFilePath)
	if err != nil {
		log.Panicf("Failed to load runtime configuration file: %v", err)
	}

	fmt.Printf("%+v \n", *config)

	redisHealthCheckChannel := redisInstanceHealthCheckLoop()
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	for {
		fmt.Println("Hello")
		<-time.After(5 * time.Second)

		select {
		case redisStatus := <-redisHealthCheckChannel:
			fmt.Printf("redis status %v\n", redisStatus)
		case signal := <-interruptChannel:
			fmt.Printf("We were asked to shutdown %v\n", signal)
			break
		}
	}
}

func redisInstanceHealthCheckLoop() <-chan bool {

	go func() {

	}()

	return make(chan bool)
}