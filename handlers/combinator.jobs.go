package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"jabberwocky238/console/dblayer"
	"jabberwocky238/console/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// notifyAllCombinatorPods 向所有 combinator pod 发送删除通知
func notifyAllCombinatorPods(userUID, resourceID, resourceType string) error {
	if k8s.K8sClient == nil {
		return fmt.Errorf("k8s client not available")
	}

	ctx := context.Background()

	// 获取所有 combinator pod
	pods, err := k8s.K8sClient.CoreV1().Pods("combinator").List(ctx, metav1.ListOptions{
		LabelSelector: "app=combinator",
	})
	if err != nil {
		return fmt.Errorf("failed to list combinator pods: %w", err)
	}

	// 准备请求数据
	payload := map[string]string{
		"user_uid":      userUID,
		"resource_id":   resourceID,
		"resource_type": resourceType,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 向每个 pod 发送请求
	client := &http.Client{Timeout: 5 * time.Second}
	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running" {
			continue
		}

		podIP := pod.Status.PodIP
		if podIP == "" {
			continue
		}

		url := fmt.Sprintf("http://%s:8890/webhook", podIP)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[combinator] failed to create request for pod %s: %v", pod.Name, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[combinator] failed to notify pod %s: %v", pod.Name, err)
			continue
		}
		resp.Body.Close()

		log.Printf("[combinator] notified pod %s about deletion of %s/%s", pod.Name, resourceType, resourceID)
	}

	return nil
}

// --- CreateRDBJob ---

type CreateRDBJob struct {
	RecordID   string // combinator_resources.id
	UserUID    string
	Name       string
	ResourceID string
}

func NewCreateRDBJob(recordID, userUID, name, resourceID string) *CreateRDBJob {
	return &CreateRDBJob{
		RecordID:   recordID,
		UserUID:    userUID,
		Name:       name,
		ResourceID: resourceID,
	}
}

func (j *CreateRDBJob) Type() string { return "combinator.create_rdb" }
func (j *CreateRDBJob) ID() string   { return j.RecordID }

func (j *CreateRDBJob) Do() error {
	if k8s.RDBManager == nil {
		dblayer.UpdateCombinatorResourceStatus(j.RecordID, "error", "cockroachdb not available")
		return fmt.Errorf("cockroachdb not available")
	}
	if err := k8s.RDBManager.InitUserRDB(j.UserUID); err != nil {
		dblayer.UpdateCombinatorResourceStatus(j.RecordID, "error", err.Error())
		return fmt.Errorf("init user rdb: %w", err)
	}
	if err := k8s.RDBManager.CreateSchema(j.UserUID, j.ResourceID); err != nil {
		dblayer.UpdateCombinatorResourceStatus(j.RecordID, "error", err.Error())
		return fmt.Errorf("create schema: %w", err)
	}

	dblayer.UpdateCombinatorResourceStatus(j.RecordID, "active", "")
	log.Printf("[combinator] RDB %s created for user %s", j.ResourceID, j.UserUID)
	return nil
}

// --- DeleteRDBJob ---

type DeleteRDBJob struct {
	UserUID    string
	ResourceID string
}

func NewDeleteRDBJob(userUID, resourceID string) *DeleteRDBJob {
	return &DeleteRDBJob{UserUID: userUID, ResourceID: resourceID}
}

func (j *DeleteRDBJob) Type() string { return "combinator.delete_rdb" }
func (j *DeleteRDBJob) ID() string   { return j.ResourceID }

func (j *DeleteRDBJob) Do() error {
	if k8s.RDBManager != nil {
		if err := k8s.RDBManager.DeleteSchema(j.UserUID, j.ResourceID); err != nil {
			log.Printf("[combinator] delete schema %s failed: %v", j.ResourceID, err)
		}
	}

	// 通知所有 combinator pod
	if err := notifyAllCombinatorPods(j.UserUID, j.ResourceID, "rdb"); err != nil {
		log.Printf("[combinator] failed to notify pods about RDB deletion: %v", err)
	}

	log.Printf("[combinator] RDB %s deleted for user %s", j.ResourceID, j.UserUID)
	return nil
}

// --- CreateKVJob ---

type CreateKVJob struct {
	RecordID   string // combinator_resources.id
	UserUID    string
	ResourceID string
}

func NewCreateKVJob(recordID, userUID, resourceID string) *CreateKVJob {
	return &CreateKVJob{
		RecordID:   recordID,
		UserUID:    userUID,
		ResourceID: resourceID,
	}
}

func (j *CreateKVJob) Type() string { return "combinator.create_kv" }
func (j *CreateKVJob) ID() string   { return j.RecordID }

func (j *CreateKVJob) Do() error {
	dblayer.UpdateCombinatorResourceStatus(j.RecordID, "active", "")
	log.Printf("[combinator] KV %s created for user %s", j.ResourceID, j.UserUID)
	return nil
}

// --- DeleteKVJob ---

type DeleteKVJob struct {
	UserUID    string
	ResourceID string
}

func NewDeleteKVJob(userUID, resourceID string) *DeleteKVJob {
	return &DeleteKVJob{UserUID: userUID, ResourceID: resourceID}
}

func (j *DeleteKVJob) Type() string { return "combinator.delete_kv" }
func (j *DeleteKVJob) ID() string   { return j.ResourceID }

func (j *DeleteKVJob) Do() error {
	// 通知所有 combinator pod
	if err := notifyAllCombinatorPods(j.UserUID, j.ResourceID, "kv"); err != nil {
		log.Printf("[combinator] failed to notify pods about KV deletion: %v", err)
	}

	log.Printf("[combinator] KV %s deleted for user %s", j.ResourceID, j.UserUID)
	return nil
}
