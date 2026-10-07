<template>
  <div class="page">
    <!-- top bar-->
    <div class="topbar">
      <div class="title">PicoOffice</div>
      <el-dropdown @command="onCreate">
        <el-button type="primary" size="small">
          新建 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="doc">文字文档</el-dropdown-item>
            <el-dropdown-item command="sheet">电子表格</el-dropdown-item>
            <el-dropdown-item command="slide">演示文稿</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <!-- import docx/xlsx/pptx-->
      <el-upload
        :show-file-list="false"
        :auto-upload="false"
        accept=".docx,.xlsx,.pptx"
        :on-change="onImport"
      >
        <el-button size="small" :loading="importing">导入</el-button>
      </el-upload>

      <el-input
        v-model="keyword"
        placeholder="搜索文件"
        prefix-icon="Search"
        clearable
        size="small"
        class="search"
        @keyup.enter="loadList"
        @clear="loadList"
      />

      <!-- offline status indicator-->
      <el-tag :type="net.statusType" size="small" effect="plain">{{ net.statusText }}</el-tag>

      <div class="spacer" />

      <!-- nav: trash / shared by me-->
      <el-button size="small" text @click="$router.push('/trash')">回收站</el-button>
      <el-button size="small" text @click="$router.push('/shared')">我分享的</el-button>

      <!-- usage bar-->
      <div v-if="quota" class="quota">
        <span class="quota-label">用量</span>
        <el-progress
          :percentage="quotaPercent"
          :stroke-width="10"
          style="width: 140px"
          :show-text="false"
        />
        <span class="quota-num">{{ usedText }}</span>
      </div>
      <el-button v-if="auth.isAdmin" size="small" @click="$router.push('/admin')">用户管理</el-button>
      <el-dropdown>
        <el-button size="small" circle>
          {{ avatarText }}
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="onLogout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>

    <!-- offline banner-->
    <el-alert
      v-if="!net.online"
      title="当前离线，下方展示的是本地缓存的文档，可继续编辑，联网后自动同步"
      type="warning"
      :closable="false"
      style="margin-bottom: 8px"
    />

    <!-- top tabs: all / starred / recent-->
    <el-tabs v-model="activeTab" class="tabs" @tab-change="loadList">
      <el-tab-pane label="全部" name="all" />
      <el-tab-pane label="星标" name="starred" />
      <el-tab-pane label="最近" name="recent" />
    </el-tabs>

    <div class="main">
      <!-- folder sidebar-->
      <FolderTree
        ref="folderTreeRef"
        v-model="folderId"
        @changed="loadList"
      />

      <div class="right">
        <!-- batch action bar-->
        <div v-if="selected.length" class="batch-bar">
          <span class="batch-count">已选 {{ selected.length }} 项</span>
          <el-select
            v-model="moveFolderId"
            placeholder="移动到文件夹"
            size="small"
            style="width: 160px"
            @change="batchMove"
          >
            <el-option label="根目录" :value="null" />
            <el-option v-for="f in flatFolders" :key="f.id" :label="f.name" :value="f.id" />
          </el-select>
          <el-button size="small" type="danger" @click="batchDelete">批量删除</el-button>
        </div>

        <el-table
          v-loading="loading"
          :data="list"
          class="tbl"
          @row-dblclick="openRow"
          :header-cell-style="{ background: '#fafafa' }"
          @selection-change="onSelectionChange"
        >
          <el-table-column type="selection" width="40" />
          <el-table-column label="名称" min-width="240">
            <template #default="{ row }">
              <span class="fname">
                <el-icon v-if="row.type === 'doc'" color="#409eff"><Document /></el-icon>
                <el-icon v-else-if="row.type === 'sheet'" color="#67c23a"><Grid /></el-icon>
                <el-icon v-else color="#e6a23c"><VideoPlay /></el-icon>
                {{ row.title }}
                <el-icon
                  class="star"
                  :class="{ on: row.starred }"
                  @click.stop="toggleStar(row)"
                >
                  <StarFilled />
                </el-icon>
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="80">
            <template #default="{ row }">
              <el-tag size="small" effect="plain">{{ typeText(row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="标签" width="160">
            <template #default="{ row }">
              <template v-if="row.tags && row.tags.length">
                <el-tag
                  v-for="t in row.tags"
                  :key="t.id || t"
                  size="small"
                  effect="plain"
                  class="tag"
                >{{ t.name || t }}</el-tag>
              </template>
              <span v-else class="muted">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" width="160">
            <template #default="{ row }">{{ fmtTime(row.updated_at) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="230" align="right">
            <template #default="{ row }">
              <el-button size="small" text @click="openRow(row)">打开</el-button>
              <el-button size="small" text @click="openShare(row)">分享</el-button>
              <el-button size="small" text @click="openTagDlg(row)">标签</el-button>
              <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- share dialog-->
    <el-dialog v-model="shareDlg" :title="'分享「' + (shareRow?.title || '') + '」'" width="440px">
      <div v-if="!shareRow" />
      <template v-else>
        <div v-if="shareRow.share_token" class="share-ok">
          <el-alert type="success" :closable="false" title="链接已生成" style="margin-bottom: 10px" />
          <el-input :model-value="shareLink(shareRow)" readonly>
            <template #append>
              <el-button @click="copyLink(shareRow)">复制</el-button>
            </template>
          </el-input>
          <div class="share-perm">当前权限：{{ shareRow.share_perm === 'edit' ? '可编辑' : '仅查看' }}</div>
        </div>
        <div v-else>
          <el-radio-group v-model="sharePerm">
            <el-radio value="view">仅查看</el-radio>
            <el-radio value="edit">可编辑</el-radio>
          </el-radio-group>
        </div>
      </template>
      <template #footer>
        <el-button v-if="shareRow?.share_token" size="small" type="danger" @click="cancelShare">取消分享</el-button>
        <el-button size="small" @click="shareDlg = false">关闭</el-button>
        <el-button v-if="!shareRow?.share_token" size="small" type="primary" @click="genShare">生成链接</el-button>
      </template>
    </el-dialog>

    <!-- tag dialog-->
    <el-dialog v-model="tagDlg" :title="'打标签「' + (tagRow?.title || '') + '」'" width="360px">
      <el-select v-model="tagSelected" multiple placeholder="选择标签" style="width: 100%">
        <el-option v-for="t in allTags" :key="t.id" :label="t.name" :value="t.id" />
      </el-select>
      <template #footer>
        <el-button size="small" @click="tagDlg = false">取消</el-button>
        <el-button size="small" type="primary" @click="saveTags">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useNetStore } from '@/stores/net'
import { apiListDocs, apiCreateDoc, apiDeleteDoc, apiMyQuota, apiUpdateDoc } from '@/api'
import {
  apiImportDoc,
  apiCreateShare,
  apiDeleteShare,
  apiStarDoc,
  apiUnstarDoc,
  apiListFolders,
  apiListTags,
  apiSetDocTags
} from '@/api/extras'
import FolderTree from '@/components/FolderTree.vue'

const router = useRouter()
const auth = useAuthStore()
const net = useNetStore()

const list = ref([])
const loading = ref(false)
const keyword = ref('')
const quota = ref(null)
const importing = ref(false)

// top tabs / folder filter / batch
const activeTab = ref('all')
const folderId = ref(null)
const selected = ref([])
const flatFolders = ref([])
const moveFolderId = ref(null)

// share dialog
const shareDlg = ref(false)
const shareRow = ref(null)
const sharePerm = ref('view')

// tag dialog
const tagDlg = ref(false)
const tagRow = ref(null)
const tagSelected = ref([])
const allTags = ref([])

const folderTreeRef = ref(null)

const avatarText = computed(() => {
  const name = auth.user?.nickname || auth.user?.username || '?'
  return name.slice(0, 1).toUpperCase()
})

const quotaPercent = computed(() => {
  if (!quota.value || !quota.value.quota) return 0
  const p = (quota.value.used / quota.value.quota) * 100
  return Math.min(100, Math.round(p))
})
const usedText = computed(() => {
  if (!quota.value) return ''
  return quota.value.used + ' / ' + quota.value.quota + (quota.value.unit || '')
})

function typeText(t) {
  return t === 'doc' ? '文档' : t === 'sheet' ? '表格' : t === 'slide' ? '演示' : t
}

function fmtTime(t) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return '-'
  const p = (n) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) +
    ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
}

