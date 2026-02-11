package services

import (
	"iano_server/models"

	"gorm.io/gorm"
)

type ProviderService struct {
	db *gorm.DB
}

func NewProviderService(db *gorm.DB) *ProviderService {
	return &ProviderService{db: db}
}

func (s *ProviderService) Create(provider *models.Provider) error {
	if provider.IsDefault {
		s.db.Model(&models.Provider{}).Where("is_default = ?", true).Update("is_default", false)
	}
	return s.db.Create(provider).Error
}

func (s *ProviderService) GetByID(id string) (*models.Provider, error) {
	var provider models.Provider
	if err := s.db.First(&provider, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *ProviderService) GetAll() ([]models.Provider, error) {
	var providers []models.Provider
	if err := s.db.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (s *ProviderService) Update(id string, updates map[string]interface{}) (*models.Provider, error) {
	var provider models.Provider
	if err := s.db.First(&provider, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if isDefault, ok := updates["is_default"]; ok && isDefault == true {
		s.db.Model(&models.Provider{}).Where("is_default = ? AND id != ?", true, id).Update("is_default", false)
	}

	if err := s.db.Model(&provider).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (s *ProviderService) Delete(id string) error {
	result := s.db.Delete(&models.Provider{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *ProviderService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Provider{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ProviderService) GetDefault() (*models.Provider, error) {
	var provider models.Provider
	if err := s.db.Where("is_default = ?", true).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}
