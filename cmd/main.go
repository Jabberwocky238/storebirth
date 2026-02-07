package main

import (
	"flag"
	"jabberwocky238/console/dblayer"
	"jabberwocky238/console/handlers"
	"jabberwocky238/console/k8s"
	"log"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse flags
	listen := flag.String("l", "localhost:9900", "Listen address")
	dbDSN := flag.String("d", "postgresql://myuser:your_password@localhost:5432/mydb?sslmode=disable", "Database DSN")
	kubeconfig := flag.String("k", "", "Kubeconfig path (empty for in-cluster)")
	flag.Parse()

	// Check DOMAIN environment variable (required for IngressRoute creation)
	// 检查是否为测试环境
	if os.Getenv("ENV") != "test" {
		checkEnv()
	}

	// Initialize database
	if err := dblayer.InitDB(*dbDSN); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dblayer.DB.Close()

	// Initialize K8s client
	if err := k8s.InitK8s(*kubeconfig); err != nil {
		log.Printf("Warning: K8s client init failed: %v", err)
		log.Println("Running without K8s integration")
	} else {
		log.Println("K8s client initialized")
	}

	// Start worker handler (deploy queue + periodic reconcile)
	wh := handlers.NewWorkerHandler()
	wh.Start()

	// Start periodic domain check
	k8s.StartPeriodicCheck()

	log.Println("Control plane starting...")

	// Setup Gin router
	r := gin.Default()
	if os.Getenv("ENV") == "test" {
		r.Use(crossOriginMiddleware())
	}
	// Health check endpoint
	r.GET("/health", handlers.Health)

	// Serve frontend static files from dist/
	r.Static("/assets", "./dist/assets")
	r.GET("/", func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	api := r.Group("/api")
	// Public routes
	api.POST("/auth/register", handlers.Register)
	api.POST("/auth/login", handlers.Login)
	api.POST("/auth/send-code", handlers.SendCode)
	api.POST("/auth/reset-password", handlers.ResetPassword)

	// Protected routes
	api.Use(handlers.AuthMiddleware())
	{
		api.GET("/rdb", handlers.ListRDBs)
		api.POST("/rdb", handlers.CreateRDB)
		api.DELETE("/rdb/:id", handlers.DeleteRDB)

		api.GET("/kv", handlers.ListKVs)
		api.POST("/kv", handlers.CreateKV)
		api.DELETE("/kv/:id", handlers.DeleteKV)

		api.GET("/worker", wh.ListWorkers)
		api.GET("/worker/:id", wh.GetWorker)
		api.POST("/worker", wh.CreateWorker)
		api.DELETE("/worker/:id", wh.DeleteWorker)

		api.GET("/domain", handlers.ListCustomDomains)
		api.GET("/domain/:id", handlers.GetCustomDomain)
		api.POST("/domain", handlers.AddCustomDomain)
		api.DELETE("/domain/:id", handlers.DeleteCustomDomain)
	}

	// Sensitive routes (signature required)
	sensitive := r.Group("/api")
	sensitive.Use(handlers.SignatureMiddleware())
	{
		sensitive.POST("/worker/deploy", wh.DeployWorker)
	}

	// Fallback: serve static files from dist/ or index.html for SPA
	r.NoRoute(func(c *gin.Context) {
		filePath := path.Join("./dist", c.Request.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}
		c.File("./dist/index.html")
	})

	// Start server
	log.Printf("Server listening on %s", *listen)
	r.Run(*listen)
}

func checkEnv() {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		panic("DOMAIN environment variable is required")
	}
	log.Printf("Using domain: %s", domain)
	k8s.Domain = domain
}

func crossOriginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
