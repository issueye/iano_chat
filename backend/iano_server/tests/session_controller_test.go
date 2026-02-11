package tests

import (
	"iano_server/models"
	"iano_server/pkg/config"
	"iano_server/routes"
	"iano_server/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionController(t *testing.T) {
	// 创建测试数据库
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()
	config := config.Load("")
	// 创建路由引擎
	engine := routes.SetupRoutes(testDB.DB, config)

	t.Run("Create Session", func(t *testing.T) {
		reqBody := `{
			"title": "Test Session",
			"status": "active"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/sessions", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusCreated)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get All Sessions", func(t *testing.T) {
		// Create a session first
		service := services.NewSessionService(testDB.DB)
		session := &models.Session{
			Title:  "Session 1",
			Status: models.SessionStatusActive,
		}
		service.Create(session)

		req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Session By ID - Invalid ID Format", func(t *testing.T) {
		// Controller expects numeric ID but model uses UUID
		// So using a string ID should return 400 Bad Request
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/invalid-uuid-string", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Controller tries to parse ID as int64, so invalid format returns 400
		AssertStatusCode(t, rr, http.StatusBadRequest)
	})

	t.Run("Get Session By ID - Valid Numeric ID But Not Found", func(t *testing.T) {
		// Using a numeric ID that doesn't exist in database
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/99999", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return 404 as this ID doesn't exist
		AssertStatusCode(t, rr, http.StatusNotFound)
	})

	t.Run("Get Sessions By Status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/status?status=active", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Update Session - Invalid ID", func(t *testing.T) {
		reqBody := `{"title": "Updated Title"}`

		req := httptest.NewRequest(http.MethodPut, "/api/sessions/invalid-id", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return 400 for invalid ID format
		AssertStatusCode(t, rr, http.StatusBadRequest)
	})

	t.Run("Delete Session - Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/sessions/invalid-id", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return 400 for invalid ID format
		AssertStatusCode(t, rr, http.StatusBadRequest)
	})

	t.Run("Get Session Config - Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/sessions/invalid-id/config", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return 400 for invalid ID format
		AssertStatusCode(t, rr, http.StatusBadRequest)
	})

	t.Run("Update Session Config - Invalid ID", func(t *testing.T) {
		reqBody := `{"temperature": 0.9}`

		req := httptest.NewRequest(http.MethodPut, "/api/sessions/invalid-id/config", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return 400 for invalid ID format
		AssertStatusCode(t, rr, http.StatusBadRequest)
	})
}
