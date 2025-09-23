package engine

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"os"

	"github.com/google/uuid"
)

var Logger *log.Logger

type OrderBook struct {
	levels     map[Side]map[int]*Level
	orders     map[uuid.UUID]*Order
	lowestAsk  *Level
	highestBid *Level
	trades     []Trade
	db         *sql.DB
}

type Side int

const (
	Buy Side = iota
	Sell
)

func (s Side) String() string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}

func NewOrderBook(db *sql.DB) *OrderBook {
	levels := map[Side]map[int]*Level{
		Buy:  make(map[int]*Level),
		Sell: make(map[int]*Level),
	}

	ob := &OrderBook{
		levels: levels,
		orders: make(map[uuid.UUID]*Order),
		db: db,
	}

	return ob
}

func (ob *OrderBook) NewLevel(order *Order, side Side) *Level {
	newLevel := &Level{
		Price:     order.Price,
		Volume:    order.Remaining,
		Count:     1,
		headOrder: order,
		tailOrder: order,
	}

	order.parentLevel = newLevel
	order.prevOrder = newLevel.tailOrder
	newLevel.tailOrder.nextOrder = order
	newLevel.tailOrder = order

	switch side {
	case Buy:
		if ob.highestBid == nil {
			ob.highestBid = newLevel
		} else if newLevel.Price > ob.highestBid.Price {
			newLevel.nextLevel = ob.highestBid
			ob.highestBid = newLevel
		} else {
			currentBid := ob.highestBid
			for currentBid.nextLevel != nil && newLevel.Price < currentBid.nextLevel.Price {
				currentBid = currentBid.nextLevel
			}
			newLevel.nextLevel = currentBid.nextLevel
			currentBid.nextLevel = newLevel
		}
	case Sell:
		if ob.lowestAsk == nil {
			ob.lowestAsk = newLevel
		} else if newLevel.Price < ob.lowestAsk.Price {
			newLevel.nextLevel = ob.lowestAsk
			ob.lowestAsk = newLevel
		} else {
			currentAsk := ob.lowestAsk
			for currentAsk.nextLevel != nil && newLevel.Price > currentAsk.nextLevel.Price {
				currentAsk = currentAsk.nextLevel
			}
			newLevel.nextLevel = currentAsk.nextLevel
			currentAsk.nextLevel = newLevel
		}
	}

	return newLevel
}

func (ob *OrderBook) createOrder(id uuid.UUID, side Side, price int, size int, remaining int) Order {
	return Order{
		Id:        id,
		Side:      side,
		Size:      size,
		Remaining: remaining,
		Price:     price,
		Time:      time.Now().UTC(),
	}
}

func (ob *OrderBook) AddOrder(order Order) uuid.UUID {
	level, ok := ob.levels[order.Side][order.Price]
	if ok {
		order.parentLevel = level
		order.prevOrder = level.tailOrder
		level.tailOrder.nextOrder = &order
		level.tailOrder = &order
		level.Volume += order.Remaining
		level.Count++
	} else {
		newLevel := ob.NewLevel(&order, order.Side)
		ob.levels[order.Side][newLevel.Price] = newLevel
	}

	ob.orders[order.Id] = &order
	ob.DumpOrders()
	return order.Id

}

func (ob *OrderBook) RemoveOrder(order Order) *Order {
	delete(ob.orders, order.Id)
	ob.DumpOrders()
	parentLevel := order.parentLevel
	parentLevel.Volume -= order.Remaining
	parentLevel.Count--

	if parentLevel.Count > 0 {
		if parentLevel.headOrder.Id == order.Id {
			parentLevel.headOrder = order.nextOrder
		} else if parentLevel.tailOrder.Id == order.Id {
			parentLevel.tailOrder = order.prevOrder
		} else {
			A := order.prevOrder
			C := order.nextOrder
			A.nextOrder = C
			C.prevOrder = A
		}
		return parentLevel.headOrder
	} else {
		delete(ob.levels[order.Side], order.parentLevel.Price)
		if order.Side == Buy {
			if ob.highestBid == parentLevel {
				ob.highestBid = parentLevel.nextLevel
			}
		} else {
			if ob.lowestAsk == parentLevel {
				ob.lowestAsk = parentLevel.nextLevel
			}

		}
	}
	return nil

}

