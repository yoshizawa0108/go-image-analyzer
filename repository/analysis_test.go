package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-image-analyzer/model"
)

func newMockRepository(t *testing.T) (*AnalysisRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: db,
		}),
		&gorm.Config{
			DisableAutomaticPing: true,
		},
	)
	if err != nil {
		_ = db.Close()
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	t.Cleanup(func() {
		mock.ExpectClose()

		if err := db.Close(); err != nil {
			t.Errorf("failed to close mock DB: %v", err)
		}
	})

	return NewAnalysisRepository(gormDB), mock
}

func TestAnalysisRepository_Save(t *testing.T) {
	objects := []model.DetectedObject{
		{
			Label: "cat",
			Score: 0.98,
			Box: model.BoundingBox{
				Xmin: 10,
				Ymin: 20,
				Xmax: 100,
				Ymax: 200,
			},
		},
	}

	tests := []struct {
		name    string
		setup   func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful save",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(
					regexp.QuoteMeta(`INSERT INTO "analysis_results"`),
				).
					WithArgs(
						"test.jpg",
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).AddRow(1),
					)

				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				mock.ExpectQuery(
					regexp.QuoteMeta(`INSERT INTO "analysis_results"`),
				).
					WithArgs(
						"test.jpg",
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(
						errors.New("database connection error"),
					)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newMockRepository(t)

			tt.setup(mock)

			result, err := repo.Save("test.jpg", objects)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if result == nil {
					t.Fatal("expected result, got nil")
				}

				if result.ID != 1 {
					t.Errorf("expected ID 1, got %d", result.ID)
				}

				if result.Filename != "test.jpg" {
					t.Errorf(
						"expected filename test.jpg, got %s",
						result.Filename,
					)
				}

				if len(result.Objects) != len(objects) {
					t.Errorf(
						"expected %d objects, got %d",
						len(objects),
						len(result.Objects),
					)
				}

				if result.CreatedAt.After(time.Now()) {
					t.Errorf(
						"unexpected future CreatedAt: %v",
						result.CreatedAt,
					)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled SQL expectations: %v", err)
			}
		})
	}
}
