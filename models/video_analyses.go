package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VideoAnalysisStatus represents the possible status values for video analysis jobs
type VideoAnalysisStatus string

const (
	AnalysisStatusPending    VideoAnalysisStatus = "pending"
	AnalysisStatusProcessing VideoAnalysisStatus = "processing"
	AnalysisStatusCompleted  VideoAnalysisStatus = "completed"
	AnalysisStatusFailed     VideoAnalysisStatus = "failed"
)

// VideoAnalysis represents a video analysis job
type VideoAnalysis struct {
	JobID        string              `gorm:"type:varchar(255);primaryKey;default:uuid_generate_v4()" json:"job_id"`
	VideoID      string              `gorm:"type:varchar(255);not null" json:"video_id"`
	S3URL        string              `gorm:"type:text;not null" json:"s3_url"`
	PeopleCount  *int                `gorm:"type:integer" json:"people_count"`
	Status       VideoAnalysisStatus `gorm:"type:varchar(50);not null;default:'pending'" json:"status"`
	CreatedAt    time.Time           `gorm:"type:timestamp;not null" json:"created_at"`
	CompletedAt  *time.Time          `gorm:"type:timestamp" json:"completed_at"`
	ErrorMessage *string             `gorm:"type:text" json:"error_message"`
	// Foreign key to link to the user who created the analysis can be null if not applicable
	CreatedBy    *uint               `gorm:"type:integer;index" json:"created_by,omitempty"`
}

// BeforeCreate hook to generate UUID for job_id if not provided
func (va *VideoAnalysis) BeforeCreate(tx *gorm.DB) error {
	if va.JobID == "" {
		va.JobID = uuid.New().String()
	}
	if va.CreatedAt.IsZero() {
		va.CreatedAt = time.Now()
	}
	return nil
}

// TableName returns the table name for the VideoAnalysis model
func (VideoAnalysis) TableName() string {
	return "video_analyses"
}
