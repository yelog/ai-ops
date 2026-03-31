export interface DeploymentTemplate {
  id: string
  name: string
  description: string
  type: 'dev' | 'test' | 'staging' | 'prod' | 'custom'
  provider: 'bare-metal' | 'vm' | 'cloud'
  config: string
  components: string
  is_default: boolean
  created_at: string
}

export interface DeploymentTask {
  id: string
  cluster_id: string
  template_id: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'rollback'
  current_step: string
  progress: number
  error_message?: string
  created_by: string
  started_at?: string
  finished_at?: string
}

export interface CreateTemplateRequest {
  name: string
  description?: string
  type: string
  provider: string
  config?: string
  components?: string
  is_default?: boolean
}

export interface CreateTaskRequest {
  cluster_id: string
  template_id: string
}

export interface TemplateListResponse {
  templates: DeploymentTemplate[]
}

export interface TaskListResponse {
  tasks: DeploymentTask[]
}