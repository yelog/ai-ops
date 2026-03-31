# AI-K8S-OPS 用户使用指南

## 目录

1. [系统要求](#系统要求)
2. [安装部署](#安装部署)
3. [快速开始](#快速开始)
4. [集群部署](#集群部署)
5. [AI 智能助手](#ai-智能助手)
6. [常见操作](#常见操作)
7. [故障排查](#故障排查)

---

## 系统要求

### 后端服务器

**最低配置：**
- CPU: 2 核
- 内存: 4 GB
- 磁盘: 20 GB
- 操作系统: Ubuntu 20.04+ / CentOS 7+ / Debian 10+

**推荐配置：**
- CPU: 4 核
- 内存: 8 GB
- 磁盘: 50 GB SSD

**软件依赖：**
- Go 1.21+（用于编译）
- SQLite 3
- Node.js 18+（前端构建）

### K8S 节点要求

**开发环境（单节点）：**
- CPU: 2 核
- 内存: 4 GB
- 磁盘: 20 GB

**生产环境（高可用）：**
- Master 节点: 3 台
  - CPU: 4 核
  - 内存: 8 GB
  - 磁盘: 50 GB SSD
- Worker 节点: 3+ 台
  - CPU: 8 核
  - 内存: 16 GB
  - 磁盘: 100 GB SSD

---

## 安装部署

### 方式一：二进制部署

#### 1. 下载安装包

```bash
# 下载最新版本
wget https://github.com/your-org/ai-k8s-ops/releases/download/v1.0.0/ai-k8s-ops-linux-amd64.tar.gz

# 解压
tar -xzf ai-k8s-ops-linux-amd64.tar.gz

# 进入目录
cd ai-k8s-ops
```

#### 2. 配置

```bash
# 复制示例配置
cp configs/config.example.yaml configs/config.yaml

# 编辑配置
vim configs/config.yaml
```

**关键配置项：**

```yaml
server:
  port: 8080              # 服务端口
  mode: production        # 生产模式

ai:
  provider: openai
  api_key: "sk-xxx"      # OpenAI API Key（必须配置）
  model: "gpt-4-turbo"    # AI 模型

auth:
  jwt_secret: "your-secret-key"  # JWT 密钥（必须修改）
```

#### 3. 初始化数据

```bash
# 初始化默认模板
./bin/seed

# 输出：
# Created template: 开发环境模板
# Created template: 测试环境模板
# Created template: 生产环境模板
```

#### 4. 启动服务

```bash
# 直接启动
./bin/server

# 或使用 systemd 管理
sudo cp scripts/ai-k8s-ops.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ai-k8s-ops
sudo systemctl start ai-k8s-ops
```

#### 5. 访问应用

打开浏览器访问：`http://your-server-ip:8080`

---

### 方式二：Docker 部署

#### 1. 拉取镜像

```bash
docker pull your-org/ai-k8s-ops:latest
```

#### 2. 创建配置文件

```bash
mkdir -p /opt/ai-k8s-ops/{data,configs}

# 创建配置文件
cat > /opt/ai-k8s-ops/configs/config.yaml <<EOF
server:
  port: 8080
  mode: production

database:
  type: sqlite
  path: /data/ai-k8s-ops.db

ai:
  provider: openai
  api_key: "sk-xxx"
  model: "gpt-4-turbo"

auth:
  jwt_secret: "your-secret-key"
EOF
```

#### 3. 启动容器

```bash
docker run -d \
  --name ai-k8s-ops \
  -p 8080:8080 \
  -v /opt/ai-k8s-ops/data:/data \
  -v /opt/ai-k8s-ops/configs:/configs \
  your-org/ai-k8s-ops:latest
```

---

### 方式三：Kubernetes 部署

```yaml
# ai-k8s-ops-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-k8s-ops
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ai-k8s-ops
  template:
    metadata:
      labels:
        app: ai-k8s-ops
    spec:
      containers:
      - name: ai-k8s-ops
        image: your-org/ai-k8s-ops:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /configs
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: ai-k8s-ops-pvc
      - name: config
        configMap:
          name: ai-k8s-ops-config
---
apiVersion: v1
kind: Service
metadata:
  name: ai-k8s-ops
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: ai-k8s-ops
```

---

## 快速开始

### 1. 注册账号

1. 访问应用首页，点击"立即注册"
2. 填写注册信息：
   - 用户名：至少 3 个字符
   - 邮箱：有效的邮箱地址
   - 密码：至少 8 个字符
   - 角色：选择 observer/operator/admin
3. 点击"注册"按钮

### 2. 登录系统

1. 使用注册的用户名和密码登录
2. 登录成功后跳转到仪表盘

### 3. 仪表盘概览

仪表盘显示：
- 集群总数
- 健康集群数量
- 告警数量
- 待处理事项

---

## 集群部署

### 方式一：使用部署模板

#### 1. 查看部署模板

1. 点击左侧菜单"部署中心"
2. 选择"部署模板"
3. 查看可用的预置模板：
   - **开发环境模板**: 单节点集群，适合开发测试
   - **测试环境模板**: 3 节点集群，包含完整监控栈
   - **生产环境模板**: 高可用集群，适合生产环境

#### 2. 创建部署任务

1. 在"部署中心"点击"部署任务"
2. 点击"新建部署任务"
3. 选择：
   - 目标集群（新集群选择"创建新集群"）
   - 部署模板
   - 节点信息

#### 3. 填写节点信息

**单节点部署（开发环境）：**

```yaml
节点 IP: 192.168.1.100
SSH 用户: root
SSH 密码: ********
节点角色: master + worker
```

**多节点部署（生产环境）：**

```yaml
Master 节点:
  - 192.168.1.10 (master)
  - 192.168.1.11 (master)
  - 192.168.1.12 (master)

Worker 节点:
  - 192.168.1.20 (worker)
  - 192.168.1.21 (worker)
  - 192.168.1.22 (worker)

SSH 用户: root
SSH 密钥: 选择私钥文件
```

#### 4. 开始部署

1. 确认配置信息
2. 点击"开始部署"
3. 实时查看部署进度

**部署步骤：**

```
✓ 1. 环境检测 (完成) - 10%
✓ 2. 依赖安装 (完成) - 30%
✓ 3. 网络配置 (完成) - 50%
⏳ 4. K8S 安装 (进行中) - 70%
○ 5. 组件安装 - 90%
○ 6. 集群验证 - 100%
```

#### 5. 查看部署日志

部署过程中可以查看实时日志：

```
[10:35:02] INFO: 开始安装 Kubernetes 组件
[10:35:05] INFO: 安装 kubeadm, kubelet, kubectl
[10:35:15] INFO: 初始化控制平面
[10:35:25] INFO: 配置 kubectl
[10:35:30] INFO: 安装网络插件 Calico
[10:35:45] INFO: 加入工作节点
[10:36:00] INFO: 集群部署完成
```

---

### 方式二：导入现有集群

如果你已经有运行的 K8S 集群，可以直接导入：

#### 1. 准备 kubeconfig

```bash
# 从现有集群获取 kubeconfig
cat ~/.kube/config
```

#### 2. 导入集群

1. 点击"集群管理" -> "导入集群"
2. 填写信息：
   - 集群名称：my-existing-cluster
   - 环境：生产环境
   - kubeconfig：粘贴配置内容
3. 点击"导入"

#### 3. 验证导入

导入成功后，可以查看：
- 节点信息
- 资源使用情况
- Pod 状态

---

### 方式三：使用 AI 助手部署

#### 1. 打开 AI 助手

点击左侧菜单"AI 助手"

#### 2. 自然语言创建集群

输入：

```
请帮我创建一个生产环境的 K8S 集群，要求：
- 3 个 Master 节点
- 5 个 Worker 节点
- 使用 Calico 网络插件
- 版本 1.28.0
```

AI 助手会：

1. 分析你的需求
2. 生成部署配置
3. 显示部署计划

```
AI 回复：

好的，我将为您创建一个高可用的生产环境 K8S 集群。

部署计划：
- 集群名称：prod-cluster
- 版本：v1.28.0
- 高可用模式：是
- Master 节点：3 台
- Worker 节点：5 台
- 网络插件：Calico

请提供以下信息：
1. Master 节点 IP 地址（3个）
2. Worker 节点 IP 地址（5个）
3. SSH 访问凭据

确认后我将开始部署。
```

---

## AI 智能助手

### 功能介绍

AI 助手可以帮助你：

- 🔍 **查询集群信息**: "显示生产集群的节点状态"
- 🔧 **诊断问题**: "为什么 frontend Pod 一直重启？"
- 📊 **获取建议**: "如何优化集群性能？"
- 🚀 **执行操作**: "扩容 deployment 到 5 个副本"
- 📚 **学习 K8S**: "什么是 StatefulSet？"

### 使用示例

#### 示例 1：查询集群状态

```
用户: 显示生产集群的资源使用情况

AI: 当前生产集群资源使用概览：

• CPU 使用率: 65%
• 内存使用率: 72%
• Pod 数量: 234/300
• 节点状态: 5/5 健康

资源使用 Top 3 Namespace:
1. default: 45% CPU, 55% 内存
2. monitoring: 15% CPU, 20% 内存
3. logging: 8% CPU, 12% 内存

建议：内存使用较高，建议关注 default namespace。
```

#### 示例 2：故障诊断

```
用户: 为什么 elasticsearch Pod 一直重启？

AI: 我来帮您诊断 Elasticsearch Pod 重启问题。

分析结果：
• Pod: elasticsearch-0
• 命名空间: logging
• 重启次数: 8 次
• 最后重启时间: 5 分钟前

查看日志发现：
[ERROR] OOMKilled - 内存不足

原因分析：
当前内存限制 4GB，但实际使用已达 4.2GB。

建议修复方案：
1. 增加内存限制到 8GB（推荐）
2. 优化 JVM heap 配置
3. 清理不必要的索引

是否执行方案 1？
[执行] [查看详情] [取消]
```

#### 示例 3：执行操作

```
用户: 扩容 nginx deployment 到 5 个副本

AI: 即将执行以下操作：

操作：扩容 Deployment
目标：nginx (default namespace)
当前副本数：2
目标副本数：5

影响范围：
- 增加 3 个 Pod
- 需要额外资源：1.5 CPU, 3GB 内存
- 服务可用性：不受影响

确认执行？
[确认执行] [取消]
```

---

## 常见操作

### 1. 查看集群详情

1. 点击"集群管理"
2. 选择集群名称
3. 查看：
   - 节点列表
   - 资源使用
   - 组件状态
   - 命名空间

### 2. 创建部署模板

1. 点击"部署中心" -> "部署模板"
2. 点击"创建模板"
3. 填写：

```yaml
名称: 自定义模板
描述: 适合特定业务场景
类型: custom
提供商: bare-metal
配置:
  masters: 3
  workers: 5
  version: v1.28.0
  network: calico
组件:
  - prometheus
  - grafana
  - loki
```

### 3. 监控集群

1. 点击"监控中心"
2. 查看：
   - 节点资源使用
   - Pod 状态
   - 网络流量
   - 存储使用

### 4. 配置告警

1. 点击"告警中心"
2. 创建告警规则：

```yaml
名称: CPU 使用率告警
级别: P2
条件: CPU 使用率 > 80%
持续时间: 5 分钟
通知方式: 钉钉/企业微信/邮件
```

---

## 故障排查

### 问题 1：无法访问 Web UI

**检查步骤：**

```bash
# 1. 检查服务状态
systemctl status ai-k8s-ops

# 2. 检查端口
netstat -tlnp | grep 8080

# 3. 检查防火墙
sudo ufw status
sudo ufw allow 8080

# 4. 查看日志
tail -f data/logs/server.log
```

### 问题 2：AI 助手无响应

**可能原因：**

1. OpenAI API Key 未配置
2. API Key 无效
3. 网络无法访问 OpenAI

**解决方法：**

```bash
# 检查配置
grep api_key configs/config.yaml

# 测试 API Key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer sk-xxx"

# 如果使用代理，设置环境变量
export HTTP_PROXY=http://proxy:port
export HTTPS_PROXY=http://proxy:port
```

### 问题 3：部署任务失败

**查看日志：**

```bash
# 查看部署日志
cat data/logs/deployment-{task-id}.log

# 查看系统日志
journalctl -u ai-k8s-ops -f
```

**常见错误：**

| 错误 | 原因 | 解决方法 |
|------|------|----------|
| SSH 连接失败 | 密码错误/网络不通 | 检查 SSH 凭据和网络 |
| 依赖安装失败 | 网络问题/源不可用 | 配置国内镜像源 |
| 端口被占用 | 已有服务占用端口 | 停止冲突服务 |
| 磁盘空间不足 | 磁盘已满 | 清理磁盘空间 |

---

## 最佳实践

### 1. 生产环境部署建议

- ✅ 使用高可用部署（3 Master + 3+ Worker）
- ✅ 配置负载均衡器
- ✅ 启用 etcd 备份
- ✅ 配置资源配额和限制
- ✅ 启用网络策略

### 2. 安全建议

- 🔒 修改默认 JWT 密钥
- 🔒 使用强密码策略
- 🔒 配置 HTTPS
- 🔒 限制 API 访问 IP
- 🔒 定期更新系统

### 3. 性能优化

- ⚡ 使用 SSD 存储
- ⚡ 合理配置资源限制
- ⚡ 启用 Pod 优先级
- ⚡ 配置水平自动扩缩容（HPA）

---

## 附录

### 命令行工具

```bash
# 查看版本
./bin/server version

# 初始化数据库
./bin/server init

# 种子数据
./bin/seed

# 健康检查
curl http://localhost:8080/api/v1/system/health
```

### 配置文件详解

```yaml
# configs/config.yaml

server:
  port: 8080              # 服务端口
  mode: production        # 运行模式: development/production

database:
  type: sqlite            # 数据库类型
  path: data/ai-k8s-ops.db  # 数据库路径

auth:
  jwt_secret: "xxx"       # JWT 密钥（必须修改）
  jwt_expiry_hours: 24    # Token 有效期（小时）

ai:
  provider: openai        # AI 提供商
  api_key: "sk-xxx"       # API Key（必须配置）
  model: "gpt-4-turbo"    # 模型名称
  embedding_model: "text-embedding-ada-002"  # Embedding 模型

monitoring:
  prometheus_retention_days: 15  # Prometheus 数据保留天数

logging:
  level: info             # 日志级别: debug/info/warn/error
  file: data/logs/server.log  # 日志文件路径

security:
  encryption_key: "xxx"   # 数据加密密钥（必须修改）
```

---

## 技术支持

- 📖 文档：https://ai-k8s-ops.readthedocs.io
- 💬 社区：https://github.com/your-org/ai-k8s-ops/discussions
- 🐛 问题：https://github.com/your-org/ai-k8s-ops/issues
- 📧 邮件：support@example.com

---

**祝您使用愉快！** 🎉