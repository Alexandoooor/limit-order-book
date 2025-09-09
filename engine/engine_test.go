package engine

import (
	"testing"
)

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

	ob.AddOrder(Buy, 999, 1)
	expectedPrice := 999
	if ob.HighestBid.price != expectedPrice {
		t.Fatalf("tests - HighestBid wrong. expected=%+v, got=%+v", expectedPrice, ob.HighestBid.price)
	}

}
