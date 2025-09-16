package engine

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"
	"limit-order-book/web"
)

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

type OrderBook struct {
	bids       map[int]*Level
	asks       map[int]*Level
	orders     map[uuid.UUID]*Order
	lowestAsk  *Level
	highestBid *Level
	trades	   []Trade
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

type Trade struct {
	Price    int       `json:"price"`
	Size     int       `json:"size"`
	Time     time.Time `json:"time"`
	BuyerID  uuid.UUID `json:"buyerId"`
	SellerID uuid.UUID `json:"sellerId"`
}

func (ob *OrderBook) BuildOrderBookView() web.OrderBookView {
	view := web.OrderBookView{}

	// Build bids list (highest first)
	for price, level := range ob.bids {
		view.Bids = append(view.Bids, web.LevelView{
			Price: price,
			Volume: level.volume,
		})
	}

	for price, level := range ob.asks {
		view.Asks = append(view.Asks, web.LevelView{
			Price: price,
			Volume: level.volume,
		})
	}

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "unknown"
	}

	view.Hostname = hostname

	return view
}

func NewOrderBook() *OrderBook {
	ob := &OrderBook{
		bids:       make(map[int]*Level),
		asks:       make(map[int]*Level),
		orders:     make(map[uuid.UUID]*Order),
		lowestAsk:  nil,
		highestBid: nil,
	}
	return ob
}

