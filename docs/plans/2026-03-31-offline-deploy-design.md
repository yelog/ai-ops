# 离线部署功能设计文档

## 文档信息

- **创建日期**: 2026-03-31
- **功能**: 完全离线环境下的 K8S 集群部署支持
- **K8S 版本**: v1.28.0（固定版本）
- **目标 OS**: Ubuntu 20.04/22.04, CentOS 7/8

---

## 一、需求概述

### 场景

完全离线环境（无任何外网访问），需提前在联网环境打包所有资源，传输到离线环境后部署 K8S 集群。

### 核心能力

1. **模块化离线包**：分为 core、network、monitoring、logging、tracing 五个模块，按需组合
2. **双入口**：CLI 命令行（自动化/CI）+ Web UI（手动操作）
3. **与现有部署流程集成**：部署任务支持选择"离线模式"

---

## 二、整体架构

```
离线管理模块架构

┌─────────────────────────────────────────────────┐
│                  用户入口                         │
│  ┌──────────────┐    ┌───────────────────────┐  │
│  │  Web UI      │    │  CLI                  │  │
│  │  离线管理页面 │    │  offline export/import │  │
│  └──────┬───────┘    └──────────┬────────────┘  │
└─────────┼───────────────────────┼───────────────┘
          ▼                       ▼
┌─────────────────────────────────────────────────┐
│              API 层 (/api/v1/offline)            │
│  - POST /packages/export   导出离线包            │
│  - POST /packages/import   导入离线包            │
│  - GET  /packages          离线包列表            │
│  - GET  /packages/:id      离线包详情            │
│  - DELETE /packages/:id    删除离线包            │
│  - GET  /resources         可用资源清单          │
└─────────────────┬───────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────┐
│           核心服务 (internal/offline)             │
│  ┌─────────────┐  ┌────────────┐  ┌──────────┐ │
│  │ Exporter    │  │ Importer   │  │ Registry │ │
│  │ 联网环境打包 │  │ 离线环境导入│  │ 包管理   │ │
│  └─────────────┘  └────────────┘  └──────────┘ │
└─────────────────┬───────────────────────────────┘
                  ▼
┌─────────────────────────────────────────────────┐
│              资源层                               │
│  data/offline/                                   │
│  ├── packages/       # 生成的离线包               │
│  ├── images/         # 容器镜像 tar               │
│  ├── binaries/       # 二进制文件                 │
│  └── rpms|debs/      # 系统包                    │
└─────────────────────────────────────────────────┘
```

---

## 三、离线包结构（模块化）

```
ai-k8s-ops-offline-v1.28.0/
├── manifest.yaml              # 包清单（版本、OS、组件列表、校验和）
├── core/                      # 核心包（必选）
│   ├── images/                # containerd, pause, coredns, etcd, kube-*
│   ├── binaries/              # kubeadm, kubelet, kubectl, crictl, containerd
│   └── packages/              # 系统依赖 deb/rpm
│       ├── ubuntu/
│       └── centos/
├── network/                   # 网络插件包（必选）
│   └── images/                # calico 相关镜像
├── monitoring/                # 监控包（可选）
│   └── images/                # prometheus, grafana
├── logging/                   # 日志包（可选）
│   └── images/                # loki
├── tracing/                   # 追踪包（可选）
│   └── images/                # jaeger
└── scripts/                   # 安装脚本
    ├── install.sh             # 主安装入口
    ├── load-images.sh         # 镜像加载
    └── setup-deps.sh          # 依赖安装
```

### 模块定义

| 模块 | 内容 | 必选 | 预估大小 |
|------|------|------|---------|
| core | kubeadm, kubelet, kubectl, crictl, containerd, 基础镜像 | 是 | ~2.1GB |
| network | Calico 网络插件镜像 | 是 | ~320MB |
| monitoring | Prometheus + Grafana 镜像 | 否 | ~800MB |
| logging | Loki 镜像 | 否 | ~200MB |
| tracing | Jaeger 镜像 | 否 | ~150MB |

---

## 四、数据模型

