package currencies

import "testing"

func TestDefaultCurrencyNamesAndLocales(t *testing.T) {
	currencies, err := CollectionCurrencies(CollectDefaults())
	if err != nil {
		t.Fatalf("collect default currencies: %v", err)
	}

	registry := NewCurrencyService(currencies)
	want := map[string]struct {
		name  string
		local string
	}{
		"DKK": {name: "Danish krone", local: "Denmark"},
		"NOK": {name: "Norwegian krone", local: "Norway"},
	}

	for _, currency := range registry.Slice() {
		expected, ok := want[currency.Code]
		if !ok {
			continue
		}

		if currency.Name != expected.name {
			t.Errorf("currency %s name = %q, want %q", currency.Code, currency.Name, expected.name)
		}
		if currency.Local != expected.local {
			t.Errorf("currency %s locale = %q, want %q", currency.Code, currency.Local, expected.local)
		}
		delete(want, currency.Code)
	}

	for code := range want {
		t.Errorf("currency %s not found in default registry", code)
	}
}
