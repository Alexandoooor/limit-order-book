package engine

import (
	"github.com/google/uuid"
	"time"
)
type Trade struct {
	Price    int       `json:"price"`
	Size     int       `json:"size"`
	Time     time.Time `json:"time"`
	BuyerID  uuid.UUID `json:"buyerId"`
	SellerID uuid.UUID `json:"sellerId"`
}
