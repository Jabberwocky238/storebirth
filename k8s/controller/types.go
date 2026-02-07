package controller

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group          = "console.app238.com"
	Version        = "v1"
	WorkerResource = "workerapps"
	WorkerKind     = "WorkerApp"
)

var WorkerAppGVR = schema.GroupVersionResource{
	Group:    Group,
	Version:  Version,
	Resource: WorkerResource,
}

type WorkerAppSpec struct {
	WorkerID string `json:"workerID"`
	OwnerID  string `json:"ownerID"`
	Image    string `json:"image"`
	Port     int    `json:"port"`
}

type WorkerAppStatus struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
}
