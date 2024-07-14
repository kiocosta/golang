package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	USDBRL struct {
		Bid json.Number `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", getDollarExchangeRateHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}

func createTableIfNotExists(db *sql.DB) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS exchange_rates (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		"exchange_rate" VARCHAR(10) NOT NULL
	);`

	_, err := db.Exec(createTableSQL)
	return err
}

func InsertExchangeRate(db *sql.DB, exchangeRate string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO exchange_rates VALUES(NULL,?)", exchangeRate)
	if err != nil {
		return err
	}
	return nil
}

func getDollarExchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	exchangeRate, err := getDollarExchangeRate()
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = createTableIfNotExists(db)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = InsertExchangeRate(db, exchangeRate.USDBRL.Bid.String())
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exchangeRate.USDBRL.Bid.String())
}

func getDollarExchangeRate() (*ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var exchangeRate ExchangeRate
	err = json.Unmarshal(body, &exchangeRate)
	if err != nil {
		return nil, err
	}
	return &exchangeRate, nil
}
