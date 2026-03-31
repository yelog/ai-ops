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
import DeployPage from '@/pages/Deploy'
import TemplateListPage from '@/pages/Deploy/Templates'
import TaskListPage from '@/pages/Deploy/Tasks'

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
              <Route path="deploy" element={<DeployPage />} />
              <Route path="deploy/templates" element={<TemplateListPage />} />
              <Route path="deploy/tasks" element={<TaskListPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  )
}

export default App