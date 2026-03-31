import api from './api'
import type {
  OfflinePackage,
  ResourceManifest,
  ExportRequest,
  PackageListResponse,
} from '@/types/offline'

export const offlineService = {
  async getResources(): Promise<ResourceManifest> {
    const response = await api.get<ResourceManifest>('/api/v1/offline/resources')
    return response.data
  },

  async exportPackage(data: ExportRequest): Promise<OfflinePackage> {
    const response = await api.post<OfflinePackage>('/api/v1/offline/packages/export', data)
    return response.data
  },

  async importPackage(file: File): Promise<OfflinePackage> {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post<OfflinePackage>('/api/v1/offline/packages/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data
  },

  async listPackages(): Promise<OfflinePackage[]> {
    const response = await api.get<PackageListResponse>('/api/v1/offline/packages')
    return response.data.packages
  },

  async getPackage(id: string): Promise<OfflinePackage> {
    const response = await api.get<OfflinePackage>(`/api/v1/offline/packages/${id}`)
    return response.data
  },

  async deletePackage(id: string): Promise<void> {
    await api.delete(`/api/v1/offline/packages/${id}`)
  },

  getDownloadUrl(id: string): string {
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    return `${baseUrl}/api/v1/offline/packages/${id}/download`
  },
}
