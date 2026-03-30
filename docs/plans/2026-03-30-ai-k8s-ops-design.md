# AI-K8S-OPS 产品设计文档

## 文档信息

- **项目名称**: AI-K8S-OPS (AI 驱动的 K8S 运维平台)
- **版本**: v1.0
- **创建日期**: 2026-03-30
- **目标用户**: 中小团队（10-50人）
- **部署环境**: 多集群管理

---

## 一、项目概述

### 1.1 项目定位

AI-K8S-OPS 是一个 AI 驱动的 Kubernetes 运维管理平台，旨在通过智能化手段简化 K8S 集群的部署、管理和运维工作。系统集成了快速部署、自然语言交互、实时监控与主动修复三大核心能力。

### 1.2 核心价值

- **快速部署**: 通过集成多种部署方案，快速搭建单机或集群 K8S 环境
- **智能交互**: AI 自然语言交互，降低运维门槛，提升效率
- **主动运维**: 实时监控、智能诊断、主动修复建议，降低故障风险

### 1.3 目标用户

- 中小团队（10-50人）
- 运维人员、开发人员、DevOps 工程师
- 需要管理多个 K8S 集群的团队

---

## 二、整体架构设计

### 2.1 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      用户层                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Web UI (React/Vue)                      │  │
│  │  - 集群管理界面                                    │  │
│  │  - AI 对话界面                                     │  │
│  │  - 监控仪表盘                                      │  │
│  │  - 告警中心                                        │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   API Gateway 层                        │
│  - 统一认证授权 (JWT + RBAC)                            │
│  - API 路由与版本管理                                   │
│  - 限流与熔断                                          │
│  - 审计日志记录                                        │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  核心业务模块                            │
│  ┌─────────────┬─────────────┬─────────────┐          │
│  │ 集群部署模块 │ AI 交互模块  │ 监控告警模块 │          │
│  │             │             │             │          │
│  │ - 集群创建  │ - 自然语言  │ - 指标采集  │          │
│  │ - 集群导入  │ - 智能诊断  │ - 日志分析  │          │
│  │ - 配置管理  │ - 修复建议  │ - 告警管理  │          │
│  │ - 状态备份  │ - 知识库    │ - 自动修复  │          │
│  └─────────────┴─────────────┴─────────────┘          │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│                   Agent 层                              │
│  - 部署在每个 K8S 集群的轻量级 Agent                     │
│  - 执行部署、监控数据采集、日志收集                      │
│  - 与核心系统通过 gRPC/HTTPS 通信                       │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│               K8S 集群 & 监控基础设施                    │
│  ┌──────────┬──────────┬──────────┬──────────┐        │
│  │Prometheus│  Grafana  │Loki      │ Jaeger   │        │
│  │  (监控)  │ (可视化)  │ (日志)   │ (追踪)   │        │
│  └──────────┴──────────┴──────────┴──────────┘        │
│  ┌──────────┬──────────┬──────────┐                   │
│  │ SQLite   │ 本地文件 │ FAISS    │                   │
│  │ (数据库) │ (存储)   │(向量检索)│                   │
│  └──────────┴──────────┴──────────┘                   │
└─────────────────────────────────────────────────────────┘
```

### 2.2 架构特点

- **模块化单体架构**: 初期采用模块化单体，后期可按需拆分微服务
- **Agent 架构**: 支持多集群管理，轻量级代理部署
- **轻量化存储**: SQLite + 本地文件系统，降低依赖复杂度

---

## 三、技术栈选型

### 3.1 后端技术栈

```yaml
核心框架:
  - 语言: Go 1.21+
  - Web框架: Gin / Fiber
  - CLI工具: Cobra
  
AI 集成:
  - LLM SDK: OpenAI Go SDK / LangChain Go
  - 向量检索: FAISS (文件存储)
  - Embedding: text-embedding-ada-002
  
K8S 交互:
  - 客户端库: client-go
  - 部署工具: Kubespray / RKE2 / kubeadm
  
Agent 通信:
  - 协议: gRPC
  
