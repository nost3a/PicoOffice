<template>
  <div class="page">
    <div class="topbar">
      <el-button size="small" @click="$router.push('/dashboard')">
        <el-icon class="el-icon--left"><Back /></el-icon>返回
      </el-button>
      <span class="title">回收站</span>
      <div class="spacer" />
      <span class="tip">删除的文档保留在此，还原后可继续编辑</span>
    </div>

    <el-table
      v-loading="loading"
      :data="list"
      class="tbl"
      :header-cell-style="{ background: '#fafafa' }"
    >
      <el-table-column label="名称" min-width="240">
        <template #default="{ row }">
          <span class="fname">
            <el-icon v-if="row.type === 'doc'" color="#409eff"><Document /></el-icon>
            <el-icon v-else-if="row.type === 'sheet'" color="#67c23a"><Grid /></el-icon>
            <el-icon v-else color="#e6a23c"><VideoPlay /></el-icon>
            {{ row.title }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="90">
        <template #default="{ row }">
          <el-tag size="small" effect="plain">{{ typeText(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="删除时间" width="180">
        <template #default="{ row }">{{ fmtTime(row.deleted_at || row.updated_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="right">
        <template #default="{ row }">
          <el-button size="small" text type="primary" @click="onRestore(row)">还原</el-button>
          <el-button size="small" text type="danger" @click="onPurge(row)">彻底删除</el-button>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="回收站为空" :image-size="80" />
      </template>
    </el-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiListTrash, apiRestoreTrash, apiPurgeTrash } from '@/api/extras'

const list = ref([])
const loading = ref(false)

function typeText(t) {
  return t === 'doc' ? '文档' : t === 'sheet' ? '表格' : t === 'slide' ? '演示' : t || '-'
}

function fmtTime(t) {
  if (!t) return '-'
  const d = new Date(t)
  if (isNaN(d.getTime())) return '-'
  const p = (n) => String(n).padStart(2, '0')
  return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) +
    ' ' + p(d.getHours()) + ':' + p(d.getMinutes())
}

function norm(d) {
  return { ...d, type: d.doc_kind || d.type }
}

async function load() {
  loading.value = true
  try {
    const resp = await apiListTrash()
    list.value = (resp.list || resp || []).map(norm)
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function onRestore(row) {
  try {
    await apiRestoreTrash(row.id)
    ElMessage.success('已还原')
    load()
  } catch (e) {
    // interceptor already warned
  }
}

async function onPurge(row) {
  try {
    await ElMessageBox.confirm(
      '彻底删除「' + row.title + '」后将无法恢复，确认删除？',
      '彻底删除',
      { type: 'warning', confirmButtonText: '彻底删除', confirmButtonClass: 'el-button--danger' }
    )
  } catch (e) {
    return
  }
  await apiPurgeTrash(row.id)
  ElMessage.success('已彻底删除')
  load()
}

onMounted(load)
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
  margin-bottom: 12px;
}
.title {
  font-size: 16px;
  font-weight: 600;
}
.spacer {
  flex: 1;
}
.tip {
  font-size: 12px;
  color: #909399;
}
.tbl {
  background: #fff;
}
.fname {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
</style>
