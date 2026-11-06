package service

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"go-image-analyzer/model"
)

// roundTripFunc allows tests to replace the HTTP transport.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// errorReader simulates a response body read failure.
type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("response body read error")
}

func newTestService(roundTrip roundTripFunc) *HuggingFaceService {
	s := NewHuggingFaceService("test-api-key")
	s.httpClient = &http.Client{
		Transport: roundTrip,
	}
	return s
}

func newResponse(statusCode int, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       body,
		Header:     make(http.Header),
	}
}

func TestHuggingFaceService_Analyze(t *testing.T) {
	imageData := []byte("test-image-data")

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
		name      string
		roundTrip roundTripFunc
		want      []model.DetectedObject
		wantErr   string
	}{
		{
			name: "successful analysis",
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Errorf("expected method POST, got %s", req.Method)
				}

				if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
					t.Errorf("unexpected Authorization header: %s", got)
				}

				if got := req.Header.Get("Content-Type"); got != "image/jpeg" {
					t.Errorf("unexpected Content-Type: %s", got)
				}

				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}

				if !reflect.DeepEqual(body, imageData) {
					t.Errorf("unexpected request body: %s", body)
				}

				responseBody := `[
					{
						"label": "cat",
						"score": 0.98,
						"box": {
							"xmin": 10,
							"ymin": 20,
							"xmax": 100,
							"ymax": 200
						}
					}
				]`

				return newResponse(
					http.StatusOK,
					io.NopCloser(strings.NewReader(responseBody)),
				), nil
			},
			want: expectedObjects,
		},
		{
			name: "API error status",
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newResponse(
					http.StatusUnauthorized,
					io.NopCloser(strings.NewReader("invalid API key")),
				), nil
			},
			wantErr: "hugging face API error (status 401): invalid API key",
		},
		{
			name: "invalid JSON response",
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newResponse(
					http.StatusOK,
					io.NopCloser(strings.NewReader("invalid json")),
				), nil
			},
			wantErr: "failed to parse response",
		},
		{
			name: "HTTP communication error",
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			},
			wantErr: "failed to call Hugging Face API",
		},
		{
			name: "response body read error",
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return newResponse(
					http.StatusOK,
					io.NopCloser(errorReader{}),
				), nil
			},
			wantErr: "failed to read response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestService(tt.roundTrip)

			got, err := s.Analyze(imageData)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}

				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf(
						"unexpected error:\nwant to contain: %q\ngot: %q",
						tt.wantErr,
						err.Error(),
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unexpected result:\nwant: %+v\ngot:  %+v", tt.want, got)
			}
		})
	}
}
