package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("Starting...")
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
		message := fmt.Sprintf("Producer Id: %d, request: %d", producerID, i)
		fmt.Println(message)
		reply, err := nc.Request("requestreply", []byte(message), time.Second)
		if err != nil {
			fmt.Println("Error requesting message", err)
			return
		}
		fmt.Printf("Producer %d received reply: %s\n", producerID, string(reply.Data))
		time.Sleep(1 * time.Second)
	}
}
