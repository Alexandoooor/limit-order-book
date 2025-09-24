package engine

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
)

type LevelDTO struct {
	Price   int       `json:"price"`
	Volume  int       `json:"volume"`
	Count   int       `json:"count"`
	Orders  []uuid.UUID `json:"orders"`
}

type OrderBookDTO struct {
	Levels     map[Side]map[int]*LevelDTO `json:"levels"`
	Orders 	   map[uuid.UUID]*OrderDTO `json:"orders"`
	Trades     []Trade `json:"trades"`
}

type OrderDTO struct {
	Id        uuid.UUID `json:"id"`
	Side      Side      `json:"side"`
	Size      int       `json:"size"`
	Remaining int       `json:"remaining"`
	Price     int       `json:"price"`
	Time      time.Time `json:"time"`
	NextID    *uuid.UUID `json:"next_id,omitempty"`
	PrevID    *uuid.UUID `json:"prev_id,omitempty"`
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

func (dto *OrderBookDTO) ToOrderBook() *OrderBook {
	ob := &OrderBook{
		levels: map[Side]map[int]*Level{Buy: {}, Sell: {}},
		orders: make(map[uuid.UUID]*Order),
		trades: dto.Trades,
	}

	for id, odto := range dto.Orders {
		o := &Order{
			Id:        odto.Id,
			Side:      odto.Side,
			Size:      odto.Size,
			Remaining: odto.Remaining,
			Price:     odto.Price,
			Time:      odto.Time,
		}
		ob.orders[id] = o
	}

	for id, odto := range dto.Orders {
		o := ob.orders[id]
		if odto.NextID != nil {
			o.nextOrder = ob.orders[*odto.NextID]
		}
		if odto.PrevID != nil {
			o.prevOrder = ob.orders[*odto.PrevID]
		}
	}

	for side := range dto.Levels {
		for price, ldto := range dto.Levels[side] {
			lvl := &Level{
				Price:  ldto.Price,
				Volume: ldto.Volume,
				Count:  ldto.Count,
			}

			if side == Buy {
				if ob.highestBid == nil {
					ob.highestBid = lvl
				} else if lvl.Price > ob.highestBid.Price {
					ob.highestBid = lvl
				}
			} else {
				if ob.lowestAsk == nil {
					ob.lowestAsk = lvl
				} else if lvl.Price < ob.lowestAsk.Price {
					ob. lowestAsk = lvl
				}
			}


			var prev *Order
			for _, oid := range ldto.Orders {
				o := ob.orders[oid]
				if prev == nil {
					lvl.headOrder = o
				} else {
					prev.nextOrder = o
					o.prevOrder = prev
				}
				prev = o
				o.parentLevel = lvl
			}
			lvl.tailOrder = prev

			ob.levels[side][price] = lvl
		}
	}

	return ob
}

func (ob *OrderBook) DumpOrderBook() error {
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

func (ob *OrderBook) LoadOrderBook() error {
	filename := os.Getenv("ORDERBOOK")
	if filename == "" {
		filename = "/tmp/orderbook.json"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var dto2 OrderBookDTO
	_ = json.Unmarshal(data, &dto2)
	restoredBook := dto2.ToOrderBook()

	ob.levels = restoredBook.levels
	ob.orders = restoredBook.orders
	ob.highestBid = restoredBook.highestBid
	ob.lowestAsk = restoredBook.lowestAsk
	ob.trades = restoredBook.trades

	return nil
}
