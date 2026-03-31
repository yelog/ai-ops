---
render_with_liquid: false
---

# Phase 1 Week 5-6: Frontend UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement frontend UI with authentication, main layout, and cluster management pages.

**Architecture:** React 18 + TypeScript + Ant Design Pro. Zustand for state management. React Query for API integration. React Router for routing.

**Tech Stack:**
- React 18 + TypeScript
- Ant Design Pro
- React Router DOM v6
- Zustand (state management)
- React Query (API data fetching)
- Axios (HTTP client)

---

## Task 1: Install Frontend Dependencies and Setup Project Structure

**Files:**
- Modify: `frontend/package.json`
- Create: `frontend/src/types/`
- Create: `frontend/src/services/`
- Create: `frontend/src/stores/`
- Create: `frontend/src/utils/`

**Step 1: Install additional dependencies**

```bash
cd frontend
npm install zustand @tanstack/react-query axios
```

**Step 2: Create directory structure**

```bash
cd frontend/src
mkdir -p types services stores utils constants hooks pages components/layouts
```

**Step 3: Create TypeScript types**

Create file: `frontend/src/types/auth.ts`

```typescript
export interface User {
  id: string
  username: string
  email: string
  role: 'admin' | 'operator' | 'viewer'
  created_at?: string
  last_login?: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  role?: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
}
```

Create file: `frontend/src/types/cluster.ts`

```typescript
export interface Cluster {
  id: string
  name: string
  description: string
  environment: 'dev' | 'test' | 'staging' | 'prod'
  provider: 'bare-metal' | 'vm' | 'cloud' | 'imported'
  version: string
  api_server: string
  status: 'healthy' | 'warning' | 'critical' | 'pending'
  created_at: string
  updated_at: string
}

export interface CreateClusterRequest {
  name: string
  description?: string
  environment: string
  provider: string
  version?: string
  api_server?: string
  kubeconfig?: string
}

export interface ClusterListResponse {
  clusters: Cluster[]
}
```

**Step 4: Commit**

```bash
git add frontend/
git commit -m "feat: setup frontend project structure and TypeScript types"
```

---

## Task 2: Create API Service Layer with Axios

**Files:**
- Create: `frontend/src/services/api.ts`
- Create: `frontend/src/services/auth.service.ts`
- Create: `frontend/src/services/cluster.service.ts`

**Step 1: Create API client**

Create file: `frontend/src/services/api.ts`

```typescript
import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
```

**Step 2: Create auth service**

Create file: `frontend/src/services/auth.service.ts`

```typescript
import api from './api'
import type { AuthResponse, LoginRequest, RegisterRequest, User } from '@/types/auth'

export const authService = {
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/api/v1/auth/login', data)
    return response.data
  },

  async register(data: RegisterRequest): Promise<User> {
    const response = await api.post<User>('/api/v1/auth/register', data)
    return response.data
  },

  async getProfile(): Promise<User> {
    const response = await api.get<User>('/api/v1/auth/profile')
    return response.data
  },
}
```

**Step 3: Create cluster service**

Create file: `frontend/src/services/cluster.service.ts`

```typescript
import api from './api'
import type { Cluster, CreateClusterRequest, ClusterListResponse } from '@/types/cluster'

export const clusterService = {
  async list(environment?: string): Promise<Cluster[]> {
    const params = environment ? { environment } : {}
    const response = await api.get<ClusterListResponse>('/api/v1/clusters', { params })
    return response.data.clusters
  },

  async get(id: string): Promise<Cluster> {
    const response = await api.get<Cluster>(`/api/v1/clusters/${id}`)
    return response.data
  },

  async create(data: CreateClusterRequest): Promise<Cluster> {
    const response = await api.post<Cluster>('/api/v1/clusters', data)
    return response.data
  },

  async update(id: string, data: CreateClusterRequest): Promise<Cluster> {
    const response = await api.put<Cluster>(`/api/v1/clusters/${id}`, data)
    return response.data
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/v1/clusters/${id}`)
  },
}
```

**Step 4: Commit**

```bash
git add frontend/src/services frontend/src/types
git commit -m "feat: create API service layer with axios"
```

---

## Task 3: Create Auth Store with Zustand

**Files:**
- Create: `frontend/src/stores/auth.store.ts`

**Step 1: Create auth store**

Create file: `frontend/src/stores/auth.store.ts`

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  setAuth: (token: string, user: User) => void
  logout: () => void
  updateUser: (user: User) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      setAuth: (token, user) => {
        localStorage.setItem('token', token)
        set({ token, user, isAuthenticated: true })
      },
      logout: () => {
        localStorage.removeItem('token')
        set({ token: null, user: null, isAuthenticated: false })
      },
      updateUser: (user) => set({ user }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token }),
    }
  )
)
```

**Step 2: Commit**

```bash
git add frontend/src/stores
git commit -m "feat: create auth store with zustand"
```

---

## Task 4: Create Login Page

**Files:**
- Create: `frontend/src/pages/Login/index.tsx`
- Create: `frontend/src/pages/Login/styles.css`

**Step 1: Create login page**

