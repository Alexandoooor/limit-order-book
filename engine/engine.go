package engine

import (
	"time"
	"fmt"
    	"github.com/google/uuid"
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
	bids		map[int]*Level
	asks		map[int]*Level
	orders		map[uuid.UUID]*Order
	lowestAsk	*Level
	highestBid	*Level
}

type Order struct {
	id 		uuid.UUID
	side 		Side
	size		int
	remaining	int
	price		int
	time 		time.Time
	nextOrder	*Order
	prevOrder	*Order
	parentLevel	*Level
}

type Level struct {
	price 		int
	volume 		int
	count		int
	nextLevel	*Level
	headOrder	*Order
	tailOrder	*Order
}

func NewOrderBook() *OrderBook {
	ob := &OrderBook{
		bids: make(map[int]*Level),
		asks: make(map[int]*Level),
		orders: make(map[uuid.UUID]*Order),
		lowestAsk: nil,
		highestBid: nil,
	}
	return ob
}

func (ob *OrderBook) NewLevel(order *Order, side Side) *Level {
	newLevel := &Level{
		price: order.price,
		volume: order.size,
		count: 1,
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
		} else if newLevel.price < ob.lowestAsk.price  {
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

func (ob *OrderBook) AddOrder(side Side, price int, size int, remaining int) uuid.UUID {
	newOrder := Order{
		id: 	uuid.New(),
		side: 	side,
		size:	size,
		remaining: remaining,
		price: price,
		time: time.Now().UTC(),
		nextOrder: nil,
		prevOrder: nil,
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

func (ob *OrderBook) ProcessOrder(side Side, price int, size int) {
	remaining := size
	var isOrderBetterThanBestLevel func(int, int)bool
	var currentBestLevel *Level

	if side == Buy {
		isOrderBetterThanBestLevel = func (order int, currentBestLevel int) bool  {
			return order >= currentBestLevel
		}
		// Check if the buy order price is >= the lowestAsk, if it is it can be executed
		currentBestLevel = ob.lowestAsk
	} else {
		isOrderBetterThanBestLevel = func (order int, currentBestLevel int) bool {
			return order <= currentBestLevel
		}
		// Check if the sell order price is <= the highestBid, if it is it can be executed
		currentBestLevel = ob.highestBid
	}

	if currentBestLevel == nil {
		ob.AddOrder(side, price, size, remaining)
	} else {
		// fmt.Printf("%s\n", currentBestLevel)
		// fmt.Printf("Order.price better than currentBestLevel: %t\n", isOrderBetterThanBestLevel(price, currentBestLevel.price))

		var currentLevel *Level
		for true {
			if side == Buy {
				currentLevel = ob.lowestAsk
			} else {
				currentLevel = ob.highestBid
			}
			if currentLevel == nil || !isOrderBetterThanBestLevel(price, currentLevel.price) {
				break
			}
			// existingOrder: An order already in the book and a candidate to be executed
			// against an incoming order
			existingOrder := currentLevel.headOrder
			// fmt.Printf("existingOrder %s\n", existingOrder)
			for existingOrder != nil && remaining > 0 {
				if existingOrder.remaining <= remaining {
					// If the candidate order in the book has <= remaining size than the size of the
					// incoming order, it will be filled and thus removed from the order book
					remaining -= existingOrder.remaining
					existingOrder = ob.RemoveOrder(*existingOrder)
				} else {
					// Otherwise the incoming order will be fully processed and the remaining
					// size of the existing order will remain in the book
					existingOrder.remaining -= remaining
					remaining = 0
				}
			}

			if currentLevel.count == 0 {
				if side == Buy {
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
			ob.AddOrder(side, price, size, remaining)
		}
	}

}


func (o Order) Equals(other Order) bool {
	return o.id == other.id
}

func (ob *OrderBook) PrintOrder(id uuid.UUID) {
	fmt.Printf("%s\n", ob.orders[id])
}

func (ob *OrderBook) PrintOrderBook() {
	fmt.Println("-----------------------------------------------------------------------------------")
	if len(ob.orders) == 0 {
		fmt.Println("OrderBook{}")
	} else {
		fmt.Println("OrderBook{")
		for _, order := range ob.orders {
			fmt.Printf("	%+v\n", order)
		}
		fmt.Println("}")
	}
	fmt.Println("-----------------------------------------------------------------------------------")
}


func (ob *OrderBook) String() string {
	lowestAskPrice := 0
	highestBidPrice := 0

	if ob.lowestAsk != nil {
		lowestAskPrice = ob.lowestAsk.price
	}

	if ob.highestBid != nil {
		highestBidPrice = ob.highestBid.price
	}
	return fmt.Sprintf(
		"OrderBook{\n 	lowestAsk: %+v\n	highestBid: %+v\n}",
		lowestAskPrice,
		highestBidPrice,
	)
}

func (l *Level) String() string {
	return fmt.Sprintf(
		"Level{\nprice: %d\nvolume: %d\ncount: %d\nnextLevel: %+v\nheadOrder: %+v\ntailOrder: %+v\n}\n",
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
        "Order{id: %s, side: %s, size: %d, remaining: %d, price: %d, time: %s, nextOrderId: %s, prevOrderId: %s}",
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
