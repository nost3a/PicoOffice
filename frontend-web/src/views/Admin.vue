<template>
  <div class="page">
    <div class="topbar">
      <el-button size="small" link @click="$router.push('/dashboard')">
        <el-icon><Back /></el-icon>
      </el-button>
      <span class="title">用户管理</span>
    </div>
    <el-table v-loading="loading" :data="list" class="tbl">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.role === 'admin' ? 'danger' : ''">
            {{ row.role === 'admin' ? '管理员' : '普通用户' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="注册时间" />
      <el-table-column label="操作" width="180" align="right">
        <template #default="{ row }">
          <el-button size="small" text @click="openQuota(row)">改配额</el-button>
          <el-button size="small" text type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- quota dialog-->
    <el-dialog v-model="quotaDlg" title="调整用户配额" width="380px">
      <el-form label-width="90px" size="small">
        <el-form-item label="用户">
          <span>{{ quotaRow?.username }}</span>
        </el-form-item>
        <el-form-item label="文档数">
          <el-input-number v-model="quotaForm.docs" :min="0" />
        </el-form-item>
        <el-form-item label="空间 MB">
          <el-input-number v-model="quotaForm.storage_mb" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button size="small" @click="quotaDlg = false">取消</el-button>
        <el-button size="small" type="primary" @click="saveQuota">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { apiListUsers, apiDeleteUser, apiSetUserQuota } from '@/api'

const list = ref([])
const loading = ref(false)

// quota dialog state
const quotaDlg = ref(false)
const quotaRow = ref(null)
const quotaForm = ref({ docs: 100, storage_mb: 1024 })

async function loadList() {
  loading.value = true
  try {
    const resp = await apiListUsers()
    list.value = resp.list || resp || []
  } catch (e) {
    // interceptor already warned
  } finally {
    loading.value = false
  }
}

function openQuota(row) {
  quotaRow.value = row
  // fetch user's quota, default if missing
  quotaForm.value = {
    docs: row.quota_docs ?? row.quota?.docs ?? 100,
    storage_mb: row.quota_storage_mb ?? row.quota?.storage_mb ?? 1024
  }
  quotaDlg.value = true
}

async function saveQuota() {
  try {
    await apiSetUserQuota(quotaRow.value.id, {
      docs: quotaForm.value.docs,
      storage_mb: quotaForm.value.storage_mb
    })
    ElMessage.success('已更新配额')
    quotaDlg.value = false
    loadList()
  } catch (e) {
    // interceptor already warned
  }
}

async function onDelete(row) {
  await ElMessageBox.confirm('确认删除用户 ' + row.username + '？', '提示', { type: 'warning' })
  await apiDeleteUser(row.id)
  ElMessage.success('已删除')
  loadList()
}

onMounted(loadList)
</script>

<style scoped>
.page {
  height: 100%;
  padding: 12px 16px;
  box-sizing: border-box;
}
.topbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.title {
  font-size: 16px;
  font-weight: 600;
}
.tbl {
  background: #fff;
}
</style>
