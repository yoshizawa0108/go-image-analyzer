package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go-image-analyzer/model"
)

const (
	hfAPIURL = "https://router.huggingface.co/hf-inference/models/facebook/detr-resnet-50"
	timeout  = 60 * time.Second // モデルの初回ロードに時間がかかる場合がある
)

// HuggingFaceService handles communication with the Hugging Face Inference API
type HuggingFaceService struct {
	apiKey     string
	httpClient *http.Client
}

// NewHuggingFaceService creates a new HuggingFaceService instance
func NewHuggingFaceService(apiKey string) *HuggingFaceService {
	return &HuggingFaceService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Analyze sends image bytes to the Hugging Face DETR API and returns detected objects
func (s *HuggingFaceService) Analyze(imageData []byte) ([]model.DetectedObject, error) {
	req, err := http.NewRequest(http.MethodPost, hfAPIURL, bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "image/jpeg")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Hugging Face API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close Hugging Face API response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hugging face API error (status %d): %s", resp.StatusCode, string(body))
	}

	var hfResults []model.HuggingFaceResult
	if err := json.Unmarshal(body, &hfResults); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	objects := make([]model.DetectedObject, 0, len(hfResults))
	for _, r := range hfResults {
		objects = append(objects, model.DetectedObject{
			Label: r.Label,
			Score: r.Score,
			Box:   r.Box,
		})
	}

	return objects, nil
}
