package engine

import (
	"fmt"
)

type Level struct {
	Price     int `json:"price"`
	Volume    int `json:"volume"`
	Count     int `json:"count"`
	nextLevel *Level
	headOrder *Order
	tailOrder *Order
}

func (l *Level) String() string {
	return fmt.Sprintf(
		"Level{\n\tprice: %d\n\tvolume: %d\n\tcount: %d\n\tnextLevel: %+v\n\theadOrder: %+v\n\ttailOrder: %+v\n}\n",
		l.Price,
		l.Volume,
		l.Count,
		l.nextLevel,
		l.headOrder,
		l.tailOrder,
	)
}