Create file: `frontend/src/pages/Login/index.tsx`

```tsx
import { useState } from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useNavigate, Link } from 'react-router-dom'
import { authService } from '@/services/auth.service'
import { useAuthStore } from '@/stores/auth.store'
import type { LoginRequest } from '@/types/auth'
import './styles.css'

export default function LoginPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const onFinish = async (values: LoginRequest) => {
    setLoading(true)
    try {
      const response = await authService.login(values)
      setAuth(response.token, response.user)
      message.success('登录成功')
      navigate('/')
    } catch (error: any) {
      message.error(error.response?.data?.error || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <Card className="login-card" title="AI-K8S-OPS">
        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>

          <div className="login-footer">
            还没有账号？ <Link to="/register">立即注册</Link>
          </div>
        </Form>
      </Card>
    </div>
  )
}
```

Create file: `frontend/src/pages/Login/styles.css`

```css
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.login-card .ant-card-head-title {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
}

.login-footer {
  text-align: center;
  margin-top: 16px;
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Login
git commit -m "feat: create login page with form"
```

---

## Task 5: Create Register Page

**Files:**
- Create: `frontend/src/pages/Register/index.tsx`
- Create: `frontend/src/pages/Register/styles.css`

**Step 1: Create register page**

Create file: `frontend/src/pages/Register/index.tsx`

```tsx
import { useState } from 'react'
import { Form, Input, Button, Card, message, Select } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import { useNavigate, Link } from 'react-router-dom'
import { authService } from '@/services/auth.service'
import type { RegisterRequest } from '@/types/auth'
import './styles.css'

export default function RegisterPage() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const onFinish = async (values: RegisterRequest) => {
    setLoading(true)
    try {
      await authService.register(values)
      message.success('注册成功，请登录')
      navigate('/login')
    } catch (error: any) {
      message.error(error.response?.data?.error || '注册失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="register-container">
      <Card className="register-card" title="注册账号">
        <Form
          name="register"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input prefix={<MailOutlined />} placeholder="邮箱" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 8, message: '密码至少8个字符' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="确认密码" />
          </Form.Item>

          <Form.Item
            name="role"
            initialValue="viewer"
          >
            <Select placeholder="选择角色">
              <Select.Option value="viewer">观察者</Select.Option>
              <Select.Option value="operator">运维人员</Select.Option>
              <Select.Option value="admin">管理员</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              注册
            </Button>
          </Form.Item>

          <div className="register-footer">
            已有账号？ <Link to="/login">立即登录</Link>
          </div>
        </Form>
      </Card>
    </div>
  )
}
```

Create file: `frontend/src/pages/Register/styles.css`

```css
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.register-card {
  width: 400px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.register-card .ant-card-head-title {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
}

.register-footer {
  text-align: center;
  margin-top: 16px;
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Register
git commit -m "feat: create register page with form validation"
```

---

## Task 6: Create Main Layout

**Files:**
- Create: `frontend/src/components/layouts/MainLayout.tsx`
- Create: `frontend/src/components/layouts/styles.css`

**Step 1: Create main layout component**

Create file: `frontend/src/components/layouts/MainLayout.tsx`

```tsx
import { useState } from 'react'
import { Layout, Menu, Avatar, Dropdown, Button } from 'antd'
import {
  DashboardOutlined,
  ClusterOutlined,
  CloudUploadOutlined,
  RobotOutlined,
  MonitorOutlined,
  AlertOutlined,
  BookOutlined,
  FileSearchOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth.store'
import './styles.css'

const { Header, Sider, Content } = Layout

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/clusters', icon: <ClusterOutlined />, label: '集群管理' },
  { key: '/deploy', icon: <CloudUploadOutlined />, label: '部署中心' },
  { key: '/ai', icon: <RobotOutlined />, label: 'AI 助手' },
  { key: '/monitoring', icon: <MonitorOutlined />, label: '监控中心' },
  { key: '/alerts', icon: <AlertOutlined />, label: '告警中心' },
  { key: '/knowledge', icon: <BookOutlined />, label: '知识库' },
  { key: '/audit', icon: <FileSearchOutlined />, label: '审计日志' },
]

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
      onClick: () => navigate('/settings'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <Layout className="main-layout">
      <Sider trigger={null} collapsible collapsed={collapsed} className="sider">
        <div className="logo">
          {!collapsed && <span>AI-K8S-OPS</span>}
          {collapsed && <span>K8S</span>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>
      <Layout>
        <Header className="header">
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            className="trigger"
          />
          <div className="header-right">
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div className="user-info">
                <Avatar icon={<UserOutlined />} />
                <span className="username">{user?.username}</span>
              </div>
            </Dropdown>
          </div>
        </Header>
        <Content className="content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
```

Create file: `frontend/src/components/layouts/styles.css`

```css
.main-layout {
  min-height: 100vh;
}

.sider {
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}

.header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.trigger {
  font-size: 18px;
  cursor: pointer;
  transition: color 0.3s;
}

.trigger:hover {
  color: #1890ff;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 12px;
  transition: background 0.3s;
}

.user-info:hover {
  background: rgba(0, 0, 0, 0.025);
}

.username {
  margin-left: 8px;
  font-weight: 500;
}

.content {
  margin: 24px;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  min-height: 280px;
}
```

