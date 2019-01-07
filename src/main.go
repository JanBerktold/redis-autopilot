package main // import "github.com/janberktold/redis-autopilot"

import (
	"fmt"
	"os"
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

	for {
		fmt.Println("Hello")
		<-time.After(5 * time.Second)
	}
}

