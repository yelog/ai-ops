import { useState, useEffect, useRef } from 'react'
import { Layout, Input, Button, List, Card, message, Empty, Spin } from 'antd'
import { SendOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { aiService } from '@/services/ai.service'
import type { Conversation, Message } from '@/types/ai'
import './styles.css'

const { Sider, Content } = Layout

export default function AIPage() {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [currentConversation, setCurrentConversation] = useState<Conversation | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [inputValue, setInputValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    loadConversations()
  }, [])

  useEffect(() => {
    if (currentConversation) {
      loadMessages(currentConversation.id)
    }
  }, [currentConversation])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const loadConversations = async () => {
    setLoading(true)
    try {
      const data = await aiService.listConversations()
      setConversations(data)
      if (data.length > 0 && !currentConversation) {
        setCurrentConversation(data[0])
      }
    } catch (error) {
      message.error('加载对话列表失败')
    } finally {
      setLoading(false)
    }
  }

  const loadMessages = async (conversationId: string) => {
    try {
      const data = await aiService.getConversation(conversationId)
      setMessages(data.messages || [])
    } catch (error) {
      message.error('加载消息失败')
    }
  }

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const handleNewConversation = async () => {
    try {
      const conv = await aiService.createConversation({ title: '新对话' })
      setConversations([conv, ...conversations])
      setCurrentConversation(conv)
      setMessages([])
    } catch (error) {
      message.error('创建对话失败')
    }
  }

  const handleDeleteConversation = async (id: string) => {
    try {
      await aiService.deleteConversation(id)
      setConversations(conversations.filter(c => c.id !== id))
      if (currentConversation?.id === id) {
        setCurrentConversation(conversations[0] || null)
      }
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSend = async () => {
    if (!inputValue.trim() || !currentConversation) return

    const userMessage = inputValue.trim()
    setInputValue('')
    setSending(true)

    const tempUserMsg: Message = {
      id: 'temp-user',
      conversation_id: currentConversation.id,
      role: 'user',
      content: userMessage,
      tokens: 0,
      metadata: '',
      created_at: new Date().toISOString(),
    }
    setMessages([...messages, tempUserMsg])

    try {
      const response = await aiService.chat({
        conversation_id: currentConversation.id,
        message: userMessage,
      })
      setMessages(prev => [...prev, response])
    } catch (error) {
      message.error('发送失败')
      setMessages(prev => prev.filter(m => m.id !== 'temp-user'))
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="ai-page">
      <Layout className="ai-layout">
        <Sider width={280} className="sider">
          <div className="sider-header">
            <Button type="primary" icon={<PlusOutlined />} onClick={handleNewConversation} block>
              新建对话
            </Button>
          </div>
          <div className="conversation-list">
            <Spin spinning={loading}>
              <List
                dataSource={conversations}
                renderItem={(item) => (
                  <List.Item
                    className={`conversation-item ${currentConversation?.id === item.id ? 'active' : ''}`}
                    onClick={() => setCurrentConversation(item)}
                  >
                    <div className="conversation-title">{item.title || '新对话'}</div>
                    <Button
                      type="text"
                      icon={<DeleteOutlined />}
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteConversation(item.id)
                      }}
                    />
                  </List.Item>
                )}
              />
            </Spin>
          </div>
        </Sider>

        <Content className="content">
          {currentConversation ? (
            <div className="chat-container">
              <div className="messages">
                {messages.length === 0 ? (
                  <Empty description="开始对话吧" style={{ marginTop: 100 }} />
                ) : (
                  messages.map((msg) => (
                    <div key={msg.id} className={`message ${msg.role}`}>
                      <Card size="small" className="message-card">
                        <div className="message-content">{msg.content}</div>
                      </Card>
                    </div>
                  ))
                )}
                <div ref={messagesEndRef} />
              </div>

              <div className="input-area">
                <Input.TextArea
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="输入消息..."
                  autoSize={{ minRows: 1, maxRows: 4 }}
                  onPressEnter={(e) => {
                    if (!e.shiftKey) {
                      e.preventDefault()
                      handleSend()
                    }
                  }}
                />
                <Button type="primary" icon={<SendOutlined />} onClick={handleSend} loading={sending}>
                  发送
                </Button>
              </div>
            </div>
          ) : (
            <Empty description="选择或创建一个对话" style={{ marginTop: 100 }} />
          )}
        </Content>
      </Layout>
    </div>
  )
}