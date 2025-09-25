package util

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"
	"limit-order-book/engine"

	"github.com/google/uuid"
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
	Database *sql.DB
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

	var dto engine.OrderBookDTO
	err = json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}
	restoredBook := dto.ToOrderBook()

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
	levelDTO, err := getLevels(s.Database)
	if err != nil {
		return nil, err
	}

	orderDTO, err := getAllOrders(s.Database)
	if err != nil {
		return nil, err
	}

	tradesDTO, err := getAllTrades(s.Database)
	if err != nil {
		return nil, err
	}

	obDTO := engine.OrderBookDTO {
		Levels: levelDTO,
		Orders: orderDTO,
		Trades: tradesDTO,

	}
	return obDTO.ToOrderBook(), nil
}

func (s *SqlStorage) DumpOrderBook() error {
	dto := s.OrderBook.ToDTO()
	db := s.Database

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	insertLevels(db, dto)
	return tx.Commit()
}

func SetupDB() *sql.DB {
	db, err := sql.Open("sqlite3", "orderbook.db")
	if err != nil {
		Logger.Fatal(err)
	}
	defer db.Close()
	return db
}

func uuidToString(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}

func insertLevels(db *sql.DB, dto *engine.OrderBookDTO) {
	for side := range dto.Levels {
		for _, level := range dto.Levels[side] {
			insertLevel(db, side, level)
		}
	}
}

func insertLevel(db *sql.DB, side engine.Side, l *engine.LevelDTO) error {
	res, err := db.Exec(`
		INSERT INTO levels (side, price, volume, count)
		VALUES (?, ?, ?, ?)`,
		side, l.Price, l.Volume, l.Count,
	)
	if err != nil {
		return err
	}

	rowid, err := res.LastInsertId()
	if err != nil {
		return err
	}

	for _, oid := range l.Orders {
		_, err := db.Exec(`INSERT INTO level_orders (level_rowid, order_id) VALUES (?, ?)`,
			rowid, oid.String(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertOrder(db *sql.DB, o *engine.OrderDTO) error {
	_, err := db.Exec(`
		INSERT INTO orders (id, side, size, remaining, price, time, next_id, prev_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.Id.String(), o.Side, o.Size, o.Remaining, o.Price, o.Time.Format(time.RFC3339),
		uuidToString(o.NextID), uuidToString(o.PrevID),
	)
	return err
}

func insertTrade(db *sql.DB, t *engine.Trade) error {
	_, err := db.Exec(`
		INSERT INTO trades (buy_order_id, sell_order_id, price, size, time)
		VALUES (?, ?, ?, ?, ?)`,
		t.BuyOrderID.String(), t.SellOrderID.String(), t.Price, t.Size, t.Time.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return err
}

func getLevels(db *sql.DB) (map[engine.Side]map[int]*engine.LevelDTO, error) {
	rows, err := db.Query(`SELECT rowid, side, price, volume, count FROM levels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	book := map[engine.Side]map[int]*engine.LevelDTO{
		engine.Buy:  {},
		engine.Sell: {},
	}

	for rows.Next() {
		var rowid int64
		var sideInt int
		var l engine.LevelDTO
		if err := rows.Scan(&rowid, &sideInt, &l.Price, &l.Volume, &l.Count); err != nil {
			return nil, err
		}

		orderRows, err := db.Query(`SELECT order_id FROM level_orders WHERE level_rowid = ?`, rowid)
		if err != nil {
			return nil, err
		}
		for orderRows.Next() {
			var oid string
			if err := orderRows.Scan(&oid); err != nil {
				orderRows.Close()
				return nil, err
			}
			l.Orders = append(l.Orders, uuid.MustParse(oid))
		}
		orderRows.Close()

		side := engine.Side(sideInt)
		if book[side] == nil {
			book[side] = make(map[int]*engine.LevelDTO)
		}
		book[side][l.Price] = &l
	}

	return book, nil
}

func getOrder(db *sql.DB, id uuid.UUID) (*engine.OrderDTO, error) {
	row := db.QueryRow(`
		SELECT id, side, size, remaining, price, time, next_id, prev_id
		FROM orders
		WHERE id = ?`, id.String())

	var o engine.OrderDTO
	var nextID, prevID sql.NullString
	var orderTime string

	err := row.Scan(&o.Id, &o.Side, &o.Size, &o.Remaining, &o.Price, &orderTime, &nextID, &prevID)
	if err != nil {
		return nil, err
	}

	o.Time, _ = time.Parse(time.RFC3339, orderTime)

	if nextID.Valid {
		nid := uuid.MustParse(nextID.String)
		o.NextID = &nid
	}
	if prevID.Valid {
		pid := uuid.MustParse(prevID.String)
		o.PrevID = &pid
	}

	return &o, nil
}

// Fetch all orders
func getAllOrders(db *sql.DB) (map[uuid.UUID]*engine.OrderDTO, error) {
	rows, err := db.Query(`SELECT id FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make(map[uuid.UUID]*engine.OrderDTO)
	for rows.Next() {
		var idStr string
		if err := rows.Scan(&idStr); err != nil {
			return nil, err
		}
		o, err := getOrder(db, uuid.MustParse(idStr))
		if err != nil {
			return nil, err
		}
		orders[uuid.MustParse(idStr)] = o
	}
	return orders, nil
}

// Fetch a trade by ID
func getTrade(db *sql.DB, id int) (*engine.Trade, error) {
	row := db.QueryRow(`
		SELECT id, buy_order_id, sell_order_id, price, size, time
		FROM trades
		WHERE id = ?`, id)

	var t engine.Trade
	var buyID, sellID string
	var ts string

	err := row.Scan(&t.ID, &buyID, &sellID, &t.Price, &t.Size, &ts)
	if err != nil {
		return nil, err
	}

	t.BuyOrderID = uuid.MustParse(buyID)
	t.SellOrderID = uuid.MustParse(sellID)
	t.Time, _ = time.Parse(time.RFC3339, ts)

	return &t, nil
}

// Fetch all trades
func getAllTrades(db *sql.DB) ([]engine.Trade, error) {
	rows, err := db.Query(`SELECT id FROM trades`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []engine.Trade
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		t, err := getTrade(db, id)
		if err != nil {
			return nil, err
		}
		trades = append(trades, *t)
	}

	return trades, nil
}
