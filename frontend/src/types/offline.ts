export interface OfflinePackage {
  id: string
  name: string
  version: string
  os_list: string      // JSON string
  modules: string      // JSON string
  status: 'pending' | 'exporting' | 'ready' | 'failed'
  size: number
  checksum: string
  storage_path: string
  error_message?: string
  created_by: string
  created_at: string
}

export interface ModuleInfo {
  name: string
  required: boolean
  description: string
  images: string[]
  binaries?: string[]
  estimated_size: string
}

export interface ResourceManifest {
  version: string
  modules: ModuleInfo[]
}

export interface ExportRequest {
  name: string
  os_list: string[]
  modules: string[]
}

export interface PackageListResponse {
  packages: OfflinePackage[]
}
