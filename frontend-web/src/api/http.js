import axios from 'axios'
import { ElMessage } from 'element-plus'

// shared axios instance, baseURL same-origin /api; dev proxied to 127.0.0.1:8080
const req = axios.create({
  baseURL: '/api',
  timeout: 20000
})

// request interceptor: attach token
req.interceptors.request.use((config) => {
  const token = localStorage.getItem('pico_token')
  if (token) {
    config.headers.Authorization = 'Bearer ' + token
  }
  return config
})

// response interceptor: unwrap + 401 to login + 429 rate-limit toast
req.interceptors.response.use(
  (resp) => resp.data,
  (error) => {
    const status = error.response?.status
    const msg = error.response?.data?.message || error.message || '请求失败'
    if (status === 401) {
      localStorage.removeItem('pico_token')
      localStorage.removeItem('pico_user')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/login'
      }
    } else if (status === 429) {
      // login/API rate-limited, unified toast
      ElMessage.warning('尝试过于频繁，请稍后再试')
    } else {
      ElMessage.error(msg)
    }
    return Promise.reject(error)
  }
)

export default req