```sql
CREATE TABLE offline_packages (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    os_list TEXT NOT NULL,              -- JSON ["ubuntu","centos"]
    modules TEXT NOT NULL,              -- JSON ["core","network","monitoring"]
    status TEXT NOT NULL,               -- pending, exporting, ready, failed
    size INTEGER DEFAULT 0,
    checksum TEXT,
    storage_path TEXT,
    error_message TEXT,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX idx_offline_packages_status ON offline_packages(status);
```

---

## 五、API 设计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/offline/resources` | 获取可用资源清单 |
| POST | `/api/v1/offline/packages/export` | 创建导出任务 |
| POST | `/api/v1/offline/packages/import` | 上传并导入离线包 |
| GET | `/api/v1/offline/packages` | 离线包列表 |
| GET | `/api/v1/offline/packages/:id` | 离线包详情 |
| DELETE | `/api/v1/offline/packages/:id` | 删除离线包 |
| GET | `/api/v1/offline/packages/:id/download` | 下载离线包 |

### 导出请求

```json
{
  "name": "prod-offline-pkg",
  "os_list": ["ubuntu", "centos"],
  "modules": ["core", "network", "monitoring"]
}
```

### 资源清单响应

```json
{
  "version": "v1.28.0",
  "modules": [
    {
      "name": "core",
      "required": true,
      "description": "K8S 核心组件",
      "images": ["registry.k8s.io/kube-apiserver:v1.28.0"],
      "binaries": ["kubeadm", "kubelet", "kubectl", "containerd", "crictl"],
      "estimated_size": "2.1GB"
    }
  ]
}
```

---

## 六、CLI 设计

```bash
# 导出离线包
ai-k8s-ops offline export \
  --name prod-offline-pkg \
  --os ubuntu,centos \
  --modules core,network,monitoring \
  --output ./output/

# 导入离线包
ai-k8s-ops offline import \
  --file ./ai-k8s-ops-offline-v1.28.0.tar.gz

# 查看离线包
ai-k8s-ops offline list
ai-k8s-ops offline inspect --file ./ai-k8s-ops-offline-v1.28.0.tar.gz
```

---

## 七、前端 UI 设计

### 页面位置

左侧菜单"部署中心" > "离线管理"，包含两个 Tab：

### Tab 1：离线包管理

- 离线包列表（名称、版本、模块、OS、状态、大小、操作）
- 导出中的包显示进度条
- 操作：下载、删除
- 顶部按钮：导出离线包、导入离线包、刷新

### 导出离线包弹窗

- 包名称输入框
- 目标 OS 复选框（Ubuntu、CentOS）
- 模块选择列表（core/network 必选且禁用取消，其他可选）
- 显示预估总大小
- 开始导出按钮

### 导入离线包弹窗

- 拖拽上传区域，支持 .tar.gz 文件
- 上传进度条
- 文件信息展示

### Tab 2：资源清单

- 按模块折叠展示
- 每个模块显示镜像列表、二进制列表、系统包列表及大小

---

## 八、与部署流程集成

部署任务创建时新增：

- 部署模式选择：在线部署 / 离线部署
- 选择"离线部署"时，出现离线包下拉选择框

### 离线部署执行流程

1. 校验离线包完整性（SHA256）
2. 传输离线包到目标节点（SCP/SFTP）
3. 解压离线包
4. 安装系统依赖（dpkg/rpm）
5. 加载容器镜像（ctr image import）
6. 安装二进制文件
7. 执行 kubeadm init/join
8. 安装网络插件
9. 安装可选组件
10. 集群验证

---

## 九、错误处理

| 场景 | 处理方式 |
|------|---------|
| 导出时网络中断 | 支持断点续传，记录已下载资源，重试继续 |
| 离线包校验失败 | 拒绝导入，提示用户重新传输 |
| 磁盘空间不足 | 导出/导入前预检查，提前告警 |
| 目标节点 OS 不匹配 | 导入时检测 OS，与包支持列表比对 |
| 镜像加载失败 | 逐个加载，失败记录日志，支持重试 |
| SSH 连接失败 | 部署前连通性检测，失败提示 |

---

## 十、安全考虑

- 离线包使用 SHA256 校验完整性
- 传输过程使用 SCP/SFTP（SSH 加密通道）
- 离线包中不包含敏感信息
- 导入操作需要 admin/operator 权限
