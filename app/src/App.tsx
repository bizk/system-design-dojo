import { useState, type FormEvent } from 'react'
import {
  useCreateSession,
  useDeleteSessionMedia,
  useSession,
  useSessions,
  useTranscribeSessionMedia,
  useUpdateSession,
} from './sessions/queries'
import './App.css'

function App() {
  const [selectedId, setSelectedId] = useState('')
  const [title, setTitle] = useState('')
  const sessions = useSessions()
  const activeId = sessions.data?.some((session) => session.id === selectedId)
    ? selectedId
    : sessions.data?.[0]?.id ?? ''
  const session = useSession(activeId)
  const createSession = useCreateSession()
  const updateSession = useUpdateSession()
  const deleteMedia = useDeleteSessionMedia()
  const transcribeMedia = useTranscribeSessionMedia()

  const submitSession = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const value = title.trim()
    if (!value) return
    createSession.mutate({ title: value }, {
      onSuccess: (created) => {
        setTitle('')
        setSelectedId(created.id)
      },
    })
  }

  const uploadAudio = (files: FileList | null) => {
    if (!files || !activeId) return
    updateSession.mutate({ id: activeId, input: { audio: Array.from(files) } })
  }

  return (
    <main className="app-shell">
      <header className="app-header">
        <div>
          <p className="eyebrow">Audio notebook</p>
          <h1>Keep the conversation.</h1>
          <p className="intro">Store session recordings, then turn them into searchable notes.</p>
        </div>
        <span className="status-mark" aria-hidden="true">REC</span>
      </header>

      <div className="workspace">
        <aside className="sidebar panel">
          <div className="panel-heading">
            <div>
              <p className="eyebrow">Sessions</p>
              <h2>Your recordings</h2>
            </div>
            <span className="count">{sessions.data?.length ?? 0}</span>
          </div>

          <form className="new-session" onSubmit={submitSession}>
            <label htmlFor="session-title">Start a session</label>
            <div className="inline-form">
              <input
                id="session-title"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="e.g. Design review"
              />
              <button className="button button-primary" type="submit" disabled={createSession.isPending}>
                {createSession.isPending ? '...' : 'Add'}
              </button>
            </div>
          </form>

          {sessions.isPending && <p className="muted">Loading sessions...</p>}
          {sessions.error && <p className="error" role="alert">{sessions.error.message}</p>}
          <div className="session-list">
            {sessions.data?.map((item) => (
              <button
                className={`session-link${item.id === activeId ? ' selected' : ''}`}
                key={item.id}
                type="button"
                onClick={() => setSelectedId(item.id)}
              >
                <strong>{item.title}</strong>
                <span>{item.audio.length} audio {item.audio.length === 1 ? 'file' : 'files'}</span>
              </button>
            ))}
          </div>
          {sessions.data?.length === 0 && !sessions.isPending && <p className="empty">No sessions yet. Add one above.</p>}
        </aside>

        <section className="detail panel">
          {!activeId && <div className="empty-detail"><p className="eyebrow">Nothing selected</p><h2>Choose a session to begin.</h2></div>}
          {activeId && session.isPending && <p className="muted">Loading recording...</p>}
          {session.error && <p className="error" role="alert">{session.error.message}</p>}
          {session.data && (
            <>
              <div className="detail-heading">
                <div>
                  <p className="eyebrow">{session.data.status}</p>
                  <h2>{session.data.title}</h2>
                  {session.data.description && <p className="muted">{session.data.description}</p>}
                </div>
                <label className="upload-button button button-primary">
                  <input
                    type="file"
                    accept="audio/*"
                    multiple
                    onChange={(event) => {
                      uploadAudio(event.currentTarget.files)
                      event.currentTarget.value = ''
                    }}
                  />
                  {updateSession.isPending ? 'Uploading...' : 'Upload audio'}
                </label>
              </div>

              {updateSession.error && <p className="error" role="alert">{updateSession.error.message}</p>}
              {session.data.audio.length === 0 && <div className="empty-audio"><strong>No recordings attached.</strong><span>Upload an audio file to this session to get started.</span></div>}
              <div className="audio-list">
                {session.data.audio.map((media) => {
                  const isDeleting = deleteMedia.isPending && deleteMedia.variables?.mediaId === media.id
                  const isTranscribing = transcribeMedia.isPending && transcribeMedia.variables?.mediaId === media.id
                  return (
                    <article className="audio-card" key={media.id}>
                      <div className="audio-card-header">
                        <div>
                          <p className="audio-name">{media.filename}</p>
                          <p className="muted">{formatBytes(media.size)} - {media.contentType}</p>
                        </div>
                        <span className="audio-dot" aria-hidden="true" />
                      </div>
                      <audio controls preload="metadata" src={media.url}>
                        Your browser does not support audio playback.
                      </audio>
                      <div className="audio-actions">
                        <button
                          className="button button-secondary"
                          type="button"
                          disabled={isTranscribing || isDeleting}
                          onClick={() => transcribeMedia.mutate({ sessionId: activeId, mediaId: media.id })}
                        >
                          {isTranscribing ? 'Transcribing...' : 'Transcribe'}
                        </button>
                        <button
                          className="button button-danger"
                          type="button"
                          disabled={isDeleting || isTranscribing}
                          onClick={() => {
                            if (window.confirm(`Delete ${media.filename}?`)) {
                              deleteMedia.mutate({ sessionId: activeId, mediaId: media.id })
                            }
                          }}
                        >
                          {isDeleting ? 'Deleting...' : 'Delete'}
                        </button>
                      </div>
                      {media.transcription && <div className="transcription"><p className="eyebrow">Transcription</p><p>{media.transcription}</p></div>}
                    </article>
                  )
                })}
              </div>
              {transcribeMedia.error && <p className="error" role="alert">{transcribeMedia.error.message}</p>}
              {deleteMedia.error && <p className="error" role="alert">{deleteMedia.error.message}</p>}
            </>
          )}
        </section>
      </div>
    </main>
  )
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export default App
