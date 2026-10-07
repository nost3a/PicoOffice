// lightweight IndexedDB wrapper, avoids idb dep
// two stores: offline_doc (recent drafts), sync_queue (pending PUT tasks)

const DB_NAME = 'picooffice'
const DB_VERSION = 2
const STORE_DOC = 'offline_doc'
const STORE_QUEUE = 'sync_queue'
const STORE_ANN = 'pdf_annotation'

let dbPromise = null

function openDb() {
  if (dbPromise) return dbPromise
  dbPromise = new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE_DOC)) {
        // keyPath = doc_id; keep only latest draft per doc
        db.createObjectStore(STORE_DOC, { keyPath: 'doc_id' })
      }
      if (!db.objectStoreNames.contains(STORE_QUEUE)) {
        // auto-increment PK; multiple rows per doc allowed
        const q = db.createObjectStore(STORE_QUEUE, { keyPath: 'id', autoIncrement: true })
        q.createIndex('by_doc', 'doc_id', { unique: false })
      }
      if (!db.objectStoreNames.contains(STORE_ANN)) {
        // PDF annotation: indexed by doc_id; one row = highlight rect + text
        const a = db.createObjectStore(STORE_ANN, { keyPath: 'id', autoIncrement: true })
        a.createIndex('by_doc', 'doc_id', { unique: false })
      }
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
  return dbPromise
}

function tx(storeName, mode, fn) {
  return openDb().then(
    (db) =>
      new Promise((resolve, reject) => {
        const t = db.transaction(storeName, mode)
        const store = t.objectStore(storeName)
        const result = fn(store)
        t.oncomplete = () => resolve(result)
        t.onerror = () => reject(t.error)
      })
  )
}

// save offline draft (doc_id, title, doc_kind, content, updated_at)
export async function idbPutDoc(rec) {
  await tx(STORE_DOC, 'readwrite', (s) => s.put(rec))
}

export async function idbGetDoc(docId) {
  return openDb().then(
    (db) =>
      new Promise((resolve, reject) => {
        const r = db.transaction(STORE_DOC).objectStore(STORE_DOC).get(docId)
        r.onsuccess = () => resolve(r.result || null)
        r.onerror = () => reject(r.error)
      })
  )
}

export async function idbListDocs() {
  return openDb().then(
    (db) =>
      new Promise((resolve, reject) => {
        const r = db.transaction(STORE_DOC).objectStore(STORE_DOC).getAll()
        r.onsuccess = () => resolve(r.result || [])
        r.onerror = () => reject(r.error)
      })
  )
}

// sync queue: push one row (doc_id, content, doc_kind, created_at)
export async function idbPushQueue(rec) {
  let id = null
  await tx(STORE_QUEUE, 'readwrite', (s) => {
    const r = s.add(rec)
    r.onsuccess = () => { id = r.result }
  })
  return id
}

export async function idbListQueue() {
  return openDb().then(
    (db) =>
      new Promise((resolve, reject) => {
        const r = db.transaction(STORE_QUEUE).objectStore(STORE_QUEUE).getAll()
        r.onsuccess = () => resolve(r.result || [])
        r.onerror = () => reject(r.error)
      })
  )
}

export async function idbDelQueue(id) {
  await tx(STORE_QUEUE, 'readwrite', (s) => s.delete(id))
}

// PDF annotation: one row (doc_id, x, y, w, h, text, created_at)
export async function idbPutAnnotation(rec) {
  let id = null
  await tx(STORE_ANN, 'readwrite', (s) => {
    const r = s.put(rec)
    r.onsuccess = () => { id = r.result }
  })
  return id
}

// list all annotations for a doc
export async function idbListAnnotations(docId) {
  return openDb().then(
    (db) =>
      new Promise((resolve, reject) => {
        const t = db.transaction(STORE_ANN)
        const idx = t.objectStore(STORE_ANN).index('by_doc')
        const r = idx.getAll(docId)
        r.onsuccess = () => resolve(r.result || [])
        r.onerror = () => reject(r.error)
      })
  )
}

export async function idbDelAnnotation(id) {
  await tx(STORE_ANN, 'readwrite', (s) => s.delete(id))
}