// normalize backend doc: doc_kind -> type
function normDoc(d) {
  return {
    ...d,
    id: d.id,
    type: d.doc_kind || d.type,
    title: d.title || '未命名',
    updated_at: d.updated_at,
    starred: !!d.starred,
    tags: d.tags || []
  }
}

async function loadQuota() {
  try {
    quota.value = await apiMyQuota()
  } catch (e) {
    // hide if quota API missing
  }
}

async function loadFolders() {
  try {
    const resp = await apiListFolders()
    flatFolders.value = resp.list || resp || []
  } catch (e) {
    flatFolders.value = []
  }
}

async function loadTags() {
  try {
    const resp = await apiListTags()
    allTags.value = resp.list || resp || []
  } catch (e) {
    allTags.value = []
  }
}

async function loadList() {
  loading.value = true
  try {
    if (!net.online) {
      const offline = await net.offlineDocs()
      list.value = offline.map((d) => normDoc({ id: d.doc_id, doc_kind: d.doc_kind, title: d.title, updated_at: d.updated_at }))
      return
    }
    const params = { keyword: keyword.value }
    if (folderId.value) params.folder_id = folderId.value
    if (activeTab.value === 'starred') params.starred = 1
    if (activeTab.value === 'recent') params.recent = 1
    const resp = await apiListDocs(params)
    let rows = (resp.list || resp || []).map(normDoc)
    // star tab falls back to local filter
    if (activeTab.value === 'starred') rows = rows.filter((r) => r.starred)
    // recent: by updated_at desc
    rows = rows.slice().sort((a, b) => new Date(b.updated_at || 0) - new Date(a.updated_at || 0))
    list.value = rows
  } catch (e) {
    const offline = await net.offlineDocs()
    list.value = offline.map((d) => normDoc({ id: d.doc_id, doc_kind: d.doc_kind, title: d.title, updated_at: d.updated_at }))
  } finally {
    loading.value = false
  }
}

