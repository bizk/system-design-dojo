package sessions

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

func validStatus(status Status) bool {
	return status == StatusDraft || status == StatusActive || status == StatusCompleted
}

type Session struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title       string
	Description string
	Status      Status
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Media       []Media `gorm:"foreignKey:SessionID"`
}

type Media struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	SessionID     uuid.UUID `gorm:"type:uuid;index"`
	Type          string
	ObjectKey     string
	Filename      string
	ContentType   string
	Size          int64
	Transcription *string
	CreatedAt     time.Time
}

func (Media) TableName() string {
	return "session_media"
}
