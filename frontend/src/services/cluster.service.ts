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