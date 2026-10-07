<template>
  <div class="ed-page">
    <!-- top toolbar-->
    <div class="toolbar">
      <el-button size="small" link @click="$router.push('/dashboard')">
        <el-icon><Back /></el-icon>
      </el-button>
      <span class="doc-title">{{ doc.title || '未命名' }}</span>

      <el-divider direction="vertical" />
      <!-- page setup-->
      <el-button size="small" @click="pageDlg = true">页面设置</el-button>
      <!-- style dropdown: paragraph/heading/quote-->
      <el-select size="small" class="style-sel" :model-value="curStyle" @change="applyStyle" placeholder="样式">
        <el-option label="正文" value="p" />
        <el-option label="标题1" value="h1" />
        <el-option label="标题2" value="h2" />
        <el-option label="标题3" value="h3" />
        <el-option label="引用" value="blockquote" />
      </el-select>
      <!-- align-->
      <el-button size="small" :disabled="!canEdit" @click="setAlign('left')" title="左对齐"><el-icon><AlignLeft /></el-icon></el-button>
      <el-button size="small" :disabled="!canEdit" @click="setAlign('center')" title="居中"><el-icon><AlignCenter /></el-icon></el-button>
      <el-button size="small" :disabled="!canEdit" @click="setAlign('right')" title="右对齐"><el-icon><AlignRight /></el-icon></el-button>
      <el-button size="small" :disabled="!canEdit" @click="setAlign('justify')" title="两端"><el-icon><AlignJustify /></el-icon></el-button>

      <el-divider direction="vertical" />
      <el-button size="small" :disabled="!canEdit" @click="runCmd('bold')"><b>B</b></el-button>
      <el-button size="small" :disabled="!canEdit" @click="runCmd('italic')"><i>I</i></el-button>
      <el-button size="small" :disabled="!canEdit" @click="runCmd('strike')"><s>S</s></el-button>
      <!-- undo/redo: Tiptap history built-in-->
      <el-button size="small" :disabled="!canEdit" @click="runCmd('undo')" title="撤销 Ctrl+Z">撤销</el-button>
      <el-button size="small" :disabled="!canEdit" @click="runCmd('redo')" title="重做 Ctrl+Shift+Z">重做</el-button>
      <el-divider direction="vertical" />
      <el-button size="small" :disabled="!canEdit" @click="runCmd('bulletList')">列表</el-button>
      <el-button size="small" :disabled="!canEdit" @click="runCmd('table')">表格</el-button>
      <el-button size="small" :disabled="!canEdit" @click="addImage">图片</el-button>
      <!-- page break / line break-->
      <el-button size="small" :disabled="!canEdit" @click="insertPageBreak">分页</el-button>
      <el-button size="small" :disabled="!canEdit" @click="insertLineBreak" title="软换行">换行</el-button>

      <el-divider direction="vertical" />
      <!-- layout structure blocks-->
      <el-button size="small" :disabled="!canEdit" @click="insertHeader">页眉</el-button>
      <el-button size="small" :disabled="!canEdit" @click="insertFooter">页脚</el-button>
      <el-button size="small" :disabled="!canEdit" @click="insertPageNum">页码</el-button>
      <el-button size="small" :disabled="!canEdit" @click="insertToc">目录</el-button>
      <!-- image wrap (only when image selected): 5 modes-->
      <el-dropdown trigger="click" @command="imgWrap">
        <el-button size="small" :disabled="!imgActive">环绕</el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="inline">嵌入型</el-dropdown-item>
            <el-dropdown-item command="float-left">左环绕</el-dropdown-item>
            <el-dropdown-item command="float-right">右环绕</el-dropdown-item>
            <el-dropdown-item command="around">四周环绕</el-dropdown-item>
            <el-dropdown-item command="tight">紧密环绕</el-dropdown-item>
            <el-dropdown-item command="behind">衬于文字下方</el-dropdown-item>
            <el-dropdown-item command="above">浮于文字上方</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <!-- sheet enhancements: merge/split/bg/border-->
      <el-dropdown trigger="click" @command="tblAction">
        <el-button size="small" :disabled="!tblActive">表格操作</el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="merge">合并单元格</el-dropdown-item>
            <el-dropdown-item command="split">拆分单元格</el-dropdown-item>
            <el-dropdown-item command="bg">单元格底色…</el-dropdown-item>
            <el-dropdown-item command="border-all">全框线</el-dropdown-item>
            <el-dropdown-item command="border-none">无框线</el-dropdown-item>
            <el-dropdown-item command="border-outer">外框线</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <!-- style manager / outline / find-replace / version history-->
      <el-button size="small" @click="styleDlg = true">样式管理</el-button>
      <el-button size="small" :type="outlineOn ? 'primary' : ''" @click="toggleOutline">大纲</el-button>
      <el-button size="small" @click="openFindReplace('find')">查找</el-button>
      <el-button size="small" @click="openFindReplace('replace')">替换</el-button>
      <el-button size="small" @click="openHistory">历史</el-button>

      <div class="spacer" />
      <!-- online collab avatars-->
      <el-avatar
        v-for="u in onlineUsers"
        :key="u"
        :size="24"
        class="avatar"
      >{{ (u || '?').slice(0, 1).toUpperCase() }}</el-avatar>

      <span class="word-count">字数 {{ wordCount }} | 字符 {{ charCount }}</span>
      <span class="save-tip">{{ saveTip }}</span>
      <el-button size="small" type="primary" :loading="saving" @click="doSave(true)">保存</el-button>
      <el-button size="small" @click="exportFile('pdf')">导出 PDF</el-button>
      <el-button size="small" @click="$router.push('/viewer/pdf/' + docId)">PDF 批注</el-button>
      <el-button size="small" @click="exportFile('docx')">导出 docx</el-button>
      <el-button size="small" @click="doPrint">打印</el-button>
    </div>

    <!-- find/replace bar-->
    <div v-if="frBar.visible" class="fr-bar">
      <el-input size="small" v-model="frBar.find" class="fr-input" placeholder="查找内容" @keyup.enter="findNext" />
      <template v-if="frBar.mode === 'replace'">
        <el-input size="small" v-model="frBar.replace" class="fr-input" placeholder="替换为" />
      </template>
      <span class="fr-count">{{ frBar.count }}</span>
      <el-button size="small" @click="findNext">下一个</el-button>
      <el-button v-if="frBar.mode === 'replace'" size="small" type="primary" @click="replaceAll">全部替换</el-button>
      <el-button size="small" @click="closeFindReplace">关闭</el-button>
    </div>

    <!-- editing area: left outline + center page-->
    <div class="editor-wrap">
      <!-- outline panel: scan h1/h2/h3, click to jump-->
      <div v-if="outlineOn" class="outline-tree">
        <div class="ot-title">大纲</div>
        <div
          v-for="(n, i) in outlineTree"
          :key="i"
          class="ot-item"
          :class="'lv' + n.level"
          @click="jumpOutline(n)"
        >{{ n.text || '(空标题)' }}</div>
        <div v-if="!outlineTree.length" class="ot-empty">暂无标题</div>
      </div>

      <div class="paper-wrap">
        <editor-content
          class="tiptap"
          :editor="editor"
          :style="pageStyle"
        />
      </div>
    </div>

    <!-- page setup dialog-->
    <el-dialog v-model="pageDlg" title="页面设置" width="420px">
      <el-form label-width="90px" size="small">
        <el-form-item label="纸张大小">
          <el-radio-group v-model="pageSetup.paper">
            <el-radio value="a4">A4</el-radio>
            <el-radio value="a5">A5</el-radio>
            <el-radio value="letter">Letter</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="纸张方向">
          <el-radio-group v-model="pageSetup.orient">
            <el-radio value="portrait">纵向</el-radio>
            <el-radio value="landscape">横向</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="页边距">
          <el-radio-group v-model="pageSetup.margin">
            <el-radio value="narrow">窄</el-radio>
            <el-radio value="normal">普通</el-radio>
            <el-radio value="wide">宽</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分栏">
          <el-radio-group v-model.number="pageSetup.columns">
            <el-radio :value="1">1栏</el-radio>
            <el-radio :value="2">2栏</el-radio>
            <el-radio :value="3">3栏</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="行距">
          <el-select v-model.number="pageSetup.lineHeight" style="width: 120px">
            <el-option :value="1.0" label="1.0" />
            <el-option :value="1.25" label="1.25" />
            <el-option :value="1.5" label="1.5" />
            <el-option :value="2.0" label="2.0" />
          </el-select>
        </el-form-item>
        <el-form-item label="首行缩进">
          <el-radio-group v-model="pageSetup.indent">
            <el-radio value="0">无</el-radio>
            <el-radio value="2em">2字符</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="pageDlg = false">取消</el-button>
        <el-button size="small" type="primary" @click="applyPageSetup">应用</el-button>
      </template>
    </el-dialog>

    <!-- style manager: H1/H2/H3/body size/color/weight, document-wide-->
    <el-dialog v-model="styleDlg" title="样式管理" width="460px">
      <el-form label-width="80px" size="small">
        <el-form-item v-for="(cfg, key) in styleCfg" :key="key" :label="styleName(key)">
          <el-input-number v-model="cfg.size" :min="10" :max="48" :step="1" controls-position="right" style="width: 90px" />
          <el-color-picker v-model="cfg.color" />
          <el-select v-model="cfg.weight" style="width: 90px; margin-left: 8px">
            <el-option label="常规" :value="400" />
            <el-option label="加粗" :value="700" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="styleDlg = false">取消</el-button>
        <el-button size="small" type="primary" @click="applyStyleCfg">应用到全文</el-button>
      </template>
    </el-dialog>

    <!-- version history drawer-->
    <el-drawer v-model="historyOn" title="版本历史" size="360px">
      <div v-loading="historyLoading" class="hist-list">
        <div
          v-for="v in versionList"
          :key="v.id"
          class="hist-item"
          :class="{ active: previewVid === v.id }"
          @click="previewVersion(v)"
        >
          <div class="hi-who">{{ v.editor_name || '我' }} · {{ fmtSize(v.size) }}</div>
          <div class="hi-time">{{ fmtTime(v.created_at) }}</div>
        </div>
        <div v-if="!versionList.length && !historyLoading" class="ot-empty">暂无历史版本</div>
      </div>
      <template #footer>
        <el-button size="small" type="primary" :disabled="!previewVid" @click="restoreVersion">
          恢复此版本
        </el-button>
      </template>
    </el-drawer>

    <!-- version preview: overlay shows selected version body-->
    <el-dialog v-model="previewDlg" title="版本预览" width="640px">
      <div class="hist-preview" v-html="previewHtml"></div>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref, shallowRef, watch } from 'vue'
