// SPDX-File-CopyrightText: (c) 2025 Andrei Ilin <ortfero@gmail.com>
// SPDX-License-Identifier: MIT

// Package orderstracker provides functionality for tracking market-making order state.
//
// The package supports operations such as:
//   - Registering new orders with the OrderPlacing function.
//   - Confirming order placements with OrderPlaceConfirmed.
//   - Handling order rejections via OrderRejected.
//   - Initiating and confirming order modifications with OrderMoving and OrderMoveConfirmed.
//   - Processing order cancellations using OrderCancelling and OrderCancelConfirmed.
//   - Updating orders as they are filled with OrderFilled, incorporating a VWAP calculation for aggregating trade fills.
//   - Retrieving current order status along with its execution report via GetCurrentStatus.
//   - Updating market quotes using PushQuote.
//
// Designed for trading platforms or order management systems, this package ensures that
// all operations are executed in a thread-safe manner by using a mutex (guard). It is optimized to efficiently
// track and update the lifecycle of orders alongside the dynamic market data from multiple exchanges.
package orderstracker

import (
	"fmt"
	"sync"
	"time"
)

// orderContext holds the context and execution state of an order.
// It contains the current order status, the original order details,
// and the most recent execution report.
type orderContext struct {
	Status     OrderStatus
	Order      Order
	LastReport ExecutionReport
}

// marketData holds the latest market quote data for a symbol.
// It includes bid and ask prices and an optional pointer to an order context
// that may be associated with the market data.
type marketData struct {
	bid          uint64
	ask          uint64
	orderContext *orderContext
}

// Tracker is responsible for tracking the state of orders and market data.
// It maintains a synchronized view of orders across different exchanges and symbols.
type Tracker struct {
	guard     sync.Mutex
	exchanges map[ExchangeID]map[SymbolID]marketData
	orders    map[OrderClientID]*orderContext
}

// NewTracker creates and initializes a new Tracker instance.
// It returns a pointer to a Tracker with properly initialized maps for exchanges and orders.
func NewTracker() *Tracker {
	return &Tracker{
		exchanges: make(map[ExchangeID]map[SymbolID]marketData),
		orders:    make(map[OrderClientID]*orderContext),
	}
}

// OrderPlacing registers a new order in the tracker as pending placement.
// If the order already exists, it returns an error.
func (t *Tracker) OrderPlacing(order Order) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	if _, exists := t.orders[order.ClientID]; exists {
		return fmt.Errorf("order already placed (clid %v)", order.ClientID)
	}

	orderContext := &orderContext{
		Status: OrderPlacing,
		Order:  order,
	}
	t.orders[order.ClientID] = orderContext

	exchange := t.exchanges[order.Exchange]
	if exchange == nil {
		exchange = make(map[SymbolID]marketData)
		t.exchanges[order.Exchange] = exchange
	}
	symbolContext := exchange[order.Symbol]
	symbolContext.orderContext = orderContext
	exchange[order.Symbol] = symbolContext
	return nil
}

// OrderPlaceConfirmed confirms that an order has been successfully placed.
// It takes the order's client ID and the confirmation time as parameters.
// Returns an error if the order is not found or if the current status is not OrderPlacing.
func (t *Tracker) OrderPlaceConfirmed(clid OrderClientID, time time.Time) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}
	orderContext.LastReport.Kind = ReportPlaced
	orderContext.LastReport.Time = time

	if orderContext.Status != OrderPlacing {
		return fmt.Errorf("order status is not 'OrderPlacing' (clid %v, status '%s')",
			clid, orderContext.Status)
	}

	orderContext.Status = OrderPlaced
	return nil
}

// OrderRejected updates an order's state to indicate that it has been rejected.
// It accepts the order's client ID, the time of rejection, and a reason message.
// Returns an error if the order is not found or if the status does not allow for rejection.
func (t *Tracker) OrderRejected(clid OrderClientID, time time.Time, reason string) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}
	orderContext.LastReport.Kind = ReportRejected
	orderContext.LastReport.Time = time
	orderContext.LastReport.Message = reason
	if orderContext.Status == OrderPlacing {
		orderContext.Status = OrderUnplaced
		return nil
	}
	if orderContext.Status == OrderModifying || orderContext.Status == OrderCanceling {
		orderContext.Status = OrderPlaced
		return nil
	}

	return fmt.Errorf("order status should be 'OrderPlacing', 'OrderModifying' or 'OrderCanceling' to reject (clid %v, status '%s')",
		clid, orderContext.Status)
}

// OrderMoving initiates the order price modification.
// It accepts the order's client ID.
// Returns an error if the order is not found or if the order status is not OrderPlaced.
func (t *Tracker) OrderMoving(clid OrderClientID) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}
	if orderContext.Status != OrderPlaced {
		return fmt.Errorf("orderContext status is not 'OrderPlaced' (clid %v, status '%s')",
			clid, orderContext.Status)
	}
	orderContext.Status = OrderModifying
	orderContext.LastReport.Kind = ReportNone
	return nil
}

