package sessions

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("session not found")

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List() ([]Session, error) {
	var sessions []Session
	err := s.db.Preload("Media").Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func (s *Store) Get(id uuid.UUID) (Session, error) {
	var session Session
	if err := s.db.Preload("Media").First(&session, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Session{}, ErrNotFound
		}
		return Session{}, err
	}
	return session, nil
}

func (s *Store) Create(session *Session) error {
	return s.db.Create(session).Error
}

func (s *Store) Update(id uuid.UUID, fields map[string]any) error {
	result := s.db.Model(&Session{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Delete(id uuid.UUID) error {
	result := s.db.Delete(&Session{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AddMedia(media *Media) error {
	return s.db.Create(media).Error
}

func (s *Store) GetMedia(sessionID, mediaID uuid.UUID) (Media, error) {
	var media Media
	if err := s.db.First(&media, "id = ? AND session_id = ?", mediaID, sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Media{}, ErrNotFound
		}
		return Media{}, err
	}
	return media, nil
}

func (s *Store) DeleteMedia(sessionID, mediaID uuid.UUID) error {
	result := s.db.Delete(&Media{}, "id = ? AND session_id = ?", mediaID, sessionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpdateMediaTranscription(sessionID, mediaID uuid.UUID, transcription string) error {
	result := s.db.Model(&Media{}).
		Where("id = ? AND session_id = ?", mediaID, sessionID).
		Updates(map[string]any{"transcription": transcription})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func newSession(title, description, content string, status Status) Session {
	now := time.Now().UTC()
	return Session{ID: uuid.New(), Title: title, Description: description, Status: status, Content: content, CreatedAt: now, UpdatedAt: now}
}
