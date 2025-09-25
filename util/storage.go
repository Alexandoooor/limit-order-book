package util

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"limit-order-book/engine"
	_ "github.com/mattn/go-sqlite3"
)

var Logger *log.Logger

type Storage interface {
	ResetOrderBook() error
	RestoreOrderBook() (*engine.OrderBook, error)
	DumpOrderBook() error
}

type JsonStorage struct {
	OrderBook *engine.OrderBook
}

type SqlStorage struct {
	OrderBook *engine.OrderBook
}

func (j *JsonStorage) ResetOrderBook() error {
	j.OrderBook.ResetOrderBook()

	ordersFile := os.Getenv("ORDERS")
	if ordersFile == "" {
		ordersFile = "/tmp/orderbook.json"
	}
	Logger.Printf("Wiping orders from %s\n", ordersFile)
	err := os.WriteFile(ordersFile, []byte("[]"), 0644)

	if err != nil {
		Logger.Printf("Failed to reset OrderBook: %s", err)
		return err
	}

	return nil
}

func (j *JsonStorage) RestoreOrderBook() (*engine.OrderBook, error) {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dto2 engine.OrderBookDTO
	err = json.Unmarshal(data, &dto2)
	if err != nil {
		return nil, err
	}
	restoredBook := dto2.ToOrderBook()

	return restoredBook, nil
}

func (j *JsonStorage) DumpOrderBook() error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	dto := j.OrderBook.ToDTO()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}


func (s *SqlStorage) ResetOrderBook() error {
	return nil
}

func (s *SqlStorage) RestoreOrderBook() (*engine.OrderBook, error) {
	return nil, nil
}

func (s *SqlStorage) DumpOrderBook() error {
	dto := s.OrderBook.ToDTO()

	db, err := sql.Open("sqlite3", "orderbook.db")
	if err != nil {
		Logger.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, _ = tx.Exec("DELETE FROM trades")

	for _, trade := range dto.Trades {
		_, err := tx.Exec(
			"INSERT INTO trades(price, size, time, buyerID, sellerID) VALUES(?, ?, ?, ?, ?)",
			trade.Price, trade.Size, trade.Time, trade.BuyerID, trade.SellerID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
