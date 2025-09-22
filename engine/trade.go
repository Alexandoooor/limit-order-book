package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)
type Trade struct {
	Price    int       `json:"price"`
	Size     int       `json:"size"`
	Time     time.Time `json:"time"`
	BuyerID  uuid.UUID `json:"buyerId"`
	SellerID uuid.UUID `json:"sellerId"`
}

func (t *Trade) String() string {
	return fmt.Sprintf(
		"Trade{\n\tprice: %d\n\tsize: %d\n\ttime: %s\n\tbuyerId: %s\n\tsellerId: %s\n\t}\n",
		t.Price,
		t.Size,
		t.Time.Format(time.RFC3339Nano),
		t.BuyerID,
		t.SellerID,
	)
}
