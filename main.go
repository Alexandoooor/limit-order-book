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
	port      = flag.Int("port", 3000, "HTTP port")
)

func main() {
	flag.Parse()
	addr := ":" + strconv.Itoa(*port)

	output := os.Stdout
	if logFile := os.Getenv("LOGFILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		output = f
		log.Printf("LimitOrderBook running on http://localhost%s\n", addr)
		log.Printf("Logging to file: %s\n", logFile)
	} else {
		log.Println("Logging to Stdout")
	}
	logger := log.New(output, "", log.LstdFlags|log.Lshortfile)
	engine.Logger = logger
	server.Logger = logger

	ob := engine.NewOrderBook()

	err := ob.LoadTrades()
	if err != nil {
		logger.Fatal(err)
	}
	err = ob.LoadOrders()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("LimitOrderBook running on http://localhost%s\n", addr)
	if err := server.Serve(addr, ob); err != nil {
		logger.Fatal(err)
	}
}
