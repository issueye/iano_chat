### Service Template

```go
package services

import (
	"iano_chat/models"

	"gorm.io/gorm"
)

type {ModelName}Service struct {
	db *gorm.DB
}

func New{ModelName}Service(db *gorm.DB) *{ModelName}Service {
	return &{ModelName}Service{db: db}
}

func (s *{ModelName}Service) Create({modelVar} *models.{ModelName}) error {
	return s.db.Create({modelVar}).Error
}

func (s *{ModelName}Service) GetByID(id int64) (*models.{ModelName}, error) {
	var {modelVar} models.{ModelName}
	if err := s.db.First(&{modelVar}, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &{modelVar}, nil
}

func (s *{ModelName}Service) GetAll() ([]models.{ModelName}, error) {
	var {modelVar}s []models.{ModelName}
	if err := s.db.Find(&{modelVar}s).Error; err != nil {
		return nil, err
	}
	return {modelVar}s, nil
}

func (s *{ModelName}Service) GetByUserID(userID int64) ([]models.{ModelName}, error) {
	var {modelVar}s []models.{ModelName}
	if err := s.db.Where("user_id = ?", userID).Find(&{modelVar}s).Error; err != nil {
		return nil, err
	}
	return {modelVar}s, nil
}

func (s *{ModelName}Service) Update(id int64, updates map[string]interface{}) (*models.{ModelName}, error) {
	var {modelVar} models.{ModelName}
	if err := s.db.First(&{modelVar}, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&{modelVar}).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &{modelVar}, nil
}

func (s *{ModelName}Service) Delete(id int64) error {
	result := s.db.Delete(&models.{ModelName}{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *{ModelName}Service) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.{ModelName}{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
```