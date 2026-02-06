package handlers

import (
	"log"
	"strconv"

	"jabberwocky238/console/dblayer"
	"jabberwocky238/console/k8s"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkerHandler struct {
	queue chan int
}

func NewWorkerHandler() *WorkerHandler {
	return &WorkerHandler{
		queue: make(chan int, 100),
	}
}

func (h *WorkerHandler) Start() {
	go func() {
		for versionID := range h.queue {
			h.deploy(versionID)
		}
	}()
	log.Println("[worker-handler] started")
}

func (h *WorkerHandler) deploy(versionID int) {
	v, err := dblayer.GetDeployVersion(versionID)
	if err != nil {
		log.Printf("[worker-handler] get version %d failed: %v", versionID, err)
		return
	}
	w, err := dblayer.GetWorkerByID(v.WorkerID)
	if err != nil {
		log.Printf("[worker-handler] get worker %s failed: %v", v.WorkerID, err)
		dblayer.UpdateDeployVersionStatus(versionID, "error", err.Error())
		return
	}
	worker := &k8s.Worker{
		WorkerID: w.WorkerID,
		OwnerID:  w.UserUID,
		Image:    v.Image,
		Port:     v.Port,
	}
	if err := worker.Deploy(); err != nil {
		log.Printf("[worker-handler] deploy version %d failed: %v", versionID, err)
		dblayer.UpdateDeployVersionStatus(versionID, "error", err.Error())
		return
	}
	log.Printf("[worker-handler] deploy version %d success", versionID)
	dblayer.UpdateDeployVersionStatus(versionID, "success", "")
	dblayer.SetWorkerActiveVersion(v.WorkerID, versionID)
}

// CreateWorker 创建 worker 记录，返回 worker_id，status=unloaded
func (h *WorkerHandler) CreateWorker(c *gin.Context) {
	userUID := c.GetString("user_id")

	var req struct {
		WorkerName string `json:"worker_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	workerID := uuid.New().String()[:8]

	if err := dblayer.CreateWorker(userUID, workerID, req.WorkerName); err != nil {
		c.JSON(500, gin.H{"error": "failed to create worker"})
		return
	}

	c.JSON(200, gin.H{
		"worker_id":   workerID,
		"worker_name": req.WorkerName,
	})
}

// DeleteWorker 删除 worker（库 + K8s 资源）
func (h *WorkerHandler) DeleteWorker(c *gin.Context) {
	userUID := c.GetString("user_id")
	workerID := c.Param("id")

	// 验证 worker 属于该用户
	w, err := dblayer.GetWorkerByID(workerID)
	if err != nil || w.UserUID != userUID {
		c.JSON(404, gin.H{"error": "worker not found"})
		return
	}

	worker := &k8s.Worker{WorkerID: workerID, OwnerID: userUID}
	if err := worker.Delete(); err != nil {
		log.Printf("failed to delete k8s resources for worker %s: %v", workerID, err)
		c.JSON(500, gin.H{"error": "failed to delete worker resources"})
		return
	}

	// 删除数据库记录
	if err := dblayer.DeleteWorker(workerID); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete worker"})
		return
	}

	c.JSON(200, gin.H{"message": "worker deleted"})
}

// ListWorkers 列出用户所有 worker（从数据库读）
func (h *WorkerHandler) ListWorkers(c *gin.Context) {
	userUID := c.GetString("user_id")

	workers, err := dblayer.ListWorkersByUser(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list workers"})
		return
	}

	c.JSON(200, workers)
}

// GetWorker 获取单个 worker 详情，附带最近10条 version
func (h *WorkerHandler) GetWorker(c *gin.Context) {
	userUID := c.GetString("user_id")
	workerID := c.Param("id")

	w, err := dblayer.GetWorkerByID(workerID)
	if err != nil || w.UserUID != userUID {
		c.JSON(404, gin.H{"error": "worker not found"})
		return
	}

	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	versions, err := dblayer.ListDeployVersions(workerID, 10, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list versions"})
		return
	}

	c.JSON(200, gin.H{
		"worker":   w,
		"versions": versions,
	})
}

// DeployWorker 触发 worker 部署，立刻返回 200，异步执行
func (h *WorkerHandler) DeployWorker(c *gin.Context) {
	var req struct {
		UserUID  string `json:"user_uid" binding:"required"`
		WorkerID string `json:"worker_id" binding:"required"`
		Image    string `json:"image" binding:"required"`
		Port     int    `json:"port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 验证 worker 属于该用户
	w, err := dblayer.GetWorkerByID(req.WorkerID)
	if err != nil || w.UserUID != req.UserUID {
		c.JSON(404, gin.H{"error": "worker not found"})
		return
	}

	// 创建部署版本
	versionID, err := dblayer.CreateDeployVersion(req.WorkerID, req.Image, req.Port)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create deploy version"})
		return
	}

	// 入队，由后台 goroutine 异步执行部署
	h.queue <- versionID

	// 立刻返回 200
	c.JSON(200, gin.H{
		"worker_id":  req.WorkerID,
		"version_id": versionID,
		"status":     "loading",
	})
}
