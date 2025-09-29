package engine

import (
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	ResetOrderBook() error
	RestoreOrderBook() (*OrderBook, error)
	InsertLevel(side Side, l *LevelDTO) error
	InsertTrade(t *Trade) error
	InsertOrder(o *OrderDTO) error
	DeleteOrder(ob *OrderBookDTO, o *OrderDTO) error
	UpdateOrder(ob *OrderBookDTO, o *OrderDTO) error
}

type NilStorage struct {}

type JsonStorage struct {}

type SqlStorage struct {
	Database *sql.DB
}

func (n *NilStorage) InsertLevel(side Side, l *LevelDTO) error {
	return nil
}

func (n *NilStorage) InsertTrade(t *Trade) error {
	return nil
}

func (n *NilStorage) InsertOrder(o *OrderDTO) error {
	return nil

}
func (n *NilStorage) DeleteOrder(ob *OrderBookDTO, o *OrderDTO) error {
	return nil
}

func (n *NilStorage) UpdateOrder(ob *OrderBookDTO, o *OrderDTO) error {
	return nil
}

func (n *NilStorage) ResetOrderBook() error {
	return nil
}

func (n *NilStorage) RestoreOrderBook() (*OrderBook, error) {
	return NewOrderBook(), nil
}

func (j *JsonStorage) InsertLevel(side Side, l *LevelDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Levels[side][l.Price] = l
	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) InsertTrade(t *Trade) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Trades = append(dto.Trades, *t)
	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) InsertOrder(o *OrderDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Orders[o.Id] = o

	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) DeleteOrder(ob *OrderBookDTO, o *OrderDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}

	delete(dto.Orders, o.Id)
	parentLevel := dto.Levels[o.Side][o.Price]
	if parentLevel != nil {
		parentLevel.Count--
		if parentLevel.Count <= 0 {
			delete(dto.Levels[o.Side], o.Price)
		}
	}

	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) UpdateOrder(ob *OrderBookDTO, o *OrderDTO) error {
	return j.DeleteOrder(ob, o)
}

func (j *JsonStorage) WriteDTOToJson(dto *OrderBookDTO) error {
	filename := j.getFilename()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (j *JsonStorage) getDTO() (*OrderBookDTO, error) {
	filename := j.getFilename()
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dto *OrderBookDTO
	err = json.Unmarshal(data, &dto)
	if err != nil {
		return &OrderBookDTO{
			Levels: map[Side]map[int]*LevelDTO{Buy: {}, Sell: {}},
			Orders: make(map[uuid.UUID]*OrderDTO),
			Trades: []Trade{},
		}, nil
	}
	return dto, nil
}

func (j *JsonStorage) getFilename() string {
	orderBookFile := os.Getenv("ORDERBOOK")
	if orderBookFile == "" {
		orderBookFile = "/tmp/orderbook.json"
	}
	return orderBookFile
}


func (j *JsonStorage) ResetOrderBook() error {
	filename := j.getFilename()
	Logger.Printf("Wiping orders from %s\n", filename)
	err := os.WriteFile(filename, []byte("[]"), 0644)

	if err != nil {
		Logger.Printf("Failed to reset OrderBook: %s", err)
		return err
	}

	return nil
}

func (j *JsonStorage) RestoreOrderBook() (*OrderBook, error) {
	filename := j.getFilename()
	data, err := os.ReadFile(filename)
	if err != nil {
		_, err := os.Create(filename)
		if err != nil {
			Logger.Fatal(err)
		}
	}
	data, err = os.ReadFile(filename)

	var dto OrderBookDTO
	err = json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}
	restoredBook := dto.ToOrderBook()

	return restoredBook, nil
}

func (j *JsonStorage) DumpOrderBook(ob *OrderBook) error {
	filename := j.getFilename()

	dto := ob.ToDTO()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}


