package services

import (
	"iano_server/models"

	"gorm.io/gorm"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

func (s *MessageService) Create(message *models.Message) error {
	return s.db.Create(message).Error
}

func (s *MessageService) GetByID(id string) (*models.Message, error) {
	var message models.Message
	if err := s.db.First(&message, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *MessageService) GetAll() ([]models.Message, error) {
	var messages []models.Message
	if err := s.db.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageService) GetBySessionID(sessionID string) ([]models.Message, error) {
	var messages []models.Message
	if err := s.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageService) GetByType(messageType models.MessageType) ([]models.Message, error) {
	var messages []models.Message
	if err := s.db.Where("type = ?", messageType).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *MessageService) Update(id string, updates map[string]interface{}) (*models.Message, error) {
	var message models.Message
	if err := s.db.First(&message, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&message).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *MessageService) Delete(id string) error {
	result := s.db.Delete(&models.Message{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *MessageService) DeleteBySessionID(sessionID int64) error {
	result := s.db.Where("session_id = ?", sessionID).Delete(&models.Message{})
	return result.Error
}

func (s *MessageService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Message{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MessageService) CountBySessionID(sessionID int64) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Message{}).Where("session_id = ?", sessionID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MessageService) AddFeedback(id string, rating models.FeedbackRating, comment string) (*models.Message, error) {
	var message models.Message
	if err := s.db.First(&message, "id = ?", id).Error; err != nil {
		return nil, err
	}
	message.AddFeedback(rating, comment)
	if err := s.db.Save(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
