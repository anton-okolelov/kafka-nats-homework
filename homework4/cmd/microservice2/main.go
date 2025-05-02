package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"homework4/internal"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

var ErrTransactionAlreadyExists = fmt.Errorf("transaction already exists")

func main() {
	ctx := context.Background()
	seeds := []string{"broker1:9092", "broker2:9092", "broker3:9092"}
	client, err := kgo.NewClient(
		kgo.DisableAutoCommit(),
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("mygroup"),
		kgo.ConsumeTopics("transactions"),
	)

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err != nil {
		log.Error("Error creating Kafka client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("Creating db tables...")
	db, err := initDb(ctx)
	if err != nil {
		log.Error("Error initializing database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	log.Info("Kafka client created")
	for {
		fetches := client.PollRecords(ctx, 1)

		fetches.EachRecord(func(r *kgo.Record) {
			// unmarshall r.Value
			var object internal.KafkaMessage
			err := json.Unmarshal(r.Value, &object)
			if err != nil {
				log.Error("Error", slog.String("error", err.Error()))
			}

			err = db.BeginFunc(ctx, func(tx pgx.Tx) error {
				_, txErr := tx.Exec(ctx, "INSERT INTO inbox (transaction_id) VALUES ($1)", object.TransactionID)
				if txErr != nil {
					if pgx.ErrNoRows != nil {
						return ErrTransactionAlreadyExists
					}
					return fmt.Errorf("error inserting: %w", txErr)
				}

				_, txErr = tx.Exec(ctx,
					`INSERT INTO balances (account_id, balance) 
					VALUES ($1, $2) 
						ON CONFLICT (account_id) 
						DO UPDATE SET balance = balances.balance + $2`,
					object.AccountID,
					object.AmountChange,
				)
				if txErr != nil {
					return fmt.Errorf("error updating balance: %w", txErr)
				}

				return nil
			})
			if err != nil {
				if errors.Is(err, ErrTransactionAlreadyExists) {
					log.Info("Transaction already exists", slog.String("transaction_id", object.TransactionID))
					return
				}
				log.Error("Error creating transaction", slog.String("error", err.Error()))
				return
			}

			// commit kafka offset
			err = client.CommitRecords(ctx, r)
			if err != nil {
				log.Error("Error committing record", slog.String("error", err.Error()))
			}

			log.Info("Recieved message", slog.Any("object", object))
		})
	}

}

func initDb(ctx context.Context) (*pgxpool.Pool, error) {
	db, err := pgxpool.Connect(ctx, "postgres://user:pass@pg2:5432/db")
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS balances (
		    account_id bigint PRIMARY KEY,
		    balance numeric NOT NULL                       
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS inbox (
		    transaction_id text PRIMARY KEY 
		)
	`)
	return db, nil
}
