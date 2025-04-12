package orderstracker

import "testing"

func TestTracker_OrderPlacing(t *testing.T) {
	tracker := NewTracker()
	wantSymbol := SymbolID("TEST")
	wantOrder := GenerateOrderWithSymbol(wantSymbol)
	if e := tracker.OrderPlacing(wantOrder); e != nil {
		t.Error(e)
	}
	if tracker.GetOrdersCount() != 1 {
		t.Error("Should contain one order after placing")
	}
	var gotOrder Order
	var gotReport ExecutionReport
	gotStatus, e := tracker.GetOrderStatus(wantOrder.ClientID, &gotOrder, &gotReport)
	if e != nil {
		t.Error(e)
	}
	if gotStatus != OrderPlacing {
		t.Error("Order should have 'Placing' status")
	}
}

func BenchmarkTracker_OrderGenerateAndPlace(b *testing.B) {
	tracker := NewTracker()
	wantSymbol := SymbolID("TEST")
	for b.Loop() {
		_ = tracker.OrderPlacing(GenerateOrderWithSymbol(wantSymbol))
	}
}
