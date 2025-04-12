// SPDX-File-CopyrightText: (c) 2025 Andrei Ilin <ortfero@gmail.com>
// SPDX-License-Identifier: MIT

package orderstracker

import (
	"math/rand/v2"
	"strconv"
	"sync/atomic"
	"time"
)

type OrderStatus int

const (
	OrderUnplaced OrderStatus = iota
	OrderPlacing
	OrderPlaced
	OrderModifying
	OrderCanceling
	OrderFilled
)

func (o OrderStatus) String() string {
	switch o {
	case OrderUnplaced:
		return "Unplaced"
	case OrderPlacing:
		return "Placing"
	case OrderPlaced:
		return "Placed"
	case OrderModifying:
		return "Modifying"
	case OrderCanceling:
		return "Canceling"
	case OrderFilled:
		return "Filled"
	default:
		return "Unknown"
	}
}

type OrderClientID string
type ExchangeID int

const (
	ExchangeNone ExchangeID = iota
	ExchangeBinance
	ExchangeKraken
	ExchangeCount
)

func (eid ExchangeID) String() string {
	switch eid {
	case ExchangeNone:
		return "None"
	case ExchangeBinance:
		return "Binance"
	case ExchangeKraken:
		return "Kraken"
	default:
		return "Unknown"
	}
}

type SymbolID string

type Order struct {
	ClientID OrderClientID
	Exchange ExchangeID
	Symbol   SymbolID
	Amount   uint64
	Price    uint64
}

func NewOrder(clid OrderClientID, exchange ExchangeID, symbol SymbolID, amount uint64, price uint64) Order {
	return Order{
		ClientID: clid,
		Exchange: exchange,
		Symbol:   symbol,
		Amount:   amount,
		Price:    price,
	}
}

var clientIDCounter atomic.Uint32

func GenerateClientOrderID() OrderClientID {
	id := uint64(time.Now().Unix()<<16) | uint64(clientIDCounter.Add(1)&0xFFFF)
	return OrderClientID(strconv.FormatUint(id, 16))
}

func GenerateOrderWithSymbol(symbol SymbolID) Order {
	return Order{
		ClientID: GenerateClientOrderID(),
		Exchange: ExchangeID(rand.IntN(int(ExchangeCount)-1) + 1),
		Symbol:   symbol,
		Amount:   rand.Uint64N(1000000000) + 1,
		Price:    rand.Uint64N(1000000) + 1,
	}
}
