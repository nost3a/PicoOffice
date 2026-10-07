<template>
  <div class="pdf-page">
    <div class="toolbar">
      <el-button size="small" link @click="$router.push('/dashboard')">
        <el-icon><Back /></el-icon>
      </el-button>
      <span class="doc-title">PDF 批注：{{ doc.title || id }}</span>

      <div class="spacer" />
      <el-button
        size="small"
        :type="annotateMode ? 'primary' : ''"
        @click="annotateMode = !annotateMode"
      >
        {{ annotateMode ? '退出批注' : '批注模式' }}
      </el-button>
      <el-button size="small" @click="reloadPdf" :loading="loading">重新加载</el-button>
      <el-button size="small" @click="clearAll">清空批注</el-button>
    </div>

    <!-- PDF area: embed + transparent overlay-->
    <div ref="stageRef" class="stage">
      <embed v-if="pdfUrl" :src="pdfUrl" type="application/pdf" class="pdf-embed" />
      <div v-if="!pdfUrl" class="pdf-empty" v-loading="loading">
        {{ loading ? 'PDF 加载中…' : 'PDF 不可用（导出服务未就绪或无权限）' }}
      </div>

      <!-- annotation overlay: captures mouse only in annotate mode-->
      <div
        v-if="pdfUrl"
        class="annotation-layer"
        :class="{ on: annotateMode }"
        @mousedown="onDown"
        @mousemove="onMove"
        @mouseup="onUp"
      >
        <!-- saved highlights-->
        <div
          v-for="a in annotations"
          :key="a.id"
          class="hl-box"
          :style="{ left: a.x + 'px', top: a.y + 'px', width: a.w + 'px', height: a.h + 'px' }"
          :title="a.text"
          @dblclick.stop="editNote(a)"
        >{{ a.text ? '📝' : '' }}</div>
        <!-- in-progress drag rectangle-->
        <div
          v-if="drag.w"
          class="hl-box dragging"
          :style="{ left: drag.x + 'px', top: drag.y + 'px', width: drag.w + 'px', height: drag.h + 'px' }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGetDoc, exportDocBlob } from '@/api'
import {
  idbPutAnnotation,
  idbListAnnotations,
  idbDelAnnotation
} from '@/lib/idb'

const route = useRoute()
const id = route.params.id

const doc = ref({})
const loading = ref(false)
const pdfUrl = ref('')
const annotateMode = ref(false)
const annotations = ref([])
const stageRef = ref(null)

// drag-select state
const drag = ref({ x: 0, y: 0, w: 0, h: 0 })
let startX = 0
let startY = 0

// fetch PDF blob (with token), objectURL for embed
async function reloadPdf() {
  loading.value = true
  try {
    if (pdfUrl.value) URL.revokeObjectURL(pdfUrl.value)
    const blob = await exportDocBlob(id, 'pdf')
    pdfUrl.value = URL.createObjectURL(blob)
  } catch (e) {
    pdfUrl.value = ''
  } finally {
    loading.value = false
  }
}

async function loadAnnos() {
  annotations.value = await idbListAnnotations(id)
}

// drag-select: press sets origin, drag computes size, release saves highlight
function onDown(e) {
  if (!annotateMode.value) return
  const rect = stageRef.value.getBoundingClientRect()
  startX = e.clientX - rect.left
  startY = e.clientY - rect.top
  drag.value = { x: startX, y: startY, w: 0, h: 0 }
}
function onMove(e) {
  if (!annotateMode.value || !drag.value) return
  const rect = stageRef.value.getBoundingClientRect()
  const cx = e.clientX - rect.left
  const cy = e.clientY - rect.top
  drag.value = {
    x: Math.min(startX, cx),
    y: Math.min(startY, cy),
    w: Math.abs(cx - startX),
    h: Math.abs(cy - startY)
  }
}
async function onUp() {
  if (!annotateMode.value) return
  const r = drag.value
  drag.value = { x: 0, y: 0, w: 0, h: 0 }
  if (r.w < 12 || r.h < 8) return // too small = accidental tap
  // on release, save a highlight; text annotation optional
  let text = ''
  try {
    text = await ElMessageBox.prompt('给这块高亮写条批注（可留空）', '批注', {
      confirmButtonText: '保存',
      cancelButtonText: '只加高亮'
    }).then((x) => x.value).catch(() => '')
  } catch (e) {
    text = ''
  }
  await idbPutAnnotation({
    doc_id: id,
    x: Math.round(r.x), y: Math.round(r.y),
    w: Math.round(r.w), h: Math.round(r.h),
    text: text || '',
    created_at: Date.now()
  })
  ElMessage.success('已添加高亮')
  loadAnnos()
}

// double-click highlight: edit text / delete
async function editNote(a) {
  try {
    const text = await ElMessageBox.prompt('编辑批注', '批注', {
      inputValue: a.text || '',
      confirmButtonText: '保存',
      cancelButtonText: '删除此高亮'
    }).then((x) => x.value).catch(() => null)
    if (text === null) {
      await idbDelAnnotation(a.id)
      ElMessage.info('已删除')
    } else {
      await idbPutAnnotation({ ...a, text })
    }
  } catch (e) {
    // cancel
  }
  loadAnnos()
}

async function clearAll() {
  try {
    await ElMessageBox.confirm('确定清空本文档全部批注？', '清空批注', { type: 'warning' })
  } catch (e) {
    return
  }
  for (const a of annotations.value) await idbDelAnnotation(a.id)
  annotations.value = []
  ElMessage.success('已清空')
}

onMounted(async () => {
  try { doc.value = await apiGetDoc(id) } catch (e) { /* noop */ }
  await reloadPdf()
  await loadAnnos()
})
onBeforeUnmount(() => {
  if (pdfUrl.value) URL.revokeObjectURL(pdfUrl.value)
})
</script>

<style scoped>
.pdf-page {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}
.doc-title {
  font-weight: 600;
  font-size: 14px;
}
.spacer {
  flex: 1;
}
.stage {
  flex: 1;
  position: relative;
  background: #e9edf2;
  margin: 16px;
}
.pdf-embed {
  width: 100%;
  height: 100%;
  border: none;
}
.pdf-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 14px;
}
/* annotation overlay: transparent by default, captures mouse in annotate mode*/
.annotation-layer {
  position: absolute;
  inset: 0;
  z-index: 5;
  pointer-events: none;
}
.annotation-layer.on {
  pointer-events: auto;
  cursor: crosshair;
}
.hl-box {
  position: absolute;
  background: rgba(255, 235, 59, 0.35);
  border: 1px solid rgba(255, 193, 7, 0.6);
  pointer-events: auto;
  font-size: 12px;
  line-height: 1;
}
.hl-box.dragging {
  background: rgba(64, 158, 255, 0.2);
  border: 1px dashed #409eff;
}
</style>
