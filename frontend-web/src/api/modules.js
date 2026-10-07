// new-module APIs: profile/2FA/devices/mail/calendar/search/storage
import req from './http'

// ---- profile ----
export const apiGetProfile = () => req.get('/me/profile')
export const apiPutProfile = (body) => req.put('/me/profile', body)
export const apiUploadAvatar = (file) => {
  const form = new FormData()
  form.append('file', file)
  return req.post('/me/avatar', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// ---- TOTP 2FA ----
export const apiTotpSetup = () => req.post('/me/totp/setup')
export const apiTotpEnable = (code) => req.post('/me/totp/enable', { code })
export const apiTotpDisable = (code) => req.post('/me/totp/disable', { code })

// ---- login devices ----
export const apiListDevices = () => req.get('/me/devices')
export const apiDeleteDevice = (id) => req.delete('/me/devices/' + id)

// ---- auth extensions ----
export const apiLogout = (refresh_token) =>
  req.post('/auth/logout', { refresh_token })
export const apiForgotPassword = (body) => req.post('/auth/forgot-password', body)

// ---- mail ----
export const apiListMailAccounts = () => req.get('/mail/accounts')
export const apiCreateMailAccount = (body) => req.post('/mail/accounts', body)
export const apiUpdateMailAccount = (id, body) =>
  req.put('/mail/accounts/' + id, body)
export const apiDeleteMailAccount = (id) => req.delete('/mail/accounts/' + id)
export const apiSyncMail = (id) => req.post('/mail/accounts/' + id + '/sync')
export const apiListMailMessages = (params) =>
  req.get('/mail/messages', { params })
export const apiMailBody = (id) => req.get('/mail/messages/' + id)
export const apiMailSend = (body) => req.post('/mail/send', body)

// ---- calendar ----
export const apiListEvents = (params) => req.get('/calendar/events', { params })
export const apiCreateEvent = (body) => req.post('/calendar/events', body)
export const apiUpdateEvent = (id, body) => req.put('/calendar/events/' + id, body)
export const apiDeleteEvent = (id) => req.delete('/calendar/events/' + id)

// ---- global search ----
export const apiSearch = (params) => req.get('/search', { params })

// ---- object storage (admin) ----
export const apiGetStorage = () => req.get('/admin/storage')
export const apiSaveStorage = (body) => req.put('/admin/storage', body)
export const apiTestStorage = (body) => req.post('/admin/storage/test', body)
