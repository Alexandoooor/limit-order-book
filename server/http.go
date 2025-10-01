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
	"net/http"

	"github.com/gorilla/mux"
)

var Logger *log.Logger

type Server struct {
	addr 	string
	ob   	*engine.OrderBook
}

type PlaceOrderRequest struct {
	Side  string `json:"side"`
	Price int    `json:"price"`
	Size  int    `json:"size"`
}

func NewServer(addr string, ob *engine.OrderBook) *Server {
	return &Server{
		addr: addr,
		ob: ob,
	}
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func (s *Server) Serve() error {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		s.ob.RestoreOrderBook()
		view := engine.BuildOrderBookView(s.ob)
		tmpl := template.Must(template.New("index").Parse(web.IndexTemplate()))
		tmpl.Execute(w, view)
	})

	r.HandleFunc("/headers", headers)

	r.HandleFunc("/ob", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, s.ob.String())
	})

	r.HandleFunc("/api/order", func(w http.ResponseWriter, r *http.Request) {
		s.ob.RestoreOrderBook()
		placeOrder(w, r, s.ob)
	})

	r.HandleFunc("/api/wipe", func(w http.ResponseWriter, r *http.Request) {
		err := s.ob.ResetOrderBook()
		if err != nil {
			json.NewEncoder(w).Encode(map[string]bool{"ok": false})
		} else {
			json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		}
	})

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	http.Handle("/", r)

	return http.ListenAndServe(s.addr, nil)

}

func placeOrder(w http.ResponseWriter, r *http.Request, ob *engine.OrderBook) {
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
}
