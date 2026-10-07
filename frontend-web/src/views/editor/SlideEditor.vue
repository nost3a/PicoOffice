<template>
  <div class="slide-page">
    <!-- top toolbar-->
    <div class="toolbar">
      <el-button size="small" link @click="$router.push('/dashboard')">
        <el-icon><Back /></el-icon>
      </el-button>
      <span class="doc-title">{{ doc.title || '未命名演示' }}</span>

      <el-divider direction="vertical" />
      <el-button size="small" :type="masterMode ? 'primary' : ''" @click="masterMode = !masterMode">
        {{ masterMode ? '退出母版' : '母版' }}
      </el-button>
      <el-select size="small" v-model="transition" style="width: 110px">
        <el-option label="淡入" value="fade" />
        <el-option label="滑动" value="slide" />
      </el-select>

      <div class="spacer" />
      <span class="save-tip">{{ saveTip }}</span>
      <el-button size="small" type="primary" :loading="saving" @click="doSave(true)">保存</el-button>
      <el-button size="small" type="success" @click="startPlay">
        <el-icon><VideoPlay /></el-icon>&nbsp;播放
      </el-button>
      <el-button size="small" @click="doPrint">打印</el-button>
    </div>

    <!-- master toolbar: visible only in master mode-->
    <div v-if="masterMode" class="master-bar">
      <span class="mb-label">母版：</span>
      <!-- theme presets: 5 palettes, one-click switch-->
      <span>主题</span>
      <el-select size="small" :model-value="master.themeKey" @change="applyTheme" style="width: 100px">
        <el-option v-for="t in themePresets" :key="t.key" :label="t.name" :value="t.key" />
      </el-select>
      <span>背景色</span>
      <el-color-picker size="small" v-model="master.bgColor" />
      <span>标题位置</span>
      <el-select size="small" v-model="master.titlePos" style="width: 110px">
        <el-option label="顶部" value="top" />
        <el-option label="居中" value="center" />
        <el-option label="底部" value="bottom" />
      </el-select>
      <span>正文字号</span>
      <el-select size="small" v-model.number="master.fontSize" style="width: 90px">
        <el-option :value="16" label="小" />
        <el-option :value="20" label="中" />
        <el-option :value="24" label="大" />
      </el-select>
      <span class="mb-tip">改完对所有页生效</span>
    </div>

    <div class="body">
      <!-- left slide thumbnail list-->
      <div class="side">
        <div
          v-for="(s, idx) in slides"
          :key="idx"
          class="thumb"
          :class="{ active: idx === curIdx }"
          @click="curIdx = idx"
        >
          <div class="thumb-no">{{ idx + 1 }}</div>
          <div class="thumb-preview" :style="{ background: master.bgColor }">
            <div class="thumb-title">{{ s.title || '未命名' }}</div>
            <div class="thumb-body">{{ s.body || '' }}</div>
          </div>
        </div>
        <el-button size="small" class="add-btn" @click="addSlide">+ 新增一页</el-button>
      </div>

      <!-- center editing area-->
      <div class="main">
        <div
          v-if="slides[curIdx]"
          class="canvas"
          :style="{ background: master.bgColor }"
        >
          <el-input
            v-model="slides[curIdx].title"
            placeholder="标题"
            size="large"
            class="title-input"
            :class="'pos-' + master.titlePos"
            :style="{ color: master.titleColor }"
            @input="scheduleSave"
          />
          <el-input
            v-model="slides[curIdx].body"
            type="textarea"
            :rows="8"
            placeholder="正文要点"
            class="body-input"
            :style="{ fontSize: master.fontSize + 'px' }"
            @input="scheduleSave"
          />
        </div>

        <!-- notes: visible in edit mode only-->
        <div class="notes-box">
          <div class="notes-label">备注（仅编辑可见，播放时显示在屏幕下方）</div>
          <el-input
            v-model="slides[curIdx].notes"
            type="textarea"
            :rows="3"
            placeholder="给这一页写点演讲提示…"
            size="small"
            @input="scheduleSave"
          />
        </div>
      </div>
    </div>

    <!-- fullscreen play overlay-->
    <teleport to="body">
      <div v-if="playing" class="play-mask" @click.self="playing = false">
        <div
          :key="playIdx"
          class="play-slide"
          :class="'anim-' + transition"
          :style="{ background: master.bgColor }"
        >
          <h1 class="play-title" :class="'pos-' + master.titlePos" :style="{ color: master.titleColor }">
            {{ slides[playIdx]?.title }}
          </h1>
          <p class="play-body" :style="{ fontSize: master.fontSize + 2 + 'px' }">
            {{ slides[playIdx]?.body }}
          </p>
        </div>
        <!-- notes shown below screen-->
        <div class="play-notes" v-if="slides[playIdx]?.notes">
          {{ slides[playIdx].notes }}
        </div>
        <div class="play-nav">
          <el-button size="small" @click="prevSlide">上一页</el-button>
          <span>{{ playIdx + 1 }} / {{ slides.length }}</span>
          <el-button size="small" @click="nextSlide">下一页</el-button>
          <el-button size="small" @click="playing = false">退出</el-button>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGetDoc, apiGetContent, apiPutContent } from '@/api'

