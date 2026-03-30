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