func (ob *OrderBook) NewLevel(order *Order, side Side) *Level {
	newLevel := &Level{
		price:     order.price,
		volume:    order.remaining,
		count:     1,
		nextLevel: nil,
		headOrder: order,
		tailOrder: order,
	}

	order.parentLevel = newLevel
	order.prevOrder = newLevel.tailOrder
	newLevel.tailOrder.nextOrder = order
	newLevel.tailOrder = order

	switch side {
	case Buy:
		// If there is no highestBid yet, i.e no bids at all
		if ob.highestBid == nil {
			ob.highestBid = newLevel
			// If the highestBid is lower than the new level,
			// Set newLevel.nextLevel to highestBid and the highestBid to the current level
		} else if newLevel.price > ob.highestBid.price {
			newLevel.nextLevel = ob.highestBid
			ob.highestBid = newLevel
			// Otherwise we need to traverse the pre-existing levels,
			// to find where the new one fits in.
		} else {
			currentBid := ob.highestBid
			for currentBid.nextLevel != nil && newLevel.price < currentBid.nextLevel.price {
				currentBid = currentBid.nextLevel
			}
			// if the newLevel.price is larger than currentBid.nextLevel.price
			// fit it in between current and next e.g levels 91 89, newLevel is 90:
			// 90 -> 89, 91 -> 90 results in 91 -> 90 -> 89
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

func (ob *OrderBook) AddOrder(id uuid.UUID, side Side, price int, size int, remaining int) uuid.UUID {
	newOrder := Order{
		id:          id,
		side:        side,
		size:        size,
		remaining:   remaining,
		price:       price,
		time:        time.Now().UTC(),
		nextOrder:   nil,
		prevOrder:   nil,
		parentLevel: nil,
	}
	switch side {
	case Buy:
		if ob.bids[price] != nil {
			level := ob.bids[price]
			newOrder.parentLevel = level
			newOrder.prevOrder = level.tailOrder
			level.tailOrder.nextOrder = &newOrder
			level.tailOrder = &newOrder
			level.volume += newOrder.remaining
			level.count++

		} else {
			level := ob.NewLevel(&newOrder, side)
			ob.bids[level.price] = level
		}
	case Sell:
		if ob.asks[price] != nil {
			level := ob.asks[price]
			newOrder.parentLevel = level
			newOrder.prevOrder = level.tailOrder
			level.tailOrder.nextOrder = &newOrder
			level.tailOrder = &newOrder
			level.volume += newOrder.remaining
			level.count++
		} else {
			level := ob.NewLevel(&newOrder, side)
			ob.asks[level.price] = level
		}
	}

	ob.orders[newOrder.id] = &newOrder
	return newOrder.id

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
			// A - B - C => A - C
			A := order.prevOrder
			C := order.nextOrder
			A.nextOrder = C
			C.prevOrder = A
		}
		return parentLevel.headOrder
	} else {
		if order.side == Buy {
			delete(ob.bids, order.parentLevel.price)
			if ob.highestBid == parentLevel {
				ob.highestBid = parentLevel.nextLevel
			}
		} else {
			delete(ob.asks, order.parentLevel.price)
			if ob.lowestAsk == parentLevel {
				ob.lowestAsk = parentLevel.nextLevel
			}

		}
	}
	return nil

}

func (ob *OrderBook) ProcessOrder(incomingSide Side, incomingPrice int, incomingSize int) uuid.UUID {
	incomingOrderId := uuid.New()
	remaining := incomingSize
	trades := []Trade{}
	var isOrderBetterThanBestLevel func(int, int) bool
	var currentBestLevel *Level

	if incomingSide == Buy {
		isOrderBetterThanBestLevel = func(order int, currentBestLevel int) bool {
			return order >= currentBestLevel
		}
		// Check if the buy order price is >= the lowestAsk, if it is it can be executed
		currentBestLevel = ob.lowestAsk
	} else {
		isOrderBetterThanBestLevel = func(order int, currentBestLevel int) bool {
			return order <= currentBestLevel
		}
		// Check if the sell order price is <= the highestBid, if it is it can be executed
		currentBestLevel = ob.highestBid
	}

	if currentBestLevel == nil {
		ob.AddOrder(incomingOrderId, incomingSide, incomingPrice, incomingSize, remaining)
	} else {
		var currentLevel *Level
		for true {
			if incomingSide == Buy {
				currentLevel = ob.lowestAsk
			} else {
				currentLevel = ob.highestBid
			}
			if currentLevel == nil || !isOrderBetterThanBestLevel(incomingPrice, currentLevel.price) {
				break
			}
			// existingOrder: An order already in the book and a candidate to be executed
			// against an incoming order
			existingOrder := currentLevel.headOrder
			for existingOrder != nil && remaining > 0 {
				if existingOrder.remaining <= remaining {
					// If the candidate order in the book has <= remaining size than the size of the
					// incoming order, it will be filled and thus removed from the order book
					var trade Trade
					if incomingSide == Buy {
						trade = Trade{
							Price: incomingPrice,
							Size: existingOrder.remaining,
							Time: time.Now().UTC(),
							BuyerID: incomingOrderId,
							SellerID: existingOrder.id,
						}
					} else {
						trade = Trade{
							Price: incomingPrice,
							Size: existingOrder.remaining,
							Time: time.Now().UTC(),
							BuyerID: existingOrder.id,
							SellerID: incomingOrderId,
						}
					}
					trades = append(trades, trade)
					fmt.Printf("Trades{%+v}", ob.trades)
					remaining -= existingOrder.remaining
					existingOrder = ob.RemoveOrder(*existingOrder)
				} else {
					// Otherwise the incoming order will be fully processed and the remaining
					// size of the existing order will remain in the book
					existingOrder.parentLevel.volume -= remaining
					existingOrder.remaining -= remaining
					remaining = 0
				}
			}

			if currentLevel.count == 0 {
				if incomingSide == Buy {
					delete(ob.asks, currentLevel.price)
					ob.lowestAsk = currentLevel.nextLevel
					currentLevel = ob.lowestAsk
				} else {
					delete(ob.bids, currentLevel.price)
					ob.highestBid = currentLevel.nextLevel
					currentLevel = ob.highestBid
				}
			}

			if remaining == 0 {
				break
			}
		}

		if remaining > 0 {
			ob.AddOrder(incomingOrderId, incomingSide, incomingPrice, incomingSize, remaining)
		}
	}
	ob.trades = append(ob.trades, trades...)
	return incomingOrderId
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

func (ob *OrderBook) CancelOrder(id uuid.UUID) bool {
	order := ob.orders[id]
	if order == nil {
		return false
	}
	ob.RemoveOrder(*order)
	return true
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

func (ob *OrderBook) GetLevel(side Side, price int) *Level {
	if side == Buy {
		if ob.bids[price] != nil {
			return ob.bids[price]
		} else {
			return &Level{}
		}
	} else {
		if ob.asks[price] != nil {
			return ob.asks[price]
		} else {
			return &Level{}
		}
	}
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
