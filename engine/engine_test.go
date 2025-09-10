package engine

import (
	"math/rand/v2"
	"testing"
)

func TestAddingOrders(t *testing.T) {
	ob := NewOrderBook()

	for x := range 10 {
		expectedSize := 1
		id := ob.AddOrder(Buy, x, expectedSize)

		if ob.orders[id].price != x {
			t.Fatalf("tests - order.price wrong. expected=%+v, got=%+v", x, ob.orders[id].price)
		}

		if ob.orders[id].side != Buy {
			t.Fatalf("tests - order.side wrong. expected=%+v, got=%+v", Buy, ob.orders[id].side)
		}

		if ob.orders[id].size != expectedSize {
			t.Fatalf("tests - order.size wrong. expected=%+v, got=%+v", expectedSize, ob.orders[id].size)
		}
	}
}

func TestHighestBid(t *testing.T) {
	ob := NewOrderBook()

	if ob.HighestBid != nil {
		t.Fatalf("tests - HighestBid wrong. expected=%+v, got=%+v", nil, ob.HighestBid)
	}

	for x := 88; x <= 92; x++ {
		ob.AddOrder(Buy, x, 1)
		if ob.HighestBid.price != x {
			t.Fatalf("tests - HighestBid wrong. expected=%+v, got=%+v", x, ob.HighestBid.price)
		}
	}

	expectedHighestBid := 92
	ob.AddOrder(Buy, 90, 1)
	if ob.HighestBid.price != expectedHighestBid {
		t.Fatalf("tests - HighestBid wrong. expected=%+v, got=%+v", expectedHighestBid, ob.HighestBid.price)
	}

	ob.AddOrder(Buy, 999, 1)
	expectedPrice := 999
	if ob.HighestBid.price != expectedPrice {
		t.Fatalf("tests - HighestBid wrong. expected=%+v, got=%+v", expectedPrice, ob.HighestBid.price)
	}

}

func TestLowestAsk(t *testing.T) {
	ob := NewOrderBook()
	if ob.LowestAsk != nil {
		t.Fatalf("tests - LowestAsk wrong. expected=%+v, got=%+v", nil, ob.LowestAsk)
	}

	ob = NewOrderBook()
	expectedLowestAsk := 88
	for x := 88; x <= 92; x++ {
		ob.AddOrder(Sell, x, 1)
		if ob.LowestAsk.price != expectedLowestAsk {
			t.Fatalf("tests - LowestAsk wrong. expected=%+v, got=%+v", expectedLowestAsk, ob.LowestAsk.price)
		}
	}

	ob = NewOrderBook()
	for x := 92; x >= 88; x-- {
		ob.AddOrder(Sell, x, 1)
		if ob.LowestAsk.price != x {
			t.Fatalf("tests - LowestAsk wrong. expected=%+v, got=%+v", x, ob.LowestAsk.price)
		}
	}

	ob = NewOrderBook()
	ob.AddOrder(Sell, 1, 1)
	expectedLowestAsk = 1
	if ob.LowestAsk.price != expectedLowestAsk {
		t.Fatalf("tests - LowestAsk wrong. expected=%+v, got=%+v", expectedLowestAsk, ob.LowestAsk.price)
	}

}


func TestMultiOrderLevel(t *testing.T) {
	ob := NewOrderBook()

	var n int = 5
	orders := make([]Order, n)

	randomPrice := rand.IntN(42) + (42 % 3)
	for x := range n {
		id := ob.AddOrder(Buy, randomPrice, randomPrice*(x+1))
		orders[x] = *ob.orders[id]
	}

	level := ob.bids[randomPrice]

	expectedHead := &orders[0]
	if level.headOrder.id != expectedHead.id {
		t.Fatalf("tests - headOrder wrong. expected=%+v, got=%+v", expectedHead, level.headOrder)
	}

	expectedTail := &orders[n-1]
	if level.tailOrder.id != expectedTail.id {
		t.Fatalf("tests - tailOrder wrong. expected=%+v, got=%+v", expectedTail, level.headOrder)
	}

	expectedVolume := 0
	for x := range 5 {
		expectedVolume += randomPrice*(x+1)
	}
	if level.volume != expectedVolume {
		t.Fatalf("tests - volume wrong. expected=%+v, got=%+v", expectedVolume, level.volume)
	}

	if level.count != n {
		t.Fatalf("tests - count wrong. expected=%+v, got=%+v", n, level.count)
	}
}

func TestRemoveTail(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(Buy, 1337, 1)
	id1 := ob.AddOrder(Buy, 1337, 2) //The second order aka the tail is the one we remove
	// ob.PrintOrderBook()

	order := ob.orders[id1]
	ob.RemoveOrder(*order)

	// ob.PrintOrderBook()
	// t.Fatalf("parentLevel: %+v\n", order.parentLevel)
	level := ob.bids[1337]
	if level.tailOrder.id != id0 {
		t.Fatalf("tests - tail wrong. expected=%+v, got=%+v", id0, level.tailOrder.id)
	}
}

func TestRemoveHead(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(Buy, 7331, 1) //The first order aka the head is the one we remove
	id1 := ob.AddOrder(Buy, 7331, 2)
	// ob.PrintOrderBook()

	order := ob.orders[id0]
	ob.RemoveOrder(*order)

	// ob.PrintOrderBook()
	// t.Fatalf("parentLevel: %+v\n", order.parentLevel)

	level := ob.bids[7331]
	if level.headOrder.id != id1 {
		t.Fatalf("tests - head wrong. expected=%+v, got=%+v", id1, level.headOrder.id)
	}
}

func TestRemoveMiddle(t *testing.T) {
	ob := NewOrderBook()

	ob.AddOrder(Buy, 7331, 3)
	id := ob.AddOrder(Buy, 7331, 1) //The middle order is removed
	ob.AddOrder(Buy, 7331, 2)
	// ob.PrintOrderBook()

	order := ob.orders[id]
	ob.RemoveOrder(*order)

	// ob.PrintOrderBook()

	// t.Fatalf("parentLevel: %+v\n", order.parentLevel)
}