import { useRoute } from 'vue-router'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import TextStyle from '@tiptap/extension-text-style'
import Color from '@tiptap/extension-color'
import Image from '@tiptap/extension-image'
import Table from '@tiptap/extension-table'
import TableRow from '@tiptap/extension-table-row'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  apiGetDoc,
  apiGetContent,
  exportDocUrl,
  apiListVersions,
  apiGetVersion,
  apiRollback
} from '@/api'
import { useNetStore } from '@/stores/net'

const route = useRoute()
const docId = route.params.id
const net = useNetStore()

const doc = ref({})
const saving = ref(false)
const saveTip = ref('')
const onlineUsers = ref([])
const canEdit = ref(true)
const imgActive = ref(false)
const tblActive = ref(false)
const curStyle = ref('p')

// word count
const wordCount = ref(0)
const charCount = ref(0)

// outline view
const outlineOn = ref(false)
const outlineTree = ref([])
let outlineEls = []

// style manager: H1/H2/H3/body font-size/color/weight
const styleDlg = ref(false)
const styleCfg = reactive({
  h1: { size: 28, color: '#1f2d3d', weight: 700 },
  h2: { size: 22, color: '#1f2d3d', weight: 700 },
  h3: { size: 18, color: '#303133', weight: 600 },
  p:  { size: 14, color: '#303133', weight: 400 }
})
function styleName(k) {
  return { h1: '标题1', h2: '标题2', h3: '标题3', p: '正文' }[k] || k
}

