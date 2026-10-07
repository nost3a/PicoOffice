import req from './http'

// auth APIs
export const apiRegister = (body) => req.post('/auth/register', body)
export const apiLogin = (body) => req.post('/auth/login', body)
export const apiMe = () => req.get('/auth/me')

// doc list and CRUD
export const apiListDocs = (params) => req.get('/docs', { params })
export const apiCreateDoc = (body) => req.post('/docs', body)
export const apiGetDoc = (id) => req.get('/docs/' + id)
export const apiUpdateDoc = (id, body) => req.put('/docs/' + id, body)
export const apiDeleteDoc = (id) => req.delete('/docs/' + id)

// doc body; meta carries page setup; client serializes into content if backend doesn't store
export const apiGetContent = (id) => req.get('/docs/' + id + '/content')
export const apiPutContent = (id, body) => req.put('/docs/' + id + '/content', body)

// export: backend returns file stream; build URL, client uses a[download]
export const exportDocUrl = (id, bizType) =>
  '/api/docs/' + id + '/export?type=' + bizType

// export is POST (backend uses POST); fetch blob via axios for embed PDF preview
export const exportDocBlob = (id, bizType) =>
  req.post('/docs/' + id + '/export?type=' + bizType, null, { responseType: 'blob' })

// version history: list / preview / rollback
export const apiListVersions = (id) => req.get('/docs/' + id + '/versions')
export const apiGetVersion = (id, vid) => req.get('/docs/' + id + '/versions/' + vid)
export const apiRollback = (id, versionId) =>
  req.post('/docs/' + id + '/rollback', { version_id: versionId })

// user management (admin)
export const apiListUsers = () => req.get('/users')
export const apiDeleteUser = (id) => req.delete('/users/' + id)
// update user quota (admin)
export const apiSetUserQuota = (id, body) => req.put('/admin/users/' + id + '/quota', body)

// current user usage (dashboard top-right usage bar)
export const apiMyQuota = () => req.get('/me/quota')

// chunked upload: large files go init->chunk->complete; small files direct
export const apiUploadInit = (body) => req.post('/fs/upload/init', body)
export const apiUploadChunk = (uploadId, chunkIndex, blob) => {
  const form = new FormData()
  form.append('chunkIndex', chunkIndex)
  form.append('data', blob)
  return req.post('/fs/upload/chunk', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
export const apiUploadComplete = (uploadId) => req.post('/fs/upload/complete', { uploadId })
export const apiUploadStatus = (uploadId) => req.get('/fs/upload/status', { params: { uploadId } })

// small file direct upload
export const apiUploadSmall = (blob, filename) => {
  const form = new FormData()
  form.append('file', blob, filename)
  return req.post('/fs/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