// OrderMoveConfirmed confirms a previously initiated order modification.
// It takes the order's client ID, the confirmation time, and the new price.
// Returns an error if the order is not found or if the order is not in the OrderModifying state.
func (t *Tracker) OrderMoveConfirmed(clid OrderClientID, time time.Time, price uint64) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}

	orderContext.LastReport.Kind = ReportModified
	orderContext.LastReport.Time = time
	orderContext.LastReport.Price = price

	if orderContext.Status != OrderModifying {
		return fmt.Errorf("order status is not 'OrderModifying' (clid %v, status '%s')",
			clid, orderContext.Status)
	}

	orderContext.Status = OrderPlaced
	orderContext.Order.Price = price
	return nil
}

// OrderCancelling initiates the cancellation process for an active order.
// It takes the order's client ID and validates that the order exists and is in the OrderPlaced state.
// Returns an error if the order does not exist or is not in an appropriate state for cancellation.
func (t *Tracker) OrderCancelling(clid OrderClientID) error {
	t.guard.Lock()
	defer t.guard.Unlock()
	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}
	if orderContext.Status != OrderPlaced {
		return fmt.Errorf("order status is not 'OrderPlaced' (clid %v, status '%s')",
			clid, orderContext.Status)
	}
	orderContext.Status = OrderCanceling
	orderContext.LastReport.Kind = ReportNone
	return nil
}

// OrderCancelConfirmed finalizes an order cancellation.
// It takes the order's client ID and the confirmation time as parameters.
// Returns an error if the order is not found or if the order is not in the OrderCanceling state.
func (t *Tracker) OrderCancelConfirmed(clid OrderClientID, time time.Time) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}

	orderContext.LastReport.Kind = ReportCanceled
	orderContext.LastReport.Time = time

	if orderContext.Status != OrderCanceling {
		return fmt.Errorf("order status is not 'OrderCanceling' (clid %v, status '%s')",
			clid, orderContext.Status)
	}

	orderContext.Status = OrderUnplaced
	return nil
}

// OrderFilled updates an order's state to reflect that it has been filled,
// either fully or partially.
// It accepts the order's client ID, the execution time, the executed amount, and the average price.
// If multiple fills occur, it aggregates the executed amounts and recalculates the price
// using a Volume Weighted Average Price (VWAP) calculation.
// Returns an error if the order is not found.
func (t *Tracker) OrderFilled(clid OrderClientID, time time.Time, executedAmount uint64, avgPrice uint64) error {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return fmt.Errorf("order not found (clid %v)", clid)
	}

	orderContext.Status = OrderFilled
	orderContext.LastReport.Time = time

	// Aggregating trades here with VWAP price
	// Alternative is to store information about each trade
	if orderContext.LastReport.Kind == ReportFilled {
		vwap := (orderContext.LastReport.Amount*orderContext.LastReport.Price + executedAmount*avgPrice) / (orderContext.LastReport.Amount + executedAmount)
		orderContext.LastReport.Price = vwap
		orderContext.LastReport.Amount += executedAmount
	} else { // Single trade
		orderContext.LastReport.Kind = ReportFilled
		orderContext.LastReport.Amount = executedAmount
		orderContext.LastReport.Price = avgPrice
	}

	return nil
}

// GetOrderStatus retrieves the current state and details of an order.
// It takes the order's client ID and pointers to an Order and an ExecutionReport,
// which will be updated with the current order and its latest execution report.
// Returns the current OrderStatus and an error if the order does not exist.
func (t *Tracker) GetOrderStatus(clid OrderClientID, order *Order, executionReport *ExecutionReport) (OrderStatus, error) {
	t.guard.Lock()
	defer t.guard.Unlock()

	orderContext := t.orders[clid]
	if orderContext == nil {
		return OrderUnplaced, fmt.Errorf("order not found (clid %v)", clid)
	}
	*order = orderContext.Order
	*executionReport = orderContext.LastReport
	return orderContext.Status, nil
}

// PushQuote updates the market data for a specific symbol on a specific exchange.
// It accepts the ExchangeID, SymbolID, bid price, and ask price as parameters.
// If no market data exists for the exchange or symbol, new data is created.
// The function also potentially trigger order movements based on the current spread.
func (t *Tracker) PushQuote(exchangeID ExchangeID, symbolID SymbolID, bid uint64, ask uint64) {
	t.guard.Lock()
	defer t.guard.Unlock()

	exchange := t.exchanges[exchangeID]
	if exchange == nil {
		exchange = make(map[SymbolID]marketData)
		t.exchanges[exchangeID] = exchange
	}
	symbolContext := exchange[symbolID]
	symbolContext.bid = bid
	symbolContext.ask = ask
	exchange[symbolID] = symbolContext

	/// TODO: Get signals to move order based on current spread
}

// GetOrdersCount returns the number of tracked orders.
func (t *Tracker) GetOrdersCount() int {
	t.guard.Lock()
	defer t.guard.Unlock()
	return len(t.orders)
}
