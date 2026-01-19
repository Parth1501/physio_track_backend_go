package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"phsio_track_backend/internal/config"
	"phsio_track_backend/internal/http/handlers"
	"phsio_track_backend/internal/http/middleware"
	"phsio_track_backend/internal/repo"
)

func main() {
	// Load configuration (env vars + .env if present)
	cfg := config.Load()

	// Logging: write to stdout; optionally also to a log file if LOG_FILE is set
	var logWriters []io.Writer
	logWriters = append(logWriters, os.Stdout)
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open log file %s: %v", cfg.LogFile, err)
		}
		defer f.Close()
		logWriters = append(logWriters, f)
	}
	mw := io.MultiWriter(logWriters...)
	log.SetOutput(mw)
	gin.DefaultWriter = mw
	gin.DefaultErrorWriter = mw

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Init DB (Oracle via wallet/TNS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbpool, err := repo.NewDB(ctx, repo.DBConfig{
		User:          cfg.DBUser,
		Password:      cfg.DBPassword,
		ConnectString: cfg.DBConnectString,
		TNSAdmin:      cfg.TNSAdmin,
	})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbpool.Close()

	if err := repo.BootstrapSchema(ctx, dbpool); err != nil {
		log.Fatalf("failed to bootstrap schema: %v", err)
	}

	// Repos
	userRepo := repo.NewUserRepo(dbpool)
	patientRepo := repo.NewPatientRepo(dbpool)
	paymentRepo := repo.NewPaymentRepo(dbpool)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiry)
	patientHandler := handlers.NewPatientHandler(patientRepo)
	paymentHandler := handlers.NewPaymentHandler(paymentRepo)

	router := gin.Default()

	// Health
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/register", authHandler.RegisterUser)

	// Protected
	authz := middleware.JWTAuth(cfg.JWTSecret, cfg.JWTIssuer)
	api := router.Group("/")
	api.Use(authz)

	// Users
	api.POST("/users", authHandler.CreateUser)

	// Patients
	api.POST("/patients", patientHandler.Create)
	api.GET("/patients", patientHandler.List)
	api.GET("/patients/:id", patientHandler.GetByID)
	api.PATCH("/patients/:id", patientHandler.Update)

	// Payments
	api.POST("/payments", paymentHandler.Create)
	api.GET("/payments", paymentHandler.List)
	api.PATCH("/payments/:id", paymentHandler.Update)
	api.DELETE("/payments/:id", paymentHandler.Delete)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
