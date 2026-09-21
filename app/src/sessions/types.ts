export type SessionStatus = 'draft' | 'active' | 'completed'

export type SessionMediaType = 'image' | 'audio'

export interface SessionMedia {
  id: string
  filename: string
  contentType: string
  size: number
  type: SessionMediaType
  url: string
  transcription: string | null
}

export interface Session {
  id: string
  title: string
  description: string
  status: SessionStatus
  content: string
  images: SessionMedia[]
  audio: SessionMedia[]
  createdAt: string
  updatedAt: string
}

export interface SessionInput {
  title: string
  description?: string
  status?: SessionStatus
  content?: string
  images?: File[]
  audio?: File[]
}

export type SessionUpdateInput = Partial<SessionInput>
