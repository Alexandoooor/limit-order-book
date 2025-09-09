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

	// ob.PrintOrderBook()
	// t.Fatalf("%+v", level)
}

func TestLiftOrder(t *testing.T) {
	ob := NewOrderBook()

	id := ob.AddOrder(Buy, 1337, 1)
	ob.PrintOrderBook()
	order := ob.orders[id]
	ob.LiftOrder(*order)
	ob.PrintOrderBook()

	// ob.PrintOrderBook()
	t.Fatalf("%+v", id)
}
