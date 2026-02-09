package services

import (
	"iano_chat/models"

	"gorm.io/gorm"
)

type AgentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) *AgentService {
	return &AgentService{db: db}
}

func (s *AgentService) Create(agent *models.Agent) error {
	return s.db.Create(agent).Error
}

func (s *AgentService) GetByID(id string) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.First(&agent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *AgentService) GetAll() ([]models.Agent, error) {
	var agents []models.Agent
	if err := s.db.Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (s *AgentService) GetByType(agentType models.AgentType) ([]models.Agent, error) {
	var agents []models.Agent
	if err := s.db.Where("type = ?", agentType).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (s *AgentService) Update(id string, updates map[string]interface{}) (*models.Agent, error) {
	var agent models.Agent
	if err := s.db.First(&agent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&agent).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (s *AgentService) Delete(id string) error {
	result := s.db.Delete(&models.Agent{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *AgentService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Agent{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
