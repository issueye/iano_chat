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

type ProviderDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	APIKey      string  `json:"-"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	IsDefault   bool    `json:"is_default"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (s *ProviderService) ToDTO(provider *models.Provider) *ProviderDTO {
	return &ProviderDTO{
		ID:          provider.ID,
		Name:        provider.Name,
		BaseURL:     provider.BaseUrl,
		Model:       provider.Model,
		APIKey:      provider.ApiKey,
		Temperature: provider.Temperature,
		MaxTokens:   provider.MaxTokens,
		IsDefault:   provider.IsDefault,
		CreatedAt:   provider.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   provider.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *ProviderService) ToDTOList(providers []models.Provider) []ProviderDTO {
	dtos := make([]ProviderDTO, len(providers))
	for i, p := range providers {
		dto := s.ToDTO(&p)
		dtos[i] = *dto
	}
	return dtos
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

func (s *ProviderService) GetAllDTO() ([]ProviderDTO, error) {
	providers, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	return s.ToDTOList(providers), nil
}

func (s *ProviderService) GetByIDDTO(id string) (*ProviderDTO, error) {
	provider, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.ToDTO(provider), nil
}

func (s *ProviderService) GetDefault() (*ProviderDTO, error) {
	var provider models.Provider
	if err := s.db.Where("is_default = ?", true).First(&provider).Error; err != nil {
		return nil, err
	}
	return s.ToDTO(&provider), nil
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
