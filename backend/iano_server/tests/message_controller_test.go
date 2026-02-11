package tests

import (
	"iano_server/models"
	"iano_server/routes"
	"iano_server/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMessageController(t *testing.T) {
	// 创建测试数据库
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// 创建路由引擎
	engine := routes.SetupRoutes(testDB.DB)

	// 使用固定的 session ID (string) 用于消息测试
	var testSessionID string = "123456"

	t.Run("Create Message", func(t *testing.T) {
		reqBody := `{
			"session_id": "` + testSessionID + `",
			"type": "user",
			"content": "{\"text\": \"Hello, this is a test message\"}",
			"status": "completed"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusCreated)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get All Messages", func(t *testing.T) {
		service := services.NewMessageService(testDB.DB)
		message := models.CreateUserMessage(testSessionID, "Test message 1")
		service.Create(message)

		req := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Messages By SessionID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/messages/session?session_id="+testSessionID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Message By ID", func(t *testing.T) {
		service := services.NewMessageService(testDB.DB)
		message := models.CreateAssistantMessage(testSessionID, models.MessageStatusCompleted)
		message.Content = `{"text": "Test message by ID"}`
		service.Create(message)

		req := httptest.NewRequest(http.MethodGet, "/api/messages/"+message.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Update Message", func(t *testing.T) {
		service := services.NewMessageService(testDB.DB)
		message := models.CreateUserMessage(testSessionID, "Original content")
		service.Create(message)

		reqBody := `{"content": "{\"text\": \"Updated content\"}"}`

		req := httptest.NewRequest(http.MethodPut, "/api/messages/"+message.ID, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Delete Message", func(t *testing.T) {
		service := services.NewMessageService(testDB.DB)
		message := models.CreateUserMessage(testSessionID, "Message to delete")
		service.Create(message)

		req := httptest.NewRequest(http.MethodDelete, "/api/messages/"+message.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Messages By Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/messages/type?type=user", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Add Message Feedback", func(t *testing.T) {
		service := services.NewMessageService(testDB.DB)
		message := models.CreateAssistantMessage(testSessionID, models.MessageStatusCompleted)
		message.Content = `{"text": "Assistant response"}`
		service.Create(message)

		reqBody := `{"rating": "like", "comment": "Very helpful response"}`

		req := httptest.NewRequest(http.MethodPost, "/api/messages/"+message.ID+"/feedback", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Note: The feedback endpoint might have different behavior
		// Adjust expected status code based on actual implementation
		if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
			t.Errorf("Expected status code %d or %d, got %d", http.StatusOK, http.StatusCreated, rr.Code)
		}
	})
}
