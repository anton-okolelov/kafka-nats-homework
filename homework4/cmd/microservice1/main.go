package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"homework4/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/jackc/pgx/v4/pgxpool"
)

type AddTransactionRequest struct {
	AccountId       int64   `json:"account_id"`
	TransactionId   string  `json:"transaction_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
}

var ErrTransactionAlreadyExists = fmt.Errorf("transaction already exists")

func main() {

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()

	db, err := initDb(ctx)
	if err != nil {
		log.Error("Error initializing database", slog.String("error", err.Error()))
		return
	}
	defer db.Close()

	go outboxSender(ctx, log, db)

	r := chi.NewRouter()
	r.Post("/transactions", func(w http.ResponseWriter, r *http.Request) {

		var request AddTransactionRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"message": "Invalid request"})
			return
		}

		err = tryCreateTransaction(r.Context(), db, request)
		if errors.Is(err, ErrTransactionAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, map[string]string{"message": "Transaction already exists"})
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"message": "Internal server error"})
			return
		}

		render.JSON(w, r, map[string]string{"message": "Transaction created"})
	})

	log.Info("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Error("Error starting server", slog.String("error", err.Error()))
		return
	}

	log.Info("Server stopped")
}

func initDb(ctx context.Context) (*pgxpool.Pool, error) {
	db, err := pgxpool.Connect(ctx, "postgres://user:pass@pg1:5432/db")
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS transactions (
		    transaction_id text PRIMARY KEY,
		    account_id bigint NOT NULL,
		    amount numeric NOT NULL,
		    transaction_date timestamp NOT NULL
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("error creating table: %w", err)
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
		CREATE TABLE IF NOT EXISTS outbox (
		    transaction_id text PRIMARY KEY ,
			account_id bigint NOT NULL,
			amount numeric NOT NULL,
			CREATED_AT timestamp DEFAULT now()
		)
	`)
	return db, nil
}

func tryCreateTransaction(ctx context.Context, db *pgxpool.Pool, request AddTransactionRequest) error {
	// check if transaction already exists
	var exists bool
	err := db.QueryRow(ctx, "SELECT EXISTS(SELECT FROM transactions WHERE transaction_id = $1)", request.TransactionId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if transaction exists: %w", err)
	}

	if exists {
		return ErrTransactionAlreadyExists
	}

	// all queries in one transaction
	err = db.BeginFunc(ctx, func(tx pgx.Tx) error {
		_, errTx := tx.Exec(
			ctx,
			`
			INSERT INTO transactions (transaction_id, account_id, amount, transaction_date)
			VALUES ($1, $2, $3, $4)
		`,
			request.TransactionId,
			request.AccountId,
			request.Amount,
			request.TransactionDate,
		)

		if errTx != nil {
			return fmt.Errorf("error inserting transaction: %w", errTx)
		}

		// upsert balance
		_, errTx = tx.Exec(
			ctx,
			`
			INSERT INTO balances (account_id, balance)
			VALUES ($1, $2)
			ON CONFLICT (account_id) DO UPDATE
			SET balance = balances.balance + $2
		`,
			request.AccountId,
			request.Amount,
		)

		if errTx != nil {
			return fmt.Errorf("error updatings balance: %w", errTx)
		}

		// insert into outbox
		_, errTx = tx.Exec(
			ctx,
			`
			INSERT INTO outbox (transaction_id, account_id, amount)
			VALUES ($1, $2, $3)
		`,
			request.TransactionId,
			request.AccountId,
			request.Amount,
		)

		if errTx != nil {
			return fmt.Errorf("error inserting into outbox: %w", errTx)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating transaction: %w", err)
	}

	return nil

}

func outboxSender(ctx context.Context, log *slog.Logger, db *pgxpool.Pool) {
	seeds := []string{"broker1:9092", "broker2:9092", "broker3:9092"}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		log.Error("Error creating kafka client", slog.String("error", err.Error()))
		return
	}
	defer client.Close()

	// in loop get the oldest transaction from outbox. Also wait context done
	for {
		select {
		case <-ctx.Done():
			log.Info("Context done")
			return
		default:
			var transactionId string
			var accountId int64
			var amount float64

			err := db.QueryRow(ctx, `
				SELECT transaction_id, account_id, amount 
					FROM outbox
					ORDER BY created_at
					LIMIT 1`,
			).Scan(&transactionId, &accountId, &amount)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					log.Info("No transactions in outbox")
					break
				}
				log.Error("Error getting transaction from outbox", slog.String("error", err.Error()))
				break
			}

			message, err := json.Marshal(internal.KafkaMessage{
				TransactionID: transactionId,
				AccountID:     accountId,
				AmountChange:  amount,
			})

			record := &kgo.Record{
				Topic: "transactions",
				Key:   []byte(transactionId),
				Value: message,
			}

			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			log.Info("Sending message", slog.String("message", string(message)))

			if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
				log.Error("Error producing message to kafka", slog.String("error", err.Error()))
				break
			}

			// delete from outbox
			_, err = db.Exec(ctx, "DELETE FROM outbox WHERE transaction_id = $1", transactionId)
			if err != nil {
				log.Error("Error deleting transaction from outbox", slog.String("error", err.Error()))
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}
