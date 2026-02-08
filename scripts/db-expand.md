# CockroachDB 存储卷扩容手册（local-path）

## 概述

kubectl patch storageclass local-path -p '{"allowVolumeExpansion": true}'


本手册用于 local-path 存储的 CockroachDB PVC 扩容操作。

**重要：local-path 存储的特点**
- local-path 使用宿主机本地目录
- PVC 中的容量（如 10Gi）只是"软限制"
- 实际可用空间 = 宿主机磁盘空间
- 无真正的容量限制

---

## 当前 PVC 状态检查

```bash
kubectl get pvc -n cockroachdb
```

**示例输出：**
```
NAME                    STATUS   VOLUME                                     CAPACITY   STORAGECLASS
datadir-cockroachdb-0   Bound    pvc-a9d7f823-7b5a-4a24-ae16-0d988bac7800   10Gi       local-path
datadir-cockroachdb-1   Bound    pvc-2ca32050-92fe-44bd-a841-01e370ff4d68   10Gi       local-path
datadir-cockroachdb-2   Bound    pvc-d27409b6-a99e-4cc7-87a5-9d35b35408f6   10Gi       local-path
```

---

## 步骤 1：查找 PV 对应的宿主机路径

### 1.1 查看 PV 列表

```bash
kubectl get pv
```

**示例输出：**
```
NAME                                       CAPACITY   STORAGECLASS   STATUS
pvc-a9d7f823-7b5a-4a24-ae16-0d988bac7800   10Gi       local-path     Bound
```

### 1.2 查看 PV 的宿主机路径

```bash
kubectl get pv pvc-a9d7f823-7b5a-4a24-ae16-0d988bac7800 -o yaml | grep path
```

**示例输出：**
```
path: /opt/local-path-provisioner/pvc-a9d7f823-7b5a-4a24-ae16-0d988bac7800
```

---

## 步骤 2：验证宿主机磁盘空间

### 2.1 查看 Pod 运行在哪个节点

```bash
kubectl get pod -n cockroachdb -o wide
```

**示例输出：**
```
NAME             READY   STATUS    NODE
cockroachdb-0    1/1     Running   node-1
cockroachdb-1    1/1     Running   node-2
cockroachdb-2    1/1     Running   node-3
```

### 2.2 SSH 登录到节点检查磁盘空间

```bash
# SSH 登录到对应节点（例如 node-1）
ssh root@node-1

# 检查磁盘空间
df -h /opt/local-path-provisioner
```

**示例输出：**
```
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1       500G   50G  450G  10% /
```

**结论：** 宿主机有 450G 可用空间，CockroachDB 可以继续使用。

---

## 总结

### local-path 存储的特点

1. **无真正容量限制**
   - PVC 显示的 10Gi 只是声明值
   - 实际可用空间 = 宿主机磁盘空间
   - CockroachDB 可以使用超过 10Gi 的空间

2. **不需要扩容 PVC**
   - 只需确保宿主机有足够空间
   - 无需修改 PVC 配置
   - 无需重启 Pod

3. **监控建议**
   - 定期检查宿主机磁盘使用率
   - 避免磁盘空间耗尽
   - 建议保留至少 20% 空闲空间

