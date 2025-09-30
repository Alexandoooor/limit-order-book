package engine

type NilStorage struct {}

func (n *NilStorage) InsertLevel(side Side, l *LevelDTO) error {
	return nil
}

func (n *NilStorage) InsertTrade(t *Trade) error {
	return nil
}

func (n *NilStorage) InsertOrder(o *OrderDTO) error {
	return nil

}
func (n *NilStorage) DeleteOrder(ob *OrderBookDTO, o *OrderDTO) error {
	return nil
}

func (n *NilStorage) UpdateOrder(ob *OrderBookDTO, o *OrderDTO) error {
	return nil
}

func (n *NilStorage) ResetOrderBook() error {
	return nil
}

func (n *NilStorage) RestoreOrderBook() (*OrderBook, error) {
	return NewOrderBook(), nil
}