const route = useRoute()
const docId = route.params.id

const doc = ref({})
const slides = ref([])
const curIdx = ref(0)
const saving = ref(false)
const saveTip = ref('')
const playing = ref(false)
const playIdx = ref(0)

// master + transition
const masterMode = ref(false)
const transition = ref('fade')
const master = reactive({
  bgColor: '#ffffff',
  titlePos: 'top',
  fontSize: 20,
  themeKey: 'blue',
  titleColor: '#1f2d3d',
  accentColor: '#409eff'
})

// 5 preset themes: change background + title color, all slides follow
const themePresets = [
  { key: 'blue', name: '商务蓝', bgColor: '#f0f5ff', titleColor: '#1d4ed8', accentColor: '#409eff' },
  { key: 'green', name: '清新绿', bgColor: '#f0fbf4', titleColor: '#15803d', accentColor: '#67c23a' },
  { key: 'purple', name: '优雅紫', bgColor: '#f6f3ff', titleColor: '#6d28d9', accentColor: '#8b5cf6' },
  { key: 'orange', name: '活力橙', bgColor: '#fff7ed', titleColor: '#c2410c', accentColor: '#e6a23c' },
  { key: 'gray', name: '极简灰', bgColor: '#f5f5f5', titleColor: '#333333', accentColor: '#606266' }
]
function applyTheme(key) {
  const t = themePresets.find((x) => x.key === key)
  if (!t) return
  master.bgColor = t.bgColor
  master.titleColor = t.titleColor
  master.accentColor = t.accentColor
  scheduleSave()
}

let saveTimer = null

async function loadDoc() {
  doc.value = await apiGetDoc(docId)
  const resp = await apiGetContent(docId)
  try {
    const data = resp.content ? JSON.parse(resp.content) : null
    // legacy data is plain array; new is { slides, master }
    if (Array.isArray(data)) {
      slides.value = data
    } else if (data && Array.isArray(data.slides)) {
      slides.value = data.slides
      if (data.master) Object.assign(master, data.master)
      if (data.transition) transition.value = data.transition
    }
  } catch (e) {
    slides.value = []
  }
  if (!slides.value.length) slides.value.push({ title: '', body: '', notes: '' })
}

function scheduleSave() {
  saveTip.value = '编辑中…'
  clearTimeout(saveTimer)
  saveTimer = setTimeout(() => doSave(false), 5000)
}