// find/replace bar
const frBar = reactive({ visible: false, mode: 'find', find: '', replace: '', count: '' })

// version history
const historyOn = ref(false)
const historyLoading = ref(false)
const versionList = ref([])
const previewVid = ref(0)
const previewDlg = ref(false)
const previewHtml = ref('')

// page setup (meta): default A4 portrait, normal margins, single column
const pageDlg = ref(false)
const pageSetup = reactive({
  paper: 'a4',
  orient: 'portrait',
  margin: 'normal',
  columns: 1,
  lineHeight: 1.5,
  indent: '2em'
})

// paper size (px @96dpi): A4 794x1123, A5 595x842, Letter 816x1056
const PAPER_SIZE = {
  a4: [794, 1123],
  a5: [595, 842],
  letter: [816, 1056]
}
const MARGIN_PX = { narrow: 28, normal: 56, wide: 96 }

// compute container inline styles from pageSetup
const pageStyle = computed(() => {
  const [w, h] = PAPER_SIZE[pageSetup.paper] || PAPER_SIZE.a4
  const pw = pageSetup.orient === 'landscape' ? h : w
  const ph = pageSetup.orient === 'landscape' ? w : h
  const pad = MARGIN_PX[pageSetup.margin] || MARGIN_PX.normal
  return {
    width: pw + 'px',
    minHeight: ph + 'px',
    padding: pad + 'px ' + (pad + 24) + 'px',
    columnCount: pageSetup.columns,
    columnGap: '32px'
  }
})