数据访问:
  - SQLite: go-sqlite3 / modernc.org/sqlite
  - 文件系统: 本地存储
  - 缓存: 内存缓存
```

### 3.2 前端技术栈

```yaml
框架: React 18+
UI 组件库: Ant Design Pro
状态管理: Zustand / React Query
可视化: ECharts
终端: XTerm.js
构建工具: Vite
国际化: react-i18next
```

### 3.3 监控与可观测性

```yaml
监控: Prometheus (内置存储，15天数据)
可视化: Grafana
日志: Loki (轻量级) 或 文件日志
追踪: Jaeger (可选)
```

### 3.4 数据存储（轻量化）

```yaml
SQLite (单文件数据库):
  - 用户与权限
  - 集群元数据
  - AI 对话历史
  - 部署模板与任务
  - 告警规则
  - 修复记录
  - 审计日志
  
本地文件系统:
  data/
    backups/        # 集群备份
    configs/        # 配置文件
    logs/           # 日志归档
    templates/      # 部署模板
    knowledge/      # 知识库文档
    embeddings/     # 向量索引
```

---

## 四、核心功能模块设计

### 4.1 快速部署模块

#### 功能架构

```
集群部署工作流引擎
├─ 1. 环境准备
│  - 节点检测
│  - 依赖安装
│  - 网络规划
├─ 2. 集群初始化
│  - 选择部署方案
│  - 生成配置
│  - 部署控制平面
│  - 添加工作节点
├─ 3. 组件安装
│  - 网络插件
│  - 存储插件
│  - 监控组件
│  - 日志组件
├─ 4. 集群验证
│  - 健康检查
│  - 功能测试
│  - 性能基准测试
└─ 5. 状态持久化
   - 导出 kubeconfig
   - 注册到管理平台
   - 初始化监控采集
```

#### 支持的部署场景

| 部署类型 | 支持平台 | 部署方式 |
|---------|---------|---------|
| 裸金属部署 | 物理服务器 | Kubespray/kubeadm |
| 虚拟机部署 | VMware, OpenStack | Kubespray/RKE2 |
| 云平台部署 | AWS, Azure, GCP, 阿里云 | 托管/自建 |
| 导入现有集群 | 任意 K8S 集群 | kubeconfig |

#### 部署模板

- 开发环境模板: 单节点，基础组件
- 测试环境模板: 3节点，完整监控栈
- 生产环境模板: 高可用，完整可观测性栈，安全加固
- 自定义模板: 用户自定义配置

#### 关键设计

1. Agent 机制: 轻量级 Agent 部署在目标节点
2. 幂等性保证: 所有操作支持重试，状态持久化
3. 回滚机制: 一键回滚到之前状态
4. 进度可视化: WebSocket 实时推送日志

---

### 4.2 AI 交互模块

#### 功能架构

```
AI 智能助手
├─ 自然语言理解层
│  - 意图识别
│  - 实体抽取
│  - 上下文管理
├─ 知识库层
│  - K8S 官方文档
│  - 最佳实践库
│  - 故障案例库
│  - 集群特定知识（RAG）
├─ 工具调用层
│  - kubectl 封装
│  - 监控数据查询
│  - 日志分析工具
└─ 响应生成层
   - 结果解释
   - 可视化建议
   - 操作确认
