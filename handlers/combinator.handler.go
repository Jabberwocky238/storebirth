package handlers

import (
	"jabberwocky238/console/k8s"

	"github.com/gin-gonic/gin"
)

// CreateRDB creates a new RDB resource
func CreateRDB(c *gin.Context) {
	userUID := c.GetString("user_id")
	var req struct {
		Type string `json:"rdb_type" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// rdbUID, err := dblayer.CreateRDB(userUID, GenerateResourceUID(), req.Name, req.Type, req.URL)
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": "failed to create RDB: " + err.Error()})
	// 	return
	// }

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	newRDB := k8s.RDBItem{
		ID:   GenerateResourceUID(),
		Type: req.Type,
		URL:  req.URL,
	}
	combinator.RDBs = append(combinator.RDBs, newRDB)

	// Reload combinator config
	if err := combinator.UpdateConfig(); err != nil {
		// dblayer.SetRDBStatus(rdbUID, "error", err.Error())
		c.JSON(200, gin.H{"id": newRDB.ID, "error": "RDB created but failed to update config, err: " + err.Error()})
	} else {
		// dblayer.SetRDBStatus(rdbUID, "active", "")
		c.JSON(200, gin.H{"id": newRDB.ID, "message": "RDB created successfully"})
	}
}

// ListRDBs lists all RDB resources for user
func ListRDBs(c *gin.Context) {
	userUID := c.GetString("user_id")

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	rdbs := combinator.RDBs
	// rdbs, err := dblayer.ListRDBsByUser(userUID)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "failed to query"})
	// 	return
	// }
	c.JSON(200, gin.H{"rdbs": rdbs})
}

// CreateKV creates a new KV resource
func CreateKV(c *gin.Context) {
	userUID := c.GetString("user_id")
	var req struct {
		Type string `json:"kv_type" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// kvUID, err := dblayer.CreateKV(userUID, GenerateResourceUID(), req.Name, req.Type, req.URL)
	// if err != nil {
	// 	c.JSON(400, gin.H{"error": "failed to create KV: " + err.Error()})
	// 	return
	// }

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	newKV := k8s.KVItem{
		ID:   GenerateResourceUID(),
		Type: req.Type,
		URL:  req.URL,
	}
	combinator.KVs = append(combinator.KVs, newKV)

	// Reload combinator config
	if err := combinator.UpdateConfig(); err != nil {
		// dblayer.SetKVStatus(kvUID, "error", err.Error())
		c.JSON(200, gin.H{"id": newKV.ID, "error": "KV created but failed to update config, err: " + err.Error()})
	} else {
		// dblayer.SetKVStatus(kvUID, "active", "")
		c.JSON(200, gin.H{"id": newKV.ID, "message": "KV created successfully"})
	}
}

// ListKVs lists all KV resources for user
func ListKVs(c *gin.Context) {
	userUID := c.GetString("user_id")

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	// kvs, err := dblayer.ListKVsByUser(userUID)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "failed to query"})
	// 	return
	// }
	c.JSON(200, gin.H{"kvs": combinator.KVs})
}

// DeleteRDB deletes an RDB resource
func DeleteRDB(c *gin.Context) {
	userUID := c.GetString("user_id")
	rdbUID := c.Param("id")

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	// 先查找有没有
	newRDBs := []k8s.RDBItem{}
	isExist := false
	for _, rdb := range combinator.RDBs {
		if rdb.ID == rdbUID {
			isExist = true
			continue
		}
		newRDBs = append(newRDBs, rdb)
	}
	if !isExist {
		c.JSON(404, gin.H{"error": "not found this RDB: " + rdbUID})
		return
	}
	combinator.RDBs = newRDBs

	// Reload combinator config
	if err := combinator.UpdateConfig(); err != nil {
		c.JSON(500, gin.H{"error": "failed to update combinator config: " + err.Error()})
		return
	}
	// rows, err := dblayer.DeleteRDB(rdbUID, userUID)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "failed to delete: " + err.Error()})
	// 	return
	// }
	// if rows == 0 {
	// 	c.JSON(404, gin.H{"error": "not found"})
	// 	return
	// }

	c.JSON(200, gin.H{"message": "deleted"})
}

// DeleteKV deletes a KV resource
func DeleteKV(c *gin.Context) {
	userUID := c.GetString("user_id")
	kvUID := c.Param("id")

	combinator, err := k8s.GetCombinatorConfig(userUID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get combinator config: " + err.Error()})
		return
	}

	// 先查找有没有
	newKVs := []k8s.KVItem{}
	isExist := false
	for _, kv := range combinator.KVs {
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
	combinator.KVs = newKVs

	// Reload combinator config
	if err := combinator.UpdateConfig(); err != nil {
		c.JSON(500, gin.H{"error": "failed to update combinator config: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "deleted"})
}
