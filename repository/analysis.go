package repository

import (
	"fmt"

	"gorm.io/gorm"

	"go-image-analyzer/model"
)

// AnalysisRepository handles database operations for analysis results
type AnalysisRepository struct {
	db *gorm.DB
}

// NewAnalysisRepository creates a new AnalysisRepository
func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

// Save stores the analysis result and returns the inserted record
func (r *AnalysisRepository) Save(filename string, objects []model.DetectedObject) (*model.AnalysisResult, error) {
	result := &model.AnalysisResult{
		Filename: filename,
		Objects:  model.DetectedObjects(objects),
	}

	if err := r.db.Create(result).Error; err != nil {
		return nil, fmt.Errorf("failed to save analysis result: %w", err)
	}

	return result, nil
}
