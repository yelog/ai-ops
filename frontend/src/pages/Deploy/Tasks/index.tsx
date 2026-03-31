import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Progress, message } from 'antd'
import { ReloadOutlined, EyeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { deployService } from '@/services/deploy.service'
import type { DeploymentTask } from '@/types/deploy'

const statusColors: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  success: 'success',
  failed: 'error',
  rollback: 'warning',
}

const statusTexts: Record<string, string> = {
  pending: '待执行',
  running: '执行中',
  success: '成功',
  failed: '失败',
  rollback: '已回滚',
}

export default function TaskListPage() {
  const [tasks, setTasks] = useState<DeploymentTask[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchTasks = async () => {
    setLoading(true)
    try {
      const data = await deployService.listTasks()
      setTasks(data)
    } catch (error: unknown) {
      message.error('获取任务列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTasks()
    const interval = setInterval(fetchTasks, 5000)
    return () => clearInterval(interval)
  }, [])

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'id',
      key: 'id',
      width: 280,
      render: (id: string) => <span style={{ fontFamily: 'monospace' }}>{id.substring(0, 8)}...</span>,
    },
    {
      title: '集群ID',
      dataIndex: 'cluster_id',
      key: 'cluster_id',
      width: 280,
      render: (id: string) => <span style={{ fontFamily: 'monospace' }}>{id.substring(0, 8)}...</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={statusColors[status]}>{statusTexts[status]}</Tag>,
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      width: 200,
      render: (progress: number, record: DeploymentTask) => (
        <Progress percent={progress} size="small" status={record.status === 'failed' ? 'exception' : 'active'} />
      ),
    },
    {
      title: '当前步骤',
      dataIndex: 'current_step',
      key: 'current_step',
      render: (step: string) => step || '-',
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: DeploymentTask) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/deploy/tasks/${record.id}`)}>
            详情
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="部署任务"
      extra={
        <Button icon={<ReloadOutlined />} onClick={fetchTasks}>
          刷新
        </Button>
      }
    >
      <Table
        columns={columns}
        dataSource={tasks}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  )
}