// body paragraph styles (line-height/indent) on global class
watch(pageStyle, () => {
  applyTypographyCss()
})

function applyTypographyCss() {
  let styleEl = document.getElementById('pico-doc-typo')
  if (!styleEl) {
    styleEl = document.createElement('style')
    styleEl.id = 'pico-doc-typo'
    document.head.appendChild(styleEl)
  }
  styleEl.textContent =
    '.tiptap p, .tiptap li { line-height: ' + pageSetup.lineHeight +
    '; text-indent: ' + pageSetup.indent + '; }'
}

// body debounce save handle
let saveTimer = null
let ws = null
let heartbeatTimer = null
let reconnectAttempt = 0
let reconnectTimer = null

const editor = shallowRef(
  useEditor({
    extensions: [
      StarterKit,
      TextStyle,
      Color,
      Image.configure({ inline: false, allowBase64: true }),
      Table.configure({ resizable: true }),
      TableRow,
      TableCell,
      TableHeader
    ],
    content: '',
    editable: true,
    onUpdate: () => {
      scheduleSave()
      updateWordCount()
      scheduleOutline()
    },
    onSelectionUpdate: ({ editor }) => {
      imgActive.value = editor.isActive('image')
      tblActive.value = editor.isActive('table')
      // sync current style dropdown
      if (editor.isActive('heading')) {
        curStyle.value = 'h' + editor.getAttributes('heading').level
      } else if (editor.isActive('blockquote')) {
        curStyle.value = 'blockquote'
      } else {
        curStyle.value = 'p'
      }
    }
  })
)

// serialize page setup into leading HTML comment in content; backend ignores
const META_PREFIX = '<!--pico-page:'
function packContent(html) {
  const meta = JSON.stringify(pageSetup)
  // strip old comments before merging to avoid duplicates
  const bare = html.replace(/^<!--pico-page:.*?-->\n?/, '')
  return META_PREFIX + meta + '-->\n' + bare
}
function unpackContent(html) {
  const m = html && html.match(/^<!--pico-page:(.*?)-->\n?/)
  if (m) {
    try {
      Object.assign(pageSetup, JSON.parse(m[1]))
      applyTypographyCss()
    } catch (e) {
      // fall back to default if comment is malformed
    }
    return html.replace(/^<!--pico-page:.*?-->\n?/, '')
  }
  return html || ''
}

async function loadDoc() {
  doc.value = await apiGetDoc(docId)
  let html = ''
  try {
    const resp = await apiGetContent(docId)
    html = resp.content || ''
  } catch (e) {
    // online fetch failed: read from offline draft
    const draft = await net.getOfflineDoc(docId)
    if (draft) html = draft.content || ''
  }
  const bare = unpackContent(html)
  editor.value?.commands.setContent(bare)
  fillToc()
}

function scheduleSave() {
  saveTip.value = '编辑中…'
  clearTimeout(saveTimer)
  saveTimer = setTimeout(() => doSave(false), 5000)
}

