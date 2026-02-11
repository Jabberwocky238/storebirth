package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"jabberwocky238/console/dblayer"
	"jabberwocky238/console/handlers/jobs"
	"jabberwocky238/console/k8s"

	"github.com/gin-gonic/gin"
)

// Health handles health check endpoint
func HealthInner(c *gin.Context) {
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}

	// Check database connection
	if dblayer.DB != nil {
		if err := dblayer.DB.Ping(); err != nil {
			status["database"] = "unhealthy"
			status["database_error"] = err.Error()
			c.JSON(503, status)
			return
		}
		status["database"] = "healthy"
	} else {
		status["database"] = "not_initialized"
	}

	// Check K8s client
	if k8s.K8sClient != nil {
		status["kubernetes"] = "healthy"
	} else {
		status["kubernetes"] = "not_initialized"
	}

	c.JSON(200, status)
}

func HealthOuter(c *gin.Context) {
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}

	// Check database connection
	if dblayer.DB != nil {
		if err := dblayer.DB.Ping(); err != nil {
			status["database"] = "unhealthy"
			status["database_error"] = err.Error()
			c.JSON(503, status)
			return
		}
		status["database"] = "healthy"
	} else {
		status["database"] = "not_initialized"
	}

	c.JSON(200, status)
}

type AcceptTaskRequest struct {
	TaskType  k8s.JobType `json:"task_type" binding:"required"`
	Timestamp int64       `json:"timestamp" binding:"required"`
	Data      []byte      `json:"data" binding:"required"`
}

type JobsHandler struct {
	processor *k8s.Processor
	cron      *k8s.CronScheduler
}

func NewTaskHandler(proc *k8s.Processor, cron *k8s.CronScheduler) *JobsHandler {
	return &JobsHandler{
		processor: proc,
		cron:      cron,
	}
}

func (h *JobsHandler) AcceptTask(c *gin.Context) {
	var req AcceptTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate timestamp
	if req.Timestamp <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
		return
	}

	// 使用 JobFactory 反序列化 Job
	job, err := jobs.CreateJob(req.TaskType, req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to deserialize job: %v", err)})
		return
	}

	// 提交到 processor
	h.processor.Submit(job)

	c.JSON(http.StatusOK, gin.H{
		"message":     "task accepted",
		"task_type":   req.TaskType,
		"timestamp":   req.Timestamp,
		"received_at": time.Now().Unix(),
	})
}

// SendTask sends a task to the inner control plane endpoint
// Uses Kubernetes internal service: control-plane-inner.console.svc.cluster.local
func SendTask(job k8s.Job) error {
	endpoint := fmt.Sprintf("%s/api/acceptTask", k8s.ControlPlaneInnerEndpoint)

	var jobData []byte
	var err error
	jobData, err = json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	req := AcceptTaskRequest{
		TaskType:  k8s.JobType(job.Type()),
		Timestamp: time.Now().Unix(),
		Data:      jobData,
	}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("task rejected with status: %d", resp.StatusCode)
	}

	return nil
}
