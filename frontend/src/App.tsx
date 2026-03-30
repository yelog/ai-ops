import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <div style={{ padding: '50px', textAlign: 'center' }}>
        <h1>AI-K8S-OPS</h1>
        <p>AI-driven Kubernetes Operations Platform</p>
      </div>
    </ConfigProvider>
  )
}

export default App