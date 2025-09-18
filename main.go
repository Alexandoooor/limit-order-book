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
	logToFile = flag.String("logfile", "", "log file")
)

func main() {
	flag.Parse()

	output := os.Stdout
	if *logToFile != "" {
		f, err := os.OpenFile(*logToFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		output = f
	}
	logger := log.New(output, "", log.LstdFlags|log.Lshortfile)
	engine.Logger = logger
	server.Logger = logger

	ob := engine.NewOrderBook()

	addr := ":" + strconv.Itoa(*port)
	log.Printf("LimitOrderBook running on http://localhost%s\n", addr)

	if err := server.Serve(addr, ob); err != nil {
		log.Fatal(err)
	}
}
