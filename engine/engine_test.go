package engine

import (
	"math/rand/v2"
	"testing"
	"github.com/google/uuid"
)

func TestAddingOrders(t *testing.T) {
	ob := NewOrderBook()

	for x := range 10 {
		expectedSize := 1
		id := ob.AddOrder(uuid.New(), Buy, x, expectedSize, expectedSize)

		if ob.orders[id].Price != x {
			t.Fatalf("tests - order.price wrong. expected=%+v, got=%+v", x, ob.orders[id].Price)
		}

		if ob.orders[id].Side != Buy {
			t.Fatalf("tests - order.side wrong. expected=%+v, got=%+v", Buy, ob.orders[id].Side)
		}

		if ob.orders[id].Size != expectedSize {
			t.Fatalf("tests - order.size wrong. expected=%+v, got=%+v", expectedSize, ob.orders[id].Size)
		}
	}
}

func TestHighestBid(t *testing.T) {
	ob := NewOrderBook()

	if ob.highestBid != nil {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", nil, ob.highestBid)
	}

	for x := 88; x <= 92; x++ {
		ob.AddOrder(uuid.New(), Buy, x, 1, 1)
		if ob.highestBid.Price != x {
			t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", x, ob.highestBid.Price)
		}
	}

	expectedhighestBid := 92
	ob.AddOrder(uuid.New(), Buy, 90, 1, 1)
	if ob.highestBid.Price != expectedhighestBid {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", expectedhighestBid, ob.highestBid.Price)
	}

	ob.AddOrder(uuid.New(), Buy, 999, 1, 1)
	expectedPrice := 999
	if ob.highestBid.Price != expectedPrice {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", expectedPrice, ob.highestBid.Price)
	}

}

func TestLowestAsk(t *testing.T) {
	ob := NewOrderBook()
	if ob.lowestAsk != nil {
		t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", nil, ob.lowestAsk)
	}

	ob = NewOrderBook()
	expectedlowestAsk := 88
	for x := 88; x <= 92; x++ {
		ob.AddOrder(uuid.New(), Sell, x, 1, 1)
		if ob.lowestAsk.Price != expectedlowestAsk {
			t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", expectedlowestAsk, ob.lowestAsk.Price)
		}
	}

	ob = NewOrderBook()
	for x := 92; x >= 88; x-- {
		ob.AddOrder(uuid.New(), Sell, x, 1, 1)
		if ob.lowestAsk.Price != x {
			t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", x, ob.lowestAsk.Price)
		}
	}

	ob = NewOrderBook()
	ob.AddOrder(uuid.New(), Sell, 1, 1, 1)
	expectedlowestAsk = 1
	if ob.lowestAsk.Price != expectedlowestAsk {
		t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", expectedlowestAsk, ob.lowestAsk.Price)
	}

}


func TestMultiOrderLevel(t *testing.T) {
	ob := NewOrderBook()

	var n int = 5
	orders := make([]Order, n)

	randomPrice := rand.IntN(42) + (42 % 3)
	for x := range n {
		id := ob.AddOrder(uuid.New(), Buy, randomPrice, randomPrice*(x+1), randomPrice*(x+1))
		orders[x] = *ob.orders[id]
	}

	level := ob.bids[randomPrice]

	expectedHead := &orders[0]
	if level.HeadOrder.Id != expectedHead.Id {
		t.Fatalf("tests - headOrder wrong. expected=%+v, got=%+v", expectedHead, level.HeadOrder)
	}

	expectedTail := &orders[n-1]
	if level.TailOrder.Id != expectedTail.Id {
		t.Fatalf("tests - tailOrder wrong. expected=%+v, got=%+v", expectedTail, level.HeadOrder)
	}

	expectedVolume := 0
	for x := range 5 {
		expectedVolume += randomPrice*(x+1)
	}
	if level.Volume != expectedVolume {
		t.Fatalf("tests - volume wrong. expected=%+v, got=%+v", expectedVolume, level.Volume)
	}

	if level.Count != n {
		t.Fatalf("tests - count wrong. expected=%+v, got=%+v", n, level.Count)
	}
}

func TestRemoveTail(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 1337, 1, 1)
	id1 := ob.AddOrder(uuid.New(), Buy, 1337, 2, 2) //The second order aka the tail is the one we remove

	order := ob.orders[id1]
	ob.RemoveOrder(*order)

	level := ob.bids[1337]
	if level.TailOrder.Id != id0 {
		t.Fatalf("tests - tail wrong. expected=%+v, got=%+v", id0, level.TailOrder.Id)
	}
}

func TestRemoveHead(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 7331, 1, 1) //The first order aka the head is the one we remove
	id1 := ob.AddOrder(uuid.New(), Buy, 7331, 2, 2)

	order := ob.orders[id0]
	ob.RemoveOrder(*order)

	level := ob.bids[7331]
	if level.HeadOrder.Id != id1 {
		t.Fatalf("tests - head wrong. expected=%+v, got=%+v", id1, level.HeadOrder.Id)
	}
}

func TestRemoveMiddle(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 7331, 3, 3)
	id1 := ob.AddOrder(uuid.New(), Buy, 7331, 1, 1) //The middle order is removed
	id2 := ob.AddOrder(uuid.New(), Buy, 7331, 2, 2)

	order := ob.orders[id1]
	ob.RemoveOrder(*order)

	if ob.orders[id0].NextOrder.Id != id2 {
		t.Fatalf("tests - first.nextOrder wrong. expected=%+v, got=%+v", id2, ob.orders[id0].NextOrder.Id)
	}

	if ob.orders[id2].PrevOrder.Id != id0 {
		t.Fatalf("tests - last.prevOrder wrong. expected=%+v, got=%+v", id0, ob.orders[id2].PrevOrder.Id)
	}
}

