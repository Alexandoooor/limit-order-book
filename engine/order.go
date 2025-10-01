package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	Id          uuid.UUID `json:"id"`
	Side        Side `json:"side"`
	Size        int `json:"size"`
	Remaining   int `json:"remaining"`
	Price       int `json:"price"`
	Time        time.Time `json:"time"`
	nextOrder   *Order
	prevOrder   *Order
	parentLevel *Level
}

func (o Order) Equals(other Order) bool {
	return o.Id == other.Id
}

func (o *Order) String() string {
	var nextID, prevID string

	if o.nextOrder != nil {
		nextID = o.nextOrder.Id.String()
	} else {
		nextID = "nil"
	}

	if o.prevOrder != nil {
		prevID = o.prevOrder.Id.String()
	} else {
		prevID = "nil"
	}

	return fmt.Sprintf(
		"Order{\n\tid: %s\n\tside: %s\n\tsize: %d\n\tremaining: %d\n\tprice: %d\n\ttime: %s\n\tnextOrderId: %s\n\tprevOrderId: %s\n\t}\n",
		o.Id.String(),
		o.Side,
		o.Size,
		o.Remaining,
		o.Price,
		o.Time.Format(time.RFC3339),
		nextID,
		prevID,
	)
}

func (o *Order) ToDTO() *OrderDTO {
	orderDTO := &OrderDTO{
		Id:        o.Id,
		Side:      o.Side,
		Size:      o.Size,
		Remaining: o.Remaining,
		Price:     o.Price,
		Time:      o.Time,
	}
	if o.nextOrder != nil {
		orderDTO.NextID = &o.nextOrder.Id
	}
	if o.prevOrder != nil {
		orderDTO.PrevID = &o.prevOrder.Id
	}

	return orderDTO
}
