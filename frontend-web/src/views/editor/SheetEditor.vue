<template>
  <div class="sheet-page">
    <div class="toolbar">
      <el-button size="small" link @click="$router.push('/dashboard')">
        <el-icon><Back /></el-icon>
      </el-button>
      <span class="doc-title">{{ doc.title || '未命名表格' }}</span>

      <el-divider direction="vertical" />
      <!-- sheet enhancement tools: official luckysheet API with fallbacks-->
      <el-button size="small" @click="freezeFirstRow">冻结首行</el-button>
      <el-button size="small" @click="freezeToSelected">冻结到选中</el-button>
      <!-- number format: apply to selection-->
      <el-select size="small" :model-value="numFmt" @change="applyNumFmt" style="width: 130px">
        <el-option label="常规" value="general" />
        <el-option label="数字(千分位)" value="number" />
        <el-option label="货币(¥)" value="currency" />
        <el-option label="百分比" value="percent" />
        <el-option label="日期" value="date" />
      </el-select>
      <el-button size="small" @click="toggleFilter">开启筛选</el-button>
      <el-button size="small" @click="sortRange(1)">升序</el-button>
      <el-button size="small" @click="sortRange(-1)">降序</el-button>
      <el-button size="small" @click="applyColorScale">色阶</el-button>
      <el-button size="small" type="primary" @click="chartDlg = true">插入图表</el-button>

      <div class="spacer" />
      <span class="save-tip">{{ saveTip }}</span>
      <el-button size="small" type="primary" :loading="saving" @click="doSave(true)">保存</el-button>
      <el-button size="small" @click="exportFile('xlsx')">导出 xlsx</el-button>
      <el-button size="small" @click="doPrint">打印</el-button>
    </div>

    <!-- Luckysheet mount point: fixed id, luckysheet.create selects by id-->
    <div id="luckysheet-container" class="sheet-container"></div>

    <!-- chart dialog: canvas bar/line/pie-->
    <el-dialog v-model="chartDlg" title="插入图表" width="640px">
      <div class="chart-bar">
        <el-radio-group v-model="chartType" size="small">
          <el-radio-button value="bar">柱状图</el-radio-button>
          <el-radio-button value="line">折线图</el-radio-button>
          <el-radio-button value="pie">饼图</el-radio-button>
        </el-radio-group>
        <el-button size="small" @click="drawChart">用选中区域重绘</el-button>
      </div>
      <canvas ref="chartCanvas" width="560" height="320" class="chart-canvas"></canvas>
    </el-dialog>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiGetDoc, apiGetContent, apiPutContent, exportDocUrl } from '@/api'

const route = useRoute()
const docId = route.params.id

const doc = ref({})
const saving = ref(false)
const saveTip = ref('')

// chart dialog state
const chartDlg = ref(false)
const chartType = ref('bar')
const chartCanvas = ref(null)

// static assets: load from public/luckysheet
const CSS_FILES = [
  '/luckysheet/css/luckysheet.css',
  '/luckysheet/plugins/css/plugins.css'
]
const JS_FILES = [
  '/luckysheet/plugins/js/plugin.js',
  '/luckysheet/luckysheet.umd.js'
]

function loadCss(href) {
  return new Promise((resolve, reject) => {
    if (document.querySelector('link[href="' + href + '"]')) return resolve()
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = href
    link.onload = resolve
    link.onerror = reject
    document.head.appendChild(link)
  })
}

function loadScript(src) {
  return new Promise((resolve, reject) => {
    if (window.luckysheet) return resolve()
    const old = document.querySelector('script[src="' + src + '"]')
    if (old) return resolve()
    const s = document.createElement('script')
    s.src = src
    s.onload = resolve
    s.onerror = reject
    document.body.appendChild(s)
  })
}

async function loadLuckysheet() {
  for (const href of CSS_FILES) await loadCss(href)
  for (const src of JS_FILES) await loadScript(src)
}

async function loadDoc() {
  doc.value = await apiGetDoc(docId)
  const resp = await apiGetContent(docId)
  let sheets = []
  try {
    sheets = resp.content ? JSON.parse(resp.content) : []
  } catch (e) {
    sheets = []
  }
  if (!Array.isArray(sheets) || !sheets.length) {
    sheets = [{ name: 'Sheet1', color: '', status: 1, order: 0, data: [], config: {}, index: 0 }]
  }

  await loadLuckysheet()
  window.luckysheet.create({
    container: 'luckysheet-container',
    lang: 'zh',
    showinfobar: false,
    showsheetbar: true, // bottom sheet tab bar visible, multi-sheet
    allowEdit: true,
    enableCalculation: true, // formula engine: =SUM/AVERAGE/COUNT/VLOOKUP/IF auto-computed
    data: sheets,
    // freeze first row on init
    freezen: { type: 'row', range: [{ row_focus: '0', column_focus: '0' }] }
  })
  // fallback: call freeze API once more
  freezeFirstRow()
}

