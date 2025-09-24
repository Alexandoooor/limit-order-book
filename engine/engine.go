package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

var Logger *log.Logger

type OrderBook struct {
	levels     map[Side]map[int]*Level
	orders     map[uuid.UUID]*Order
	lowestAsk  *Level
	highestBid *Level
	trades     []Trade
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

func NewOrderBook() *OrderBook {
	levels := map[Side]map[int]*Level{
		Buy:  make(map[int]*Level),
		Sell: make(map[int]*Level),
	}

	ob := &OrderBook{
		levels: levels,
		orders: make(map[uuid.UUID]*Order),
	}

	return ob
}

func (ob *OrderBook) ResetOrderBook() {
	ob.levels = map[Side]map[int]*Level{Buy: {}, Sell: {}}
	ob.orders = make(map[uuid.UUID]*Order)
	ob.trades = []Trade{}
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
	return order.Id

}

func (ob *OrderBook) RemoveOrder(order Order) *Order {
	delete(ob.orders, order.Id)
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
					ob.trades = append(ob.trades, trade)

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
					ob.trades = append(ob.trades, trade)

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

func (ob *OrderBook) ToDTO() *OrderBookDTO {
	dto := &OrderBookDTO{
		Levels: map[Side]map[int]*LevelDTO{Buy: {}, Sell: {}},
		Orders: make(map[uuid.UUID]*OrderDTO),
		Trades: ob.trades,
	}

	for id, o := range ob.orders {
		orderDTO := &OrderDTO{
			Id:        o.Id,
			Side:      o.Side,
			Size:      o.Size,
			Remaining: o.Remaining,
			Price:     o.Price,
			Time:      o.Time,
		}
		if o.nextOrder != nil {
			orderDTO.NextID = &o.nextOrder.Id
		}
		if o.prevOrder != nil {
			orderDTO.PrevID = &o.prevOrder.Id
		}
		dto.Orders[id] = orderDTO
	}

	for side := range ob.levels {
		for price, lvl := range ob.levels[side] {
			levelDTO := &LevelDTO{
				Price:  lvl.Price,
				Volume: lvl.Volume,
				Count:  lvl.Count,
				Orders: []uuid.UUID{},
			}
			for o := lvl.headOrder; o != nil; o = o.nextOrder {
				levelDTO.Orders = append(levelDTO.Orders, o.Id)
				if o == lvl.tailOrder {
					break
				}
			}
			dto.Levels[side][price] = levelDTO
		}
	}

	return dto
}
