package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	apiUrl   = "http://localhost:8080/cotacao"
	filename = "cotacao.txt"
)

type QuotationResponse struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiUrl, nil)
	select {
	case <-ctx.Done():
		fmt.Println("Timeout ao buscar a cotação")
		return
	default:
	}

	if err != nil {
		fmt.Printf("Error creating request: %v\n", err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err.Error())
		return
	}
	var response QuotationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}
	err = appendToFile(filename, fmt.Sprintf("Dólar: %.2f\n", response.Bid))
	if err != nil {
		fmt.Println("Erro saving quotation on file:", err.Error())
		return
	}
	fmt.Println("Quotation saved!")
}

func appendToFile(fileName string, line string) error {
	content, err := os.ReadFile(fileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content = append(content, []byte(line)...)
	err = os.WriteFile(fileName, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
