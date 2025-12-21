package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/config"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/database"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/handler"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/middleware"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/service"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/storage"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize S3 client
	s3Client, err := storage.NewS3Client(cfg.AWSConfig)
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// Initialize Qdrant client
	qdrantClient, err := storage.NewQdrantClient(cfg.QdrantURL)
	if err != nil {
		log.Fatalf("Failed to initialize Qdrant client: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	vectorRepo := repository.NewVectorRepository(qdrantClient)

	// Initialize services
	embeddingService := service.NewEmbeddingService(cfg.OpenAIKey)
	documentService := service.NewDocumentService(documentRepo, vectorRepo, s3Client, embeddingService)
	ragService := service.NewRAGService(vectorRepo, embeddingService, cfg.OpenAIKey, documentRepo)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "RAG Personal Assistant",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	documentHandler := handler.NewDocumentHandler(documentService)
	queryHandler := handler.NewQueryHandler(ragService)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "rag-personal-assistant",
			"time":    time.Now().Unix(),
		})
	})

	// API routes
	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := api.Group("", middleware.AuthRequired(cfg.JWTSecret))

	// Document routes
	documents := protected.Group("/documents")
	documents.Post("/upload", documentHandler.Upload)
	documents.Get("", documentHandler.List)
	documents.Get("/:id", documentHandler.Get)
	documents.Delete("/:id", documentHandler.Delete)

	// Query routes
	query := protected.Group("/query")
	query.Post("", queryHandler.Query)
	query.Get("/stream", queryHandler.StreamQuery)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("🚀 Server started on port %s", port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
