package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TranscodingJobStatus represents the possible status values for transcoding jobs
type TranscodingJobStatus string

const (
	StatusPending    TranscodingJobStatus = "pending"
	StatusProcessing TranscodingJobStatus = "processing"
	StatusCompleted  TranscodingJobStatus = "completed"
	StatusFailed     TranscodingJobStatus = "failed"
	StatusCancelled  TranscodingJobStatus = "cancelled"
)

// TranscodingJob represents a video transcoding job
type TranscodingJob struct {
	ID              uuid.UUID            `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	JobID           string               `gorm:"type:varchar(255);not null;uniqueIndex:transcoding_jobs_job_id_index" json:"job_id"`
	SourcePath      string               `gorm:"type:text;not null" json:"source_path"`
	TargetCodec     string               `gorm:"type:varchar(50);not null;index:transcoding_jobs_target_codec_index" json:"target_codec"`
	TargetContainer string               `gorm:"type:varchar(50);not null" json:"target_container"`
	SourceCodec     *string              `gorm:"type:varchar(50)" json:"source_codec"`
	SourceContainer *string              `gorm:"type:varchar(50)" json:"source_container"`
	OutputURL       *string              `gorm:"type:text" json:"output_url"`
	DurationSeconds *int                 `gorm:"check:duration_seconds IS NULL OR duration_seconds > 0" json:"duration_seconds"`
	Status          TranscodingJobStatus `gorm:"type:varchar(50);default:'pending';check:status IN ('pending', 'processing', 'completed', 'failed', 'cancelled');index:transcoding_jobs_status_index,transcoding_jobs_status_inserted_at_index" json:"status"`
	ErrorMessage    *string              `gorm:"type:text" json:"error_message"`
	GPUUsed         *string              `gorm:"type:varchar(100);index:transcoding_jobs_gpu_used_index" json:"gpu_used"`
	QualityPreset   *string              `gorm:"type:varchar(50)" json:"quality_preset"`
	Bitrate         *int                 `gorm:"check:bitrate IS NULL OR bitrate > 0" json:"bitrate"`
	FileSizeBytes   *int64               `gorm:"check:file_size_bytes IS NULL OR file_size_bytes > 0" json:"file_size_bytes"`
	SourceDuration  *float64             `gorm:"type:numeric(10,3)" json:"source_duration"`
	SourceBitrate   *int                 `gorm:"check:source_bitrate IS NULL OR source_bitrate > 0" json:"source_bitrate"`
	SourceWidth     *int                 `gorm:"check:source_width IS NULL OR source_width > 0" json:"source_width"`
	SourceHeight    *int                 `gorm:"check:source_height IS NULL OR source_height > 0" json:"source_height"`
	InsertedAt      time.Time            `gorm:"type:timestamp(0);not null;index:transcoding_jobs_inserted_at_index,transcoding_jobs_status_inserted_at_index" json:"inserted_at"`
	UpdatedAt       time.Time            `gorm:"type:timestamp(0);not null" json:"updated_at"`
	CreatedBy      *uint               `gorm:"type:integer;index" json:"created_by,omitempty"`
}

// TableName returns the table name for the TranscodingJob model
func (TranscodingJob) TableName() string {
	return "transcoding_jobs"
}

// BeforeCreate sets the default values before creating a new transcoding job
func (tj *TranscodingJob) BeforeCreate(tx *gorm.DB) error {
	if tj.ID == uuid.Nil {
		tj.ID = uuid.New()
	}
	if tj.Status == "" {
		tj.Status = StatusPending
	}
	now := time.Now()
	tj.InsertedAt = now
	tj.UpdatedAt = now
	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp before updating
func (tj *TranscodingJob) BeforeUpdate(tx *gorm.DB) error {
	tj.UpdatedAt = time.Now()
	return nil
}

// IsValidStatus checks if the provided status is valid
func (s TranscodingJobStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status
func (s TranscodingJobStatus) String() string {
	return string(s)
}
