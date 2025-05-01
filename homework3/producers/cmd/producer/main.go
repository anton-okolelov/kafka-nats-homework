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

const topicClicks = "clicks_v1"
const topicViews = "views_v1"

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

	go func() {
		for {
			// клик по рандомному  advertising_id (от 0 до 9)
			produceClicks(ctx, log, client)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			// просмотр по рандомному advertising_id (от 0 до 9)
			produceViews(ctx, log, client)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Wait for Ctrl+C
	select {}
}

func produceViews(ctx context.Context, log *slog.Logger, client *kgo.Client) {
	object := internal.View{
		AdvertisingId: rand.Int31n(10),
		TimeStamp:     time.Now().Unix(),
	}

	data, _ := json.Marshal(object)

	record := &kgo.Record{Topic: topicViews, Value: data}

	log.Info("producing view", slog.Any("message", object))
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Error("record had a produceViews error while synchronously producing", slog.String("error", err.Error()))
	}
}

func produceClicks(ctx context.Context, log *slog.Logger, client *kgo.Client) {
	object := internal.Click{
		AdvertisingId: rand.Int31n(10),
		TimeStamp:     time.Now().Unix(),
	}

	data, _ := json.Marshal(object)

	record := &kgo.Record{Topic: topicClicks, Value: data}

	log.Info("producing click", slog.Any("message", object))
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
		log.Error("record had a produceClicks error while synchronously producing", slog.String("error", err.Error()))
	}
}