```

#### AI 能力矩阵

| 能力类别 | 具体功能 | 示例 |
|---------|---------|-----|
| 集群信息查询 | 节点状态、资源使用 | "显示生产集群的节点状态" |
| 故障诊断 | 分析日志、指标，定位问题 | "为什么 pod 一直重启？" |
| 操作执行 | 创建/删除/更新资源 | "扩容 deployment 到 5 个副本" |
| 最佳实践建议 | 配置优化、安全建议 | "如何优化资源配置？" |
| 容量规划 | 资源预测、扩容建议 | "何时需要扩容？" |

#### 操作安全分级

```
L1 - 只读操作: 自动执行，无需确认
L2 - 低风险操作: 单次确认
L3 - 高风险操作: 二次确认
L4 - 危险操作: 多级审批
```

---

### 4.3 监控与主动修复模块

#### 功能架构

```
监控与智能修复系统
├─ 数据采集层
│  - Prometheus (指标)
│  - Loki (日志)
│  - Jaeger (追踪)
│  - Events (事件流)
├─ 智能分析引擎
│  ├─ 异常检测
│  │  - 基于规则
│  │  - 基于机器学习
│  │  - 时序分析
│  ├─ 根因定位
│  │  - 关联分析
│  │  - 拓扑分析
│  │  - AI 诊断
│  └─ 修复建议生成
│     - 预定义修复方案库
│     - AI 生成修复建议
├─ 告警与响应
│  - 告警路由
│  - 通知渠道（邮件/钉钉/企业微信）
│  - 告警抑制与聚合
└─ 自动修复执行器
   ├─ 自动修复动作库
   │  - 重启 Pod/Deployment
   │  - 扩缩容
   │  - 清理资源
   │  - 节点隔离/恢复
   └─ 执行策略
      - 手动确认执行
      - 半自动（低风险自动）
```

#### 告警级别

| 级别 | 类型 | 响应时间 |
|-----|------|---------|
| P0 - 紧急 | 集群不可用、核心故障 | < 5分钟 |
| P1 - 严重 | 节点 NotReady、存储不足 | < 15分钟 |
| P2 - 警告 | 资源使用高、Pod 重启频繁 | < 1小时 |
| P3 - 提示 | 非核心服务异常 | < 4小时 |

---

## 五、数据模型设计（轻量化）

### 5.1 SQLite 数据模型

```sql
-- 用户与权限
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

