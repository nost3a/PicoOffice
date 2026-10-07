<template>
  <div class="page">
    <div class="topbar">
      <el-button size="small" @click="$router.push('/dashboard')">
        <el-icon class="el-icon--left"><Back /></el-icon>返回
      </el-button>
      <span class="title">我分享的</span>
    </div>

    <el-table
      v-loading="loading"
      :data="list"
      class="tbl"
      :header-cell-style="{ background: '#fafafa' }"
    >
      <el-table-column label="名称" min-width="220">
        <template #default="{ row }">
          <span class="fname">
            <el-icon v-if="row.type === 'doc'" color="#409eff"><Document /></el-icon>
            <el-icon v-else-if="row.type === 'sheet'" color="#67c23a"><Grid /></el-icon>
            <el-icon v-else color="#e6a23c"><VideoPlay /></el-icon>
            {{ row.title }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="权限" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.share_perm === 'edit' ? 'warning' : 'info'" effect="plain">
            {{ row.share_perm === 'edit' ? '可编辑' : '仅查看' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="分享链接" min-width="220">
        <template #default="{ row }">
          <span class="link mono">{{ '/s/' + row.share_token }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" align="right">
        <template #default="{ row }">
          <el-button size="small" text @click="copyLink(row)">复制链接</el-button>
          <el-button size="small" text type="danger" @click="cancelShare(row)">取消分享</el-button>
        </template>
      </el-table-column>

      <template #empty>
        <el-empty description="暂无分享" :image-size="80" />
      </template>
    </el-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiListShared, apiDeleteShare } from '@/api/extras'

const list = ref([])
const loading = ref(false)

function norm(d) {
  return { ...d, type: d.doc_kind || d.type }
}

async function load() {
  loading.value = true
  try {
    const resp = await apiListShared()
    list.value = (resp.list || resp || []).map(norm)
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

function copyLink(row) {
  const url = location.origin + '/s/' + row.share_token
  navigator.clipboard.writeText(url).then(
    () => ElMessage.success('链接已复制'),
    () => ElMessage.warning('复制失败')
  )
}

async function cancelShare(row) {
  try {
    await ElMessageBox.confirm('取消「' + row.title + '」的分享？链接将立即失效。', '取消分享', {
      type: 'warning'
    })
  } catch (e) {
    return
  }
  await apiDeleteShare(row.id)
  ElMessage.success('已取消分享')
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
.tbl {
  background: #fff;
}
.fname {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.link {
  font-size: 12px;
  color: #409eff;
}
</style>
