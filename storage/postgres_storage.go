package storage

import (
	"database/sql"
	"context"
	"fmt"
	"log"
	"os"

	"limit-order-book/engine"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var Logger *log.Logger

type PostgresStorage struct {
	Database *pgx.Conn
}

func (s *PostgresStorage) ResetOrderBook() error {
	ctx := context.Background()

	if _, err := s.Database.Exec(ctx, `DELETE FROM level_orders`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(ctx, `DELETE FROM levels`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(ctx, `DELETE FROM orders`); err != nil {
		return err
	}

	if _, err := s.Database.Exec(ctx, `DELETE FROM trades`); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) RestoreOrderBook() (*engine.OrderBook, error) {
	levelDTO, err := getPostgresLevels(s.Database)
	if err != nil {
		Logger.Printf("Error getting levels from db: %s", err)
		return nil, err
	}

	orderDTO, err := getAllPostgresOrders(s.Database)
	if err != nil {
		Logger.Printf("Error getting orders from db: %s", err)
		return nil, err
	}

	tradeDTO, err := getAllPostgresTrades(s.Database)
	if err != nil {
		Logger.Printf("Error getting trades from db: %s", err)
		return nil, err
	}


	obDTO := engine.OrderBookDTO {
		Levels: levelDTO,
		Orders: orderDTO,
		Trades: tradeDTO,
	}

	return obDTO.ToOrderBook(), nil
}

func InitPostgres() *pgx.Conn {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		Logger.Fatalf("failed to connect to db: %s", err)
	}
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS levels (
		    side INTEGER NOT NULL,
		    price INTEGER NOT NULL,
		    volume INTEGER NOT NULL,
		    count INTEGER NOT NULL,
		    PRIMARY KEY (side, price)
		);
	`)
	if err != nil {
		Logger.Fatalf("failed to create levels table: %s", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
		    id TEXT PRIMARY KEY,
		    side INTEGER NOT NULL,
		    size INTEGER NOT NULL,
		    remaining INTEGER NOT NULL,
		    price INTEGER NOT NULL,
		    time TIMESTAMP NOT NULL,
		    next_id TEXT,
		    prev_id TEXT,
		    CONSTRAINT fk_next FOREIGN KEY (next_id) REFERENCES orders(id),
		    CONSTRAINT fk_prev FOREIGN KEY (prev_id) REFERENCES orders(id)
		);
	`)
	if err != nil {
		Logger.Fatalf("failed to create orders table: %s", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS level_orders (
		    level_side INTEGER NOT NULL,
		    level_price INTEGER NOT NULL,
		    order_id TEXT NOT NULL,
		    PRIMARY KEY(level_side, level_price, order_id),
		    CONSTRAINT fk_level FOREIGN KEY(level_side, level_price) REFERENCES levels(side, price),
		    CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(id)
		);
	`)
	if err != nil {
		Logger.Fatalf("failed to create level_orders table: %s", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS trades (
		    id TEXT PRIMARY KEY,
		    buy_order_id TEXT NOT NULL,
		    sell_order_id TEXT NOT NULL,
		    price INTEGER NOT NULL,
		    size INTEGER NOT NULL,
		    time TIMESTAMP NOT NULL
		);
	`)

	if err != nil {
		Logger.Fatalf("failed to create trades table: %s", err)
	}

	return db
}


func (s *PostgresStorage) InsertLevel(side engine.Side, l *engine.LevelDTO) error {
	ctx := context.Background()
	tx, err := s.Database.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO levels (side, price, volume, count)
		VALUES ($1, $2, $3, $4)`,
		side, l.Price, 0, 0, //InsertOrder takes care of updating volume, count
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) InsertOrder(o *engine.OrderDTO) error {
	ctx := context.Background()
	tx, err := s.Database.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, side, size, remaining, price, time, next_id, prev_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		o.Id.String(), o.Side, o.Size, o.Remaining, o.Price, o.Time,
		uuidToString(o.NextID), uuidToString(o.PrevID),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE orders SET next_id = $1 WHERE id = $2`,
		o.Id.String(), uuidToString(o.PrevID),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO level_orders (level_side, level_price, order_id) VALUES ($1, $2, $3)`,
		o.Side, o.Price, o.Id,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE levels SET count = count + 1 WHERE side = $1 AND price = $2`,
		o.Side, o.Price,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE levels SET volume = volume + $1 WHERE side = $2 AND price = $3`,
		o.Remaining, o.Side, o.Price,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) DeleteOrder(ob *engine.OrderBookDTO, o *engine.OrderDTO) error {
	ctx := context.Background()
	tx, err := s.Database.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM level_orders WHERE order_id = $1`, o.Id.String()); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
	       `UPDATE levels SET count = count - 1, volume = volume - $1 WHERE side = $2 AND price = $3`,
	       o.Remaining, o.Side, o.Price,
	); err != nil {
	       return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE orders SET prev_id = $1 WHERE id = $2`,
		uuidToString(o.NextID), uuidToString(o.NextID),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM orders WHERE id = $1`, o.Id.String()); err != nil {
		return err
	}

	row := tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM level_orders
		WHERE level_side = $1 AND level_price = $2`, o.Side, o.Price)

	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		_, err := tx.Exec(ctx, `DELETE FROM levels WHERE side = $1 AND price = $2`, o.Side, o.Price)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) UpdateOrder(ob *engine.OrderBookDTO, o *engine.OrderDTO) error {
	ctx := context.Background()
	tx, err := s.Database.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
	 	SELECT remaining FROM orders
		WHERE id = $1`,
		o.Id.String(),
	)
	var oldRemaining int
	if err := row.Scan(&oldRemaining); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE orders
		SET remaining = $1
		WHERE id = $2`,
		o.Remaining, o.Id.String(),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx,
		`UPDATE levels SET volume = volume + $1 WHERE side = $2 AND price = $3`,
		(o.Remaining - oldRemaining), o.Side, o.Price,
	); err != nil {
		return err
	}

	if o.Remaining <= 0 {
		if _, err := tx.Exec(ctx, `
			DELETE FROM level_orders
			WHERE level_side = $1 AND level_price = $2 AND order_id = $3`,
			o.Side, o.Price, o.Id.String(),
		); err != nil {
			return err
		}

		row := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM level_orders
			WHERE level_side = $1 AND level_price = $2`, o.Side, o.Price)
		var count int
		if err := row.Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			if _, err := tx.Exec(ctx, `
				DELETE FROM levels
				WHERE side = $1 AND price = $2`,
				o.Side, o.Price,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) InsertTrade(t *engine.Trade) error {
	ctx := context.Background()
	tx, err := s.Database.Begin(ctx)

	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO trades (id, buy_order_id, sell_order_id, price, size, time)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.BuyOrderID.String(), t.SellOrderID.String(), t.Price, t.Size, t.Time,
	); err != nil {
		Logger.Printf("Error inserting trade: %s", err)
		return err
	}

	return tx.Commit(ctx)
}