**Step 2: Commit**

```bash
git add frontend/src/components/layouts
git commit -m "feat: create main layout with sidebar and header"
```

---

## Task 7: Create Dashboard Page

**Files:**
- Create: `frontend/src/pages/Dashboard/index.tsx`

**Step 1: Create dashboard page**

Create file: `frontend/src/pages/Dashboard/index.tsx`

```tsx
import { Card, Row, Col, Statistic } from 'antd'
import { ClusterOutlined, CheckCircleOutlined, WarningOutlined, AlertOutlined } from '@ant-design/icons'

export default function DashboardPage() {
  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>仪表盘</h2>
      
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="集群总数"
              value={5}
              prefix={<ClusterOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="健康集群"
              value={4}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="告警数量"
              value={12}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待处理"
              value={3}
              prefix={<AlertOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="最近活动">
            <p>暂无最近活动</p>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="AI 智能建议">
            <p>暂无建议</p>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Dashboard
git commit -m "feat: create dashboard page with statistics"
```

---

## Task 8: Create Cluster List Page

**Files:**
- Create: `frontend/src/pages/Clusters/List/index.tsx`
- Create: `frontend/src/pages/Clusters/List/styles.css`

**Step 1: Create cluster list page**

Create file: `frontend/src/pages/Clusters/List/index.tsx`

```tsx
import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Popconfirm, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { clusterService } from '@/services/cluster.service'
import type { Cluster } from '@/types/cluster'
import './styles.css'

const statusColors = {
  healthy: 'success',
  warning: 'warning',
  critical: 'error',
  pending: 'default',
}

const environmentColors = {
  dev: 'blue',
  test: 'cyan',
  staging: 'purple',
  prod: 'red',
}

export default function ClusterListPage() {
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchClusters = async () => {
    setLoading(true)
    try {
      const data = await clusterService.list()
      setClusters(data)
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取集群列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchClusters()
  }, [])

  const handleDelete = async (id: string) => {
    try {
      await clusterService.delete(id)
      message.success('删除成功')
      fetchClusters()
    } catch (error: any) {
      message.error(error.response?.data?.error || '删除失败')
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Cluster) => (
        <a onClick={() => navigate(`/clusters/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: '环境',
      dataIndex: 'environment',
      key: 'environment',
      render: (env: string) => <Tag color={environmentColors[env as keyof typeof environmentColors]}>{env}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={statusColors[status as keyof typeof statusColors]}>{status}</Tag>,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Cluster) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/clusters/${record.id}`)}>
            详情
          </Button>
          <Popconfirm
            title="确定删除此集群？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="cluster-list-page">
      <Card
        title="集群管理"
        extra={
          <Space>
            <Button icon={<ReloadOutlined />} onClick={fetchClusters}>
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/clusters/create')}>
              创建集群
            </Button>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={clusters}
          rowKey="id"
          loading={loading}
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  )
}
```

Create file: `frontend/src/pages/Clusters/List/styles.css`

```css
.cluster-list-page {
  /* Add custom styles if needed */
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Clusters
git commit -m "feat: create cluster list page with table"
```

---

## Task 9: Setup React Router and Protected Routes

**Files:**
- Modify: `frontend/src/App.tsx`
- Create: `frontend/src/components/ProtectedRoute.tsx`

**Step 1: Create protected route component**

Create file: `frontend/src/components/ProtectedRoute.tsx`

```tsx
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth.store'

interface ProtectedRouteProps {
  children: React.ReactNode
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
```

**Step 2: Setup routing in App.tsx**

Update file: `frontend/src/App.tsx`

```tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'

import ProtectedRoute from '@/components/ProtectedRoute'
import MainLayout from '@/components/layouts/MainLayout'
import LoginPage from '@/pages/Login'
import RegisterPage from '@/pages/Register'
import DashboardPage from '@/pages/Dashboard'
import ClusterListPage from '@/pages/Clusters/List'

import './index.css'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={zhCN}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <MainLayout />
                </ProtectedRoute>
              }
            >
              <Route index element={<DashboardPage />} />
              <Route path="clusters" element={<ClusterListPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  )
}

export default App
```

**Step 3: Test the application**

```bash
cd frontend
npm run dev
```

Verify:
- Login page loads at /login
- Register page loads at /register
- Protected routes redirect to /login when not authenticated
- After login, dashboard and cluster pages work

**Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/components
git commit -m "feat: setup react router with protected routes"
```

---

## Summary

This plan implements the frontend UI foundation:

✅ Frontend dependencies and project structure
✅ TypeScript types for API data
✅ API service layer with axios
✅ Auth store with zustand
✅ Login page with form validation
✅ Register page with password confirmation
✅ Main layout with sidebar navigation
✅ Dashboard page with statistics
✅ Cluster list page with table
✅ React router with protected routes

**Next Phase**: Week 7-8 will implement deployment functionality and complete the MVP.

---

**Plan complete and saved to `docs/plans/2026-03-30-phase1-frontend.md`**