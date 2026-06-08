package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-image-analyzer/model"
	"go-image-analyzer/repository"
	"go-image-analyzer/service"
)

const maxImageSize = 10 << 20 // 10MB

// AnalyzeHandler handles POST /api/analyze requests
type AnalyzeHandler struct {
	hfService    *service.HuggingFaceService
	analysisRepo *repository.AnalysisRepository
}

// NewAnalyzeHandler creates a new AnalyzeHandler
func NewAnalyzeHandler(hfService *service.HuggingFaceService, repo *repository.AnalysisRepository) *AnalyzeHandler {
	return &AnalyzeHandler{
		hfService:    hfService,
		analysisRepo: repo,
	}
}

// Handle processes the image analysis request
func (h *AnalyzeHandler) Handle(c *gin.Context) {
	imageData, filename, err := readImageData(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	objects, err := h.hfService.Analyze(imageData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	// 検出結果が空の場合はエラーを返してDBに登録しない
	if len(objects) == 0 {
		c.JSON(http.StatusUnprocessableEntity, model.ErrorResponse{
			Error: "no objects detected: the image could not be analyzed or contains no recognizable objects",
		})
		return
	}

	// DBに保存（失敗してもレスポンスは返す）
	result, err := h.analysisRepo.Save(filename, objects)
	if err != nil {
		log.Printf("failed to save analysis result: %v", err)
	} else {
		log.Printf("saved analysis result: id=%d filename=%s objects=%d", result.ID, filename, len(objects))
	}

	c.JSON(http.StatusOK, model.AnalyzeResponse{Objects: objects})
}

// readImageData reads image bytes and filename from multipart form or raw request body
func readImageData(c *gin.Context) ([]byte, string, error) {
	contentType := c.GetHeader("Content-Type")

	// multipart/form-data の場合
	if isMultipart(contentType) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize)
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			return nil, "", fmt.Errorf("failed to read 'image' field from form: %w", err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read image file: %w", err)
		}
		return data, header.Filename, nil
	}

	// raw bytes の場合
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read request body: %w", err)
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("request body is empty: send image as multipart form (field: 'image') or raw bytes")
	}
	return data, "", nil
}

func isMultipart(contentType string) bool {
	return len(contentType) >= 19 && contentType[:19] == "multipart/form-data"
}