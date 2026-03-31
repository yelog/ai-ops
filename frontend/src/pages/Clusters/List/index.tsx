import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Popconfirm, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { clusterService } from '@/services/cluster.service'
import type { Cluster } from '@/types/cluster'
import './styles.css'

const statusColors: Record<string, string> = {
  healthy: 'success',
  warning: 'warning',
  critical: 'error',
  pending: 'default',
}

const environmentColors: Record<string, string> = {
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
      render: (env: string) => <Tag color={environmentColors[env]}>{env}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={statusColors[status]}>{status}</Tag>,
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
      render: (_: unknown, record: Cluster) => (
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