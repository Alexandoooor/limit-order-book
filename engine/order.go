package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	id          uuid.UUID
	side        Side
	size        int
	remaining   int
	price       int
	time        time.Time
	nextOrder   *Order
	prevOrder   *Order
	parentLevel *Level
}

func (o *Order) String() string {
	var nextID, prevID string

	if o.nextOrder != nil {
		nextID = o.nextOrder.id.String()
	} else {
		nextID = "nil"
	}

	if o.prevOrder != nil {
		prevID = o.prevOrder.id.String()
	} else {
		prevID = "nil"
	}

	return fmt.Sprintf(
		"Order{\n\tid: %s\n\tside: %s\n\tsize: %d\n\tremaining: %d\n\tprice: %d\n\ttime: %s\n\tnextOrderId: %s\n\tprevOrderId: %s\n\t}\n",
		o.id.String(),
		o.side,
		o.size,
		o.remaining,
		o.price,
		o.time.Format(time.RFC3339Nano),
		nextID,
		prevID,
	)
}
