package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func getOrCreateFile(filePath string) (*os.File, error) {
	_, error := os.Stat(filePath)

	if errors.Is(error, os.ErrNotExist) {
		f, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		return f, nil
	}

	return os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func main() {
	filePath := "cotacao.txt"
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatal("Server error")
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var exchangeRate string
	json.Unmarshal(body, &exchangeRate)

	f, err := getOrCreateFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString("Dólar: " + exchangeRate + "\n")

	if err != nil {
		log.Fatal(err)
	}
}
