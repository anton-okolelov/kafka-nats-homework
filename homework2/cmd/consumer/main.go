package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"homework2/internal"

	"github.com/twmb/franz-go/pkg/kgo"
)

const topic = "clicks_v1"

func main() {
	ctx := context.Background()
	seeds := []string{"localhost:10002", "localhost:10003", "localhost:10004"}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("mygroup"),
		kgo.ConsumeTopics(topic),
	)

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err != nil {
		panic(err)
	}

	for {
		fetches := client.PollRecords(ctx, 1)

		fetches.EachRecord(func(r *kgo.Record) {
			// unmarshall r.Value
			var object internal.Click
			err := json.Unmarshal(r.Value, &object)
			if err != nil {
				log.Error("Error", slog.String("error", err.Error()))
			}

			log.Info("Recieved message", slog.Any("object", object))
		})
	}

}
