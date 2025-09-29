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


	db := engine.SetupDB()
	defer db.Close()
	storage := engine.SqlStorage{Database: db}
	// storage := engine.JsonStorage{}
	ob := engine.NewOrderBook()
	ob.AddStorage(&storage)
	ob.RestoreOrderBook()

	logger.Printf("LimitOrderBook running on http://%s\n", addr)
	server := server.NewServer(addr, ob)
	if err := server.Serve(); err != nil {
		logger.Fatal(err)
	}
}
