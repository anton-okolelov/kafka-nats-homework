package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// connect to nats
	port := "4222"
	natsURL := "nats://nats:" + port

	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Println("Error connecting to NATS", err)
		return
	}
	producerID := rand.Intn(100)
	i := 0
	for {
		i++
		message := fmt.Sprintf("Producer Id: %d, message: %d", producerID, i)
		fmt.Println(message)
		err := nc.Publish("foo", []byte(message))
		if err != nil {
			fmt.Println("Error publishing message", err)
			return
		}
		time.Sleep(1 * time.Second)
	}

}
