package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"go-image-analyzer/model"
)

// FakeAnalyzer is a fake implementation of Analyzer for testing.
type FakeAnalyzer struct {
	objects   []model.DetectedObject
	err       error
	callCount int
}

func (f *FakeAnalyzer) Analyze(imageData []byte) ([]model.DetectedObject, error) {
	f.callCount++
	return f.objects, f.err
}

// FakeAnalysisSaver is a fake implementation of AnalysisSaver for testing.
type FakeAnalysisSaver struct {
	result    *model.AnalysisResult
	err       error
	callCount int
	filename  string
	objects   []model.DetectedObject
}

func (f *FakeAnalysisSaver) Save(filename string, objects []model.DetectedObject) (*model.AnalysisResult, error) {
	f.callCount++
	f.filename = filename
	f.objects = objects
	return f.result, f.err
}

func TestAnalyzeHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedObjects := []model.DetectedObject{
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
		name             string
		imageData        []byte
		analyzerObjects  []model.DetectedObject
		analyzerErr      error
		saveErr          error
		expectedStatus   int
		expectedAnalyze  int
		expectedSave     int
		expectedFilename string
	}{
		{
			name:            "successful image analysis",
			imageData:       []byte("test-image"),
			analyzerObjects: expectedObjects,
			expectedStatus:  http.StatusOK,
			expectedAnalyze: 1,
			expectedSave:    1,
		},
		{
			name:            "empty request body",
			imageData:       []byte{},
			expectedStatus:  http.StatusBadRequest,
			expectedAnalyze: 0,
			expectedSave:    0,
		},
		{
			name:            "Hugging Face API error",
			imageData:       []byte("test-image"),
			analyzerErr:     errors.New("Hugging Face API error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedAnalyze: 1,
			expectedSave:    0,
		},
		{
			name:            "no objects detected",
			imageData:       []byte("test-image"),
			analyzerObjects: []model.DetectedObject{},
			expectedStatus:  http.StatusUnprocessableEntity,
			expectedAnalyze: 1,
			expectedSave:    0,
		},
		{
			name:            "database save error",
			imageData:       []byte("test-image"),
			analyzerObjects: expectedObjects,
			saveErr:         errors.New("database connection error"),
			expectedStatus:  http.StatusOK,
			expectedAnalyze: 1,
			expectedSave:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &FakeAnalyzer{
				objects: tt.analyzerObjects,
				err:     tt.analyzerErr,
			}

			saver := &FakeAnalysisSaver{
				result: &model.AnalysisResult{
					ID:       1,
					Filename: "test.jpg",
				},
				err: tt.saveErr,
			}

			handler := NewAnalyzeHandler(analyzer, saver)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/analyze",
				bytes.NewReader(tt.imageData),
			)
			req.Header.Set("Content-Type", "image/jpeg")

			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = req

			handler.Handle(context)

			if recorder.Code != tt.expectedStatus {
				t.Errorf(
					"expected status %d, got %d",
					tt.expectedStatus,
					recorder.Code,
				)
			}

			if analyzer.callCount != tt.expectedAnalyze {
				t.Errorf(
					"expected Analyze to be called %d times, got %d",
					tt.expectedAnalyze,
					analyzer.callCount,
				)
			}

			if saver.callCount != tt.expectedSave {
				t.Errorf(
					"expected Save to be called %d times, got %d",
					tt.expectedSave,
					saver.callCount,
				)
			}

			if tt.expectedStatus == http.StatusOK {
				var response model.AnalyzeResponse

				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if !reflect.DeepEqual(response.Objects, tt.analyzerObjects) {
					t.Errorf(
						"expected objects %+v, got %+v",
						tt.analyzerObjects,
						response.Objects,
					)
				}
			}

			if tt.expectedStatus != http.StatusOK {
				var response model.ErrorResponse

				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if response.Error == "" {
					t.Error("expected error message, got empty string")
				}
			}

			if saver.callCount > 0 && saver.filename != "" {
				if saver.filename != tt.expectedFilename && tt.expectedFilename != "" {
					t.Errorf(
						"expected filename %q, got %q",
						tt.expectedFilename,
						saver.filename,
					)
				}
			}
		})
	}
}
