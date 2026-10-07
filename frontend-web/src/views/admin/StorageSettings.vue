<template>
  <div class="page">
    <div class="page-head">
      <span class="title">{{ $t('storage.title') }}</span>
    </div>

    <el-alert :title="$t('storage.envTip')" type="warning" :closable="false" class="mb" />

    <el-card shadow="never">
      <el-form label-position="top" style="max-width: 520px">
        <el-form-item :label="$t('storage.endpoint')">
          <el-input v-model="form.endpoint" placeholder="https://s3.example.com" />
        </el-form-item>
        <el-form-item :label="$t('storage.bucket')">
          <el-input v-model="form.bucket" />
        </el-form-item>
        <el-form-item :label="$t('storage.region')">
          <el-input v-model="form.region" placeholder="us-east-1" />
        </el-form-item>
        <el-form-item :label="$t('storage.ak')">
          <el-input v-model="form.ak" />
        </el-form-item>
        <el-form-item :label="$t('storage.sk')">
          <el-input v-model="form.sk" type="password" show-password />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.ssl">{{ $t('storage.ssl') }}</el-checkbox>
        </el-form-item>
        <el-button @click="onTest">{{ $t('storage.test') }}</el-button>
        <el-button type="primary" @click="onSave">{{ $t('common.save') }}</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/i18n'
import { apiGetStorage, apiSaveStorage, apiTestStorage } from '@/api/modules'

const { t } = useI18n()
const form = reactive({ endpoint: '', bucket: '', region: '', ak: '', sk: '', ssl: true })

onMounted(async () => {
  try {
    const r = await apiGetStorage()
    Object.assign(form, r.config || r)
  } catch (e) {
    // read route may not exist; keep empty form
  }
})

async function onTest() {
  try {
    await apiTestStorage({ ...form })
    ElMessage.success(t('storage.testOk'))
  } catch (e) {}
}
async function onSave() {
  try {
    await apiSaveStorage({ ...form })
    ElMessage.success(t('storage.saved'))
  } catch (e) {
    // interceptor already warns when save route missing; echoes envTip above
  }
}
</script>

<style scoped>
.page { padding: 16px; max-width: 900px; margin: 0 auto; }
.page-head { margin-bottom: 12px; }
.title { font-size: 16px; font-weight: 600; }
.mb { margin-bottom: 12px; }
</style>