async function doSave(manual) {
  if (!editor.value) return
  saving.value = true
  try {
    const full = packContent(editor.value.getHTML())
    const r = await net.saveDoc(docId, 'doc', doc.value.title, full)
    saveTip.value =
      r === 'queued'
        ? '离线，已存本地待同步'
        : '已保存 ' + new Date().toLocaleTimeString()
    if (manual) ElMessage.success(r === 'queued' ? '已存本地，联网后自动同步' : '已保存')
  } catch (e) {
    // 409: doc modified elsewhere; prompt refresh/merge
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

// print: use browser print; print CSS keeps only page content
function doPrint() {
  window.print()
}

function runCmd(cmd) {
  const e = editor.value
  if (!e) return
  if (cmd === 'bold') e.chain().focus().toggleBold().run()
  else if (cmd === 'italic') e.chain().focus().toggleItalic().run()
  else if (cmd === 'strike') e.chain().focus().toggleStrike().run()
  else if (cmd === 'bulletList') e.chain().focus().toggleBulletList().run()
  else if (cmd === 'table') e.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
  else if (cmd === 'undo') e.chain().focus().undo().run()
  else if (cmd === 'redo') e.chain().focus().redo().run()
}

// page break: insert empty paragraph marked page-break, starts new page on print
function insertPageBreak() {
  editor.value.chain().focus()
    .insertContent('<div class="page-break" style="page-break-after:always"></div><p></p>')
    .run()
}
// line break: soft <br>
function insertLineBreak() {
  editor.value.chain().focus().setHardBreak().run()
}

// word count: Chinese = non-whitespace chars; chars = total length
function updateWordCount() {
  const e = editor.value
  if (!e) return
  const text = e.getText() || ''
  charCount.value = text.length
  wordCount.value = (text.replace(/\s/g, '').length)
}

// outline: scan h1/h2/h3 into tree (debounced)
let outlineTimer = null
function scheduleOutline() {
  clearTimeout(outlineTimer)
  outlineTimer = setTimeout(buildOutline, 400)
}
function buildOutline() {
  const e = editor.value
  if (!e) return
  const heads = e.view?.dom?.querySelectorAll('h1,h2,h3') || []
  outlineTree.value = []
  outlineEls = []
  heads.forEach((h) => {
    outlineTree.value.push({ level: Number(h.tagName[1]) - 1, text: h.textContent, el: h })
    outlineEls.push(h)
  })
}
function jumpOutline(n) {
  n.el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
function toggleOutline() {
  outlineOn.value = !outlineOn.value
  if (outlineOn.value) buildOutline()
}

// apply paragraph style in one click
function applyStyle(val) {
  const e = editor.value
  if (!e) return
  e.chain().focus().clearNodes().run()
  if (val === 'p') e.chain().focus().setParagraph().run()
  else if (val === 'h1') e.chain().focus().toggleHeading({ level: 1 }).run()
  else if (val === 'h2') e.chain().focus().toggleHeading({ level: 2 }).run()
  else if (val === 'h3') e.chain().focus().toggleHeading({ level: 3 }).run()
  else if (val === 'blockquote') e.chain().focus().toggleBlockquote().run()
}

function setAlign(dir) {
  const e = editor.value
  if (!e) return
  e.chain().focus().updateAttributes('paragraph', { style: 'text-align:' + dir }).run()
  e.chain().focus().updateAttributes('heading', { style: 'text-align:' + dir }).run()
}

function addImage() {
  const url = prompt('图片 URL')
  if (url) editor.value.chain().focus().setImage({ src: url }).run()
}

// image wrap: change selected image class
function imgWrap(mode) {
  const e = editor.value
  if (!e) return
  e.chain().focus().updateAttributes('image', { class: mode }).run()
}

// table ops: merge/split/bg/border style
function tblAction(kind) {
  const e = editor.value
  if (!e) return
  if (kind === 'merge') {
    // select adjacent cells -> merge
    e.chain().focus().mergeCells().run()
  } else if (kind === 'split') {
    e.chain().focus().splitCell().run()
  } else if (kind === 'bg') {
    // open color picker for selected cell background
    const color = prompt('单元格底色（十六进制，如 #fff3cd）', '#fff3cd')
    if (color) e.chain().focus().updateAttributes('tableCell', { background: color }).run()
  } else if (kind === 'border-all') {
    e.chain().focus().updateAttributes('table', { class: 'pico-tbl-border pico-tbl-all' }).run()
  } else if (kind === 'border-none') {
    e.chain().focus().updateAttributes('table', { class: 'pico-tbl-none' }).run()
  } else if (kind === 'border-outer') {
    e.chain().focus().updateAttributes('table', { class: 'pico-tbl-outer' }).run()
  }
}

// style manager: write config into global <style>, applies document-wide
function applyStyleCfg() {
  let el = document.getElementById('pico-doc-styles')
  if (!el) {
    el = document.createElement('style')
    el.id = 'pico-doc-styles'
    document.head.appendChild(el)
  }
  const c = styleCfg
  el.textContent =
    '.tiptap h1{font-size:' + c.h1.size + 'px!important;color:' + c.h1.color +
      '!important;font-weight:' + c.h1.weight + '!important;}' +
    '.tiptap h2{font-size:' + c.h2.size + 'px!important;color:' + c.h2.color +
      '!important;font-weight:' + c.h2.weight + '!important;}' +
    '.tiptap h3{font-size:' + c.h3.size + 'px!important;color:' + c.h3.color +
      '!important;font-weight:' + c.h3.weight + '!important;}' +
    '.tiptap p{font-size:' + c.p.size + 'px!important;color:' + c.p.color +
      '!important;font-weight:' + c.p.weight + '!important;}'
  styleDlg.value = false
  scheduleSave()
}

// header/footer/page-number/TOC
function insertHeader() {
  editor.value.chain().focus().insertContentAt(0, '<header class="doc-header" contenteditable="true">页眉：在此输入</header>').run()
}
function insertFooter() {
  editor.value.chain().focus().insertContent('<footer class="doc-footer" contenteditable="true">页脚：{{page}}</footer>').run()
}
function insertPageNum() {
  // insert page-number placeholder at cursor; replaced on print/play
  editor.value.chain().focus().insertContent('{{page}}').run()
}
function insertToc() {
  editor.value.chain().focus().insertContent('<div class="toc"><p class="toc-title">目录</p></div>').run()
  setTimeout(fillToc, 100)
}

// scan h1/h2/h3 to auto-fill TOC
function fillToc() {
  const e = editor.value
  if (!e) return
  const tocEl = e.view?.dom?.querySelector('.toc')
  if (!tocEl) return
  const headings = e.view.dom.querySelectorAll('h1,h2,h3')
  if (!headings.length) return
  let html = '<p class="toc-title">目录</p>'
  headings.forEach((h) => {
    const lv = h.tagName === 'H1' ? 0 : h.tagName === 'H2' ? 1 : 2
    const txt = h.textContent || ''
    html += '<p class="toc-item lv' + lv + '">' + txt + '</p>'
  })
  tocEl.innerHTML = html
}

function applyPageSetup() {
  pageDlg.value = false
  applyTypographyCss()
  scheduleSave()
}

function exportFile(bizType) {
  doSave(false)
  window.open(exportDocUrl(docId, bizType), '_blank')
}

// ---- find/replace: find uses window.find; replace rewrites DOM clone ----
function openFindReplace(mode) {
  frBar.visible = true
  frBar.mode = mode
  updateFrCount()
  setTimeout(() => document.querySelector('.fr-bar input')?.focus(), 50)
}
function closeFindReplace() {
  frBar.visible = false
}
function updateFrCount() {
  const e = editor.value
  if (!e || !frBar.find) { frBar.count = ''; return }
  const n = (e.getText().split(frBar.find).length - 1)
  frBar.count = n + ' 处'
}
// next: use native find to jump to next match (with highlight)
function findNext() {
  updateFrCount()
  if (!frBar.find) return
  window.find(frBar.find, false, false, true, false, true)
}
// replace all: clone editing DOM, walk text nodes, then setContent to rebuild
function replaceAll() {
  const e = editor.value
  if (!e || !frBar.find) return
  const clone = e.view.dom.cloneNode(true)
  const walker = document.createTreeWalker(clone, NodeFilter.SHOW_TEXT)
  const nodes = []
  while (walker.nextNode()) nodes.push(walker.currentNode)
  let hit = 0
  nodes.forEach((n) => {
    if (n.nodeValue && n.nodeValue.includes(frBar.find)) {
      n.nodeValue = n.nodeValue.split(frBar.find).join(frBar.replace)
      hit++
    }
  })
  e.commands.setContent(clone.innerHTML)
  ElMessage.success('已替换 ' + hit + ' 处')
  frBar.find = ''
  updateFrCount()
}

// ---- version history ----
function fmtSize(n) {
  if (!n) return '0 B'
  if (n > 1024) return (n / 1024).toFixed(1) + ' KB'
  return n + ' B'
}
function fmtTime(t) {
  if (!t) return ''
  return new Date(t).toLocaleString()
}
async function openHistory() {
  historyOn.value = true
  historyLoading.value = true
  previewVid.value = 0
  try {
    const r = await apiListVersions(docId)
    versionList.value = r.items || []
  } catch (e) {
    versionList.value = []
  } finally {
    historyLoading.value = false
  }
}
async function previewVersion(v) {
  previewVid.value = v.id
  try {
    const r = await apiGetVersion(docId, v.id)
    // preview strips page-setup comment, display only
    previewHtml.value = (r.content || '').replace(/^<!--pico-page:.*?-->\n?/, '')
    previewDlg.value = true
  } catch (e) {
    ElMessage.warning('预览失败')
  }
}
async function restoreVersion() {
  try {
    await ElMessageBox.confirm('确定恢复到此版本？当前内容会被覆盖。', '恢复版本', { type: 'warning' })
  } catch (e) {
    return // user cancelled
  }
  try {
    const r = await apiRollback(docId, previewVid.value)
    // backend returns latest content after rollback; setContent directly
    const bare = unpackContent(r.content || '')
    editor.value?.commands.setContent(bare)
    previewDlg.value = false
    historyOn.value = false
    fillToc()
    updateWordCount()
    buildOutline()
    ElMessage.success('已恢复此版本')
  } catch (e) {
    // error already toasted by interceptor
  }
}

// collab: WS, 30s heartbeat, exponential backoff reconnect (max 5)
function connectWs() {
  const token = localStorage.getItem('pico_token') || ''
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = proto + '://' + location.host + '/api/ws/docs/' + docId + '?token=' + encodeURIComponent(token)
  try {
    ws = new WebSocket(url)
  } catch (e) {
    scheduleReconnect()
    return
  }
  ws.onopen = () => {
    reconnectAttempt = 0
    ws.send(JSON.stringify({ type: 'join' }))
    // send pong every 30s to keep alive
    clearInterval(heartbeatTimer)
    heartbeatTimer = setInterval(() => {
      if (ws && ws.readyState === 1) ws.send(JSON.stringify({ type: 'pong' }))
    }, 30000)
  }
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data)
      if (msg.type === 'presence' && Array.isArray(msg.users)) {
        onlineUsers.value = msg.users
      }
    } catch (e) {
      // ignore non-JSON
    }
  }
  ws.onclose = () => {
    clearInterval(heartbeatTimer)
    scheduleReconnect()
  }
  ws.onerror = () => {
    try { ws.close() } catch (e) { /* noop */ }
  }
}

