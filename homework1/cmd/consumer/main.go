package main

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

func main() {
	// connect to nats
	port := "4222"
	natsURL := "nats://nats:" + port

	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Printf("Error connecting to NATS: %v\n", err)
	}

	_, err = nc.QueueSubscribe("foo", "mygroup", func(m *nats.Msg) {
		fmt.Printf("Консюмер получил сообщение: %s\n", string(m.Data))
		m.Ack()
	})

	if err != nil {
		fmt.Printf("Error subscribing to NATS: %v\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()

}
