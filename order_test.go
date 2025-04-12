package orderstracker

import "testing"

func Test_GenerateClientOrderID(t *testing.T) {
	got := GenerateClientOrderID()
	if got == "" {
		t.Error("Should not return empty string")
	}
	gotNext := GenerateClientOrderID()
	if got == gotNext {
		t.Error("Should return unique id")
	}
}

func Test_GenerateOrderWithSymbol(t *testing.T) {
	wantSymbol := SymbolID("TEST")
	got := GenerateOrderWithSymbol(wantSymbol)
	if got.ClientID == "" {
		t.Error("Should not return order with empty id")
	}
	if got.Exchange == ExchangeNone {
		t.Error("Should not return order with empty exchange")
	}
	if got.Exchange >= ExchangeCount {
		t.Error("Should not return order with invalid exchange")
	}
	if got.Symbol != wantSymbol {
		t.Errorf("Should have specified symbol: %v != %v", got.Symbol, wantSymbol)
	}
	if got.Price == 0 {
		t.Error("Price should not be zero")
	}
	if got.Amount == 0 {
		t.Error("Amount should not be zero")
	}
}
