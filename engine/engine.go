package engine

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
	"os"
)

var Logger *log.Logger

type OrderBook struct {
	levels     map[Side]map[int]*Level
	orders     map[uuid.UUID]*Order
	lowestAsk  *Level
	highestBid *Level
	trades     []Trade
}

type Order struct {
	id          uuid.UUID
	side        Side
	size        int
	remaining   int
	price       int
	time        time.Time
	nextOrder   *Order
	prevOrder   *Order
	parentLevel *Level
}

type Level struct {
	price     int
	volume    int
	count     int
	nextLevel *Level
	headOrder *Order
	tailOrder *Order
}

type Side int

const (
	Buy Side = iota
	Sell
)

func NewOrderBook() *OrderBook {
	levels := map[Side]map[int]*Level{
		Buy:  make(map[int]*Level),
		Sell: make(map[int]*Level),
	}

	ob := &OrderBook{
		levels: levels,
		orders: make(map[uuid.UUID]*Order),
	}

	ob.readTrades()

	return ob
}

func (ob *OrderBook) NewLevel(order *Order, side Side) *Level {
	newLevel := &Level{
		price:     order.price,
		volume:    order.remaining,
		count:     1,
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
		} else if newLevel.price > ob.highestBid.price {
			newLevel.nextLevel = ob.highestBid
			ob.highestBid = newLevel
		} else {
			currentBid := ob.highestBid
			for currentBid.nextLevel != nil && newLevel.price < currentBid.nextLevel.price {
				currentBid = currentBid.nextLevel
			}
			newLevel.nextLevel = currentBid.nextLevel
			currentBid.nextLevel = newLevel
		}
	case Sell:
		if ob.lowestAsk == nil {
			ob.lowestAsk = newLevel
		} else if newLevel.price < ob.lowestAsk.price {
			newLevel.nextLevel = ob.lowestAsk
			ob.lowestAsk = newLevel
		} else {
			currentAsk := ob.lowestAsk
			for currentAsk.nextLevel != nil && newLevel.price > currentAsk.nextLevel.price {
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
		id:        id,
		side:      side,
		size:      size,
		remaining: remaining,
		price:     price,
		time:      time.Now().UTC(),
	}
}

func (ob *OrderBook) AddOrder(order Order) uuid.UUID {
	level, ok := ob.levels[order.side][order.price]
	if ok {
		order.parentLevel = level
		order.prevOrder = level.tailOrder
		level.tailOrder.nextOrder = &order
		level.tailOrder = &order
		level.volume += order.remaining
		level.count++
	} else {
		newLevel := ob.NewLevel(&order, order.side)
		ob.levels[order.side][newLevel.price] = newLevel
	}

	ob.orders[order.id] = &order
	return order.id

}

func (ob *OrderBook) RemoveOrder(order Order) *Order {
	delete(ob.orders, order.id)
	parentLevel := order.parentLevel
	parentLevel.volume -= order.remaining
	parentLevel.count--

	if parentLevel.count > 0 {
		if parentLevel.headOrder.id == order.id {
			parentLevel.headOrder = order.nextOrder
		} else if parentLevel.tailOrder.id == order.id {
			parentLevel.tailOrder = order.prevOrder
		} else {
			A := order.prevOrder
			C := order.nextOrder
			A.nextOrder = C
			C.prevOrder = A
		}
		return parentLevel.headOrder
	} else {
		delete(ob.levels[order.side], order.parentLevel.price)
		if order.side == Buy {
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

	if incomingOrder.side == Buy {
		currentBestLevel = ob.lowestAsk
	} else {
		currentBestLevel = ob.highestBid
	}

	if currentBestLevel == nil {
		ob.AddOrder(incomingOrder)
	} else {
		var currentLevel *Level
		for true {
			if incomingOrder.side == Buy {
				currentLevel = ob.lowestAsk
				if ob.lowestAsk == nil || incomingOrder.price < ob.lowestAsk.price {
					break
				}
			} else {
				currentLevel = ob.highestBid
				if ob.highestBid == nil || incomingOrder.price > ob.highestBid.price {
					break
				}
			}
			existingOrder := currentLevel.headOrder
			for existingOrder != nil && incomingOrder.remaining > 0 {
				if existingOrder.remaining <= incomingOrder.remaining {
					var trade Trade
					if incomingOrder.side == Buy {
						trade = Trade{
							Price:    existingOrder.price,
							Size:     existingOrder.remaining,
							Time:     time.Now().UTC(),
							BuyerID:  incomingOrder.id,
							SellerID: existingOrder.id,
						}
					} else {
						trade = Trade{
							Price:    existingOrder.price,
							Size:     existingOrder.remaining,
							Time:     time.Now().UTC(),
							BuyerID:  existingOrder.id,
							SellerID: incomingOrder.id,
						}
					}
					ob.recordTrade(trade)

					incomingOrder.remaining -= existingOrder.remaining
					existingOrder = ob.RemoveOrder(*existingOrder)
				} else {
					var trade Trade
					if incomingOrder.side == Buy {
						trade = Trade{
							Price:    existingOrder.price,
							Size:     incomingOrder.remaining,
							Time:     time.Now().UTC(),
							BuyerID:  incomingOrder.id,
							SellerID: existingOrder.id,
						}
					} else {
						trade = Trade{
							Price:    existingOrder.price,
							Size:     incomingOrder.remaining,
							Time:     time.Now().UTC(),
							BuyerID:  existingOrder.id,
							SellerID: incomingOrder.id,
						}
					}
					ob.recordTrade(trade)

					existingOrder.parentLevel.volume -= incomingOrder.remaining
					existingOrder.remaining -= incomingOrder.remaining
					incomingOrder.remaining = 0
				}
			}

			if currentLevel.count == 0 {
				delete(ob.levels[incomingOrder.side], currentLevel.price)
				if incomingOrder.side == Buy {
					ob.lowestAsk = currentLevel.nextLevel
					currentLevel = ob.lowestAsk
				} else {
					ob.highestBid = currentLevel.nextLevel
					currentLevel = ob.highestBid
				}
			}

			if incomingOrder.remaining == 0 {
				break
			}
		}

		if incomingOrder.remaining > 0 {
			ob.AddOrder(incomingOrder)
		}
	}
	return incomingOrder.id
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
	return o.id == other.id
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
	ob.dumpTrades()
	Logger.Println(ob.GetTrades())
}

func (ob *OrderBook) readTrades() {
	tradesFile := os.Getenv("TRADES")
	if tradesFile == "" {
		tradesFile = "/tmp/trades.json"
	}

	file, err := os.Open(tradesFile)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&ob.trades); err != nil {
			fmt.Println("Warning: could not decode trades file:", err)
		} else {
			fmt.Printf("Loaded %d trades from %s\n", len(ob.trades), tradesFile)
		}
	} else if !os.IsNotExist(err) {
		fmt.Println("Warning: could not open trades file:", err)
	}

}

func (ob *OrderBook) dumpTrades() error {
	tradesFile := os.Getenv("TRADES")
	if tradesFile == "" {
		tradesFile = "/tmp/trades.json"
	}

	file, err := os.Create(tradesFile)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(ob.trades); err != nil {
		return fmt.Errorf("cannot encode trades: %w", err)
	}

	return nil
}

func (o *Order) String() string {
	var nextID, prevID string

	if o.nextOrder != nil {
		nextID = o.nextOrder.id.String()
	} else {
		nextID = "nil"
	}

	if o.prevOrder != nil {
		prevID = o.prevOrder.id.String()
	} else {
		prevID = "nil"
	}

	return fmt.Sprintf(
		"Order{\n\tid: %s\n\tside: %s\n\tsize: %d\n\tremaining: %d\n\tprice: %d\n\ttime: %s\n\tnextOrderId: %s\n\tprevOrderId: %s\n\t}\n",
		o.id.String(),
		o.side,
		o.size,
		o.remaining,
		o.price,
		o.time.Format(time.RFC3339Nano),
		nextID,
		prevID,
	)
}

func (l *Level) String() string {
	return fmt.Sprintf(
		"Level{\n\tprice: %d\n\tvolume: %d\n\tcount: %d\n\tnextLevel: %+v\n\theadOrder: %+v\n\ttailOrder: %+v\n}\n",
		l.price,
		l.volume,
		l.count,
		l.nextLevel,
		l.headOrder,
		l.tailOrder,
	)
}

func (t *Trade) String() string {
	return fmt.Sprintf(
		"Trade{\n\tprice: %d\n\tsize: %d\n\ttime: %s\n\tbuyerId: %s\n\tsellerId: %s\n\t}\n",
		t.Price,
		t.Size,
		t.Time.Format(time.RFC3339Nano),
		t.BuyerID,
		t.SellerID,
	)
}

func (s Side) String() string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}
