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

// TestHealthCheck 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	engine := routes.SetupRoutes(testDB.DB)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	engine.ServeHTTP(rr, req)

	AssertStatusCode(t, rr, http.StatusOK)
}

// TestIntegrationCompleteWorkflow 测试完整的工作流程
func TestIntegrationCompleteWorkflow(t *testing.T) {
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	engine := routes.SetupRoutes(testDB.DB)

	// 1. 创建一个 Agent
	t.Run("Step1_CreateAgent", func(t *testing.T) {
		reqBody := `{
			"name": "Integration Test Agent",
			"description": "Test agent for integration",
			"type": "main",
			"is_sub_agent": false,
			"provider_id": "provider-1",
			"model": "gpt-4",
			"instructions": "You are a helpful assistant",
			"tools": "[]"
		}`

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusCreated)
		response := ParseResponse(t, rr)
		AssertSuccess(t, response)
	})

	// 2. 创建一个 Session
	t.Run("Step2_CreateSession", func(t *testing.T) {
		reqBody := `{
			"user_id": 1,
			"title": "Integration Test Session",
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

	// 3. 创建一个 Tool
	t.Run("Step3_CreateTool", func(t *testing.T) {
		reqBody := `{
			"name": "Integration Test Tool",
			"desc": "Test tool for integration",
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

	// 4. 验证所有数据都已创建
	t.Run("Step4_VerifyData", func(t *testing.T) {
		// 检查 Agents
		req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		AssertStatusCode(t, rr, http.StatusOK)

		// 检查 Sessions
		req = httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
		rr = httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		AssertStatusCode(t, rr, http.StatusOK)

		// 检查 Tools
		req = httptest.NewRequest(http.MethodGet, "/api/tools", nil)
		rr = httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		AssertStatusCode(t, rr, http.StatusOK)
	})
}

// TestSequentialOperations 测试顺序操作
func TestSequentialOperations(t *testing.T) {
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// Test sequential creation using service layer
	agentService := services.NewAgentService(testDB.DB)
	sessionService := services.NewSessionService(testDB.DB)

	t.Run("SequentialCreateAgents", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			agent := &models.Agent{
				Name:        "Sequential Agent",
				Description: "Test sequential creation",
				Type:        models.AgentTypeMain,
			}
			agent.NewID()
			if err := agentService.Create(agent); err != nil {
				t.Errorf("Failed to create agent %d: %v", i, err)
			}
		}
	})

	t.Run("SequentialCreateSessions", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			session := &models.Session{
				UserID: 1,
				Title:  "Sequential Session",
				Status: models.SessionStatusActive,
			}
			session.NewID() // Generate UUID for ID
			if err := sessionService.Create(session); err != nil {
				t.Errorf("Failed to create session %d: %v", i, err)
			}
		}
	})

	// Verify data
	t.Run("VerifySequentialData", func(t *testing.T) {
		agents, err := agentService.GetAll()
		if err != nil {
			t.Fatalf("Failed to get agents: %v", err)
		}
		if len(agents) < 5 {
			t.Errorf("Expected at least 5 agents, got %d", len(agents))
		}

		sessions, err := sessionService.GetAll()
		if err != nil {
			t.Fatalf("Failed to get sessions: %v", err)
		}
		if len(sessions) < 5 {
			t.Errorf("Expected at least 5 sessions, got %d", len(sessions))
		}
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	testDB, err := NewTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	engine := routes.SetupRoutes(testDB.DB)

	t.Run("InvalidJSON", func(t *testing.T) {
		reqBody := `{"invalid json}`

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return bad request
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d for invalid JSON, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/agents/non-existent-id-12345", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		AssertStatusCode(t, rr, http.StatusNotFound)
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/agents", nil)
		rr := httptest.NewRecorder()

		engine.ServeHTTP(rr, req)

		// Should return method not allowed or not found
		if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d or %d, got %d", http.StatusMethodNotAllowed, http.StatusNotFound, rr.Code)
		}
	})
}
