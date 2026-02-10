package services

import (
	"iano_server/models"

	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) Create(session *models.Session) error {
	return s.db.Create(session).Error
}

func (s *SessionService) GetByID(id int64) (*models.Session, error) {
	var session models.Session
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) GetAll() ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *SessionService) GetByKeyID(keyID string) ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Where("key_id = ?", keyID).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *SessionService) GetByStatus(status models.SessionStatus) ([]models.Session, error) {
	var sessions []models.Session
	if err := s.db.Where("status = ?", status).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *SessionService) Update(id int64, updates map[string]interface{}) (*models.Session, error) {
	var session models.Session
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&session).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionService) Delete(id int64) error {
	result := s.db.Delete(&models.Session{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *SessionService) DeleteByKeyID(keyID string) error {
	return s.db.Where("key_id = ?", keyID).Delete(&models.Session{}).Error
}

func (s *SessionService) Count() (int64, error) {
	var count int64
	if err := s.db.Model(&models.Session{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SessionService) CountByKeyID(keyID string) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Session{}).Where("key_id = ?", keyID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