-- 集群管理
CREATE TABLE clusters (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    environment TEXT,
    provider TEXT,
    version TEXT,
    api_server TEXT,
    kubeconfig TEXT,
    status TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 节点信息
CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    cluster_id TEXT,
    name TEXT,
    ip TEXT,
    role TEXT,
    status TEXT,
    cpu_capacity INTEGER,
    memory_capacity INTEGER,
    os_info TEXT,
    labels TEXT,
    last_heartbeat TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

-- AI 对话
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    cluster_id TEXT,
    title TEXT,
    context TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT,
    role TEXT,
    content TEXT,
    tokens INTEGER,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- 知识库
CREATE TABLE knowledge_base (
    id TEXT PRIMARY KEY,
    title TEXT,
    content TEXT,
    category TEXT,
    tags TEXT,
    source TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 部署模板
CREATE TABLE deployment_templates (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    type TEXT,
    provider TEXT,
    config TEXT,
    components TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 部署任务
CREATE TABLE deployments (
    id TEXT PRIMARY KEY,
    cluster_id TEXT,
    template_id TEXT,
    status TEXT,
    current_step TEXT,
    progress INTEGER,
    error_message TEXT,
    created_by TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id),
    FOREIGN KEY (template_id) REFERENCES deployment_templates(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- 告警规则
CREATE TABLE alert_rules (
    id TEXT PRIMARY KEY,
    cluster_id TEXT,
    name TEXT,
    severity TEXT,
    promql TEXT,
    threshold REAL,
    duration INTEGER,
    labels TEXT,
    annotations TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

-- 告警事件
CREATE TABLE alerts (
    id TEXT PRIMARY KEY,
    rule_id TEXT,
    cluster_id TEXT,
    status TEXT,
    severity TEXT,
    message TEXT,
    labels TEXT,
    started_at TIMESTAMP,
    resolved_at TIMESTAMP,
    root_cause TEXT,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id),
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

-- 修复记录
CREATE TABLE remediations (
    id TEXT PRIMARY KEY,
    alert_id TEXT,
    cluster_id TEXT,
    action_type TEXT,
    action_params TEXT,
    status TEXT,
    result TEXT,
    executed_by TEXT,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    FOREIGN KEY (alert_id) REFERENCES alerts(id),
    FOREIGN KEY (cluster_id) REFERENCES clusters(id),
    FOREIGN KEY (executed_by) REFERENCES users(id)
);

-- 备份记录
CREATE TABLE backups (
    id TEXT PRIMARY KEY,
    cluster_id TEXT,
    type TEXT,
    status TEXT,
    size INTEGER,
    storage_path TEXT,
    retention_days INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

-- 审计日志
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    cluster_id TEXT,
    action TEXT,
    resource_type TEXT,
    resource_name TEXT,
    request TEXT,
    response TEXT,
    status TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);

-- 创建索引
CREATE INDEX idx_clusters_status ON clusters(status);
CREATE INDEX idx_nodes_cluster ON nodes(cluster_id);
CREATE INDEX idx_conversations_user ON conversations(user_id);
CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_audit_logs_time ON audit_logs(created_at);
```

### 5.2 文件系统结构

```
ai-k8s-ops/
├─ data/
│  ├─ ai-k8s-ops.db
│  ├─ backups/
│  ├─ configs/
│  ├─ logs/
│  ├─ templates/
│  ├─ knowledge/
│  ├─ embeddings/
│  └─ cache/
```

---

## 六、API 接口设计

### 6.1 RESTful API 结构

```
/api/v1/
├─ auth/
│  ├─ POST /login
│  ├─ POST /logout
│  ├─ GET /profile
│
├─ clusters/
│  ├─ GET /
│  ├─ POST /
│  ├─ GET /{id}
│  ├─ PUT /{id}
│  ├─ DELETE /{id}
│  ├─ POST /import
│  ├─ GET /{id}/nodes
│  ├─ GET /{id}/status
│  ├─ POST /{id}/backup
│
├─ deploy/
│  ├─ GET /templates
│  ├─ POST /templates
│  ├─ GET /tasks
│  ├─ GET /tasks/{id}/logs
│
├─ ai/
│  ├─ POST /chat
│  ├─ GET /conversations
│  ├─ POST /execute
│
├─ monitoring/
│  ├─ GET /metrics
│  ├─ GET /alerts
│  ├─ POST /rules
│
├─ remediation/
│  ├─ GET /history
│  ├─ POST /{id}/approve
│
├─ knowledge/
│  ├─ GET /entries
│  ├─ POST /entries
│
├─ audit/
│  ├─ GET /logs
│
└─ system/
   ├─ GET /health
   ├─ GET /users
```

### 6.2 WebSocket 接口

```
/ws/
├─ /deploy/{task_id}   # 部署实时日志
├─ /alerts             # 告警实时推送
├─ /chat               # AI 实时对话
```

---

## 七、前端 UI 设计

### 7.1 页面结构

- 登录页
- Dashboard (仪表盘)
- Clusters (集群管理)
- Deploy (部署中心)
- AI Assistant (AI助手)
- Monitoring (监控中心)
- Alerts (告警中心)
- Knowledge (知识库)
- Audit (审计日志)
- Settings (系统设置)

### 7.2 UI 组件库

- Ant Design Pro (企业级 UI)
- ECharts (图表)
- XTerm.js (终端)
- React Query (数据管理)

---

## 八、安全设计

### 8.1 安全架构层次

1. 网络安全层: HTTPS、TLS 1.3
2. 应用安全层: JWT 认证、RBAC、API 限流
3. 数据安全层: 敏感数据加密 (AES-256-GCM)
4. 操作安全层: AI 操作分级 (L1-L4)
5. 集群安全层: K8S RBAC 集成

### 8.2 认证与授权

```yaml
认证: JWT Token (RS256)
有效期: 
  - Access Token: 24小时
  - Refresh Token: 7天
  
授权:
  Admin: 所有权限
  Operator: 集群管理、AI L1-L2、部署、告警
  Viewer: 查看权限、AI L1
```

### 8.3 审计日志

- 用户登录/登出
- 集群创建/删除/导入
- 部署任务
- AI 操作执行
- 告警处理
- 配置变更

---

## 九、Agent 架构设计

### 9.1 Agent 系统架构

```
管理中心 (Manager)
├─ Web UI
├─ API Server
├─ AI Engine
├─ SQLite 数据库
└─ 监控数据汇聚

        ▼ gRPC/HTTPS

Agent 层（部署在每个集群）
├─ 部署执行器
├─ 监控采集器
├─ 日志收集器
├─ 命令执行器
└─ 状态同步器

        ▼

K8S 集群
├─ Master Nodes
├─ Worker Nodes
└─ Monitoring Stack
```

### 9.2 Agent 通信协议

- gRPC 服务
- 心跳管理 (30秒间隔)
- 部署任务流式推送
- 命令执行
- 监控数据上传

### 9.3 Agent 部署方式

1. 新集群部署: Agent 安装包 + 安装脚本
2. 导入现有集群: DaemonSet 部署
3. 离线部署: 离线安装包

---

## 十、项目实施计划

### 10.1 开发阶段

```
Phase 1 - MVP 基础版本（8-10周）
├─ 项目初始化 (Week 1-2)
├─ 后端 API 基础 (Week 3-4)
├─ 前端 UI 基础 (Week 5-6)
├─ 部署功能基础 (Week 7-8)
└─ AI 交互基础 + 监控 (Week 9-10)

Phase 2 - 功能完善版本（6-8周）
├─ 部署功能完善 (Week 11-12)
├─ AI 能力增强 (Week 13-14)
├─ 告警系统 (Week 15-16)
└─ 权限与安全 (Week 17-18)

Phase 3 - 生产级版本（4-6周）
├─ AI 修复能力 (Week 19-20)
├─ 高级部署功能 (Week 21-22)
└─ 安全加固与性能优化 (Week 23-24)

Phase 4 - 运维增强版本（持续迭代）
```

### 10.2 发布计划

- v0.1.0 (MVP): 第10周
- v0.5.0 (功能完善): 第18周
- v1.0.0 (生产级): 第24周
- v1.x.x: 持续迭代

---

## 十一、附加功能需求

### 11.1 审计日志

- 全操作审计记录
- 支持查询、导出
- 保留策略: L1-L2 1年, L3-L4 3年

### 11.2 国际化

- 支持中英文界面
- react-i18next

### 11.3 集群状态备份与回滚

- etcd 备份
- 配置备份
- 一键回滚

### 11.4 开放 API 接口

- RESTful API
- WebSocket 实时推送
- API 文档

---

## 十二、技术栈实现细节

### 12.1 后端项目结构

```
cmd/
  ├─ server/
  ├─ agent/
  └─ cli/

internal/
  ├─ api/
  ├─ auth/
  ├─ cluster/
  ├─ deploy/
  ├─ ai/
  ├─ monitor/
  ├─ alert/
  ├─ remediation/
  ├─ knowledge/
  ├─ audit/
  ├─ storage/
  ├─ grpc/
  ├─ k8s/
  ├─ llm/
  └─ utils/

pkg/
  ├─ config/
  ├─ logger/
  ├─ crypto/
  └─ version/

configs/
scripts/
data/
tests/
```

### 12.2 前端项目结构

```
src/
  ├─ components/
  ├─ pages/
  ├─ services/
  ├─ stores/
  ├─ hooks/
  ├─ utils/
  ├─ constants/
  ├─ types/
  ├─ i18n/
  ├─ App.tsx
  └─ main.tsx
```

---

## 十三、测试策略

### 13.1 测试类型

- 单元测试: Go testing + Jest, 覆盖率 > 70%
- 集成测试: API、数据库、K8S 交互
- 性能测试: API 响应、并发、大数据量
- 安全测试: 权限边界、加密验证
- 用户验收测试: 功能完整性、流程测试

### 13.2 测试环境

- 开发环境: 本地 + Mock K8S
- 测试环境: 单节点 K8S
- 预生产环境: 多节点 K8S
- 生产环境: 生产级 K8S

---

## 十四、总结

AI-K8S-OPS 是一个面向中小团队的 AI 驱动 K8S 运维平台，采用模块化单体架构、轻量化存储方案，具备快速部署、智能交互、监控告警、主动修复等核心能力。系统设计注重安全性、可扩展性和用户体验，预计 24 周完成生产级版本开发。

---

**文档状态**: 已确认
**下一步**: 创建详细实施计划