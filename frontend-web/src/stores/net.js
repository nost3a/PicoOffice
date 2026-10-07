import { defineStore } from 'pinia'
import {
  idbPutDoc,
  idbGetDoc,
  idbListDocs,
  idbPushQueue,
  idbListQueue,
  idbDelQueue
} from '@/lib/idb'
import { apiPutContent } from '@/api'

// network state + offline drafts + pending sync queue
export const useNetStore = defineStore('net', {
  state: () => ({
    online: typeof navigator === 'undefined' ? true : navigator.onLine,
    pending: 0 // pending sync count
  }),
  getters: {
    // top badge: online / offline / N pending
    statusText(s) {
      if (!s.online) return '离线'
      if (s.pending > 0) return '待同步 ' + s.pending + ' 条'
      return '在线'
    },
    statusType(s) {
      if (!s.online) return 'warning'
      if (s.pending > 0) return 'danger'
      return 'success'
    }
  },
  actions: {
    init() {
      // listen for online/visibility changes, trigger sync
      window.addEventListener('online', () => {
        this.online = true
        this.flushQueue()
      })
      window.addEventListener('offline', () => {
        this.online = false
      })
      document.addEventListener('visibilitychange', () => {
        if (!document.hidden && this.online) this.flushQueue()
      })
      this.refreshPending()
    },
    async refreshPending() {
      const q = await idbListQueue()
      this.pending = q.length
    },
    // save doc: PUT online; offline writes IndexedDB + sync queue
    async saveDoc(docId, docKind, title, content) {
      // always save a local draft; dashboard shows it offline
      await idbPutDoc({
        doc_id: docId,
        doc_kind: docKind,
        title: title || '未命名',
        content,
        updated_at: Date.now()
      })
      if (this.online) {
        try {
          await apiPutContent(docId, { content })
          return 'ok'
        } catch (e) {
          // 409 optimistic-lock: content changed elsewhere; propagate so editor prompts refresh/merge
          if (e.response?.status === 409) throw e
          // other API errors (incl. offline moments) go to sync queue
        }
      }
      await idbPushQueue({ doc_id: docId, content, created_at: Date.now() })
      await this.refreshPending()
      return 'queued'
    },
    // back online: replay PUTs from queue one by one
    async flushQueue() {
      const q = await idbListQueue()
      for (const item of q) {
        try {
          await apiPutContent(item.doc_id, { content: item.content })
          await idbDelQueue(item.id)
        } catch (e) {
          // keep on failure; retry next time online
          break
        }
      }
      await this.refreshPending()
    },
    // offline dashboard list
    async offlineDocs() {
      const list = await idbListDocs()
      return list.sort((a, b) => (b.updated_at || 0) - (a.updated_at || 0))
    },
    async getOfflineDoc(docId) {
      return idbGetDoc(docId)
    }
  }
})
