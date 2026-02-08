# CockroachDB PVC 迁移手册

## 概述

本手册用于处理 CockroachDB 节点故障后的 PVC 迁移和恢复操作。

当 Kubernetes 节点死亡时，绑定在该节点上的 PVC（使用 local-path 存储）无法迁移到其他节点，导致 Pod 无法启动。需要手动删除孤立的 PVC，让 StatefulSet 重新创建。

---

## 架构说明

### 当前配置
- **存储类型**: local-path (本地文件系统)
- **副本策略**: CockroachDB 自身管理 3 副本
- **故障检测**: 节点死亡 2 分钟后自动重新分配副本
- **PVC 大小**: 40Gi per Pod

### 数据安全性
- CockroachDB 默认 3 副本分布在不同节点
- 单节点故障不会导致数据丢失
- 节点故障后，其他节点的副本仍然可用

---

## 故障场景

### 场景 1：单节点故障

**初始状态：**
```
Node-A: cockroachdb-0 + PVC-0 ✅
Node-B: cockroachdb-1 + PVC-1 ✅
Node-C: cockroachdb-2 + PVC-2 ✅
```

**Node-B 死亡后：**
```
Node-A: cockroachdb-0 + PVC-0 ✅
Node-B: [死亡] cockroachdb-1 (NotReady) + PVC-1 (孤立) ❌
Node-C: cockroachdb-2 + PVC-2 ✅
```

**问题：**
- PVC-1 绑定在 Node-B 上，无法迁移
- cockroachdb-1 无法在其他节点启动
- 集群只有 2 个节点运行（仍然可用）

---

## 恢复操作步骤

### 步骤 1：确认节点状态

```bash
# 查看节点状态
kubectl get nodes

# 查看 Pod 状态
kubectl get pods -n cockroachdb -o wide

# 查看 PVC 状态
kubectl get pvc -n cockroachdb
```

**预期输出：**
```
NAME                              STATUS    NODE
datadir-cockroachdb-0            Bound     node-a
datadir-cockroachdb-1            Bound     node-b  (节点已死亡)
datadir-cockroachdb-2            Bound     node-c
```

### 步骤 2：强制删除死亡节点上的 Pod

```bash
# 强制删除 Pod（假设是 cockroachdb-1）
kubectl delete pod cockroachdb-1 -n cockroachdb --force --grace-period=0
```

### 步骤 3：删除孤立的 PVC（关键步骤）

```bash
# 删除绑定在死亡节点上的 PVC
kubectl delete pvc datadir-cockroachdb-1 -n cockroachdb
```

**警告：** 删除 PVC 会丢失该节点的本地数据，但 CockroachDB 的其他副本仍然保留完整数据。

### 步骤 4：等待自动恢复

StatefulSet 会自动执行以下操作：

1. 创建新的 PVC（datadir-cockroachdb-1）
2. 在可用节点上启动新的 Pod
3. CockroachDB 自动从其他节点同步数据

```bash
# 监控 Pod 状态
kubectl get pods -n cockroachdb -w
```

### 步骤 5：验证集群健康状态

```bash
# 检查所有 Pod 是否 Running
kubectl get pods -n cockroachdb

# 连接到 CockroachDB 查看集群状态
kubectl exec -it cockroachdb-0 -n cockroachdb -- /cockroach/cockroach node status --insecure
```

**预期输出：** 所有节点状态应为 `live`

---

## 监控副本同步进度

### 方法 1：Web UI（推荐）

访问 CockroachDB Web 界面：

```bash
# 端口转发到本地
kubectl port-forward -n cockroachdb svc/cockroachdb-public 8080:8080
```

然后访问：`http://localhost:8080`

查看以下页面：
- **Replication Dashboard** → "Snapshot Data Received" 图表
- **Queues Dashboard** → "Raft Snapshot Queue" 状态

### 方法 2：SQL 查询

```bash
# 查看副本分布情况
kubectl exec -it cockroachdb-0 -n cockroachdb -- /cockroach/cockroach sql --insecure -e "SELECT * FROM crdb_internal.ranges WHERE under_replicated = true;"
```

如果返回空结果，说明所有副本已同步完成。

