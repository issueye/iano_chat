package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"iano_server/models"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
)

type MCPService struct {
	db                *gorm.DB
	ClientManager     *MCPClientManager
	ServerService     *MCPServerService
	ServerToolService *MCPServerToolService
}

func NewMCPService(db *gorm.DB) *MCPService {
	return &MCPService{
		db:                db,
		ClientManager:     NewMCPClientManager(),
		ServerService:     NewMCPServerService(db),
		ServerToolService: NewMCPServerToolService(db),
	}
}

type MCPServerService struct {
	db *gorm.DB
}

func NewMCPServerService(db *gorm.DB) *MCPServerService {
	return &MCPServerService{db: db}
}

func (s *MCPServerService) Create(server *models.MCPServer) error {
	return s.db.Create(server).Error
}

func (s *MCPServerService) GetByID(id string) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.First(&server, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *MCPServerService) GetAll() ([]models.MCPServer, error) {
	var servers []models.MCPServer
	if err := s.db.Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *MCPServerService) GetEnabled() ([]models.MCPServer, error) {
	var servers []models.MCPServer
	if err := s.db.Where("enabled = ?", true).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *MCPServerService) Update(id string, updates map[string]interface{}) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.First(&server, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&server).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *MCPServerService) Delete(id string) error {
	result := s.db.Delete(&models.MCPServer{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *MCPServerService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.MCPServer{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type MCPServerToolService struct {
	db *gorm.DB
}

func NewMCPServerToolService(db *gorm.DB) *MCPServerToolService {
	return &MCPServerToolService{db: db}
}

func (s *MCPServerToolService) Create(tool *models.MCPServerTool) error {
	return s.db.Create(tool).Error
}

func (s *MCPServerToolService) GetByServerID(serverID string) ([]models.MCPServerTool, error) {
	var tools []models.MCPServerTool
	if err := s.db.Where("server_id = ?", serverID).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *MCPServerToolService) DeleteByServerID(serverID string) error {
	return s.db.Where("server_id = ?", serverID).Delete(&models.MCPServerTool{}).Error
}

type MCPClientManager struct {
	clients map[string]client.MCPClient
	mu      sync.RWMutex
}

func NewMCPClientManager() *MCPClientManager {
	return &MCPClientManager{
		clients: make(map[string]client.MCPClient),
	}
}

func (m *MCPClientManager) GetClient(serverID string) (client.MCPClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cli, ok := m.clients[serverID]
	return cli, ok
}

func (m *MCPClientManager) SetClient(serverID string, cli client.MCPClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[serverID] = cli
}

func (m *MCPClientManager) RemoveClient(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, serverID)
}

func (m *MCPClientManager) CloseClient(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cli, ok := m.clients[serverID]; ok {
		delete(m.clients, serverID)
		return cli.Close()
	}
	return nil
}

func (s *MCPService) ConnectServer(ctx context.Context, serverID string) error {
	server, err := s.ServerService.GetByID(serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	if err := s.DisconnectServer(serverID); err != nil {
	}

	var mcpClient client.MCPClient
	var errCreate error

	switch server.Transport {
	case models.MCPTransportStdio:
		var cmdArgs []string
		if server.Args != "" {
			json.Unmarshal([]byte(server.Args), &cmdArgs)
		}
		var envVars []string
		if server.Env != "" {
			json.Unmarshal([]byte(server.Env), &envVars)
		}
		mcpClient, errCreate = client.NewStdioMCPClient(server.Command, envVars, cmdArgs...)
	case models.MCPTransportSSE, models.MCPTransportHTTP:
		mcpClient, errCreate = client.NewSSEMCPClient(server.URL)
	default:
		return fmt.Errorf("unsupported transport type: %s", server.Transport)
	}

	if errCreate != nil {
		s.ServerService.Update(serverID, map[string]interface{}{
			"status":     models.MCPServerStatusError,
			"last_error": errCreate.Error(),
		})
		return fmt.Errorf("failed to create client: %w", errCreate)
	}

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: struct {
			ProtocolVersion string                 `json:"protocolVersion"`
			Capabilities    mcp.ClientCapabilities `json:"capabilities"`
			ClientInfo      mcp.Implementation     `json:"clientInfo"`
		}{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "iano_server",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		mcpClient.Close()
		s.ServerService.Update(serverID, map[string]interface{}{
			"status":     models.MCPServerStatusError,
			"last_error": err.Error(),
		})
		return fmt.Errorf("failed to initialize: %w", err)
	}

	s.ClientManager.SetClient(serverID, mcpClient)

	result, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err == nil && result.Tools != nil {
		s.ServerToolService.DeleteByServerID(serverID)
		for _, tool := range result.Tools {
			schema, _ := json.Marshal(tool.InputSchema)
			s.ServerToolService.Create(&models.MCPServerTool{
				ServerID:    serverID,
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: string(schema),
			})
		}
		s.ServerService.Update(serverID, map[string]interface{}{
			"status":      models.MCPServerStatusConnected,
			"tools_count": len(result.Tools),
		})
	} else {
		s.ServerService.Update(serverID, map[string]interface{}{
			"status": models.MCPServerStatusConnected,
		})
	}

	return nil
}

func (s *MCPService) DisconnectServer(serverID string) error {
	return s.ClientManager.CloseClient(serverID)
}

func (s *MCPService) ConnectAllEnabled(ctx context.Context) error {
	servers, err := s.ServerService.GetEnabled()
	if err != nil {
		return fmt.Errorf("failed to get enabled servers: %w", err)
	}

	for _, server := range servers {
		if err := s.ConnectServer(ctx, server.ID); err != nil {
			continue
		}
	}
	return nil
}

func (s *MCPService) CallTool(ctx context.Context, serverID string, toolName string, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	mcpClient, ok := s.ClientManager.GetClient(serverID)
	if !ok {
		return nil, fmt.Errorf("client not connected for server: %s", serverID)
	}

	return mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: arguments,
		},
	})
}

func (s *MCPService) ListTools(ctx context.Context, serverID string) (*mcp.ListToolsResult, error) {
	mcpClient, ok := s.ClientManager.GetClient(serverID)
	if !ok {
		return nil, fmt.Errorf("client not connected for server: %s", serverID)
	}

	return mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
}

func (s *MCPService) Ping(ctx context.Context, serverID string) error {
	mcpClient, ok := s.ClientManager.GetClient(serverID)
	if !ok {
		return fmt.Errorf("client not connected for server: %s", serverID)
	}

	return mcpClient.Ping(ctx)
}
