import { Card, Row, Col } from 'antd'
import { FileTextOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

export default function DeployPage() {
  const navigate = useNavigate()

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>部署中心</h2>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card hoverable onClick={() => navigate('/deploy/templates')}>
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <FileTextOutlined style={{ fontSize: 48, color: '#1890ff' }} />
              <h3 style={{ marginTop: 16 }}>部署模板</h3>
              <p style={{ color: '#666' }}>管理和配置部署模板</p>
            </div>
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card hoverable onClick={() => navigate('/deploy/tasks')}>
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <PlayCircleOutlined style={{ fontSize: 48, color: '#52c41a' }} />
              <h3 style={{ marginTop: 16 }}>部署任务</h3>
              <p style={{ color: '#666' }}>查看和管理部署任务</p>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}