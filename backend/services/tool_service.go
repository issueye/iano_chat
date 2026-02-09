package services

import (
	"iano_chat/models"

	"gorm.io/gorm"
)

type ToolService struct {
	db *gorm.DB
}

func NewToolService(db *gorm.DB) *ToolService {
	return &ToolService{db: db}
}

func (s *ToolService) Create(tool *models.Tool) error {
	return s.db.Create(tool).Error
}

func (s *ToolService) GetByID(id string) (*models.Tool, error) {
	var tool models.Tool
	if err := s.db.First(&tool, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}

func (s *ToolService) GetAll() ([]models.Tool, error) {
	var tools []models.Tool
	if err := s.db.Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *ToolService) GetByType(toolType models.ToolType) ([]models.Tool, error) {
	var tools []models.Tool
	if err := s.db.Where("type = ?", toolType).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *ToolService) GetByStatus(status models.ToolStatus) ([]models.Tool, error) {
	var tools []models.Tool
	if err := s.db.Where("status = ?", status).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *ToolService) Update(id string, updates map[string]interface{}) (*models.Tool, error) {
	var tool models.Tool
	if err := s.db.First(&tool, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&tool).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}

func (s *ToolService) UpdateConfig(id string, config string) (*models.Tool, error) {
	return s.Update(id, map[string]interface{}{"config": config})
}

func (s *ToolService) Delete(id string) error {
	result := s.db.Delete(&models.Tool{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ToolService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Tool{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ToolService) IncrementCallCount(id string) error {
	return s.db.Model(&models.Tool{}).Where("id = ?", id).
		UpdateColumn("call_count", gorm.Expr("call_count + ?", 1)).Error
}

func (s *ToolService) IncrementErrorCount(id string) error {
	return s.db.Model(&models.Tool{}).Where("id = ?", id).
		UpdateColumn("error_count", gorm.Expr("error_count + ?", 1)).Error
}
