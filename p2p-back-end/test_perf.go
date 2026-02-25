package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	payload := map[string]interface{}{}
	body, _ := json.Marshal(payload)

	start := time.Now()
	resp, err := http.Post("http://127.0.0.1:8000/v1/budgets/details", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error details: %v\n", err)
	} else {
		defer resp.Body.Close()
		buf, _ := io.ReadAll(resp.Body)
		fmt.Printf("Budget Details: HTTP %d, %d bytes, took %v\n", resp.StatusCode, len(buf), time.Since(start))
	}

	start = time.Now()
	resp2, err := http.Post("http://127.0.0.1:8000/v1/budgets/actuals-transactions", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error actuals: %v\n", err)
	} else {
		defer resp2.Body.Close()
		buf2, _ := io.ReadAll(resp2.Body)
		fmt.Printf("Actuals Details: HTTP %d, %d bytes, took %v\n", resp2.StatusCode, len(buf2), time.Since(start))
	}
}
