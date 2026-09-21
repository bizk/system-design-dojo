import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createSession, deleteSession, deleteSessionMedia, getSession, listSessions, updateSession } from './api'
import type { SessionInput, SessionUpdateInput } from './types'

export const sessionKeys = {
  all: ['sessions'] as const,
  detail: (id: string) => ['sessions', id] as const,
}

export const useSessions = () => useQuery({ queryKey: sessionKeys.all, queryFn: listSessions })

export const useSession = (id: string) =>
  useQuery({ queryKey: sessionKeys.detail(id), queryFn: () => getSession(id), enabled: Boolean(id) })

export const useCreateSession = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: SessionInput) => createSession(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: sessionKeys.all }),
  })
}

export const useUpdateSession = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: SessionUpdateInput }) => updateSession(id, input),
    onSuccess: (session) => {
      queryClient.setQueryData(sessionKeys.detail(session.id), session)
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export const useDeleteSession = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteSession,
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: sessionKeys.detail(id) })
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export const useDeleteSessionMedia = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ sessionId, mediaId }: { sessionId: string; mediaId: string }) => deleteSessionMedia(sessionId, mediaId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.detail(variables.sessionId) })
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}
