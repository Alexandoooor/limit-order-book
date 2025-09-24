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
	addr := ":" + strconv.Itoa(*port)

	logger := util.SetupLogging()
	engine.Logger = logger
	server.Logger = logger
	util.Logger = logger

	ob := engine.NewOrderBook()
	restoredOrderBook, err := util.RestoreOrderBook()
	if err != nil {
		logger.Printf("Failed to restore OrderBook from storage. Continue with new OrderBook. %s", err)
	}
	ob = restoredOrderBook

	logger.Printf("LimitOrderBook running on http://localhost%s\n", addr)
	server := server.NewServer(addr, ob)
	if err := server.Serve(); err != nil {
		logger.Fatal(err)
	}
}