async function onCreate(docKind) {
  const ext = { doc: '未命名文档', sheet: '未命名表格', slide: '未命名演示' }[docKind]
  try {
    const resp = await apiCreateDoc({ title: ext, type: docKind, folder_id: folderId.value })
    const id = resp.id || resp.doc?.id
    openEditor(docKind, id)
  } catch (e) {
    // interceptor already warned
  }
}

// import: pick docx/xlsx/pptx, POST /docs/import, route by doc_kind
async function onImport(file) {
  importing.value = true
  try {
    const resp = await apiImportDoc(file.raw)
    const doc = resp.doc || resp
    const id = doc.id
    const docKind = doc.doc_kind || doc.type
    ElMessage.success('导入成功')
    openEditor(docKind, id)
  } catch (e) {
    // interceptor already warned on error
  } finally {
    importing.value = false
  }
}

function openEditor(docKind, id) {
  router.push('/editor/' + docKind + '/' + id)
}

function openRow(row) {
  openEditor(row.type, row.id)
}

async function onDelete(row) {
  if (!net.online) {
    ElMessage.warning('离线状态不能删除，请联网后操作')
    return
  }
  await ElMessageBox.confirm('确认删除「' + row.title + '」？', '提示', { type: 'warning' })
  await apiDeleteDoc(row.id)
  ElMessage.success('已删除')
  loadList()
}

