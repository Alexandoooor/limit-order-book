package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID 	 uuid.UUID `json:"id"`
	Price    int       `json:"price"`
	Size     int       `json:"size"`
	Time     time.Time `json:"time"`
	BuyOrderID  uuid.UUID `json:"buyerId"`
	SellOrderID uuid.UUID `json:"sellerId"`
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