func TestRemoveOnlyOrderInLevel(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 42, 9, 9)
	ob.AddOrder(uuid.New(), Buy, 41, 9, 9)

	order := ob.orders[id0]
	ob.RemoveOrder(*order)

	if ob.bids[42] != nil {
		t.Fatalf("tests - removing last order in level should delete level. expected=%+v, got=%+v", nil, ob.bids[42])
	}

	if ob.highestBid != ob.bids[41] {
		t.Fatalf("tests - highestBid not moved to nextLevel. expected=%+v, got=%+v", ob.bids[41], ob.highestBid)
	}

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Sell, 40, 2)

	isCanceled := ob.CancelOrder(id0)


	if len(ob.orders) != 0 {
		t.Fatalf("tests - order book should be empty. expected=%+v, got=%+v", 0, len(ob.orders))
	}

	if !isCanceled {
		t.Fatalf("tests - Order should be canceled. expected=%t, got=%t", true, isCanceled)
	}

	isCanceled = ob.CancelOrder(id0)

	if isCanceled {
		t.Fatalf("tests - Order should no longer exist. expected=%t, got=%t", false, isCanceled)
	}

}

func TestRemoveOnlyOrderInBook(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 42, 9, 9)

	order := ob.orders[id0]
	ob.RemoveOrder(*order)

	if ob.bids[42] != nil {
		t.Fatalf("tests - removing last order in level should delete level. expected=%+v, got=%+v", nil, ob.bids[42])
	}

	if ob.highestBid != nil {
		t.Fatalf("tests - highestBid not nil after removing only order in book. expected=%+v, got=%+v", nil, ob.highestBid)
	}

}


func TestProcessPartialOrder(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Buy, 42, 1)
	id1 := ob.ProcessOrder(Sell, 40, 2)

	o0 := ob.orders[id0]
	o1 := ob.orders[id1]

	expectedRemaining := 1
	if o1.Remaining != expectedRemaining {
		t.Fatalf("tests - incorrect remaning size of partially filled order. expected=%d, got=%d",
			expectedRemaining, o1.Remaining)
	}

	if o0 != nil {
		t.Fatalf("tests - order %d was completely filled and should have" +
			" been removed from the orderbook. expected=%v, got=%+v",
			id0, nil, o1)
	}
}

func TestProcessWholeOrder(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Buy, 42, 2)
	id1 := ob.ProcessOrder(Sell, 40, 2)

	o0 := ob.orders[id0]
	o1 := ob.orders[id1]

	if o0 != nil {
		t.Fatalf("tests - order 0 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o0)
	}

	if o1 != nil {
		t.Fatalf("tests - order 1 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o1)
	}

	if len(ob.bids) != 0 {
		t.Fatalf("tests - bids should be empty. expected=%+v, got=%+v", 0, len(ob.bids))
	}

	if len(ob.asks) != 0 {
		t.Fatalf("tests - asks should be empty. expected=%+v, got=%+v", 0, len(ob.asks))
	}

	if len(ob.orders) != 0 {
		t.Fatalf("tests - order book should be empty. expected=%+v, got=%+v", 0, len(ob.orders))
	}


}

func TestProcessMultiLevelOrder(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Buy, 42, 3)
	id1 := ob.ProcessOrder(Sell, 40, 2)

	if ob.orders[id0].Remaining != 1 {
		t.Fatalf("tests - order 0 should be partially filled." + " expected=%d, got=%+v", 1, ob.orders[id0].Remaining)
	}

	id2 := ob.ProcessOrder(Sell, 41, 1)

	o0 := ob.orders[id0]
	o1 := ob.orders[id1]
	o2 := ob.orders[id2]

	if o0 != nil {
		t.Fatalf("tests - order 0 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o0)
	}

	if o1 != nil {
		t.Fatalf("tests - order 1 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o1)
	}

	if o2 != nil {
		t.Fatalf("tests - order 2 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o2)
	}

	if len(ob.bids) != 0 {
		t.Fatalf("tests - bids should be empty. expected=%+v, got=%+v", 0, len(ob.bids))
	}

	if len(ob.asks) != 0 {
		t.Fatalf("tests - asks should be empty. expected=%+v, got=%+v", 0, len(ob.asks))
	}

	if len(ob.orders) != 0 {
		t.Fatalf("tests - order book should be empty. expected=%+v, got=%+v", 0, len(ob.orders))
	}

}

func TestProcessMultiLevelOrder2(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Sell, 40, 2)
	id1 := ob.ProcessOrder(Buy, 42, 3)

	if ob.orders[id1].Remaining != 1 {
		t.Fatalf("tests - order 1 should be partially filled." + " expected=%d, got=%+v", 1, ob.orders[id1].Remaining)
	}

	id2 := ob.ProcessOrder(Sell, 41, 1)

	o0 := ob.orders[id0]
	o1 := ob.orders[id1]
	o2 := ob.orders[id2]

	if o0 != nil {
		t.Fatalf("tests - order 0 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o0)
	}

	if o1 != nil {
		t.Fatalf("tests - order 1 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o1)
	}

	if o2 != nil {
		t.Fatalf("tests - order 2 should be executed and removed from the order book." +
			" expected=%+v, got=%+v", nil, o2)
	}

	if len(ob.bids) != 0 {
		t.Fatalf("tests - bids should be empty. expected=%+v, got=%+v", 0, len(ob.bids))
	}

	if len(ob.asks) != 0 {
		t.Fatalf("tests - asks should be empty. expected=%+v, got=%+v", 0, len(ob.asks))
	}

	if len(ob.orders) != 0 {
		t.Fatalf("tests - order book should be empty. expected=%+v, got=%+v", 0, len(ob.orders))
	}

}
