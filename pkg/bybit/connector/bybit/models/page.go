package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

type CursorPage[T any] struct {
	Category       string `json:"category"`
	List           []T    `json:"list"`
	NextPageCursor string `json:"nextPageCursor"`
}

type Flex string

func (f *Flex) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" {
		*f = ""
		return nil
	}
	if strings.HasPrefix(s, "\"") {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*f = Flex(str)
		return nil
	}
	*f = Flex(s)
	return nil
}

func (f Flex) Int64() int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(string(f)), 10, 64)
	return v
}

func (f Flex) Float() float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(string(f)), 64)
	return v
}

func (f Flex) String() string {
	return string(f)
}
