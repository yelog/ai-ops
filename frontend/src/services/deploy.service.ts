import api from './api'
import type {
  DeploymentTemplate,
  DeploymentTask,
  CreateTemplateRequest,
  CreateTaskRequest,
  TemplateListResponse,
  TaskListResponse,
} from '@/types/deploy'

export const deployService = {
  async listTemplates(): Promise<DeploymentTemplate[]> {
    const response = await api.get<TemplateListResponse>('/api/v1/deploy/templates')
    return response.data.templates
  },

  async getTemplate(id: string): Promise<DeploymentTemplate> {
    const response = await api.get<DeploymentTemplate>(`/api/v1/deploy/templates/${id}`)
    return response.data
  },

  async createTemplate(data: CreateTemplateRequest): Promise<DeploymentTemplate> {
    const response = await api.post<DeploymentTemplate>('/api/v1/deploy/templates', data)
    return response.data
  },

  async updateTemplate(id: string, data: CreateTemplateRequest): Promise<DeploymentTemplate> {
    const response = await api.put<DeploymentTemplate>(`/api/v1/deploy/templates/${id}`, data)
    return response.data
  },

  async deleteTemplate(id: string): Promise<void> {
    await api.delete(`/api/v1/deploy/templates/${id}`)
  },

  async listTasks(): Promise<DeploymentTask[]> {
    const response = await api.get<TaskListResponse>('/api/v1/deploy/tasks')
    return response.data.tasks
  },

  async getTask(id: string): Promise<DeploymentTask> {
    const response = await api.get<DeploymentTask>(`/api/v1/deploy/tasks/${id}`)
    return response.data
  },

  async createTask(data: CreateTaskRequest): Promise<DeploymentTask> {
    const response = await api.post<DeploymentTask>('/api/v1/deploy/tasks', data)
    return response.data
  },
}