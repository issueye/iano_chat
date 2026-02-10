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

func TestToolController(t *testing.T) {
	// 创建测试数据库
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// 创建路由引擎
	engine := routes.SetupRoutes(testDB.DB)

	t.Run("Create Tool", func(t *testing.T) {
		reqBody := `{
			"name": "Test Tool",
			"desc": "A test tool",
			"type": "builtin",
			"status": "enabled",
			"config": "{}"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/tools", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusCreated)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get All Tools", func(t *testing.T) {
		service := services.NewToolService(testDB.DB)
		tool := &models.Tool{
			Name:   "Tool 1",
			Desc:   "Description 1",
			Type:   models.ToolTypeBuiltin,
			Status: models.ToolStatusEnabled,
		}
		tool.NewID() // Generate UUID for ID
		service.Create(tool)

		req := httptest.NewRequest(http.MethodGet, "/api/tools", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Tool By ID", func(t *testing.T) {
		service := services.NewToolService(testDB.DB)
		tool := &models.Tool{
			Name:   "Tool By ID",
			Desc:   "Description",
			Type:   models.ToolTypeBuiltin,
			Status: models.ToolStatusEnabled,
		}
		tool.NewID() // Generate UUID for ID
		service.Create(tool)

		req := httptest.NewRequest(http.MethodGet, "/api/tools/"+tool.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Tools By Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tools/type?type=builtin", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Tools By Status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tools/status?status=enabled", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Update Tool", func(t *testing.T) {
		service := services.NewToolService(testDB.DB)
		tool := &models.Tool{
			Name:   "Tool To Update",
			Desc:   "Original Description",
			Type:   models.ToolTypeBuiltin,
			Status: models.ToolStatusEnabled,
		}
		tool.NewID() // Generate UUID for ID
		service.Create(tool)

		reqBody := `{"name": "Updated Tool Name"}`

		req := httptest.NewRequest(http.MethodPut, "/api/tools/"+tool.ID, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Update Tool Config", func(t *testing.T) {
		service := services.NewToolService(testDB.DB)
		tool := &models.Tool{
			Name:   "Tool Update Config",
			Desc:   "Description",
			Type:   models.ToolTypeBuiltin,
			Status: models.ToolStatusEnabled,
			Config: `{}`,
		}
		tool.NewID() // Generate UUID for ID
		service.Create(tool)

		reqBody := `{"timeout": 30, "retry_count": 3}`

		req := httptest.NewRequest(http.MethodPut, "/api/tools/"+tool.ID+"/config", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Delete Tool", func(t *testing.T) {
		service := services.NewToolService(testDB.DB)
		tool := &models.Tool{
			Name:   "Tool To Delete",
			Desc:   "Will be deleted",
			Type:   models.ToolTypeBuiltin,
			Status: models.ToolStatusEnabled,
		}
		tool.NewID() // Generate UUID for ID
		service.Create(tool)

		req := httptest.NewRequest(http.MethodDelete, "/api/tools/"+tool.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)

		// Verify deletion
		_, err := service.GetByID(tool.ID)
		if err == nil {
			t.Error("Tool should have been deleted")
		}
	})
}
