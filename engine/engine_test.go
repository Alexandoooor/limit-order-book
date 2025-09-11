package engine

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

func TestAddingOrders(t *testing.T) {
	ob := NewOrderBook()

	for x := range 10 {
		expectedSize := 1
		id := ob.AddOrder(uuid.New(), Buy, x, expectedSize, expectedSize)

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

	if ob.highestBid != nil {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", nil, ob.highestBid)
	}

	for x := 88; x <= 92; x++ {
		ob.AddOrder(uuid.New(), Buy, x, 1, 1)
		if ob.highestBid.price != x {
			t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", x, ob.highestBid.price)
		}
	}

	expectedhighestBid := 92
	ob.AddOrder(uuid.New(), Buy, 90, 1, 1)
	if ob.highestBid.price != expectedhighestBid {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", expectedhighestBid, ob.highestBid.price)
	}

	ob.AddOrder(uuid.New(), Buy, 999, 1, 1)
	expectedPrice := 999
	if ob.highestBid.price != expectedPrice {
		t.Fatalf("tests - highestBid wrong. expected=%+v, got=%+v", expectedPrice, ob.highestBid.price)
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
		if ob.lowestAsk.price != expectedlowestAsk {
			t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", expectedlowestAsk, ob.lowestAsk.price)
		}
	}

	ob = NewOrderBook()
	for x := 92; x >= 88; x-- {
		ob.AddOrder(uuid.New(), Sell, x, 1, 1)
		if ob.lowestAsk.price != x {
			t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", x, ob.lowestAsk.price)
		}
	}

	ob = NewOrderBook()
	ob.AddOrder(uuid.New(), Sell, 1, 1, 1)
	expectedlowestAsk = 1
	if ob.lowestAsk.price != expectedlowestAsk {
		t.Fatalf("tests - lowestAsk wrong. expected=%+v, got=%+v", expectedlowestAsk, ob.lowestAsk.price)
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

	id0 := ob.AddOrder(uuid.New(), Buy, 1337, 1, 1)
	id1 := ob.AddOrder(uuid.New(), Buy, 1337, 2, 2) //The second order aka the tail is the one we remove

	order := ob.orders[id1]
	ob.RemoveOrder(*order)

	level := ob.bids[1337]
	if level.tailOrder.id != id0 {
		t.Fatalf("tests - tail wrong. expected=%+v, got=%+v", id0, level.tailOrder.id)
	}
}

func TestRemoveHead(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 7331, 1, 1) //The first order aka the head is the one we remove
	id1 := ob.AddOrder(uuid.New(), Buy, 7331, 2, 2)

	order := ob.orders[id0]
	ob.RemoveOrder(*order)

	level := ob.bids[7331]
	if level.headOrder.id != id1 {
		t.Fatalf("tests - head wrong. expected=%+v, got=%+v", id1, level.headOrder.id)
	}
}

func TestRemoveMiddle(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.AddOrder(uuid.New(), Buy, 7331, 3, 3)
	id1 := ob.AddOrder(uuid.New(), Buy, 7331, 1, 1) //The middle order is removed
	id2 := ob.AddOrder(uuid.New(), Buy, 7331, 2, 2)

	order := ob.orders[id1]
	ob.RemoveOrder(*order)

	if ob.orders[id0].nextOrder.id != id2 {
		t.Fatalf("tests - first.nextOrder wrong. expected=%+v, got=%+v", id2, ob.orders[id0].nextOrder.id)
	}

	if ob.orders[id2].prevOrder.id != id0 {
		t.Fatalf("tests - last.prevOrder wrong. expected=%+v, got=%+v", id0, ob.orders[id2].prevOrder.id)
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
	// ob.PrintOrderBook()

	expectedRemaining := 1
	if o1.remaining != expectedRemaining {
		t.Fatalf("tests - incorrect remaning size of partially filled order. expected=%d, got=%d",
			expectedRemaining, o1.remaining)
	}

	if o0 != nil {
		t.Fatalf("tests - order %d was completely filled and should have " +
			"been removed from the orderbook. expected=%v, got=%+v",
			id0, nil, o1)
	}
}

func TestProcessWholeBuyOrder(t *testing.T) {
	ob := NewOrderBook()

	id0 := ob.ProcessOrder(Buy, 42, 2)
	id1 := ob.ProcessOrder(Sell, 40, 2)

	o0 := ob.orders[id0]
	o1 := ob.orders[id1]

	fmt.Println(o0)
	fmt.Println(o1)
	fmt.Println(ob.bids)
	fmt.Println(ob.asks)

	t.Fatalf("")
	// if ob.bids[42] != nil {
	// 	t.Fatalf("tests - removing last order in level should delete level. expected=%+v, got=%+v", nil, ob.bids[42])
	// }
	//
	// if ob.highestBid != nil {
	// 	t.Fatalf("tests - highestBid not nil after removing only order in book. expected=%+v, got=%+v", nil, ob.highestBid)
	// }

}
//
// func TestProcessWholeSellOrder(t *testing.T) {
// 	ob := NewOrderBook()
//
// 	ob.AddOrder(uuid.New(), Buy, 43, 1, 1)
// 	id1 := ob.AddOrder(uuid.New(), Sell, 42, 1, 1)
//
// 	order := ob.orders[id1]
// 	ob.ProcessOrder(*order)
//
// 	t.Fatalf("")
// 	// if ob.bids[42] != nil {
// 	// 	t.Fatalf("tests - removing last order in level should delete level. expected=%+v, got=%+v", nil, ob.bids[42])
// 	// }
// 	//
// 	// if ob.highestBid != nil {
// 	// 	t.Fatalf("tests - highestBid not nil after removing only order in book. expected=%+v, got=%+v", nil, ob.highestBid)
// 	// }
//
// }