func (s *SqlStorage) ResetOrderBook() error {
	if _, err := s.Database.Exec(`DELETE FROM levels`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(`DELETE FROM level_orders`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(`DELETE FROM orders`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(`DELETE FROM trades`); err != nil {
		return err
	}

	return nil
}

func (s *SqlStorage) RestoreOrderBook() (*OrderBook, error) {
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


	obDTO := OrderBookDTO {
		Levels: levelDTO,
		Orders: orderDTO,
		Trades: tradeDTO,
	}

	return obDTO.ToOrderBook(), nil
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

func (s *SqlStorage) InsertLevel(side Side, l *LevelDTO) error {
	_, err := s.Database.Exec(`
		INSERT INTO levels (side, price, volume, count)
		VALUES (?, ?, ?, ?)`,
		side, l.Price, 0, 0, //SQL-trigger takes care of updating volume and count when inserting an order
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqlStorage) InsertOrder(o *OrderDTO) error {
	_, err := s.Database.Exec(`
		INSERT INTO orders (id, side, size, remaining, price, time, next_id, prev_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		o.Id.String(), o.Side, o.Size, o.Remaining, o.Price, o.Time.Format(time.RFC3339),
		uuidToString(o.NextID), uuidToString(o.PrevID),
	)
	if err != nil {
		return err
	}

	_, err = s.Database.Exec(`INSERT INTO level_orders (level_side, level_price, order_id) VALUES (?, ?, ?)`,
		o.Side, o.Price, o.Id,
	)
	if err != nil {
		return err
	}



	return nil
}

func (s *SqlStorage) DeleteOrder(ob *OrderBookDTO, o *OrderDTO) error {
	row := s.Database.QueryRow(`
		SELECT o.side, o.price
		FROM orders o
		WHERE o.id = ?`, o.Id.String())

	var side string
	var price int
	if err := row.Scan(&side, &price); err != nil {
		return err
	}

	if _, err := s.Database.Exec(`DELETE FROM level_orders WHERE order_id = ?`, o.Id.String()); err != nil {
		return err
	}

	if _, err := s.Database.Exec(`DELETE FROM orders WHERE id = ?`, o.Id.String()); err != nil {
		return err
	}

	row = s.Database.QueryRow(`
		SELECT COUNT(*)
		FROM level_orders
		WHERE level_side = ? AND level_price = ?`, side, price)

	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		_, err := s.Database.Exec(`DELETE FROM levels WHERE side = ? AND price = ?`, side, price)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlStorage) UpdateOrder(ob *OrderBookDTO, o *OrderDTO) error {
	_, err := s.Database.Exec(`
		UPDATE orders
		SET remaining = ?
		WHERE id = ?`,
		o.Remaining, o.Id.String(),
	)

	if err != nil {
		return err
	}

	if o.Remaining <= 0 {
		if _, err := s.Database.Exec(`
			DELETE FROM level_orders
			WHERE level_side = ? AND level_price = ? AND order_id = ?`,
			o.Side, o.Price, o.Id.String(),
		); err != nil {
			return err
		}

		row := s.Database.QueryRow(`
			SELECT COUNT(*) FROM level_orders
			WHERE level_side = ? AND level_price = ?`, o.Side, o.Price)
		var count int
		if err := row.Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := s.Database.Exec(`
				DELETE FROM levels
				WHERE side = ? AND price = ?`,
				o.Side, o.Price,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SqlStorage) InsertTrade(t *Trade) error {
	_, err := s.Database.Exec(`
		INSERT INTO trades (id, buy_order_id, sell_order_id, price, size, time)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.BuyOrderID.String(), t.SellOrderID.String(), t.Price, t.Size, t.Time.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return err
}

func getLevels(db *sql.DB) (map[Side]map[int]*LevelDTO, error) {
	rows, err := db.Query(`SELECT side, price, volume, count FROM levels`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	book := map[Side]map[int]*LevelDTO{
		Buy:  {},
		Sell: {},
	}

	for rows.Next() {
		var sideInt int
		var l LevelDTO
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

		side := Side(sideInt)
		if book[side] == nil {
			book[side] = make(map[int]*LevelDTO)
		}
		book[side][l.Price] = &l
	}

	return book, nil
}

func getOrder(db *sql.DB, id uuid.UUID) (*OrderDTO, error) {
	row := db.QueryRow(`
		SELECT id, side, size, remaining, price, time, next_id, prev_id
		FROM orders
		WHERE id = ?`, id.String())

	var o OrderDTO
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

func getAllOrders(db *sql.DB) (map[uuid.UUID]*OrderDTO, error) {
	rows, err := db.Query(`SELECT id FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make(map[uuid.UUID]*OrderDTO)
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

func getTrade(db *sql.DB, id uuid.UUID) (*Trade, error) {
	row := db.QueryRow(`
		SELECT id, buy_order_id, sell_order_id, price, size, time
		FROM trades
		WHERE id = ?`, id.String())

	var t Trade
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

func getAllTrades(db *sql.DB) ([]Trade, error) {
	rows, err := db.Query(`SELECT id FROM trades`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []Trade
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
