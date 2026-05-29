import { client } from './client'

export interface Collection {
  id: string
  name: string
  description: string
  type: string
  status: string
  doc_count: number
  created_at: string
  updated_at: string
}

export interface CollectionDocument {
  collection_id: string
  document_id: string
  title: string
  status: string
  sort_order: number
  note: string
  created_at: string
}

export interface CollectionDetailResult {
  collection: Collection
  documents: CollectionDocument[]
  total: number
}

export const listCollections = () =>
  client.get<{ items: Collection[]; total: number }>('/api/v1/collections')

export const createCollection = (req: { name: string; description?: string; type?: string }) =>
  client.post<Collection>('/api/v1/collections', req)

export const getCollection = (id: string) =>
  client.get<CollectionDetailResult>(`/api/v1/collections/${id}`)

export const updateCollection = (id: string, req: { name: string; description?: string }) =>
  client.put<Collection>(`/api/v1/collections/${id}`, req)

export const deleteCollection = (id: string) => client.delete(`/api/v1/collections/${id}`)

export const addDocumentToCollection = (collectionId: string, documentId: string, note?: string) =>
  client.post<void>(`/api/v1/collections/${collectionId}/documents`, {
    document_id: documentId,
    note: note ?? '',
  })

export const removeDocumentFromCollection = (collectionId: string, documentId: string) =>
  client.delete(`/api/v1/collections/${collectionId}/documents/${documentId}`)

export const reorderCollectionDocuments = (collectionId: string, documentIds: string[]) =>
  client.put<void>(`/api/v1/collections/${collectionId}/documents/reorder`, {
    document_ids: documentIds,
  })

export const createCollectionFromSearch = (req: {
  name: string
  description?: string
  document_ids: string[]
}) => client.post<Collection>('/api/v1/collections/from-search', req)
