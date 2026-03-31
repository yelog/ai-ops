import { useState, useEffect } from 'react'
import {
  Card, Tabs, Table, Button, Tag, Space, Progress, Modal,
  Form, Input, Checkbox, Upload, Collapse, message, Popconfirm,
} from 'antd'
import {
  ExportOutlined, ImportOutlined, ReloadOutlined,
  DownloadOutlined, DeleteOutlined, InboxOutlined,
} from '@ant-design/icons'
import { offlineService } from '@/services/offline.service'
import type { OfflinePackage, ResourceManifest } from '@/types/offline'

const { TabPane } = Tabs
const { Dragger } = Upload

const statusConfig: Record<string, { color: string; text: string }> = {
  pending: { color: 'default', text: '等待中' },
  exporting: { color: 'processing', text: '导出中' },
  ready: { color: 'success', text: '就绪' },
  failed: { color: 'error', text: '失败' },
}

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return bytes + ' B'
}

export default function OfflinePage() {
  const [packages, setPackages] = useState<OfflinePackage[]>([])
  const [manifest, setManifest] = useState<ResourceManifest | null>(null)
  const [loading, setLoading] = useState(false)
  const [exportModalOpen, setExportModalOpen] = useState(false)
  const [importModalOpen, setImportModalOpen] = useState(false)
  const [exportForm] = Form.useForm()

  const fetchPackages = async () => {
    setLoading(true)
    try {
      const data = await offlineService.listPackages()
      setPackages(data || [])
    } catch {
      message.error('获取离线包列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchManifest = async () => {
    try {
      const data = await offlineService.getResources()
      setManifest(data)
    } catch {
      message.error('获取资源清单失败')
    }
  }

  useEffect(() => {
    fetchPackages()
    fetchManifest()
    const interval = setInterval(fetchPackages, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleExport = async (values: any) => {
    try {
      await offlineService.exportPackage({
        name: values.name,
        os_list: values.os_list,
        modules: values.modules,
      })
      message.success('导出任务已创建')
      setExportModalOpen(false)
      exportForm.resetFields()
      fetchPackages()
    } catch {
      message.error('创建导出任务失败')
    }
  }

  const handleImport = async (file: File) => {
    try {
      await offlineService.importPackage(file)
      message.success('导入成功')
      setImportModalOpen(false)
      fetchPackages()
    } catch {
      message.error('导入失败')
    }
    return false
  }

  const handleDelete = async (id: string) => {
    try {
      await offlineService.deletePackage(id)
      message.success('删除成功')
      fetchPackages()
    } catch {
      message.error('删除失败')
    }
  }

  const handleDownload = (id: string) => {
    const token = localStorage.getItem('token')
    const url = offlineService.getDownloadUrl(id)
    const a = document.createElement('a')
    a.href = `${url}?token=${token}`
    a.click()
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
    {
      title: '模块',
      dataIndex: 'modules',
      key: 'modules',
      render: (val: string) => {
        try {
          const modules: string[] = JSON.parse(val)
          return modules.map(m => <Tag key={m}>{m}</Tag>)
        } catch {
          return val
        }
      },
    },
    {
      title: 'OS',
      dataIndex: 'os_list',
      key: 'os_list',
      render: (val: string) => {
        try {
          const osList: string[] = JSON.parse(val)
          return osList.map(os => <Tag key={os} color="blue">{os}</Tag>)
        } catch {
          return val
        }
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        const config = statusConfig[status] || { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (size: number) => size > 0 ? formatSize(size) : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: OfflinePackage) => (
        <Space>
          {record.status === 'ready' && (
            <Button
              type="link"
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(record.id)}
            >
              下载
            </Button>
          )}
          <Popconfirm
            title="确定删除此离线包？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const moduleOptions = manifest?.modules.map(m => ({
    label: `${m.name} - ${m.description} (~${m.estimated_size})`,
    value: m.name,
    disabled: m.required,
  })) || []

  return (
    <Card
      title="离线管理"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchPackages}>
            刷新
          </Button>
          <Button icon={<ImportOutlined />} onClick={() => setImportModalOpen(true)}>
            导入离线包
          </Button>
          <Button type="primary" icon={<ExportOutlined />} onClick={() => setExportModalOpen(true)}>
            导出离线包
          </Button>
        </Space>
      }
    >
      <Tabs defaultActiveKey="packages">
        <TabPane tab="离线包管理" key="packages">
          <Table
            columns={columns}
            dataSource={packages}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10 }}
            expandable={{
              expandedRowRender: (record: OfflinePackage) => (
                <div>
                  {record.status === 'exporting' && (
                    <Progress percent={30} status="active" style={{ maxWidth: 400 }} />
                  )}
                  {record.error_message && (
                    <p style={{ color: '#ff4d4f' }}>错误: {record.error_message}</p>
                  )}
                  {record.checksum && <p>校验和: <code>{record.checksum}</code></p>}
                </div>
              ),
            }}
          />
        </TabPane>

        <TabPane tab="资源清单" key="resources">
          {manifest && (
            <div>
              <p style={{ marginBottom: 16 }}>
                <strong>K8S 版本:</strong> {manifest.version}
              </p>
              <Collapse>
                {manifest.modules.map(mod => (
                  <Collapse.Panel
                    key={mod.name}
                    header={
                      <span>
                        <Tag color={mod.required ? 'red' : 'blue'}>
                          {mod.required ? '必选' : '可选'}
                        </Tag>
                        {mod.name} - {mod.description} ({mod.estimated_size})
                      </span>
                    }
                  >
                    <h4>容器镜像</h4>
                    <Table
                      size="small"
                      pagination={false}
                      dataSource={mod.images.map((img, i) => ({ key: i, type: '镜像', name: img }))}
                      columns={[
                        { title: '类型', dataIndex: 'type', width: 80 },
                        { title: '名称', dataIndex: 'name' },
                      ]}
                    />
                    {mod.binaries && mod.binaries.length > 0 && (
                      <>
                        <h4 style={{ marginTop: 16 }}>二进制文件</h4>
                        <Table
                          size="small"
                          pagination={false}
                          dataSource={mod.binaries.map((b, i) => ({ key: i, type: '二进制', name: b }))}
                          columns={[
                            { title: '类型', dataIndex: 'type', width: 80 },
                            { title: '名称', dataIndex: 'name' },
                          ]}
                        />
                      </>
                    )}
                  </Collapse.Panel>
                ))}
              </Collapse>
            </div>
          )}
        </TabPane>
      </Tabs>

      {/* Export Modal */}
      <Modal
        title="导出离线包"
        open={exportModalOpen}
        onCancel={() => setExportModalOpen(false)}
        footer={null}
      >
        <Form
          form={exportForm}
          layout="vertical"
          onFinish={handleExport}
          initialValues={{
            os_list: ['ubuntu', 'centos'],
            modules: ['core', 'network'],
          }}
        >
          <Form.Item
            name="name"
            label="包名称"
            rules={[{ required: true, message: '请输入包名称' }]}
          >
            <Input placeholder="例: prod-offline-pkg" />
          </Form.Item>

          <Form.Item name="os_list" label="目标 OS">
            <Checkbox.Group>
              <Checkbox value="ubuntu">Ubuntu 20.04/22.04</Checkbox>
              <Checkbox value="centos">CentOS 7/8</Checkbox>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item name="modules" label="选择模块">
            <Checkbox.Group options={moduleOptions} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setExportModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">开始导出</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Import Modal */}
      <Modal
        title="导入离线包"
        open={importModalOpen}
        onCancel={() => setImportModalOpen(false)}
        footer={null}
      >
        <Dragger
          accept=".tar.gz,.tgz"
          maxCount={1}
          beforeUpload={(file) => {
            handleImport(file)
            return false
          }}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">将 .tar.gz 文件拖到此处</p>
          <p className="ant-upload-hint">或点击选择文件</p>
        </Dragger>
      </Modal>
    </Card>
  )
}
