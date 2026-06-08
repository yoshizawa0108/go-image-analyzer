package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"go-image-analyzer/db"
	"go-image-analyzer/handler"
	"go-image-analyzer/model"
	"go-image-analyzer/repository"
	"go-image-analyzer/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	apiKey := os.Getenv("HF_API_KEY")
	if apiKey == "" {
		log.Fatal("HF_API_KEY is not set.")
	}

	// DB接続
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database")

	// AutoMigrate（テーブルが未作成でも自動生成）
	if err := database.AutoMigrate(&model.AnalysisResult{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 依存関係の初期化
	hfService := service.NewHuggingFaceService(apiKey)
	analysisRepo := repository.NewAnalysisRepository(database)
	analyzeHandler := handler.NewAnalyzeHandler(hfService, analysisRepo)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/analyze", analyzeHandler.Handle)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}