package container

import (
	"context"
	"iano_server/controllers"
	"iano_server/pkg/config"
	"iano_server/services"
	web "iano_web"

	"gorm.io/gorm"
)

type Container struct {
	DB     *gorm.DB
	SSEHub *web.SSEHub
	Cfg    *config.Config

	Ctx                 context.Context
	AgentService        *services.AgentService
	MessageService      *services.MessageService
	SessionService      *services.SessionService
	ToolService         *services.ToolService
	ProviderService     *services.ProviderService
	MCPService          *services.MCPService
	AgentRuntimeService *services.AgentRuntimeService

	AgentSSEClientMap *services.AgentSSEClientMap

	AgentController    *controllers.AgentController
	MessageController  *controllers.MessageController
	SessionController  *controllers.SessionController
	ToolController     *controllers.ToolController
	ProviderController *controllers.ProviderController
	ChatController     *controllers.ChatController
	MCPController      *controllers.MCPController
	BaseController     *controllers.BaseController
}

func NewContainer(ctx context.Context, db *gorm.DB, cfg *config.Config) *Container {
	cnr := &Container{
		Ctx: ctx,
	}

	cnr.Provide(db, cfg)

	return cnr
}

func (c *Container) Provide(db *gorm.DB, cfg *config.Config) {
	c.Cfg = cfg

	c.DB = db
	c.SSEHub = web.NewSSEHubWithContext(c.Ctx)

	c.AgentService = services.NewAgentService(db)
	c.MessageService = services.NewMessageService(db)
	c.SessionService = services.NewSessionService(db)
	c.ToolService = services.NewToolService(db)
	c.ProviderService = services.NewProviderService(db)
	c.MCPService = services.NewMCPService(db)
	c.AgentRuntimeService = services.NewAgentRuntimeServiceWithMCP(
		db,
		c.AgentService,
		c.ProviderService,
		c.ToolService,
		c.MCPService,
	)
	c.AgentSSEClientMap = services.NewAgentSSEClientMap()

	c.AgentController = controllers.NewAgentController(c.AgentService, c.AgentRuntimeService)
	c.MessageController = controllers.NewMessageController(c.MessageService)
	c.SessionController = controllers.NewSessionController(c.SessionService)
	c.ToolController = controllers.NewToolController(c.ToolService)
	c.ProviderController = controllers.NewProviderController(c.ProviderService)
	c.ChatController = controllers.NewChatController(
		c.AgentService,
		c.ProviderService,
		c.MessageService,
		c.AgentRuntimeService,
		c.AgentSSEClientMap,
	)
	c.MCPController = controllers.NewMCPController(c.MCPService)
	c.BaseController = &controllers.BaseController{}
}

// GetConfig 获取配置
func (c *Container) GetConfig() *config.Config {
	return c.Cfg
}

// GetDB 获取数据库连接
func (c *Container) GetDB() *gorm.DB {
	return c.DB
}

// GetSSEHub 获取 SSE 集线器
func (c *Container) GetSSEHub() *web.SSEHub {
	return c.SSEHub
}
