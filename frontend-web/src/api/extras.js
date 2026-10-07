import req from './http'

// P0/P1 new APIs: import/trash/share/folder/star/tags.
// kept separate from src/api/index.js (touched by another agent) to avoid conflicts.

// ---- import: pick docx/xlsx/pptx -> online doc ----
export const apiImportDoc = (file) => {
  const form = new FormData()
  form.append('file', file)
  return req.post('/docs/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// ---- trash ----
export const apiListTrash = () => req.get('/trash')
export const apiRestoreTrash = (id) => req.post('/trash/' + id + '/restore')
export const apiPurgeTrash = (id) => req.delete('/trash/' + id)

// ---- share ----
// perm: view | edit
export const apiCreateShare = (id, sharePerm) =>
  req.post('/docs/' + id + '/share', { perm: sharePerm })
export const apiDeleteShare = (id) => req.delete('/docs/' + id + '/share')
// shared by me / shared with me
export const apiListShared = () => req.get('/shared')
// public share: anonymous access works without token (interceptor skips Authorization when no token)
export const apiGetPublicShare = (token) => req.get('/share/' + token)

// ---- folders ----
export const apiListFolders = () => req.get('/folders')
export const apiCreateFolder = (body) => req.post('/folders', body)
export const apiUpdateFolder = (id, body) => req.put('/folders/' + id, body)
export const apiDeleteFolder = (id) => req.delete('/folders/' + id)

// ---- star ----
export const apiStarDoc = (id) => req.post('/docs/' + id + '/star')
export const apiUnstarDoc = (id) => req.delete('/docs/' + id + '/star')

// ---- tags ----
export const apiListTags = () => req.get('/tags')
export const apiCreateTag = (name) => req.post('/tags', { name })
export const apiDeleteTag = (id) => req.delete('/tags/' + id)
// replace doc tags, body: { tag_ids: [] }
export const apiSetDocTags = (id, tagIds) => req.put('/docs/' + id + '/tags', { tag_ids: tagIds })

// batch move/delete: reuse single-doc API in parallel
