<template>
  <div class="pub-page" v-loading="loading">
    <div class="pub-top">
      <span class="pub-title">{{ doc.title || '分享文档' }}</span>
      <el-tag size="small" effect="plain">{{ permText }}</el-tag>
      <div class="spacer" />
      <router-link v-if="!isLogin" to="/login" class="login-link">登录</router-link>
    </div>

    <!-- editable but not logged in: prompt-->
    <el-alert
      v-if="needLogin"
      type="warning"
      :closable="false"
      title="该文档允许编辑，但当前未登录。登录后即可进入编辑。"
      style="margin-bottom: 12px"
    >
      <router-link to="/login">
        <el-button size="small" type="primary">去登录</el-button>
      </router-link>
    </el-alert>

    <!-- doc: read-only layout-->
    <div v-if="loadError" class="pub-empty">
      <el-empty :description="loadError" :image-size="80" />
    </div>

    <div v-else-if="docKind === 'doc'" class="paper doc-body" v-html="docHtml" />

    <!-- table: render first sheet read-only-->
    <div v-else-if="docKind === 'sheet'" class="paper">
      <table v-if="sheetRows.length" class="ro-sheet">
        <tbody>
          <tr v-for="(row, ri) in sheetRows" :key="ri">
            <td v-for="(cell, ci) in row" :key="ci">{{ cellText(cell) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-else class="pub-empty">（空表格）</div>
    </div>

    <!-- slides: read-only play-->
    <div v-else-if="docKind === 'slide'" class="paper slide-stage">
      <div v-if="slides.length" class="slide-card" :style="slideStyle">
        <div class="slide-idx">{{ curSlide + 1 }} / {{ slides.length }}</div>
        <div class="slide-title" v-html="slides[curSlide].title || ''" />
        <div class="slide-body" v-html="slides[curSlide].body || ''" />
      </div>
      <div v-else class="pub-empty">（空演示）</div>
      <div v-if="slides.length > 1" class="slide-nav">
        <el-button size="small" :disabled="curSlide === 0" @click="curSlide--">上一页</el-button>
        <el-button size="small" :disabled="curSlide === slides.length - 1" @click="curSlide++">下一页</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { apiGetPublicShare } from '@/api/extras'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()
const token = route.params.token

const loading = ref(true)
const loadError = ref('')
const doc = ref({})
const rawContent = ref('')
const docKind = ref('doc')
const perm = ref('view')
const curSlide = ref(0)

const isLogin = computed(() => auth.isLogin)
const permText = computed(() => (perm.value === 'edit' ? '可编辑' : '仅查看'))
// edit permission but not logged in -> prompt login
const needLogin = computed(() => perm.value === 'edit' && !auth.isLogin)

// doc body: strip page-setup comment, then render
const docHtml = computed(() => {
  return (rawContent.value || '').replace(/^<!--pico-page:.*?-->\n?/, '')
})

// table: luckysheet snapshot JSON; render first sheet with data
const sheetRows = computed(() => {
  try {
    const sheets = JSON.parse(rawContent.value || '[]')
    if (!Array.isArray(sheets) || !sheets.length) return []
    const sheet = sheets.find((s) => s.data && s.data.length) || sheets[0]
    return (sheet && sheet.data) || []
  } catch (e) {
    return []
  }
})
function cellText(cell) {
  if (cell == null) return ''
  if (typeof cell === 'object') return cell.v ?? cell.m ?? ''
  return cell
}

// slides: try parse as slides array; on failure treat body as one slide
const slides = computed(() => {
  try {
    const arr = JSON.parse(rawContent.value || '[]')
    if (Array.isArray(arr) && arr.length && arr[0] && (arr[0].title != null || arr[0].body != null)) {
      return arr
    }
  } catch (e) {
    // falls through below
  }
  if (rawContent.value) return [{ title: '', body: rawContent.value }]
  return []
})
const slideStyle = computed(() => ({}))

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const resp = await apiGetPublicShare(token)
    // backend may flatten doc fields or wrap in { doc, content }
    const d = resp.doc || resp
    doc.value = { title: d.title || '', id: d.id }
    docKind.value = d.doc_kind || d.kind || 'doc'
    perm.value = d.share_perm || d.perm || 'view'
    rawContent.value = resp.content != null ? resp.content : d.content || ''
  } catch (e) {
    loadError.value = '分享不存在或已失效'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.pub-page {
  min-height: 100%;
  padding: 16px;
  box-sizing: border-box;
  background: #e9edf2;
}
.pub-top {
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 900px;
  margin: 0 auto 12px;
}
.pub-title {
  font-size: 16px;
  font-weight: 600;
}
.spacer {
  flex: 1;
}
.login-link {
  font-size: 13px;
}
.paper {
  max-width: 900px;
  margin: 0 auto;
  background: #fff;
  padding: 40px 56px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}
.doc-body {
  font-size: 14px;
  line-height: 1.7;
}
.doc-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
}
.doc-body :deep(td),
.doc-body :deep(th) {
  border: 1px solid #dcdfe6;
  padding: 6px 10px;
}
.ro-sheet {
  border-collapse: collapse;
  width: 100%;
  font-size: 13px;
}
.ro-sheet td {
  border: 1px solid #dcdfe6;
  padding: 4px 8px;
  min-width: 60px;
}
.slide-stage {
  min-height: 420px;
  display: flex;
  flex-direction: column;
}
.slide-card {
  flex: 1;
  border: 1px solid #ebeef5;
  padding: 32px;
  position: relative;
}
.slide-idx {
  position: absolute;
  top: 8px;
  right: 12px;
  font-size: 12px;
  color: #909399;
}
.slide-title {
  font-size: 22px;
  font-weight: 700;
  margin-bottom: 16px;
}
.slide-body {
  font-size: 15px;
  line-height: 1.8;
}
.slide-nav {
  margin-top: 16px;
  display: flex;
  justify-content: center;
  gap: 12px;
}
.pub-empty {
  max-width: 900px;
  margin: 0 auto;
  background: #fff;
  padding: 40px;
}
</style>
