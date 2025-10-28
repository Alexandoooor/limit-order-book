package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID 	 uuid.UUID
	Price    int
	Size     int
	Time     time.Time
	BuyOrderID  uuid.UUID
	SellOrderID uuid.UUID
}

func (t *Trade) String() string {
	return fmt.Sprintf(
	"Trade{\n\tid: %s\n\tprice: %d\n\tsize: %d\n\ttime: %s\n\tbuyerId: %s\n\tsellerId: %s\n\t}\n",
		t.ID,
		t.Price,
		t.Size,
		t.Time.Format(time.RFC3339),
		t.BuyOrderID,
		t.SellOrderID,
	)
}
