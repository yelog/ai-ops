export interface Conversation {
  id: string
  user_id: string
  cluster_id?: string
  title: string
  context: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  tokens: number
  metadata: string
  created_at: string
}

export interface CreateConversationRequest {
  title?: string
  cluster_id?: string
}

export interface ChatRequest {
  conversation_id: string
  message: string
}

export interface ConversationListResponse {
  conversations: Conversation[]
}

export interface ConversationDetailResponse {
  conversation: Conversation
  messages: Message[]
}