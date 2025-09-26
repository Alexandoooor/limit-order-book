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
	ResetOrderBook(ob *engine.OrderBook) error
	RestoreOrderBook() (*engine.OrderBook, error)
	DumpOrderBook(ob *engine.OrderBook) error
}

type JsonStorage struct {}

type SqlStorage struct {
	Database *sql.DB
}

func (j *JsonStorage) ResetOrderBook(ob *engine.OrderBook) error {
	ob.ResetOrderBook()

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

func (j *JsonStorage) DumpOrderBook(ob *engine.OrderBook) error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	dto := ob.ToDTO()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}


func (s *SqlStorage) ResetOrderBook(ob *engine.OrderBook) error {
	return nil
}

func (s *SqlStorage) RestoreOrderBook() (*engine.OrderBook, error) {
	levelDTO, err := getLevels(s.Database)
	Logger.Printf("Levels: %+v", levelDTO)
	if err != nil {
		return nil, err
	}

	orderDTO, err := getAllOrders(s.Database)
	Logger.Printf("Orders: %+v", orderDTO)
	if err != nil {
		return nil, err
	}

	tradeDTO, err := getAllTrades(s.Database)
	Logger.Printf("Trades: %+v", tradeDTO)
	if err != nil {
		return nil, err
	}


	obDTO := engine.OrderBookDTO {
		Levels: levelDTO,
		Orders: orderDTO,
		Trades: tradeDTO,

	}
	return obDTO.ToOrderBook(), nil
}

func (s *SqlStorage) DumpOrderBook(ob *engine.OrderBook) error {
	dto := ob.ToDTO()
	db := s.Database

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	insertLevels(db, dto)
	insertOrders(db, dto)
	deleteOrders(db, ob.DeletedOrders) // TBD FIX DB
	insertTrades(db, dto)
	return tx.Commit()
}

func SetupDB() *sql.DB {
	db, err := sql.Open("sqlite3", "orderbook.db")
	if err != nil {
		Logger.Fatal(err)
	}
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
		INSERT OR REPLACE INTO levels (side, price, volume, count)
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
		_, err := db.Exec(`INSERT OR REPLACE INTO level_orders (level_side, level_price, order_id) VALUES (?, ?)`,
			rowid, oid.String(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertOrders(db *sql.DB, dto *engine.OrderBookDTO) {
	for _, order := range dto.Orders {
		insertOrder(db, order)
	}
}

func insertOrder(db *sql.DB, o *engine.OrderDTO) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO orders (id, side, size, remaining, price, time, next_id, prev_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.Id.String(), o.Side, o.Size, o.Remaining, o.Price, o.Time.Format(time.RFC3339),
		uuidToString(o.NextID), uuidToString(o.PrevID),
	)
	return err
}

func deleteOrders(db *sql.DB, do []uuid.UUID) {
	for _, deletedOrderID := range do {
		deleteOrder(db, deletedOrderID)
	}
}

func deleteOrder(db *sql.DB, id uuid.UUID) error {
	// Get side + price before deleting
	row := db.QueryRow(`
		SELECT o.side, o.price
		FROM orders o
		WHERE o.id = ?`, id.String())

	var side string
	var price int
	if err := row.Scan(&side, &price); err != nil {
		return err
	}

	// Delete from level_orders
	if _, err := db.Exec(`DELETE FROM level_orders WHERE order_id = ?`, id.String()); err != nil {
		return err
	}

	// Delete from orders
	if _, err := db.Exec(`DELETE FROM orders WHERE id = ?`, id.String()); err != nil {
		return err
	}

	// Check if level still has any orders
	row = db.QueryRow(`
		SELECT COUNT(*)
		FROM level_orders
		WHERE level_side = ? AND level_price = ?`, side, price)

	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}

	// If no orders left → remove the level
	if count == 0 {
		_, err := db.Exec(`DELETE FROM levels WHERE side = ? AND price = ?`, side, price)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteLevels(db *sql.DB, dl map[engine.Side]int) error {
	for side, price := range dl {
		_, err := db.Exec(`DELETE FROM levels WHERE side = ? AND price = ?`, side, price)
		return err
	}
	return nil
}

func insertTrades(db *sql.DB, dto *engine.OrderBookDTO) {
	for _, trade := range dto.Trades {
		insertTrade(db, &trade)
	}
}

func insertTrade(db *sql.DB, t *engine.Trade) error {
	_, err := db.Exec(`
		INSERT INTO trades (id, buy_order_id, sell_order_id, price, size, time)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.BuyOrderID.String(), t.SellOrderID.String(), t.Price, t.Size, t.Time.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return err
}

func getLevels(db *sql.DB) (map[engine.Side]map[int]*engine.LevelDTO, error) {
	rows, err := db.Query(`SELECT side, price, volume, count FROM levels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	book := map[engine.Side]map[int]*engine.LevelDTO{
		engine.Buy:  {},
		engine.Sell: {},
	}

	for rows.Next() {
		var sideInt int
		var l engine.LevelDTO
		if err := rows.Scan(&sideInt, &l.Price, &l.Volume, &l.Count); err != nil {
			return nil, err
		}

		orderRows, err := db.Query(`SELECT order_id FROM level_orders WHERE level_side = ? AND level_price = ?`, sideInt, l.Price)
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

func getTrade(db *sql.DB, id uuid.UUID) (*engine.Trade, error) {
	row := db.QueryRow(`
		SELECT id, buy_order_id, sell_order_id, price, size, time
		FROM trades
		WHERE id = ?`, id.String())

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
		var idStr string
		if err := rows.Scan(&idStr); err != nil {
			return nil, err
		}
		t, err := getTrade(db, uuid.MustParse(idStr))
		if err != nil {
			return nil, err
		}
		trades = append(trades, *t)
	}

	return trades, nil
}
