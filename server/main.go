package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL       = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	databaseFile = "./cotacoes.db"
)

type QuotationResponse struct {
	USDBRL Quotation
}

type Quotation struct {
	Bid float64 `json:"bid"`
}

func main() {
	http.HandleFunc("/cotacao", quotationHandler)
	http.ListenAndServe(":8080", nil)
}

func quotationHandler(w http.ResponseWriter, r *http.Request) {
	cotacao, err := getQuotationFromApi()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = persistOnDatabase(cotacao)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao)
}

func getQuotationFromApi() (*Quotation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return nil, fmt.Errorf("timeout ao buscar a cotação")
		default:
			return nil, fmt.Errorf("erro ao buscar a cotação")
		}
	}
	defer resp.Body.Close()
	var response QuotationResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o json de resposta: %w", err)
	}
	return &response.USDBRL, nil
}

func persistOnDatabase(quotation *Quotation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

	defer cancel()
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return fmt.Errorf("erro ao abrir o banco de dados: %w", err)
	}
	defer db.Close()
	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY AUTOINCREMENT, bid REAL, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return fmt.Errorf("erro ao criar a tabela: %w", err)
	}
	_, err = db.ExecContext(ctx, "INSERT INTO cotacoes (bid) VALUES (?)", quotation.Bid)
	if err != nil {
		return fmt.Errorf("erro ao inserir a cotação: %w", err)
	}
	return nil
}

func (q *Quotation) UnmarshalJSON(data []byte) error {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}
	bidString, ok := jsonData["bid"].(string)
	if !ok {
		return fmt.Errorf("bid não encontrado no formato esperado")
	}
	bid, err := strconv.ParseFloat(bidString, 64)
	if err != nil {
		return err
	}
	q.Bid = bid
	return nil
}
