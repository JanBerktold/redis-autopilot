package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Hello")
		<-time.After(5 * time.Second)
	}
}
