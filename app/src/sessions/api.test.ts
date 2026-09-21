import assert from 'node:assert/strict'
import test from 'node:test'
import { buildSessionFormData } from './api'

test('buildSessionFormData preserves text fields and repeated media', () => {
  const image = new File(['image'], 'screen.png', { type: 'image/png' })
  const audio = new File(['audio'], 'answer.mp3', { type: 'audio/mpeg' })
  const form = buildSessionFormData({
    title: 'Practice',
    description: 'Description',
    status: 'active',
    content: '# Notes',
    images: [image],
    audio: [audio, audio],
  })

  assert.equal(form.get('title'), 'Practice')
  assert.equal(form.get('content'), '# Notes')
  assert.equal(form.getAll('images').length, 1)
  assert.equal(form.getAll('audio').length, 2)
})