// ---- sheet enhancements: use official API, warn on failure ----
function freezeFirstRow() {
  try {
    // official setFreeze: freeze above row 1 (index 0)
    window.luckysheet?.setFreeze?.('row', 0, 1)
    ElMessage.success('已冻结首行')
  } catch (e) {
    ElMessage.warning('当前版本不支持冻结 API')
  }
}

// freeze at selected row/col: use selection bottom-right as freeze point
function freezeToSelected() {
  const range = getSelectedRange()
  if (!range) {
    ElMessage.warning('请先选中要冻结到的位置')
    return
  }
  try {
    const r2 = range.row[1]
    const c2 = range.column[1]
    // freeze both row+col; fall back to row-only on failure
    window.luckysheet?.setFreeze?.('both', r2, c2)
      || window.luckysheet?.setFreeze?.('row', 0, r2)
    ElMessage.success('已冻结到选中位置')
  } catch (e) {
    ElMessage.warning('冻结失败')
  }
}

// number format: apply ct (cell display format) to selection
const numFmt = ref('general')
const NUM_FMT_MAP = {
  number:   { fa: '#,##0.00', t: 'n' },
  currency: { fa: '¥#,##0.00', t: 'n' },
  percent:  { fa: '0%', t: 'n' },
  date:     { fa: 'yyyy-mm-dd', t: 'd' }
}
function applyNumFmt(kind) {
  const range = getSelectedRange()
  if (!range && kind !== 'general') {
    ElMessage.warning('请先选中区域')
    numFmt.value = 'general'
    return
  }
  try {
    const ct = NUM_FMT_MAP[kind]
    for (let r = range.row[0]; r <= range.row[1]; r++) {
      for (let c = range.column[0]; c <= range.column[1]; c++) {
        if (kind === 'general') {
          // normal: clear custom format
          window.luckysheet?.setCellFormat?.({ row: [r, r], column: [c, c] }, { ct: { fa: 'General', t: 'g' } })
        } else {
          window.luckysheet?.setCellFormat?.({ row: [r, r], column: [c, c] }, { ct })
        }
      }
    }
    ElMessage.success('已应用格式')
  } catch (e) {
    ElMessage.warning('格式应用失败')
  }
}

function toggleFilter() {
  try {
    // show filter button / enable filter
    window.luckysheet?.showFilter?.()
      || window.luckysheet?.setFilter?.()
    ElMessage.success('筛选已开启')
  } catch (e) {
    ElMessage.warning('筛选 API 不可用')
  }
}

// read current selection: luckysheet selection API varies; unified fallback
function getSelectedRange() {
  try {
    const r = window.luckysheet?.getRange?.()
    if (r && r.length) return r[0]
  } catch (e) { /* noop */ }
  // fallback: whole data range of current sheet
  return null
}

// read one cell value
function cellVal(r, c) {
  try {
    const v = window.luckysheet?.getCellValue?.(r, c)
    if (v == null) return null
    if (typeof v === 'object') return v.v ?? v.v === 0 ? v.v : null
    return v
  } catch (e) {
    return null
  }
}

// asc/desc: sort by first column of selection, write back
function sortRange(dir) {
  const range = getSelectedRange()
  if (!range) {
    ElMessage.warning('请先选中要排序的区域')
    return
  }
  try {
    const row1 = range.row[0]
    const row2 = range.row[1]
    const col = range.column[0]
    const rows = []
    for (let r = row1; r <= row2; r++) {
      rows.push({ r, key: Number(cellVal(r, col)) || 0 })
    }
    rows.sort((a, b) => (a.key - b.key) * dir)
    // keep simple: write sorted order back to that column
    rows.forEach((item, i) => {
      window.luckysheet?.setCellValue?.(row1 + i, col, item.key)
    })
    ElMessage.success(dir === 1 ? '已升序' : '已降序')
  } catch (e) {
    ElMessage.warning('排序失败')
  }
}

// color scale: blue->red background by selection value
function applyColorScale() {
  const range = getSelectedRange()
  if (!range) {
    ElMessage.warning('请先选中要上色的区域')
    return
  }
  try {
    const vals = []
    for (let r = range.row[0]; r <= range.row[1]; r++) {
      for (let c = range.column[0]; c <= range.column[1]; c++) {
        vals.push({ r, c, v: Number(cellVal(r, c)) || 0 })
      }
    }
    const min = Math.min(...vals.map((x) => x.v))
    const max = Math.max(...vals.map((x) => x.v))
    const span = max - min || 1
    vals.forEach((item) => {
      const t = (item.v - min) / span
      // blue(cold) -> red(warm)
      const rr = Math.round(64 + t * (230 - 64))
      const gg = Math.round(158 - t * 100)
      const bb = Math.round(255 - t * 200)
      window.luckysheet?.setCellFormat?.(
        { row: [item.r, item.r], column: [item.c, item.c] },
        { bg: 'rgb(' + rr + ',' + gg + ',' + bb + ')' }
      )
    })
    ElMessage.success('已应用色阶')
  } catch (e) {
    ElMessage.warning('色阶失败')
  }
}

