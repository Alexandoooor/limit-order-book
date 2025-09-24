package util

import (
	"log"
	"os"
)

var Logger *log.Logger

func WipeOrderBook() {
	ordersFile := os.Getenv("ORDERS")
	if ordersFile == "" {
		ordersFile = "/tmp/orderbook.json"
	}
	Logger.Printf("Wiping orders from %s\n", ordersFile)
	err := os.WriteFile(ordersFile, []byte("[]"), 0644)
	if err != nil {
		panic(err)
	}
}