func getPostgresLevels(db *pgx.Conn) (map[engine.Side]map[int]*engine.LevelDTO, error) {
	ctx := context.Background()

	// Single query joins levels with level_orders
	rows, err := db.Query(ctx, `
		SELECT l.side, l.price, l.volume, l.count, lo.order_id
		FROM levels l
		LEFT JOIN level_orders lo
		  ON l.side = lo.level_side AND l.price = lo.level_price
		ORDER BY l.side, l.price
	`)
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
		var price, volume, count int
		var orderID *string

		if err := rows.Scan(&sideInt, &price, &volume, &count, &orderID); err != nil {
			return nil, err
		}

		side := engine.Side(sideInt)

		if book[side] == nil {
			book[side] = make(map[int]*engine.LevelDTO)
		}

		level, exists := book[side][price]
		if !exists {
			level = &engine.LevelDTO{
				Price:  price,
				Volume: volume,
				Count:  count,
				Orders: []uuid.UUID{},
			}
			book[side][price] = level
		}

		if orderID != nil {
			level.Orders = append(level.Orders, uuid.MustParse(*orderID))
		}
	}

	return book, nil
}


func getAllPostgresOrders(db *pgx.Conn) (map[uuid.UUID]*engine.OrderDTO, error) {
	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT id, side, size, remaining, price, time, next_id, prev_id
		FROM orders
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make(map[uuid.UUID]*engine.OrderDTO)

	for rows.Next() {
		var o engine.OrderDTO
		var idStr string
		var nextID, prevID sql.NullString

		if err := rows.Scan(&idStr, &o.Side, &o.Size, &o.Remaining, &o.Price, &o.Time, &nextID, &prevID); err != nil {
			return nil, err
		}

		o.Id = uuid.MustParse(idStr)

		if nextID.Valid {
			nid := uuid.MustParse(nextID.String)
			o.NextID = &nid
		}
		if prevID.Valid {
			pid := uuid.MustParse(prevID.String)
			o.PrevID = &pid
		}

		orders[o.Id] = &o
	}

	return orders, nil
}

func getAllPostgresTrades(db *pgx.Conn) ([]engine.Trade, error) {
	ctx := context.Background()
	rows, err := db.Query(ctx, `
		SELECT id, buy_order_id, sell_order_id, price, size, time
		FROM trades
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []engine.Trade

	for rows.Next() {
		var t engine.Trade
		var buyID, sellID string

		if err := rows.Scan(&t.ID, &buyID, &sellID, &t.Price, &t.Size, &t.Time); err != nil {
			return nil, err
		}

		t.BuyOrderID = uuid.MustParse(buyID)
		t.SellOrderID = uuid.MustParse(sellID)

		trades = append(trades, t)
	}

	return trades, nil
}

func uuidToString(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}
