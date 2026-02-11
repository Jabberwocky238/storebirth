# Console Project

## 项目结构

### 入口
- `cmd/inner/main.go` - 内网网关入口
- `cmd/outer/main.go` - 外网网关入口

### 代码组织
- `dblayer/` - 所有数据库操作函数
- `k8s/` - 所有K8s资源操作
- `handlers/` - 所有HTTP端点处理器

### 部署配置
- `scripts/` - K8s部署YAML文件
- `scripts/init.sql` - 数据库初始化表结构

## 架构组件

### 核心组件
- **Worker** - 无状态工作节点，处理实际的业务逻辑和任务执行，用户侧自定义。
- **Combinator** - 合并网关层，统一管理：
  - 数据库操作（PostgreSQL）
  - KV存储（Redis）
  - 对象存储（S3）
  - 消息队列
  - 源码位于 `combinator/` 子目录（独立仓库）
- **自定义域名** - 域名管理服务，支持：
  - TXT记录验证
  - 自动创建IngressRoute
  - cert-manager证书管理
  - 代码在 `k8s/customdomain.go` 和 `handlers/customdomain.handler.go`

### 网关入口
- **Inner** (`cmd/inner/main.go`) - 内网API，用于集群内部服务间通信
- **Outer** (`cmd/outer/main.go`) - 外网API，面向用户的公开接口