// ---- charts: canvas drawn ----
function readSeries() {
  const range = getSelectedRange()
  if (!range) return null
  const labels = []
  const values = []
  for (let c = range.column[0]; c <= range.column[1]; c++) {
    labels.push(String(cellVal(range.row[0], c) ?? ''))
    values.push(Number(cellVal(range.row[0] + 1, c)) || 0)
  }
  // fallback: if only one column, read row by row
  if (values.length <= 1) {
    labels.length = 0
    values.length = 0
    for (let r = range.row[0]; r <= range.row[1]; r++) {
      labels.push(String(cellVal(r, range.column[0]) ?? r))
      values.push(Number(cellVal(r, range.column[0] + 1)) || 0)
    }
  }
  return { labels, values }
}

function drawChart() {
  const cv = chartCanvas.value
  if (!cv) return
  const ctx = cv.getContext('2d')
  ctx.clearRect(0, 0, cv.width, cv.height)
  const data = readSeries()
  if (!data || !data.values.length) {
    ctx.fillStyle = '#909399'
    ctx.font = '14px sans-serif'
    ctx.fillText('请先在表格里选中数据区域', 180, 160)
    return
  }
  const { labels, values } = data
  const max = Math.max(...values, 1)
  const colors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#b37feb']

  if (chartType.value === 'bar') {
    const bw = cv.width / (values.length * 1.6)
    values.forEach((v, i) => {
      const h = (v / max) * (cv.height - 60)
      const x = 40 + i * bw * 1.6
      const y = cv.height - 30 - h
      ctx.fillStyle = colors[i % colors.length]
      ctx.fillRect(x, y, bw, h)
      ctx.fillStyle = '#606266'
      ctx.font = '11px sans-serif'
      ctx.fillText(labels[i] || '', x, cv.height - 12)
    })
  } else if (chartType.value === 'line') {
    ctx.strokeStyle = '#409eff'
    ctx.lineWidth = 2
    ctx.beginPath()
    values.forEach((v, i) => {
      const x = 40 + (i / Math.max(values.length - 1, 1)) * (cv.width - 60)
      const y = cv.height - 30 - (v / max) * (cv.height - 60)
      if (i === 0) ctx.moveTo(x, y)
      else ctx.lineTo(x, y)
    })
    ctx.stroke()
    ctx.fillStyle = '#409eff'
    values.forEach((v, i) => {
      const x = 40 + (i / Math.max(values.length - 1, 1)) * (cv.width - 60)
      const y = cv.height - 30 - (v / max) * (cv.height - 60)
      ctx.beginPath()
      ctx.arc(x, y, 3, 0, Math.PI * 2)
      ctx.fill()
    })
  } else {
    // pie chart
    const total = values.reduce((a, b) => a + b, 0) || 1
    let start = -Math.PI / 2
    const cx = cv.width / 2
    const cy = cv.height / 2
    const r = 110
    values.forEach((v, i) => {
      const angle = (v / total) * Math.PI * 2
      ctx.fillStyle = colors[i % colors.length]
      ctx.beginPath()
      ctx.moveTo(cx, cy)
      ctx.arc(cx, cy, r, start, start + angle)
      ctx.closePath()
      ctx.fill()
      start += angle
    })
  }
}

async function openChart() {
  chartDlg.value = true
  await nextTick()
  drawChart()
}
// watch-triggered: draw when chartDlg opens
watch(chartDlg, (v) => {
  if (v) openChart()
})
watch(chartType, () => drawChart())

async function doSave(manual) {
  saving.value = true
  try {
    // snapshot whole workbook to backend
    let snapshot = []
    try {
      snapshot = window.luckysheet.getAllSheets?.() || window.luckysheet.getLuckysheetfile?.() || []
    } catch (e) {
      snapshot = []
    }
    await apiPutContent(docId, { content: JSON.stringify(snapshot) })
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

function exportFile(bizType) {
  doSave(false)
  window.open(exportDocUrl(docId, bizType), '_blank')
}

onMounted(loadDoc)
onBeforeUnmount(() => {
  try {
    window.luckysheet?.destroy?.()
  } catch (e) {
    // ignore
  }
})
</script>

<style scoped>
.sheet-page {
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
  flex-wrap: wrap;
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
.sheet-container {
  flex: 1;
  position: relative;
}
.chart-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}
.chart-canvas {
  background: #fafafa;
  border: 1px solid #ebeef5;
  border-radius: 2px;
}
</style>
