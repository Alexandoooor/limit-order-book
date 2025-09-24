package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"limit-order-book/engine"
	"limit-order-book/web"
	"log"
	"os"
	"net/http"
)

var Logger *log.Logger

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

type PlaceOrderRequest struct {
	Side  string `json:"side"`
	Price int    `json:"price"`
	Size  int    `json:"size"`
}

func Serve(addr string, ob *engine.OrderBook) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ob.DumpOrders()
		ob.DumpTrades()
		view := engine.BuildOrderBookView(ob)
		tmpl := template.Must(template.New("index").Parse(web.IndexTemplate()))
		tmpl.Execute(w, view)
	})

	http.HandleFunc("/headers", headers)

	http.HandleFunc("/wipe", func(w http.ResponseWriter, r *http.Request) {
		tradesFile := os.Getenv("TRADES")
		if tradesFile == "" {
			tradesFile = "/tmp/trades.json"
		}
		Logger.Printf("Wiping trades from %s\n", tradesFile)
		err := os.WriteFile(tradesFile, []byte("[]"), 0644)
		if err != nil {
			panic(err)
		}

		ordersFile := os.Getenv("ORDERS")
		if ordersFile == "" {
			ordersFile = "/tmp/orderbook.json"
		}
		Logger.Printf("Wiping orders from %s\n", ordersFile)
		err = os.WriteFile(ordersFile, []byte("[]"), 0644)
		if err != nil {
			panic(err)
		}
		ob.ResetOrderBook()
	})

	http.HandleFunc("/ob", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, ob.String())
	})

	http.HandleFunc("/api/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		body = bytes.TrimSpace(body)

		var req PlaceOrderRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		var side engine.Side
		switch req.Side {
		case "buy":
			side = engine.Buy
		case "sell":
			side = engine.Sell
		default:
			http.Error(w, "Invalid side, use 'buy' or 'sell'", http.StatusBadRequest)
			return
		}

		order := ob.ProcessOrder(side, req.Price, req.Size)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)

		Logger.Println(ob)
		Logger.Println(ob.GetLevel(side, req.Price))
	})

	return http.ListenAndServe(addr, nil)

}
