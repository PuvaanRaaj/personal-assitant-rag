package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/PuvaanRaaj/personal-rag-agent/internal/config"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/database"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/handler"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/logger"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/middleware"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/repository"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/service"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/storage"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/watcher"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		// This is expected when running in Docker
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize structured logger (OTel-compatible)
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	logger.InitLogger(env)

	logger.Info("Starting RAG Personal Assistant",
		"environment", env,
		"port", cfg.Port,
	)

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	logger.Info("Database connected successfully")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}
	logger.Info("Database migrations completed")

	// Initialize storage driver (local, localstack, or s3)
	storageDriver, err := storage.NewStorageDriver(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize storage driver",
			"driver", cfg.StorageDriver,
			"error", err,
		)
	}
	logger.Info("Storage driver initialized",
		"driver", cfg.StorageDriver,
		"local_path", cfg.LocalStoragePath,
	)

	// Initialize Qdrant client
	qdrantClient, err := storage.NewQdrantClient(cfg.QdrantURL)
	if err != nil {
		logger.Fatal("Failed to initialize Qdrant client",
			"url", cfg.QdrantURL,
			"error", err,
		)
	}
	logger.Info("Qdrant client initialized", "url", cfg.QdrantURL)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	vectorRepo := repository.NewVectorRepository(qdrantClient)

	// Initialize services
	embeddingService := service.NewEmbeddingService(cfg.OpenAIKey)
	documentService := service.NewDocumentService(documentRepo, vectorRepo, storageDriver, embeddingService)
	ragService := service.NewRAGService(vectorRepo, embeddingService, cfg.OpenAIKey, documentRepo)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Initialize Knowledge Base Watcher
	kbWatcher, err := watcher.NewWatcher(cfg.KnowledgeBasePath, cfg.DefaultUserID, documentService)
	if err != nil {
		logger.Fatal("Failed to initialize knowledge base watcher", "error", err)
	}
	
	// Start watcher in background
	watcherCtx, watcherCancel := context.WithCancel(context.Background())
	defer watcherCancel()
	if err := kbWatcher.Start(watcherCtx); err != nil {
		logger.Fatal("Failed to start knowledge base watcher", "error", err)
	}
	defer kbWatcher.Close()

	// Perform initial sync
	go func() {
		time.Sleep(2 * time.Second) // Wait for server to be ready
		if err := kbWatcher.Sync(context.Background()); err != nil {
			logger.Error("Initial sync failed", "error", err)
		}
	}()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "RAG Personal Assistant",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		// Allow larger uploads (e.g., PDFs)
		BodyLimit: 50 * 1024 * 1024, // 50 MB
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
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
	documents.Post("/sync", func(c *fiber.Ctx) error {
		// Manual sync trigger
		go func() {
			if err := kbWatcher.Sync(context.Background()); err != nil {
				logger.Error("Manual sync failed", "error", err)
			}
		}()
		return c.JSON(fiber.Map{
			"message": "sync triggered successfully",
		})
	})
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

	// Check for TLS configuration
	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")

	// Graceful shutdown
	go func() {
		if tlsCertFile != "" && tlsKeyFile != "" {
			// Start HTTPS server
			httpsPort := os.Getenv("HTTPS_PORT")
			if httpsPort == "" {
				httpsPort = "8443"
			}
			logger.Info("Starting HTTPS server",
				"port", httpsPort,
				"cert", tlsCertFile,
			)
			if err := app.ListenTLS(":"+httpsPort, tlsCertFile, tlsKeyFile); err != nil {
				logger.Fatal("HTTPS server failed to start", "error", err)
			}
		} else {
			// Start HTTP server
			logger.Info("Starting HTTP server", "port", port)
			if err := app.Listen(":" + port); err != nil {
				logger.Fatal("Server failed to start", "error", err)
			}
		}
	}()

	if tlsCertFile != "" && tlsKeyFile != "" {
		logger.Info("Server started with HTTPS",
			"https_port", os.Getenv("HTTPS_PORT"),
		)
	} else {
		logger.Info("Server started with HTTP", "port", port)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited gracefully")
}
