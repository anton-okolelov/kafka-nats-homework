package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"homework2/internal"

	"github.com/twmb/franz-go/pkg/kgo"
)

const topic = "clicks_v1"

func main() {

	seeds := []string{"localhost:10002", "localhost:10003", "localhost:10004"}

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	log.Info("starting producer")
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	for {
		produce(ctx, log, client)
		time.Sleep(time.Second)
	}
}

func produce(ctx context.Context, log *slog.Logger, client *kgo.Client) {
	object := internal.Click{
		SiteId:     rand.Int31n(10),
		PriceCents: rand.Int31n(1000),
		ClickedAt:  time.Now().Format(time.RFC3339),
	}

	data, _ := json.Marshal(object)

	record := &kgo.Record{Topic: topic, Value: data}

	log.Info("producing message", slog.Any("message", object))
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Error("record had a produce error while synchronously producing", slog.String("error", err.Error()))
	}
}
