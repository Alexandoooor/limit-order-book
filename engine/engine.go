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
	LowestAsk	*Level
	HighestBid	*Level
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
		LowestAsk: nil,
		HighestBid: nil,
	}
	return ob
}

func (ob *OrderBook) NewLevel(order Order, side Side) *Level {
	newLevel := &Level{
		price: order.price,
		volume: order.size,
		count: 1,
		nextLevel: nil,
		headOrder: &order,
		tailOrder: &order,
	}

	switch side {
	case Buy:
		// If there is no highestBid yet, i.e no bids at all
		if ob.HighestBid == nil {
			ob.HighestBid = newLevel
		// If the highestBid is lower than the new level,
		// Set newLevel.nextLevel to highestBid and the highestBid to the current level
		} else if newLevel.price > ob.HighestBid.price {
			newLevel.nextLevel = ob.HighestBid
			ob.HighestBid = newLevel
		// Otherwise we need to traverse the pre-existing levels,
		// to find where the new one fits in.
		} else {
			currentBid := ob.HighestBid
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
		if ob.LowestAsk == nil {
			ob.LowestAsk = newLevel
		} else if newLevel.price < ob.LowestAsk.price  {
			newLevel.nextLevel = ob.LowestAsk
			ob.LowestAsk = newLevel
		} else {
			currentAsk := ob.LowestAsk
			for currentAsk.nextLevel != nil && newLevel.price > currentAsk.nextLevel.price {
				currentAsk = currentAsk.nextLevel
			}
			newLevel.nextLevel = currentAsk.nextLevel
			currentAsk.nextLevel = newLevel
		}
	}

	return newLevel
}

func (ob *OrderBook) AddOrder(side Side, price int, size int) uuid.UUID {
	newOrder := Order{
		id: 	uuid.New(),
		side: 	side,
		size:	size,
		remaining: size,
		price: price,
		time: time.Now().UTC(),
		nextOrder: nil,
		prevOrder: nil,
		parentLevel: nil,
	}
	ob.orders[newOrder.id] = &newOrder

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
			level := ob.NewLevel(newOrder, side)
			ob.bids[level.price] = level
			newOrder.parentLevel = level
			newOrder.prevOrder = level.tailOrder
			level.tailOrder.nextOrder = &newOrder
			level.tailOrder = &newOrder
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
			level := ob.NewLevel(newOrder, side)
			ob.asks[level.price] = level
			newOrder.parentLevel = level
			newOrder.prevOrder = level.tailOrder
			level.tailOrder.nextOrder = &newOrder
			level.tailOrder = &newOrder
		}
	}

	return newOrder.id

}

func (ob *OrderBook) RemoveOrder(order Order) {
	delete(ob.orders, order.id)
	parentLevel := order.parentLevel
	parentLevel.volume -= order.size
	parentLevel.count--

	if parentLevel.count > 0 {
		// fmt.Printf("%+v\n", parentLevel)
		// fmt.Printf("Prev: %+v\n", order.prevOrder)
		// fmt.Printf("Next: %+v\n", order.nextOrder)
		// fmt.Printf("parentLevel.tailOrder == order ? %t\n", parentLevel.tailOrder.id == order.id)
		// fmt.Printf("parentLevel.headOrder == order ? %t\n", parentLevel.headOrder.id == order.id)

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
			// fmt.Printf("prevOrder: %+v\n", order.prevOrder.id)
			// fmt.Printf("currOrder: %+v\n", order.id)
			// fmt.Printf("nextOrder: %+v\n", order.nextOrder.id)
		}
	}

}

func (ob *OrderBook) CancelOrder() {

}

func (o Order) Equals(other Order) bool {
	return o.id == other.id
}

func (ob *OrderBook) PrintOrder(id uuid.UUID) {
	fmt.Printf("%+v\n", ob.orders[id])
}

func (ob *OrderBook) PrintOrderBook() {
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

	if ob.LowestAsk != nil {
		lowestAskPrice = ob.LowestAsk.price
	}

	if ob.HighestBid != nil {
		highestBidPrice = ob.HighestBid.price
	}
	return fmt.Sprintf(
		"OrderBook{\n 	lowestAsk: %+v\n	highestBid: %+v\n}",
		lowestAskPrice,
		highestBidPrice,
	)
}

func (l *Level) String() string {
	return fmt.Sprintf(
		"Level{price: %d\nvolume: %d\ncount: %d\nnextLevel: %+v\nheadOrder: %+v\ntailOrder: %+v}\n",
		l.price,
		l.volume,
		l.count,
		l.nextLevel,
		l.headOrder,
		l.tailOrder,
	)
}

// func (o *Order) String() string {
//
//     return fmt.Sprintf(
// 	    "Order{id: %s, side: %s, size: %d, remaining: %d, price: %d, time: %s, nextOrderId: %s, prevOrderId: %s}",
//         o.id.String(),
//         o.side,
//         o.size,
//         o.remaining,
//         o.price,
//         o.time.Format(time.RFC3339Nano),
// 	o.nextOrder.id,
// 	o.prevOrder.id,
//     )
// }

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
