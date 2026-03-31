import api from './api'
import type {
  Conversation,
  Message,
  CreateConversationRequest,
  ChatRequest,
  ConversationListResponse,
  ConversationDetailResponse,
} from '@/types/ai'

export const aiService = {
  async listConversations(): Promise<Conversation[]> {
    const response = await api.get<ConversationListResponse>('/api/v1/ai/conversations')
    return response.data.conversations
  },

  async createConversation(data: CreateConversationRequest): Promise<Conversation> {
    const response = await api.post<Conversation>('/api/v1/ai/conversations', data)
    return response.data
  },

  async getConversation(id: string): Promise<ConversationDetailResponse> {
    const response = await api.get<ConversationDetailResponse>(`/api/v1/ai/conversations/${id}`)
    return response.data
  },

  async deleteConversation(id: string): Promise<void> {
    await api.delete(`/api/v1/ai/conversations/${id}`)
  },

  async chat(data: ChatRequest): Promise<Message> {
    const response = await api.post<{ message: Message }>('/api/v1/ai/chat', data)
    return response.data.message
  },
}