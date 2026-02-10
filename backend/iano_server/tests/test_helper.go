package tests

import (
	"bytes"
	"encoding/json"
	"iano_server/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB 测试数据库实例
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB 创建测试数据库（使用SQLite内存模式）
func NewTestDB() (*TestDB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移模型
	err = db.AutoMigrate(
		&models.Provider{},
		&models.Session{},
		&models.Message{},
		&models.Agent{},
		&models.Tool{},
	)
	if err != nil {
		return nil, err
	}

	return &TestDB{DB: db}, nil
}

// Close 关闭测试数据库
func (td *TestDB) Close() error {
	sqlDB, err := td.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// MakeRequest 创建HTTP请求辅助函数
func MakeRequest(method, url string, body interface{}) (*httptest.ResponseRecorder, *http.Request) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	return httptest.NewRecorder(), req
}

// ParseResponse 解析响应体
func ParseResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	return response
}

// AssertStatusCode 断言状态码
func AssertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	if rr.Code != expected {
		t.Errorf("Expected status code %d, got %d", expected, rr.Code)
	}
}

// AssertSuccess 断言成功响应
func AssertSuccess(t *testing.T, response map[string]interface{}) {
	if code, ok := response["code"].(float64); !ok || code != 200 {
		t.Errorf("Expected success code 200, got %v", response["code"])
	}
}

// AssertError 断言错误响应
func AssertError(t *testing.T, response map[string]interface{}) {
	if code, ok := response["code"].(float64); !ok || code == 200 {
		t.Errorf("Expected error code, got %v", response["code"])
	}
}

// ToJSON 转换为JSON字符串
func ToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
