package models

import (
	"database/sql/driver"
	"strings"
)

type StrArray []string

func (s *StrArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = StrArray(strings.Split(v, ","))
	case []byte:
		if len(v) == 0 {
			return nil
		}
		*s = StrArray(strings.Split(string(v), ","))
	}

	return nil
}

func (s StrArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}

	return strings.Join(s, ","), nil
}
