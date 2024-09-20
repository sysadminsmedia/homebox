package repo

import (
	"bytes"
	"fmt"
	"strconv"
)

type AssetID int

func (aid AssetID) Nil() bool {
	return aid.Int() <= 0
}

func (aid AssetID) Int() int {
	return int(aid)
}

func ParseAssetIDBytes(d []byte) (aid AssetID, ok bool) {
	d = bytes.ReplaceAll(d, []byte(`"`), []byte(``))
	d = bytes.ReplaceAll(d, []byte(`-`), []byte(``))

	aidInt, err := strconv.Atoi(string(d))
	if err != nil {
		return AssetID(-1), false
	}

	return AssetID(aidInt), true
}

func ParseAssetID(s string) (aid AssetID, ok bool) {
	return ParseAssetIDBytes([]byte(s))
}

func (aid AssetID) String() string {
	if aid.Nil() {
		return ""
	}

	aidStr := fmt.Sprintf("%06d", aid)
	aidStr = fmt.Sprintf("%s-%s", aidStr[:3], aidStr[3:])
	return aidStr
}

func (aid AssetID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + aid.String() + `"`), nil
}

func (aid *AssetID) UnmarshalJSON(d []byte) error {
	if len(d) == 0 || bytes.Equal(d, []byte(`""`)) {
		*aid = -1
		return nil
	}

	d = bytes.ReplaceAll(d, []byte(`"`), []byte(``))
	d = bytes.ReplaceAll(d, []byte(`-`), []byte(``))

	aidInt, err := strconv.Atoi(string(d))
	if err != nil {
		return err
	}

	*aid = AssetID(aidInt)
	return nil
}

func (aid AssetID) MarshalCSV() (string, error) {
	return aid.String(), nil
}

func (aid *AssetID) UnmarshalCSV(d string) error {
	return aid.UnmarshalJSON([]byte(d))
}
