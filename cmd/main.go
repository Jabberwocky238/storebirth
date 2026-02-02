package main

import (
	"flag"
	"log"
	"os"

	storebirth "jabberwocky238/storebirth/core"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse flags
	listen := flag.String("l", "localhost:9900", "Listen address")
	dbDSN := flag.String("d", "", "Database DSN")
	kubeconfig := flag.String("k", "", "Kubeconfig path (empty for in-cluster)")
	flag.Parse()

	// Check DOMAIN environment variable (required for IngressRoute creation)
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("DOMAIN environment variable is required")
	}
	log.Printf("Using domain: %s", domain)

	// Initialize database
	if err := storebirth.InitDB(*dbDSN); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer storebirth.DB.Close()

	// Initialize K8s client
	if err := storebirth.InitK8s(*kubeconfig); err != nil {
		log.Printf("Warning: K8s client init failed: %v", err)
		log.Println("Running without K8s integration")
	} else {
		log.Println("K8s client initialized")
	}

	log.Println("Control plane starting...")

	// Setup Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", storebirth.Health)

	// Serve index.html at root
	r.StaticFile("/", "./index.html")

	// Public routes
	r.POST("/auth/register", storebirth.Register)
	r.POST("/auth/login", storebirth.Login)
	r.POST("/auth/send-code", storebirth.SendCode)
	r.POST("/auth/reset-password", storebirth.ResetPassword)

	// Protected routes
	api := r.Group("/api")
	api.Use(storebirth.AuthMiddleware())
	{
		api.POST("/rdb", storebirth.CreateRDB)
		api.GET("/rdb", storebirth.ListRDBs)
		api.DELETE("/rdb/:id", storebirth.DeleteRDB)
		api.POST("/kv", storebirth.CreateKV)
		api.GET("/kv", storebirth.ListKVs)
		api.DELETE("/kv/:id", storebirth.DeleteKV)
	}

	// Start server
	log.Printf("Server listening on %s", *listen)
	r.Run(*listen)
}
