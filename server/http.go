package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"limit-order-book/engine"
	"limit-order-book/web"
	"net/http"
	"log"
)


func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

type PlaceOrderRequest struct {
	Side  string	`json:"side"`
	Price int	`json:"price"`
	Size  int	`json:"size"`
}

func Serve(addr string, ob *engine.OrderBook, logger *log.Logger) error {
    	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		view := ob.BuildOrderBookView()
		tmpl := template.Must(template.New("index").Parse(web.IndexTemplate()))
		tmpl.Execute(w, view)

	})

	http.HandleFunc("/headers", headers)

	http.HandleFunc("/ob", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, ob.String())
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

		// Trim whitespace to avoid hidden CR/LF issues
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

		// orderBook := ob.String()
		logger.Println(ob)
		logger.Println(ob.GetLevel(side, req.Price))
	})

	return http.ListenAndServe(addr, nil)

}
