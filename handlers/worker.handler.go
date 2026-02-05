package handlers

import (
	"jabberwocky238/console/k8s"

	"github.com/gin-gonic/gin"
)

// RegisterWorker registers a new worker
func RegisterWorker(c *gin.Context) {
	userUID := c.GetString("user_id")

	var req struct {
		WorkerID string `json:"worker_id" binding:"required"`
		Image    string `json:"image" binding:"required"`
		Port     int    `json:"port" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Deploy to K8s
	worker := &k8s.Worker{
		WorkerID: req.WorkerID,
		OwnerID:  userUID,
		Image:    req.Image,
		Port:     req.Port,
	}

	if err := worker.Deploy(); err != nil {
		c.JSON(500, gin.H{
			"worker_id": req.WorkerID,
			"status":    "error",
			"error":     err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"worker_id": req.WorkerID,
		"status":    "active",
		"domain":    worker.Name() + ".worker." + k8s.Domain,
	})
}

// DeleteWorker deletes a worker
func DeleteWorker(c *gin.Context) {
	userUID := c.GetString("user_id")
	workerID := c.Param("id")

	worker := &k8s.Worker{
		WorkerID: workerID,
		OwnerID:  userUID,
	}

	if err := worker.Delete(); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete worker from k8s"})
		return
	}

	c.JSON(200, gin.H{"message": "worker deleted"})
}

// ListWorkers lists all workers for the user
func ListWorkers(c *gin.Context) {
	userUID := c.GetString("user_id")

	workers, err := k8s.ListWorkers("", userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list workers"})
		return
	}

	c.JSON(200, workers)
}

// GetWorker gets a single worker
func GetWorker(c *gin.Context) {
	userUID := c.GetString("user_id")
	workerID := c.Param("id")

	workers, err := k8s.ListWorkers(workerID, userUID)
	if err != nil || len(workers) == 0 {
		c.JSON(404, gin.H{"error": "worker not found"})
		return
	}

	c.JSON(200, workers[0])
}