// ---- star toggle ----
async function toggleStar(row) {
  try {
    if (row.starred) {
      await apiUnstarDoc(row.id)
      row.starred = false
    } else {
      await apiStarDoc(row.id)
      row.starred = true
    }
    if (activeTab.value === 'starred') loadList()
  } catch (e) {
    // interceptor already warned
  }
}

// ---- share ----
function shareLink(row) {
  return location.origin + '/s/' + row.share_token
}
function openShare(row) {
  shareRow.value = row
  sharePerm.value = row.share_perm || 'view'
  shareDlg.value = true
}
async function genShare() {
  try {
    const resp = await apiCreateShare(shareRow.value.id, sharePerm.value)
    shareRow.value.share_token = resp.token || resp.share_token
    shareRow.value.share_perm = sharePerm.value
    ElMessage.success('链接已生成')
  } catch (e) {
    // interceptor already warned
  }
}
async function cancelShare() {
  try {
    await apiDeleteShare(shareRow.value.id)
    shareRow.value.share_token = ''
    ElMessage.success('已取消分享')
    loadList()
  } catch (e) {
    // interceptor already warned
  }
}
function copyLink(row) {
  const url = shareLink(row)
  navigator.clipboard.writeText(url).then(
    () => ElMessage.success('链接已复制'),
    () => ElMessage.warning('复制失败，请手动复制')
  )
}

// ---- tags ----
function openTagDlg(row) {
  tagRow.value = row
  tagSelected.value = (row.tags || []).map((t) => t.id || t)
  tagDlg.value = true
}
async function saveTags() {
  try {
    await apiSetDocTags(tagRow.value.id, tagSelected.value)
    ElMessage.success('已保存标签')
    tagDlg.value = false
    loadList()
  } catch (e) {
    // interceptor already warned
  }
}

// ---- batch ----
function onSelectionChange(rows) {
  selected.value = rows
}
async function batchMove(folderIdVal) {
  if (!selected.value.length) return
  try {
    for (const row of selected.value) {
      await apiUpdateDoc(row.id, { folder_id: folderIdVal || null })
    }
    ElMessage.success('已移动 ' + selected.value.length + ' 项')
    moveFolderId.value = null
    loadList()
  } catch (e) {
    // interceptor already warned
  }
}
async function batchDelete() {
  if (!selected.value.length) return
  try {
    await ElMessageBox.confirm('确认删除选中的 ' + selected.value.length + ' 项？', '批量删除', { type: 'warning' })
  } catch (e) {
    return
  }
  for (const row of selected.value) {
    await apiDeleteDoc(row.id)
  }
  ElMessage.success('已删除')
  loadList()
}

function onLogout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  loadList()
  loadQuota()
  loadFolders()
  loadTags()
})
</script>

<style scoped>
.page {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  box-sizing: border-box;
}
.topbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}
.title {
  font-size: 16px;
  font-weight: 600;
  margin-right: 4px;
}
.search {
  width: 200px;
}
.spacer {
  flex: 1;
}
.quota {
  display: flex;
  align-items: center;
  gap: 8px;
}
.quota-label {
  font-size: 12px;
  color: #909399;
}
.quota-num {
  font-size: 12px;
  color: #606266;
}
.tabs {
  margin-bottom: 4px;
}
.tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}
.main {
  flex: 1;
  display: flex;
  gap: 10px;
  min-height: 0;
  margin-top: 4px;
}
.right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.batch-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  background: #ecf5ff;
  border: 1px solid #d9ecff;
  margin-bottom: 8px;
}
.batch-count {
  font-size: 13px;
  color: #409eff;
}
.tbl {
  flex: 1;
  background: #fff;
}
.fname {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.star {
  color: #dcdfe6;
  cursor: pointer;
}
.star.on {
  color: #e6a23c;
}
.tag {
  margin-right: 4px;
}
.muted {
  color: #c0c4cc;
  font-size: 12px;
}
.share-perm {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
