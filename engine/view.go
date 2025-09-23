package engine

import (
	"os"
	"sort"
)

type LevelView struct {
	Price  int
	Volume int
}

type OrderBookView struct {
	Bids     []LevelView
	Asks     []LevelView
	Trades   []Trade
	Hostname string
}

func BuildOrderBookView(ob *OrderBook) OrderBookView {
	view := OrderBookView{}

	for price, level := range ob.levels[Buy] {
		view.Bids = append(view.Bids, LevelView{
			Price:  price,
			Volume: level.Volume,
		})
	}

	for price, level := range ob.levels[Sell] {
		view.Asks = append(view.Asks, LevelView{
			Price:  price,
			Volume: level.Volume,
		})
	}

	view.Trades = ob.trades

	sort.Slice(view.Bids, func(i, j int) bool {
		return view.Bids[i].Price > view.Bids[j].Price
	})

	// Sort asks ascending by price
	sort.Slice(view.Asks, func(i, j int) bool {
		return view.Asks[i].Price < view.Asks[j].Price
	})

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "unknown"
	}

	view.Hostname = hostname

	return view
}
