package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

// NullFloat32 可空的 float32 类型，用于处理数据库中的 NULL 或空字符串
type NullFloat32 struct {
	Float32 float32
	Valid   bool
}

// Value 实现 driver.Valuer 接口
func (n NullFloat32) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return float64(n.Float32), nil
}

// Scan 实现 sql.Scanner 接口
func (n *NullFloat32) Scan(value interface{}) error {
	if value == nil {
		n.Float32 = 0
		n.Valid = false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case float32:
		n.Float32 = v
	case float64:
		n.Float32 = float32(v)
	case int:
		n.Float32 = float32(v)
	case int64:
		n.Float32 = float32(v)
	case string:
		if v == "" {
			n.Float32 = 0
			n.Valid = false
			return nil
		}
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return fmt.Errorf("cannot parse float32 from string '%s': %w", v, err)
		}
		n.Float32 = float32(f)
	case []byte:
		return n.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into NullFloat32", value)
	}
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (n NullFloat32) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatFloat(float64(n.Float32), 'f', -1, 32)), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (n *NullFloat32) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == "" {
		n.Float32 = 0
		n.Valid = false
		return nil
	}
	f, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return err
	}
	n.Float32 = float32(f)
	n.Valid = true
	return nil
}

// NullInt 可空的 int 类型，用于处理数据库中的 NULL 或空字符串
type NullInt struct {
	Int   int
	Valid bool
}

// Value 实现 driver.Valuer 接口
func (n NullInt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return int64(n.Int), nil
}

// Scan 实现 sql.Scanner 接口
func (n *NullInt) Scan(value interface{}) error {
	if value == nil {
		n.Int = 0
		n.Valid = false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case int:
		n.Int = v
	case int8:
		n.Int = int(v)
	case int16:
		n.Int = int(v)
	case int32:
		n.Int = int(v)
	case int64:
		n.Int = int(v)
	case uint:
		n.Int = int(v)
	case uint8:
		n.Int = int(v)
	case uint16:
		n.Int = int(v)
	case uint32:
		n.Int = int(v)
	case uint64:
		n.Int = int(v)
	case float32:
		n.Int = int(v)
	case float64:
		n.Int = int(v)
	case string:
		if v == "" {
			n.Int = 0
			n.Valid = false
			return nil
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("cannot parse int from string '%s': %w", v, err)
		}
		n.Int = i
	case []byte:
		return n.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into NullInt", value)
	}
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (n NullInt) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.Itoa(n.Int)), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (n *NullInt) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == "" {
		n.Int = 0
		n.Valid = false
		return nil
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	n.Int = i
	n.Valid = true
	return nil
}
