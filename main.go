package main

import (
	"flag"
	"limit-order-book/engine"
	"limit-order-book/server"
	"log"
	"os"
	"strconv"
)

var (
	port = flag.Int("port", 3000, "HTTP port")
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags | log.Lshortfile)

	flag.Parse()
	ob := engine.NewOrderBook()

	addr := ":" + strconv.Itoa(*port)
	logger.Printf("Stock Streamer running on http://localhost%s\n", addr)

	if err := server.Serve(addr, ob, logger); err != nil {
		logger.Fatal(err)
	}
}
