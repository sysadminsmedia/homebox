package repo

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAssetIDSearch(t *testing.T) {
	const top = AssetID(math.MaxInt64)
	valid := map[string]AssetIDRange{
		"#>=100":                {100, top},
		"#>= 100":               {100, top},
		"# >= 000-100":          {100, top},
		"asset-id >= 000-100":   {100, top},
		"Asset ID>=100":         {100, top},
		"assetid >= 100":        {100, top},
		"#>100":                 {101, top},
		"#<=200":                {1, 200},
		"#<200":                 {1, 199},
		"#>=100 <=200":          {100, 200},
		"#>=100<=200":           {100, 200},
		"#<=200 >=100":          {100, 200},
		"#>=0":                  {1, top},
		"#<1":                   {1, 0},
		"#>=200 <=100":          {200, 100},
		"#>9223372036854775807": {1, 0},
	}
	for in, want := range valid {
		got, ok := ParseAssetIDSearch(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", "#", "#000-123", "#abc", "#>=", "#>=abc", "#>=100 books", "books #>=100", "asset-id", "asset-id 100", "#>=1 <=2 <=3", "#>=99999999999999999999"} {
		_, ok := ParseAssetIDSearch(in)
		assert.False(t, ok, in)
	}
}

func TestEntityQuery_ResolveAssetIDSearch(t *testing.T) {
	exact := EntityQuery{Search: "#000-123"}
	exact.ResolveAssetIDSearch()
	assert.Equal(t, AssetID(123), exact.AssetID)
	assert.Empty(t, exact.Search)
	assert.Nil(t, exact.AssetIDRange)

	ranged := EntityQuery{Search: "#>=100"}
	ranged.ResolveAssetIDSearch()
	require.NotNil(t, ranged.AssetIDRange)
	assert.Empty(t, ranged.Search)

	text := EntityQuery{Search: "#abc"}
	text.ResolveAssetIDSearch()
	assert.Equal(t, "#abc", text.Search)
	assert.Nil(t, text.AssetIDRange)
}

func TestAssetID_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		aid     AssetID
		want    []byte
		wantErr bool
	}{
		{
			name: "basic test",
			aid:  123,
			want: []byte(`"000-123"`),
		},
		{
			name: "zero test",
			aid:  0,
			want: []byte(`""`),
		},
		{
			name: "large int",
			aid:  123456789,
			want: []byte(`"123-456789"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.aid.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetID.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssetID.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssetID_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		aid     *AssetID
		args    args
		want    AssetID
		wantErr bool
	}{
		{
			name: "basic test",
			aid:  new(AssetID),
			want: 123,
			args: args{
				data: []byte(`{"AssetID":"000123"}`),
			},
		},
		{
			name: "dashed format",
			aid:  new(AssetID),
			want: 123,
			args: args{
				data: []byte(`{"AssetID":"000-123"}`),
			},
		},
		{
			name: "no leading zeros",
			aid:  new(AssetID),
			want: 123,
			args: args{
				data: []byte(`{"AssetID":"123"}`),
			},
		},
		{
			name: "trailing zeros",
			aid:  new(AssetID),
			want: 123000,
			args: args{
				data: []byte(`{"AssetID":"000123000"}`),
			},
		},
		{
			name: "large int",
			aid:  new(AssetID),
			want: 123456789,
			args: args{
				data: []byte(`{"AssetID":"123456789"}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := struct {
				AssetID AssetID `json:"AssetID"`
			}{}

			err := json.Unmarshal(tt.args.data, &st)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssetID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if st.AssetID != tt.want {
				t.Errorf("AssetID.UnmarshalJSON() = %v, want %v", st.AssetID, tt.want)
			}
		})
	}
}
