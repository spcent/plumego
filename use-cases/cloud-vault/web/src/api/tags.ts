import { client } from './client'

export interface Tag {
  id: string
  name: string
  color?: string
  source?: string
  created_at: string
}

export interface TagListResult {
  items: Tag[]
  total: number
}

const BASE = '/api/v1'

export const tagsAPI = {
  list(): Promise<TagListResult> {
    return client.get<TagListResult>(`${BASE}/tags`)
  },

  create(name: string, color?: string): Promise<Tag> {
    return client.post<Tag>(`${BASE}/tags`, { name, color })
  },

  update(id: string, updates: { name?: string; color?: string }): Promise<Tag> {
    return client.put<Tag>(`${BASE}/tags/${id}`, updates)
  },

  delete(id: string): Promise<void> {
    return client.delete(`${BASE}/tags/${id}`)
  },

  getDocumentTags(docId: string): Promise<TagListResult> {
    return client.get<TagListResult>(`${BASE}/documents/${docId}/tags`)
  },

  setDocumentTags(docId: string, tagIds: string[]): Promise<void> {
    return client.put<void>(`${BASE}/documents/${docId}/tags`, { tag_ids: tagIds })
  },

  removeDocumentTag(docId: string, tagId: string): Promise<void> {
    return client.delete(`${BASE}/documents/${docId}/tags/${tagId}`)
  },
}
