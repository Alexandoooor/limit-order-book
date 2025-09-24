package engine

import (
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
