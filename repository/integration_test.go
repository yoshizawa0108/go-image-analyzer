package repository

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go-image-analyzer/db"
	"go-image-analyzer/model"

	"gorm.io/gorm"
)

func setupIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Prevent accidental connection to the development database.
	if os.Getenv("DB_NAME") != "image_analyzer_test" {
		t.Skip("integration test skipped: set DB_NAME=image_analyzer_test")
	}

	database, err := db.Connect()
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("failed to close database connection: %v", err)
		}
	})

	return database
}

func TestAnalysisRepository_SaveIntegration(t *testing.T) {
	database := setupIntegrationDB(t)
	repo := NewAnalysisRepository(database)

	filename := fmt.Sprintf("integration-test-%d.jpg", time.Now().UnixNano())

	objects := []model.DetectedObject{
		{
			Label: "cat",
			Score: 0.98,
			Box: model.BoundingBox{
				Xmin: 10,
				Ymin: 20,
				Xmax: 100,
				Ymax: 120,
			},
		},
		{
			Label: "dog",
			Score: 0.87,
			Box: model.BoundingBox{
				Xmin: 30,
				Ymin: 40,
				Xmax: 150,
				Ymax: 180,
			},
		},
	}

	saved, err := repo.Save(filename, objects)
	if err != nil {
		t.Fatalf("failed to save analysis result: %v", err)
	}

	if saved.ID == 0 {
		t.Fatal("expected saved record to have an ID")
	}

	// Delete only the record created by this test.
	t.Cleanup(func() {
		if err := database.Delete(&model.AnalysisResult{}, saved.ID).Error; err != nil {
			t.Errorf("failed to delete test record: %v", err)
		}
	})

	var actual model.AnalysisResult
	if err := database.First(&actual, saved.ID).Error; err != nil {
		t.Fatalf("failed to retrieve analysis result: %v", err)
	}

	if actual.ID != saved.ID {
		t.Errorf("ID mismatch: got %d, want %d", actual.ID, saved.ID)
	}

	if actual.Filename != filename {
		t.Errorf("filename mismatch: got %q, want %q", actual.Filename, filename)
	}

	if len(actual.Objects) != len(objects) {
		t.Fatalf("objects length mismatch: got %d, want %d", len(actual.Objects), len(objects))
	}

	for i, want := range objects {
		got := actual.Objects[i]

		if got.Label != want.Label {
			t.Errorf("objects[%d].Label mismatch: got %q, want %q", i, got.Label, want.Label)
		}

		if got.Score != want.Score {
			t.Errorf("objects[%d].Score mismatch: got %v, want %v", i, got.Score, want.Score)
		}

		if got.Box != want.Box {
			t.Errorf("objects[%d].Box mismatch: got %+v, want %+v", i, got.Box, want.Box)
		}
	}

	if actual.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}
