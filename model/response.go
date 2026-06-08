package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// BoundingBox represents the detected object's position in the image
type BoundingBox struct {
	Xmin float64 `json:"xmin"`
	Ymin float64 `json:"ymin"`
	Xmax float64 `json:"xmax"`
	Ymax float64 `json:"ymax"`
}

// DetectedObject represents a single object detected in the image
type DetectedObject struct {
	Label string      `json:"label"`
	Score float64     `json:"score"`
	Box   BoundingBox `json:"box"`
}

// DetectedObjects is a slice of DetectedObject that supports JSONB storage
type DetectedObjects []DetectedObject

func (d DetectedObjects) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DetectedObjects) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan DetectedObjects: expected []byte")
	}
	return json.Unmarshal(bytes, d)
}

// AnalysisResult is the GORM model for storing analysis results
type AnalysisResult struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Filename  string          `gorm:"type:text"                json:"filename"`
	Objects   DetectedObjects `gorm:"type:jsonb;not null"      json:"objects"`
	CreatedAt time.Time       `                                json:"created_at"`
}

// AnalyzeResponse is the API response returned to the client
type AnalyzeResponse struct {
	Objects []DetectedObject `json:"objects"`
}

// HuggingFaceResult represents a single result from Hugging Face DETR API
type HuggingFaceResult struct {
	Score float64     `json:"score"`
	Label string      `json:"label"`
	Box   BoundingBox `json:"box"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}