async function doSave(manual) {
  saving.value = true
  try {
    const payload = { slides: slides.value, master: { ...master }, transition: transition.value }
    await apiPutContent(docId, { content: JSON.stringify(payload) })
    saveTip.value = '已保存 ' + new Date().toLocaleTimeString()
    if (manual) ElMessage.success('已保存')
  } catch (e) {
    if (e.response?.status === 409) {
      saveTip.value = '保存冲突'
      ElMessageBox.alert('文档已被他人修改，请刷新后合并。', '版本冲突', {
        confirmButtonText: '刷新',
        type: 'warning'
      }).then(() => loadDoc()).catch(() => {})
    } else {
      saveTip.value = '保存失败'
    }
  } finally {
    saving.value = false
  }
}

function doPrint() {
  window.print()
}

function addSlide() {
  slides.value.push({ title: '', body: '', notes: '' })
  curIdx.value = slides.value.length - 1
  scheduleSave()
}

function startPlay() {
  playIdx.value = curIdx.value
  playing.value = true
}
function nextSlide() {
  if (playIdx.value < slides.value.length - 1) playIdx.value++
}
function prevSlide() {
  if (playIdx.value > 0) playIdx.value--
}

// play mode: left/right arrows switch slide
function onKeydown(e) {
  if (!playing.value) return
  if (e.key === 'ArrowRight' || e.key === ' ') nextSlide()
  if (e.key === 'ArrowLeft') prevSlide()
  if (e.key === 'Escape') playing.value = false
}

onMounted(() => {
  loadDoc()
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  clearTimeout(saveTimer)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.slide-page {
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
.save-tip {
  font-size: 12px;
  color: #909399;
  margin-right: 8px;
}
.master-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #ecf5ff;
  border-bottom: 1px solid #d9ecff;
  font-size: 12px;
}
.mb-label {
  font-weight: 600;
}
.mb-tip {
  color: #909399;
  margin-left: 8px;
}
.body {
  flex: 1;
  display: flex;
  overflow: hidden;
}
.side {
  width: 160px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  padding: 8px;
  overflow-y: auto;
}
.thumb {
  display: flex;
  gap: 6px;
  padding: 6px;
  margin-bottom: 6px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
}
.thumb.active {
  border-color: #409eff;
}
.thumb-no {
  color: #909399;
  font-size: 12px;
}
.thumb-preview {
  flex: 1;
  min-width: 0;
  padding: 4px;
}
.thumb-title {
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.thumb-body {
  font-size: 11px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.add-btn {
  width: 100%;
}
.main {
  flex: 1;
  padding: 24px;
  overflow: auto;
  background: #f5f7fa;
}
.canvas {
  max-width: 860px;
  margin: 0 auto;
  padding: 32px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  min-height: 320px;
}
.title-input :deep(.el-input__wrapper) {
  font-size: 22px;
  font-weight: 600;
  box-shadow: none;
  border-bottom: 1px solid #e4e7ed;
  border-radius: 0;
}
.title-input.pos-center {
  text-align: center;
  margin: 40px 0;
}
.body-input {
  margin-top: 16px;
}
.notes-box {
  max-width: 860px;
  margin: 16px auto 0;
}
.notes-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
</style>

<style>
.play-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  z-index: 2000;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.play-slide {
  width: 70%;
  min-height: 60%;
  padding: 40px;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
/* slide transition*/
.anim-fade {
  animation: picoFade 0.4s ease;
}
.anim-slide {
  animation: picoSlide 0.35s ease;
}
@keyframes picoFade {
  from { opacity: 0; }
  to { opacity: 1; }
}
@keyframes picoSlide {
  from { transform: translateX(40px); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
.play-title {
  font-size: 40px;
  margin-bottom: 24px;
}
.play-title.pos-center { text-align: center; }
.play-body {
  font-size: 20px;
  white-space: pre-wrap;
  line-height: 1.8;
}
.play-notes {
  position: absolute;
  bottom: 56px;
  width: 70%;
  color: #c0c4cc;
  font-size: 13px;
  text-align: center;
}
.play-nav {
  position: absolute;
  bottom: 16px;
  display: flex;
  gap: 12px;
  align-items: center;
  color: #fff;
}
</style>
