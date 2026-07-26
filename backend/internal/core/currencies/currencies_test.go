package currencies

import "testing"

func TestDefaultCurrencyNames(t *testing.T) {
	currencies, err := CollectionCurrencies(CollectDefaults())
	if err != nil {
		t.Fatalf("collect default currencies: %v", err)
	}

	registry := NewCurrencyService(currencies)
	want := map[string]string{
		"DKK": "Danish krone",
		"NOK": "Norwegian krone",
	}

	for _, currency := range registry.Slice() {
		name, ok := want[currency.Code]
		if !ok {
			continue
		}

		if currency.Name != name {
			t.Errorf("currency %s name = %q, want %q", currency.Code, currency.Name, name)
		}
		delete(want, currency.Code)
	}

	for code := range want {
		t.Errorf("currency %s not found in default registry", code)
	}
}
