package main

import (
	"flag"
	"limit-order-book/engine"
	"limit-order-book/server"
	"limit-order-book/util"
	"strconv"
)

var (
	port      = flag.Int("port", 3000, "HTTP port")
)

func main() {
	flag.Parse()
	addr := "localhost:" + strconv.Itoa(*port)

	logger := util.SetupLogging()
	engine.Logger = logger
	server.Logger = logger
	util.Logger = logger

	ob := engine.NewOrderBook()

	db := util.SetupDB()
	defer db.Close()
	storage := util.SqlStorage{Database: db}
	restoredOrderBook, err := storage.RestoreOrderBook()
	if err != nil {
		logger.Printf("Failed to restore OrderBook from storage. Continue with new OrderBook. %s", err)
	} else {
		ob = restoredOrderBook
	}

	logger.Printf("LimitOrderBook running on http://%s\n", addr)
	server := server.NewServer(addr, ob, &storage)
	if err := server.Serve(); err != nil {
		logger.Fatal(err)
	}
}