function scheduleReconnect() {
  if (reconnectAttempt >= 5) return
  const delay = 3000 * Math.pow(2, reconnectAttempt) // 3s,6s,12s,24s,48s
  reconnectAttempt++
  clearTimeout(reconnectTimer)
  reconnectTimer = setTimeout(connectWs, delay)
}

onMounted(async () => {
  applyTypographyCss()
  applyStyleCfgSilent()
  await loadDoc()
  updateWordCount()
  buildOutline()
  connectWs()
  // Ctrl+F find / Ctrl+H replace, override browser default
  window.addEventListener('keydown', onDocKeydown)
})

// shortcuts: Ctrl+F find, Ctrl+H replace (undo/redo built into Tiptap)
function onDocKeydown(e) {
  if (!(e.ctrlKey || e.metaKey)) return
  const k = e.key.toLowerCase()
  if (k === 'f') {
    e.preventDefault()
    openFindReplace('find')
  } else if (k === 'h') {
    e.preventDefault()
    openFindReplace('replace')
  }
}

// silently apply style config (no dialog/persist, just re-render)
function applyStyleCfgSilent() {
  let el = document.getElementById('pico-doc-styles')
  if (el) return
  el = document.createElement('style')
  el.id = 'pico-doc-styles'
  document.head.appendChild(el)
  const c = styleCfg
  el.textContent =
    '.tiptap h1{font-size:' + c.h1.size + 'px;color:' + c.h1.color + ';font-weight:' + c.h1.weight + ';}' +
    '.tiptap h2{font-size:' + c.h2.size + 'px;color:' + c.h2.color + ';font-weight:' + c.h2.weight + ';}' +
    '.tiptap h3{font-size:' + c.h3.size + 'px;color:' + c.h3.color + ';font-weight:' + c.h3.weight + ';}'
}

