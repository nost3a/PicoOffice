<template>
  <div class="page">
    <div class="page-head">
      <span class="title">{{ $t('search.title') }}</span>
    </div>

    <div class="search-bar">
      <el-input
        v-model="kw"
        :placeholder="$t('search.placeholder')"
        clearable
        @keyup.enter="run"
      >
        <template #append>
          <el-button @click="run">{{ $t('common.search') }}</el-button>
        </template>
      </el-input>
      <el-checkbox v-model="rebuild" size="small" class="ml">{{ $t('search.rebuild') }}</el-checkbox>
    </div>

    <div v-if="loaded && !results.length" class="empty">{{ $t('search.resultEmpty') }}</div>
    <el-alert v-if="loaded && results.length" :title="$t('search.rebuildTip')" type="info :closable=false" class="mt" />

    <div v-for="r in results" :key="r.id || r.title" class="row" @click="open(r)">
      <div class="rhead">
        <el-tag size="small">{{ r.type || 'doc' }}</el-tag>
        <span class="rtitle">{{ r.title }}</span>
      </div>
      <div class="rsum">{{ r.summary || r.snippet || '' }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/i18n'
import { apiSearch } from '@/api/modules'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const kw = ref(route.query.q || '')
const rebuild = ref(false)
const results = ref([])
const loaded = ref(false)

async function run() {
  if (!kw.value) return
  try {
    const params = { q: kw.value }
    if (rebuild.value) params.rebuild = 1
    const r = await apiSearch(params)
    results.value = r.results || r.items || r || []
  } catch (e) {
    results.value = []
  } finally {
    loaded.value = true
  }
}

// route by type to matching editor
function open(r) {
  const type = (r.type || 'doc').toLowerCase()
  if (type === 'sheet') router.push('/editor/sheet/' + r.id)
  else if (type === 'slide') router.push('/editor/slide/' + r.id)
  else router.push('/editor/doc/' + r.id)
}

onMounted(run)
</script>

<style scoped>
.page { padding: 16px; max-width: 900px; margin: 0 auto; }
.page-head { margin-bottom: 12px; }
.title { font-size: 16px; font-weight: 600; }
.search-bar { display: flex; align-items: center; gap: 8px; }
.search-bar .ml { margin-left: 8px; }
.mt { margin-top: 8px; }
.empty { color: #909399; margin-top: 24px; text-align: center; }
.row { padding: 12px; border: 1px solid #ebeef5; border-radius: 6px; margin-top: 10px; cursor: pointer; }
.row:hover { background: #f5f7fa; }
.rhead { display: flex; align-items: center; gap: 8px; }
.rtitle { font-size: 14px; font-weight: 600; }
.rsum { color: #606266; font-size: 12px; margin-top: 6px; }
</style>
