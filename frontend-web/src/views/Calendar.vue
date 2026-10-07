<template>
  <div class="page">
    <div class="page-head">
      <div class="head-left">
        <el-button size="small" @click="shift(-1)">{{ $t('calendar.prevMonth') }}</el-button>
        <span class="ym">{{ year }}年{{ month + 1 }}月</span>
        <el-button size="small" @click="shift(1)">{{ $t('calendar.nextMonth') }}</el-button>
        <el-button size="small" @click="goToday">{{ $t('calendar.today') }}</el-button>
      </div>
      <div>
        <el-button size="small" @click="reqNotify">🔔 {{ notifyOk ? $t('calendar.notifyOn') : $t('calendar.notifyOff') }}</el-button>
        <el-button size="small" type="primary" @click="openEvent(null)">{{ $t('calendar.newEvent') }}</el-button>
      </div>
    </div>

    <div class="grid">
      <div v-for="w in weekDays" :key="w" class="dow">{{ w }}</div>
      <div
        v-for="(cell, i) in cells"
        :key="i"
        class="cell"
        :class="{ muted: cell.month !== month, today: cell.isToday }"
        @click="openEvent(null, cell.date)"
      >
        <div class="dnum">{{ cell.day }}</div>
        <div
          v-for="e in eventsOf(cell.date)"
          :key="e.id"
          class="ev"
          @click.stop="openEvent(e)"
        >
          {{ e.title }}
        </div>
      </div>
    </div>

    <el-dialog v-model="dlg" :title="editing.id ? $t('common.edit') : $t('calendar.newEvent')" width="420px">
      <el-form label-position="top">
        <el-form-item :label="$t('calendar.eventTitle')">
          <el-input v-model="editing.title" />
        </el-form-item>
        <el-form-item :label="$t('calendar.eventStart')">
          <el-date-picker v-model="editing.start" type="datetime" />
        </el-form-item>
        <el-form-item :label="$t('calendar.eventEnd')">
          <el-date-picker v-model="editing.end" type="datetime" />
        </el-form-item>
        <el-form-item :label="$t('calendar.remind')">
          <el-input-number v-model="editing.remind_min" :min="0" :step="5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button v-if="editing.id" type="danger" @click="onDel">{{ $t('common.delete') }}</el-button>
        <el-button @click="dlg = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="onSave">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from '@/i18n'
import { apiListEvents, apiCreateEvent, apiUpdateEvent, apiDeleteEvent } from '@/api/modules'

const { t } = useI18n()
const now = new Date()
const year = ref(now.getFullYear())
const month = ref(now.getMonth())
const events = ref([])
const dlg = ref(false)
const notifyOk = ref(false)
const editing = reactive({ id: null, title: '', start: '', end: '', remind_min: 5 })

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

// build month grid: 42 cells
const cells = computed(() => {
  const first = new Date(year.value, month.value, 1)
  const startDay = first.getDay()
  const daysInMonth = new Date(year.value, month.value + 1, 0).getDate()
  const arr = []
  // prev-month filler
  const prevDays = new Date(year.value, month.value, 0).getDate()
  for (let i = startDay - 1; i >= 0; i--) {
    arr.push({ day: prevDays - i, month: month.value - 1, date: fmt(new Date(year.value, month.value - 1, prevDays - i)) })
  }
  const t0 = new Date()
  for (let d = 1; d <= daysInMonth; d++) {
    const date = new Date(year.value, month.value, d)
    arr.push({ day: d, month: month.value, date: fmt(date), isToday: sameDay(date, t0) })
  }
  // next-month filler to 42 cells
  let next = 1
  while (arr.length % 7 !== 0 || arr.length < 42) {
    arr.push({ day: next, month: month.value + 1, date: fmt(new Date(year.value, month.value + 1, next)) })
    next++
    if (arr.length >= 42) break
  }
  return arr
})

function fmt(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0')
}
function sameDay(a, b) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

function eventsOf(date) {
  return events.value.filter((e) => (e.start || '').startsWith(date))
}

function shift(n) {
  let m = month.value + n
  let y = year.value
  if (m < 0) { m = 11; y-- }
  if (m > 11) { m = 0; y++ }
  month.value = m; year.value = y
  load()
}
function goToday() {
  year.value = now.getFullYear(); month.value = now.getMonth()
  load()
}

async function load() {
  const from = fmt(new Date(year.value, month.value - 1, 1))
  const to = fmt(new Date(year.value, month.value + 2, 0))
  try {
    const r = await apiListEvents({ from, to })
    events.value = r.events || r || []
    scheduleNotify()
  } catch (e) {}
}

function openEvent(e, presetDate) {
  if (e) {
    Object.assign(editing, { id: e.id, title: e.title, start: e.start, end: e.end, remind_min: e.remind_min || 0 })
  } else {
    const base = presetDate || fmt(new Date())
    Object.assign(editing, { id: null, title: '', start: base + 'T09:00', end: base + 'T10:00', remind_min: 5 })
  }
  dlg.value = true
}

async function onSave() {
  try {
    if (editing.id) {
      await apiUpdateEvent(editing.id, { ...editing })
    } else {
      await apiCreateEvent({ ...editing })
    }
    ElMessage.success(t('common.success'))
    dlg.value = false
    load()
  } catch (e) {}
}
async function onDel() {
  try {
    await ElMessageBox.confirm(t('common.delete') + '?', t('common.confirm'), { type: 'warning' })
    await apiDeleteEvent(editing.id)
    dlg.value = false
    load()
  } catch (e) {}
}

// desktop notification: remind at event time
async function reqNotify() {
  if (!('Notification' in window)) return
  const p = await Notification.requestPermission()
  notifyOk.value = p === 'granted'
}
function scheduleNotify() {
  if (!('Notification' in window) || Notification.permission !== 'granted') return
  notifyOk.value = true
  const nowTs = Date.now()
  events.value.forEach((e) => {
    const remind = Number(e.remind_min || 0)
    const fireAt = new Date(e.start).getTime() - remind * 60000
    if (fireAt > nowTs && fireAt - nowTs < 6 * 3600 * 1000) {
      setTimeout(() => new Notification(e.title || '日程提醒', { body: e.start }), fireAt - nowTs)
    }
  })
}

onMounted(() => {
  load()
  if ('Notification' in window) notifyOk.value = Notification.permission === 'granted'
})
</script>

<style scoped>
.page { padding: 16px; max-width: 1100px; margin: 0 auto; }
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.head-left { display: flex; align-items: center; gap: 8px; }
.ym { font-weight: 600; min-width: 90px; }
.grid { display: grid; grid-template-columns: repeat(7, 1fr); border: 1px solid #ebeef5; border-radius: 6px; overflow: hidden; }
.dow { background: #f5f7fa; padding: 6px; text-align: center; font-size: 12px; font-weight: 600; }
.cell { min-height: 84px; border-right: 1px solid #ebeef5; border-bottom: 1px solid #ebeef5; padding: 4px; cursor: pointer; }
.cell:nth-child(7n) { border-right: none; }
.cell.muted { background: #fafafa; }
.cell.today { background: #ecf5ff; }
.dnum { font-size: 12px; color: #606266; }
.ev { background: #409eff; color: #fff; font-size: 11px; padding: 1px 4px; border-radius: 3px; margin-top: 2px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
@media (max-width: 640px) {
  .cell { min-height: 56px; }
}
</style>
