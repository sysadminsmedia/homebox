package v1

import (
	"encoding/json"
	"testing"
)

func TestEntityTemplateCreateItemRequestLowStockThresholdPresence(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		present     bool
		threshold   *float64
	}{
		{
			name:    "omitted",
			payload: `{}`,
		},
		{
			name:    "explicit null",
			payload: `{"lowStockThreshold":null}`,
			present: true,
		},
		{
			name:      "value",
			payload:   `{"lowStockThreshold":3.5}`,
			present:   true,
			threshold: func() *float64 { value := 3.5; return &value }(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var request EntityTemplateCreateItemRequest
			if err := json.Unmarshal([]byte(test.payload), &request); err != nil {
				t.Fatal(err)
			}

			if request.lowStockThresholdSet != test.present {
				t.Fatalf("lowStockThresholdSet = %t, want %t", request.lowStockThresholdSet, test.present)
			}

			if test.threshold == nil {
				if request.LowStockThreshold != nil {
					t.Fatalf("LowStockThreshold = %v, want nil", *request.LowStockThreshold)
				}
				return
			}

			if request.LowStockThreshold == nil || *request.LowStockThreshold != *test.threshold {
				t.Fatalf("LowStockThreshold = %v, want %v", request.LowStockThreshold, *test.threshold)
			}
		})
	}
}
