package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// JSONTime 自定义时间类型，用于处理 SQLite 中的时间存储
type JSONTime struct {
	time.Time
	Valid bool // 用于处理 NULL 值
}

// MarshalJSON 实现 json.Marshaler 接口
func (t JSONTime) MarshalJSON() ([]byte, error) {
	if !t.Valid || t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", t.Format(time.RFC3339))), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (t *JSONTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == "" {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}
	// 去除引号
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	// 尝试多种时间格式
	parsed, err := parseTime(str)
	if err != nil {
		return err
	}
	t.Time = parsed
	t.Valid = true
	return nil
}

// Value 实现 driver.Valuer 接口
func (t JSONTime) Value() (driver.Value, error) {
	if !t.Valid || t.IsZero() {
		return nil, nil
	}
	return t.Format("2006-01-02 15:04:05"), nil
}

// Scan 实现 sql.Scanner 接口
func (t *JSONTime) Scan(value interface{}) error {
	// 处理 NULL 值
	if value == nil {
		t.Time = time.Time{}
		t.Valid = false
		return nil
	}

	t.Valid = true
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		if v == "" {
			t.Time = time.Time{}
			t.Valid = false
			return nil
		}
		parsed, err := parseTime(v)
		if err != nil {
			return fmt.Errorf("cannot parse time '%s': %w", v, err)
		}
		t.Time = parsed
		return nil
	case []byte:
		return t.Scan(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into JSONTime", value)
	}
}

// parseTime 尝试用多种格式解析时间
func parseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05.000000000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// GormDataType 返回 gorm 数据类型
func (JSONTime) GormDataType() string {
	return "datetime"
}

// PtrJSONTime 返回指向 JSONTime 的指针
func PtrJSONTime(t time.Time) *JSONTime {
	return &JSONTime{Time: t, Valid: true}
}
