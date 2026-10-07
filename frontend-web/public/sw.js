/* PicoOffice 手写 Service Worker
 * 策略：
 *  - app shell（index.html + /assets/* + /luckysheet/*）预缓存 + 版本化
 *  - GET /api/docs 列表、GET /api/docs/:id/content：网络优先，失败回退缓存
 *  - POST/PUT 一律不缓存
 */
const VERSION = 'pico-v1'
const SHELL_CACHE = VERSION + '-shell'
const RUNTIME_CACHE = VERSION + '-runtime'

const SHELL_ASSETS = ['/', '/index.html', '/manifest.webmanifest']

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(SHELL_CACHE).then((c) => c.addAll(SHELL_ASSETS)).then(() => self.skipWaiting())
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((k) => k !== SHELL_CACHE && k !== RUNTIME_CACHE && k.startsWith('pico-'))
          .map((k) => caches.delete(k))
      )
    ).then(() => self.clients.claim())
  )
})

// 判定是否是只读的文档 GET
function isDocGet(req, url) {
  if (req.method !== 'GET') return false
  if (url.origin !== self.location.origin) return false
  // 文档列表
  if (url.pathname === '/api/docs') return true
  // 单文档正文
  if (/^\/api\/docs\/[^/]+\/content$/.test(url.pathname)) return true
  return false
}

self.addEventListener('fetch', (event) => {
  const req = event.request
  const url = new URL(req.url)

  // 只处理同源；跨域（如 CDN）直接放行
  if (url.origin !== self.location.origin) return

  // POST/PUT/DELETE 不碰缓存
  if (req.method !== 'GET') return

  // app shell：cache-first
  const isShell =
    url.pathname === '/' ||
    url.pathname === '/index.html' ||
    url.pathname.startsWith('/assets/') ||
    url.pathname.startsWith('/luckysheet/') ||
    url.pathname.startsWith('/icon-')

  if (isShell) {
    event.respondWith(
      caches.match(req).then((hit) => {
        if (hit) return hit
        return fetch(req).then((resp) => {
          const copy = resp.clone()
          caches.open(SHELL_CACHE).then((c) => c.put(req, copy))
          return resp
        })
      })
    )
    return
  }

  // 文档接口：网络优先，失败读缓存
  if (isDocGet(req, url)) {
    event.respondWith(
      fetch(req)
        .then((resp) => {
          const copy = resp.clone()
          caches.open(RUNTIME_CACHE).then((c) => c.put(req, copy))
          return resp
        })
        .catch(() => caches.match(req).then((hit) => hit || Response.error()))
    )
  }
})
