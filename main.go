package main

import (
	"context"
	"flag"
	"limit-order-book/engine"
	"limit-order-book/server"
	"limit-order-book/storage"
	"limit-order-book/util"
	"strconv"
)

var (
	port      = flag.Int("port", 3000, "HTTP port")
)

func main() {
	flag.Parse()
	addr := ":" + strconv.Itoa(*port)

	logger := util.SetupLogging()
	engine.Logger = logger
	server.Logger = logger
	storage.Logger = logger


	db := storage.InitPostgres()
	defer db.Close(context.Background())

	storage := storage.PostgresStorage{Database: db}
	ob := engine.NewOrderBook()
	ob.AddStorage(&storage)
	ob.RestoreOrderBook()

	logger.Printf("LimitOrderBook running on http://%s\n", addr)
	server := server.NewServer(addr, ob)
	if err := server.Serve(); err != nil {
		logger.Fatal(err)
	}
}
