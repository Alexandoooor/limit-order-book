package engine

import (
	"fmt"

	"github.com/google/uuid"
)

type Level struct {
	Price     int
	Volume    int
	Count     int
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

func (l *Level) ToDTO() *LevelDTO {
	levelDTO := &LevelDTO{
		Price:  l.Price,
		Volume: l.Volume,
		Count:  l.Count,
		Orders: []uuid.UUID{},
	}
	for o := l.headOrder; o != nil; o = o.nextOrder {
		levelDTO.Orders = append(levelDTO.Orders, o.Id)
		if o == l.tailOrder {
			break
		}
	}

	return levelDTO
}
