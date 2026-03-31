import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Popconfirm, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { deployService } from '@/services/deploy.service'
import type { DeploymentTemplate } from '@/types/deploy'

const typeColors: Record<string, string> = {
  dev: 'blue',
  test: 'cyan',
  staging: 'purple',
  prod: 'red',
  custom: 'default',
}

export default function TemplateListPage() {
  const [templates, setTemplates] = useState<DeploymentTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchTemplates = async () => {
    setLoading(true)
    try {
      const data = await deployService.listTemplates()
      setTemplates(data)
    } catch (error: unknown) {
      message.error('获取模板列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTemplates()
  }, [])

  const handleDelete = async (id: string) => {
    try {
      await deployService.deleteTemplate(id)
      message.success('删除成功')
      fetchTemplates()
    } catch (error: unknown) {
      message.error('删除失败')
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color={typeColors[type]}>{type}</Tag>,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
    },
    {
      title: '默认',
      dataIndex: 'is_default',
      key: 'is_default',
      render: (isDefault: boolean) => isDefault ? <Tag color="green">是</Tag> : <Tag>否</Tag>,
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
      render: (_: unknown, record: DeploymentTemplate) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/deploy/templates/${record.id}`)}>
            查看
          </Button>
          <Popconfirm
            title="确定删除此模板？"
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
    <Card
      title="部署模板"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchTemplates}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/deploy/templates/create')}>
            创建模板
          </Button>
        </Space>
      }
    >
      <Table
        columns={columns}
        dataSource={templates}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  )
}