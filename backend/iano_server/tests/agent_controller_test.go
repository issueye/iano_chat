package tests

import (
	"encoding/json"
	"iano_server/controllers"
	"iano_server/models"
	"iano_server/routes"
	"iano_server/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAgentController(t *testing.T) {
	// 创建测试数据库
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// 创建路由引擎
	engine := routes.SetupRoutes(testDB.DB)

	t.Run("Create Agent", func(t *testing.T) {
		reqBody := `{
			"name": "Test Agent",
			"description": "Test Description",
			"type": "main",
			"is_sub_agent": false,
			"provider_id": "provider-1",
			"model": "gpt-4",
			"instructions": "Test instructions",
			"tools": "[]"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusCreated)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)

		if response["data"] == nil {
			t.Error("Expected data in response")
		}
	})

	t.Run("Create Agent With Invalid JSON", func(t *testing.T) {
		reqBody := `{"invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return bad request due to JSON parse error
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("Get All Agents", func(t *testing.T) {
		// First create an agent
		service := services.NewAgentService(testDB.DB)
		agent := &models.Agent{
			Name:        "Agent 1",
			Description: "Description 1",
			Type:        models.AgentTypeMain,
		}
		agent.NewID() // Generate UUID for ID
		service.Create(agent)

		req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Agent By ID", func(t *testing.T) {
		// Create an agent first
		service := services.NewAgentService(testDB.DB)
		agent := &models.Agent{
			Name:        "Test Agent By ID",
			Description: "Test Description",
			Type:        models.AgentTypeMain,
		}
		agent.NewID() // Generate UUID for ID
		service.Create(agent)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/"+agent.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Get Agent By Non-existent ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/agents/non-existent-id-12345", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusNotFound)
	})

	t.Run("Update Agent", func(t *testing.T) {
		// Create an agent first
		service := services.NewAgentService(testDB.DB)
		agent := &models.Agent{
			Name:        "Agent To Update",
			Description: "Original Description",
			Type:        models.AgentTypeMain,
		}
		agent.NewID() // Generate UUID for ID
		service.Create(agent)

		reqBody := `{"name": "Updated Agent Name"}`

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agent.ID, strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	t.Run("Delete Agent", func(t *testing.T) {
		// Create an agent first
		service := services.NewAgentService(testDB.DB)
		agent := &models.Agent{
			Name:        "Agent To Delete",
			Description: "Will be deleted",
			Type:        models.AgentTypeMain,
		}
		agent.NewID() // Generate UUID for ID
		service.Create(agent)

		req := httptest.NewRequest(http.MethodDelete, "/api/agents/"+agent.ID, nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)

		// Verify deletion
		_, err := service.GetByID(agent.ID)
		if err == nil {
			t.Error("Agent should have been deleted")
		}
	})

	t.Run("Get Agents By Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/agents/type?type=main", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusOK)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})
}

func TestAgentControllerUnit(t *testing.T) {
	t.Run("CreateAgentRequest Validation", func(t *testing.T) {
		// Test request struct
		req := controllers.CreateAgentRequest{
			Name:       "Test",
			Type:       "main",
			ProviderID: "provider-1",
			Model:      "gpt-4",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		var decoded controllers.CreateAgentRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal request: %v", err)
		}

		if decoded.Name != req.Name {
			t.Error("Name mismatch after serialization")
		}
	})

	t.Run("UpdateAgentRequest With Partial Fields", func(t *testing.T) {
		name := "Updated Name"
		req := controllers.UpdateAgentRequest{
			Name: &name,
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		var decoded controllers.UpdateAgentRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal request: %v", err)
		}

		if decoded.Name == nil || *decoded.Name != name {
			t.Error("Name pointer mismatch after serialization")
		}
	})
}
