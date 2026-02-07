package handlers

import (
	"encoding/json"

	"jabberwocky238/console/k8s"
	"jabberwocky238/console/k8s/controller"

	"github.com/gin-gonic/gin"
)

type CombinatorHandler struct {
	proc *k8s.Processor
}

func NewCombinatorHandler(proc *k8s.Processor) *CombinatorHandler {
	return &CombinatorHandler{proc: proc}
}

// CreateRDB creates a new RDB resource
func (h *CombinatorHandler) CreateRDB(c *gin.Context) {
	userUID := c.GetString("user_id")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	id, newConfig, err := combinator.AddRDB(req.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create RDB: " + err.Error()})
		return
	}

	if err := controller.UpdateCombinatorAppConfig(k8s.DynamicClient, userUID, newConfig); err != nil {
		c.JSON(500, gin.H{"error": "failed to update CR config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"id": id, "message": "RDB created successfully"})
}

// ListRDBs lists all RDB resources for user
func (h *CombinatorHandler) ListRDBs(c *gin.Context) {
	userUID := c.GetString("user_id")

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	cfg, err := combinator.ParseConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse config: " + err.Error()})
		return
	}

	userRDB := k8s.UserRDB{UserUID: userUID}
	dbSize, _ := userRDB.DatabaseSize()

	type rdbWithSize struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
		Size int64  `json:"size"`
	}

	items := make([]rdbWithSize, 0, len(cfg.RDBs))
	for _, rdb := range cfg.RDBs {
		size, _ := userRDB.SchemaSize(rdb.ID)
		items = append(items, rdbWithSize{
			ID:   rdb.ID,
			Name: rdb.Name,
			URL:  rdb.URL,
			Size: size,
		})
	}

	c.JSON(200, gin.H{"rdbs": items, "database_size": dbSize})
}

// CreateKV creates a new KV resource
func (h *CombinatorHandler) CreateKV(c *gin.Context) {
	userUID := c.GetString("user_id")
	var req struct {
		Type string `json:"kv_type" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	cfg, err := combinator.ParseConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse config: " + err.Error()})
		return
	}

	newKV := controller.KVItem{
		ID:   GenerateResourceUID(),
		Type: req.Type,
		URL:  req.URL,
	}
	cfg.KVs = append(cfg.KVs, newKV)

	newConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to marshal config: " + err.Error()})
		return
	}

	if err := controller.UpdateCombinatorAppConfig(k8s.DynamicClient, userUID, string(newConfig)); err != nil {
		c.JSON(200, gin.H{"id": newKV.ID, "error": "KV created but failed to update config, err: " + err.Error()})
	} else {
		c.JSON(200, gin.H{"id": newKV.ID, "message": "KV created successfully"})
	}
}

// ListKVs lists all KV resources for user
func (h *CombinatorHandler) ListKVs(c *gin.Context) {
	userUID := c.GetString("user_id")

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	cfg, err := combinator.ParseConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"kvs": cfg.KVs})
}

// DeleteRDB deletes an RDB resource
func (h *CombinatorHandler) DeleteRDB(c *gin.Context) {
	userUID := c.GetString("user_id")
	rdbID := c.Param("id")

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	newConfig, err := combinator.DeleteRDB(rdbID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete RDB: " + err.Error()})
		return
	}

	if err := controller.UpdateCombinatorAppConfig(k8s.DynamicClient, userUID, newConfig); err != nil {
		c.JSON(500, gin.H{"error": "failed to update CR config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "deleted"})
}

// DeleteKV deletes a KV resource
func (h *CombinatorHandler) DeleteKV(c *gin.Context) {
	userUID := c.GetString("user_id")
	kvUID := c.Param("id")

	combinator, err := controller.GetCombinatorAppConfig(k8s.DynamicClient, userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	cfg, err := combinator.ParseConfig()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to parse config: " + err.Error()})
		return
	}

	newKVs := []controller.KVItem{}
	isExist := false
	for _, kv := range cfg.KVs {
		if kv.ID == kvUID {
			isExist = true
			continue
		}
		newKVs = append(newKVs, kv)
	}
	if !isExist {
		c.JSON(404, gin.H{"error": "not found this KV: " + kvUID})
		return
	}
	cfg.KVs = newKVs

	newConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to marshal config: " + err.Error()})
		return
	}

	if err := controller.UpdateCombinatorAppConfig(k8s.DynamicClient, userUID, string(newConfig)); err != nil {
		c.JSON(500, gin.H{"error": "failed to update combinator config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "deleted"})
}
