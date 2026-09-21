import type { Session, SessionInput, SessionUpdateInput } from './types'

const request = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init)
  const body = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(body?.error ?? 'Request failed')
  }
  return body as T
}

export const buildSessionFormData = (input: SessionInput | SessionUpdateInput) => {
  const form = new FormData()
  if (input.title !== undefined) form.set('title', input.title)
  if (input.description !== undefined) form.set('description', input.description)
  if (input.status !== undefined) form.set('status', input.status)
  if (input.content !== undefined) form.set('content', input.content)
  input.images?.forEach((file) => form.append('images', file))
  input.audio?.forEach((file) => form.append('audio', file))
  return form
}

export const listSessions = () => request<Session[]>('/api/sessions')

export const getSession = (id: string) => request<Session>(`/api/sessions/${id}`)

export const createSession = (input: SessionInput) =>
  request<Session>('/api/sessions', { method: 'POST', body: buildSessionFormData(input) })

export const updateSession = (id: string, input: SessionUpdateInput) =>
  request<Session>(`/api/sessions/${id}`, { method: 'PATCH', body: buildSessionFormData(input) })

export const deleteSession = (id: string) =>
  request<void>(`/api/sessions/${id}`, { method: 'DELETE' })

export const deleteSessionMedia = (sessionId: string, mediaId: string) =>
  request<void>(`/api/sessions/${sessionId}/media/${mediaId}`, { method: 'DELETE' })
