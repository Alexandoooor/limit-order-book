package util

import (
	"encoding/json"
	"log"
	"os"
	"limit-order-book/engine"
)

var Logger *log.Logger

func ResetOrderBook(ob *engine.OrderBook) {
	ob.ResetOrderBook()

	ordersFile := os.Getenv("ORDERS")
	if ordersFile == "" {
		ordersFile = "/tmp/orderbook.json"
	}
	Logger.Printf("Wiping orders from %s\n", ordersFile)
	err := os.WriteFile(ordersFile, []byte("[]"), 0644)
	if err != nil {
		Logger.Printf("Failed to reset OrderBook: %s", err)
	}
}

func RestoreOrderBook() (*engine.OrderBook, error) {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dto2 engine.OrderBookDTO
	err = json.Unmarshal(data, &dto2)
	if err != nil {
		return nil, err
	}
	restoredBook := dto2.ToOrderBook()

	return restoredBook, nil
}

func DumpOrderBook(ob *engine.OrderBook) error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	dto := ob.ToDTO()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