func (ob *OrderBook) ProcessOrder(incomingSide Side, incomingPrice int, incomingSize int) uuid.UUID {
	incomingOrderId := uuid.New()
	incomingRemaining := incomingSize
	incomingOrder := ob.createOrder(incomingOrderId, incomingSide, incomingPrice, incomingSize, incomingRemaining)

	var currentBestLevel *Level

	if incomingOrder.Side == Buy {
		currentBestLevel = ob.lowestAsk
	} else {
		currentBestLevel = ob.highestBid
	}

	if currentBestLevel == nil {
		ob.AddOrder(incomingOrder)
	} else {
		var currentLevel *Level
		for true {
			if incomingOrder.Side == Buy {
				currentLevel = ob.lowestAsk
				if ob.lowestAsk == nil || incomingOrder.Price < ob.lowestAsk.Price {
					break
				}
			} else {
				currentLevel = ob.highestBid
				if ob.highestBid == nil || incomingOrder.Price > ob.highestBid.Price {
					break
				}
			}
			existingOrder := currentLevel.headOrder
			for existingOrder != nil && incomingOrder.Remaining > 0 {
				if existingOrder.Remaining <= incomingOrder.Remaining {
					var trade Trade
					if incomingOrder.Side == Buy {
						trade = Trade{
							Price:    existingOrder.Price,
							Size:     existingOrder.Remaining,
							Time:     time.Now().UTC(),
							BuyerID:  incomingOrder.Id,
							SellerID: existingOrder.Id,
						}
					} else {
						trade = Trade{
							Price:    existingOrder.Price,
							Size:     existingOrder.Remaining,
							Time:     time.Now().UTC(),
							BuyerID:  existingOrder.Id,
							SellerID: incomingOrder.Id,
						}
					}
					ob.recordTrade(trade)

					incomingOrder.Remaining -= existingOrder.Remaining
					existingOrder = ob.RemoveOrder(*existingOrder)
				} else {
					var trade Trade
					if incomingOrder.Side == Buy {
						trade = Trade{
							Price:    existingOrder.Price,
							Size:     incomingOrder.Remaining,
							Time:     time.Now().UTC(),
							BuyerID:  incomingOrder.Id,
							SellerID: existingOrder.Id,
						}
					} else {
						trade = Trade{
							Price:    existingOrder.Price,
							Size:     incomingOrder.Remaining,
							Time:     time.Now().UTC(),
							BuyerID:  existingOrder.Id,
							SellerID: incomingOrder.Id,
						}
					}
					ob.recordTrade(trade)

					existingOrder.parentLevel.Volume -= incomingOrder.Remaining
					existingOrder.Remaining -= incomingOrder.Remaining
					incomingOrder.Remaining = 0
				}
			}

			if currentLevel.Count == 0 {
				delete(ob.levels[incomingOrder.Side], currentLevel.Price)
				if incomingOrder.Side == Buy {
					ob.lowestAsk = currentLevel.nextLevel
					currentLevel = ob.lowestAsk
				} else {
					ob.highestBid = currentLevel.nextLevel
					currentLevel = ob.highestBid
				}
			}

			if incomingOrder.Remaining == 0 {
				break
			}
		}

		if incomingOrder.Remaining > 0 {
			ob.AddOrder(incomingOrder)
		}
	}
	return incomingOrder.Id
}

func (ob *OrderBook) CancelOrder(id uuid.UUID) bool {
	order := ob.orders[id]
	if order == nil {
		return false
	}
	ob.RemoveOrder(*order)
	return true
}

func (o Order) Equals(other Order) bool {
	return o.Id == other.Id
}

func (ob *OrderBook) PrintOrder(id uuid.UUID) {
	fmt.Printf("%s\n", ob.orders[id])
}

func (ob *OrderBook) GetOrderBook() string {
	if len(ob.orders) == 0 {
		return ""
	} else {
		s := ""
		for _, order := range ob.orders {
			s += fmt.Sprintf("	%+v\n", order)
		}
		return s
	}
}

func (ob *OrderBook) String() string {
	if len(ob.orders) == 0 {
		return "OrderBook{}"
	} else {
		s := "OrderBook{\n"
		s += ob.GetOrderBook()
		s += "}\n"
		return s
	}
}

func (ob *OrderBook) GetTrades() string {
	if len(ob.trades) == 0 {
		return "Trades{}"
	} else {
		s := "Trades{\n"
		for _, trade := range ob.trades {
			s += fmt.Sprintf("	%+v\n", trade)
		}
		s += "}\n"
		return s
	}
}

func (ob *OrderBook) GetLevel(side Side, price int) *Level {
	val, ok := ob.levels[side][price]
	if ok {
		return val
	}
	return &Level{}
}

func (ob *OrderBook) recordTrade(trade Trade) {
	ob.trades = append(ob.trades, trade)
	ob.DumpTrades()
	err := ob.SaveToDB()
	if err != nil {
		Logger.Fatal(err)
	}
	Logger.Println(ob.GetTrades())
}

type jsonOrderBook struct {
	Levels map[Side]map[int]*Level
	Orders map[uuid.UUID]*Order
}

func (ob *OrderBook) DumpOrders() error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}

	job := jsonOrderBook{
		Orders: ob.orders,
		Levels: ob.levels,
	}

	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (ob *OrderBook) LoadOrders() error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var job jsonOrderBook
	if err := json.Unmarshal(data, &job); err != nil {
		return err
	}

	ob.levels = job.Levels
	ob.orders = job.Orders

	return nil
}

func (ob *OrderBook) DumpTrades() error {
	filename := os.Getenv("TRADES")
	if filename == "" {
		filename = "/tmp/trades.json"
	}

	data, err := json.MarshalIndent(ob.trades, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (ob *OrderBook) LoadTrades() error {
	filename := os.Getenv("TRADES")
	if filename == "" {
		filename = "/tmp/trades.json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &ob.trades); err != nil {
		return err
	}

	return nil
}

func (ob *OrderBook) SaveToDB() error {
	tx, err := ob.db.Begin()
	if err != nil {
		return err
	}

	_, _ = tx.Exec("DELETE FROM trades")

	for _, trade := range ob.trades {
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

func (ob *OrderBook) LoadFromDB() error {
	ob.trades = []Trade{}

	rows3, err := ob.db.Query("SELECT price, size, time, buyerID, sellerID FROM trades")
	if err != nil {
		return err
	}
	defer rows3.Close()

	for rows3.Next() {
		var price, size int
		var time time.Time
		var buyerID, sellerID uuid.UUID
		if err := rows3.Scan(&price, &size, &time, &buyerID, &sellerID); err != nil {
			return err
		}
		ob.trades = append(ob.trades, Trade{Price: price, Size: size, Time: time, BuyerID: buyerID, SellerID: sellerID})
	}

	return nil
}
