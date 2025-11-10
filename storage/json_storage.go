package storage

import (
	"limit-order-book/engine"

	"encoding/json"
	"os"
	"github.com/google/uuid"
)

type JsonStorage struct {}

func (j *JsonStorage) InsertLevel(side engine.Side, l *engine.LevelDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Levels[side][l.Price] = l
	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) InsertTrade(t *engine.Trade) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Trades = append(dto.Trades, *t)
	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) InsertOrder(o *engine.OrderDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}
	dto.Orders[o.Id] = o

	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) DeleteOrder(ob *engine.OrderBookDTO, o *engine.OrderDTO) error {
	dto, err := j.getDTO()
	if err != nil {
		return err
	}

	delete(dto.Orders, o.Id)
	parentLevel := dto.Levels[o.Side][o.Price]
	if parentLevel != nil {
		parentLevel.Count--
		if parentLevel.Count <= 0 {
			delete(dto.Levels[o.Side], o.Price)
		}
	}

	return j.WriteDTOToJson(dto)
}

func (j *JsonStorage) UpdateOrder(ob *engine.OrderBookDTO, o *engine.OrderDTO) error {
	return j.DeleteOrder(ob, o)
}

func (j *JsonStorage) WriteDTOToJson(dto *engine.OrderBookDTO) error {
	filename := j.getFilename()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (j *JsonStorage) getDTO() (*engine.OrderBookDTO, error) {
	filename := j.getFilename()
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dto *engine.OrderBookDTO
	err = json.Unmarshal(data, &dto)
	if err != nil {
		return &engine.OrderBookDTO{
			Levels: map[engine.Side]map[int]*engine.LevelDTO{engine.Buy: {}, engine.Sell: {}},
			Orders: make(map[uuid.UUID]*engine.OrderDTO),
			Trades: []engine.Trade{},
		}, nil
	}
	return dto, nil
}

func (j *JsonStorage) getFilename() string {
	orderBookFile := os.Getenv("ORDERBOOK")
	if orderBookFile == "" {
		orderBookFile = "/tmp/orderbook.json"
	}
	return orderBookFile
}


func (j *JsonStorage) ResetOrderBook() error {
	filename := j.getFilename()
	Logger.Printf("Wiping orders from %s\n", filename)
	err := os.WriteFile(filename, []byte("[]"), 0644)

	if err != nil {
		Logger.Printf("Failed to reset OrderBook: %s", err)
		return err
	}

	return nil
}

func (j *JsonStorage) RestoreOrderBook() (*engine.OrderBook, error) {
	filename := j.getFilename()
	data, err := os.ReadFile(filename)
	if err != nil {
		_, err := os.Create(filename)
		if err != nil {
			Logger.Fatal(err)
		}
	}
	data, err = os.ReadFile(filename)

	var dto engine.OrderBookDTO
	err = json.Unmarshal(data, &dto)
	if err != nil {
		return nil, err
	}
	restoredBook := dto.ToOrderBook()

	return restoredBook, nil
}

func (j *JsonStorage) DumpOrderBook(ob *engine.OrderBook) error {
	filename := j.getFilename()

	dto := ob.ToDTO()
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
