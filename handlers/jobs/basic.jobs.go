package jobs

import (
	"encoding/json"
	"fmt"
	"jabberwocky238/console/k8s"
)

const (
	JobTypeAuthRegisterUser     k8s.JobType = "auth.register_user"
	JobTypeAuthUserAudit        k8s.JobType = "auth.user_audit"
	JobTypeWorkerDeployWorker   k8s.JobType = "worker.deploy_worker"
	JobTypeWorkerDeleteWorkerCR k8s.JobType = "worker.delete_worker_cr"
	JobTypeWorkerSyncEnv        k8s.JobType = "worker.sync_env"
	JobTypeWorkerSyncSecret     k8s.JobType = "worker.sync_secret"
	JobTypeCombinatorCreateRDB  k8s.JobType = "combinator.create_rdb"
	JobTypeCombinatorDeleteRDB  k8s.JobType = "combinator.delete_rdb"
	JobTypeCombinatorCreateKV   k8s.JobType = "combinator.create_kv"
	JobTypeCombinatorDeleteKV   k8s.JobType = "combinator.delete_kv"
	JobTypeDomainCheck          k8s.JobType = "domain.check"
)

type ObjectBuilder func() k8s.Job

// JobFactory 用于根据类型反序列化 Job
type JobFactory struct {
	objectBuilders map[k8s.JobType]ObjectBuilder
}

// JobFactory 用于根据类型反序列化 Job
var globalFactory = &JobFactory{
	objectBuilders: make(map[k8s.JobType]ObjectBuilder),
}

// RegisterJobType 注册 Job 类型的反序列化函数
func RegisterJobType(jobType k8s.JobType, deserializer func() k8s.Job) {
	globalFactory.objectBuilders[jobType] = deserializer
}

// CreateJob 根据类型和数据创建 Job 实例
func CreateJob(jobType k8s.JobType, data []byte) (k8s.Job, error) {
	objectBuilder, ok := globalFactory.objectBuilders[jobType]
	if !ok {
		return nil, fmt.Errorf("unknown job type: %s", jobType)
	}
	object := objectBuilder()
	if err := json.Unmarshal(data, &object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job data: %w", err)
	}
	return object, nil
}
