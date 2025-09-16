package main

import (
	"flag"
	"fmt"
	"limit-order-book/engine"
	"limit-order-book/server"
	"log"
	"strconv"
)

var (
	port = flag.Int("port", 3000, "HTTP port")
)

func main() {
	flag.Parse()

	ob := engine.NewOrderBook()

	// id := ob.AddOrder(engine.Buy, 88, 1)
	// ob.PrintOrder(id)

	// for x := 88; x <= 92; x++ {
	// 	ob.ProcessOrder(engine.Buy, x, 1)
	// 	fmt.Println(ob)
	// }
	//
	// // add existing level
	// id := ob.ProcessOrder(engine.Buy, 90, 1)
	// ob.PrintOrder(id)
	fmt.Println(ob)
	// ob.PrintOrderBook()
	//
	// ob.ProcessOrder(engine.Sell, 999, 834)
	// ob.PrintOrderBook()

	addr := ":" + strconv.Itoa(*port)
	log.Printf("Stock Streamer running on http://localhost%s\n", addr)

	if err := server.Serve(addr, ob); err != nil {
		log.Fatal(err)
	}
}
