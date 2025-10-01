package engine

import (
	"os"
	"sort"
	// "time"

	"github.com/google/uuid"
)

type LevelView struct {
	Price  int
	Volume int
}

type TradeView struct {
	ID 	 uuid.UUID
	Price    int
	Size     int
	Time     string
	BuyOrderID  uuid.UUID
	SellOrderID uuid.UUID
}

type OrderBookView struct {
	Bids     []LevelView
	Asks     []LevelView
	Trades   []Trade
	// Trades   []TradeView
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

	// for _, trade := range ob.trades {
	// 	view.Trades = append(view.Trades, TradeView{
	// 		trade.ID,
	// 		trade.Price,
	// 		trade.Size,
	// 		trade.Time.Format(time.RFC3339),
	// 		trade.BuyOrderID,
	// 		trade.SellOrderID,
	//
	// 	})
	// }

	sort.Slice(view.Bids, func(i, j int) bool {
		return view.Bids[i].Price > view.Bids[j].Price
	})

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