onBeforeUnmount(() => {
  clearTimeout(saveTimer)
  clearInterval(heartbeatTimer)
  clearTimeout(reconnectTimer)
  clearTimeout(outlineTimer)
  window.removeEventListener('keydown', onDocKeydown)
  if (ws) {
    ws.onclose = null
    ws.close()
  }
  editor.value?.destroy()
})
</script>

<style scoped>
.ed-page {
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
.style-sel {
  width: 110px;
}
.doc-title {
  font-weight: 600;
  font-size: 14px;
}
.spacer {
  flex: 1;
}
.avatar {
  margin-left: 4px;
  background: #409eff;
  color: #fff;
}
.save-tip {
  font-size: 12px;
  color: #909399;
  margin-right: 8px;
}
.editor-wrap {
  flex: 1;
  overflow: auto;
  background: #e9edf2;
  padding: 24px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.paper-wrap {
  flex: 1;
  min-width: 0;
}
.word-count {
  font-size: 12px;
  color: #606266;
  margin-right: 8px;
}
.fr-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}
.fr-input {
  width: 180px;
}
.fr-count {
  font-size: 12px;
  color: #909399;
  min-width: 50px;
}
.outline-tree {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border: 1px solid #e4e7ed;
  padding: 8px;
  max-height: 70vh;
  overflow: auto;
}
.ot-title {
  font-weight: 600;
  font-size: 13px;
  margin-bottom: 6px;
}
.ot-item {
  font-size: 12px;
  padding: 3px 4px;
  cursor: pointer;
  color: #606266;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ot-item.lv1 { padding-left: 16px; }
.ot-item.lv2 { padding-left: 32px; }
.ot-item:hover { background: #ecf5ff; color: #409eff; }
.ot-empty {
  font-size: 12px;
  color: #909399;
  padding: 8px 4px;
}
.hist-list {
  padding-bottom: 12px;
}
.hist-item {
  padding: 10px;
  border: 1px solid #e4e7ed;
  margin-bottom: 8px;
  cursor: pointer;
}
.hist-item:hover, .hist-item.active {
  border-color: #409eff;
  background: #ecf5ff;
}
.hi-who { font-size: 13px; color: #303133; }
.hi-time { font-size: 12px; color: #909399; margin-top: 2px; }
.hist-preview {
  max-height: 60vh;
  overflow: auto;
  padding: 16px;
  background: #fafafa;
  border: 1px solid #ebeef5;
  font-size: 14px;
  line-height: 1.7;
}
</style>

<style>
/* Tiptap editing area white page; width controlled by inline pageStyle*/
.tiptap {
  margin: 0 auto;
  background: #fff;
  min-height: 80vh;
  outline: none;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.12);
}
.tiptap p {
  margin: 6px 0;
}
/* header/footer*/
.tiptap .doc-header,
.tiptap .doc-footer {
  color: #909399;
  font-size: 12px;
  border-bottom: 1px dashed #dcdfe6;
  padding-bottom: 6px;
  margin-bottom: 16px;
}
.tiptap .doc-footer {
  border-bottom: none;
  border-top: 1px dashed #dcdfe6;
  margin-top: 24px;
  padding-top: 6px;
}
/* table of contents*/
.tiptap .toc {
  background: #fafafa;
  border: 1px solid #ebeef5;
  padding: 12px 16px;
  margin: 12px 0;
}
.tiptap .toc-title {
  font-weight: 600;
  margin: 0 0 6px;
}
.tiptap .toc-item {
  margin: 2px 0;
  font-size: 13px;
  color: #606266;
}
.tiptap .toc-item.lv1 { padding-left: 16px; }
.tiptap .toc-item.lv2 { padding-left: 32px; }
/* image wrap*/
.tiptap img.float-left { float: left; margin: 0 12px 8px 0; max-width: 40%; }
.tiptap img.float-right { float: right; margin: 0 0 8px 12px; max-width: 40%; }
.tiptap img { max-width: 100%; }
/* table*/
.tiptap table {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
}
.tiptap td,
.tiptap th {
  padding: 6px 10px;
}
.tiptap table.pico-tbl-border td,
.tiptap table.pico-tbl-border th {
  border: 1px solid #dcdfe6;
}
.tiptap table.pico-tbl-zebra tbody tr:nth-child(even) {
  background: #f5f7fa;
}
.tiptap table.pico-tbl-headbold th {
  font-weight: 700;
  background: #f0f2f5;
}
/* page break*/
.tiptap .page-break {
  page-break-after: always;
  border-top: 1px dashed #c0c4cc;
  height: 0;
  margin: 12px 0;
}
/* image wrap: margin for square/tight; absolute for behind/front*/
.tiptap img.around { margin: 8px 12px; max-width: 50%; }
.tiptap img.tight { float: left; margin: 4px 14px 4px 0; max-width: 45%; border-radius: 6px; }
.tiptap img.behind {
  position: absolute;
  z-index: -1;
  opacity: 0.7;
  max-width: 60%;
}
.tiptap img.above {
  position: absolute;
  z-index: 10;
  max-width: 45%;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}
/* table border three states*/
.tiptap table.pico-tbl-all td,
.tiptap table.pico-tbl-all th { border: 1px solid #dcdfe6; }
.tiptap table.pico-tbl-none td,
.tiptap table.pico-tbl-none th { border: none; }
.tiptap table.pico-tbl-outer { border: 1px solid #dcdfe6; }
.tiptap table.pico-tbl-outer td,
.tiptap table.pico-tbl-outer th { border: none; }
</style>
