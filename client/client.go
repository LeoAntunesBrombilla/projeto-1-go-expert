package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type USD struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	Client()
}

func Client() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)

	defer cancel()

	req, err := http.NewRequest("GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		log.Println("Operation failed:", err)
	}

	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("Operation took longer than 300ms")
		} else {
			log.Println("Operation failed:", err)
		}
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var price USD

	err = json.Unmarshal(body, &price)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Bid", price.USDBRL.Bid)

	file, err := os.Create("cotacao.txt")

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	_, err = io.WriteString(file, "Dólar: "+price.USDBRL.Bid+"\n")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
