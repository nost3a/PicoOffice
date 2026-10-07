<template>
  <div class="page">
    <div class="page-head">
      <span class="title">{{ $t('mail.title') }}</span>
      <div>
        <el-button size="small" @click="openCompose">{{ $t('mail.compose') }}</el-button>
        <el-button size="small" type="primary" @click="openAccount">{{ $t('mail.addAccount') }}</el-button>
      </div>
    </div>

    <el-row :gutter="12">
      <!-- accounts + list-->
      <el-col :xs="24" :md="10">
        <el-card shadow="never" class="mb">
          <div class="card-title">{{ $t('mail.accounts') }}</div>
          <el-select v-model="curAccount" class="full" @change="loadMessages">
            <el-option v-for="a in accounts" :key="a.id" :label="a.addr || a.email" :value="a.id" />
          </el-select>
          <div v-if="!accounts.length" class="tip mt">{{ $t('mail.noServer') }}</div>
        </el-card>

        <el-card shadow="never">
          <div class="card-title">
            {{ $t('mail.messages') }}
            <el-button size="small" link type="primary" @click="onSync">{{ $t('mail.sync') }}</el-button>
          </div>
          <div v-if="!messages.length" class="tip">{{ $t('common.empty') }}</div>
          <div
            v-for="m in messages"
            :key="m.id"
            class="mail-item"
            :class="{ active: m.id === curMsg }"
            @click="openMail(m)"
          >
            <div class="subj">{{ m.subject || '(no subject)' }}</div>
            <div class="tip">{{ m.from || '' }} · {{ m.date || '' }}</div>
          </div>
        </el-card>
      </el-col>

      <!-- body-->
      <el-col :xs="24" :md="14">
        <el-card shadow="never" class="h-full">
          <div v-if="!cur">{{ $t('mail.selectToRead') }}</div>
          <template v-else>
            <div class="subj-lg">{{ cur.subject }}</div>
            <div class="tip">{{ $t('mail.from') }}: {{ cur.from }} · {{ cur.date }}</div>
            <el-divider />
            <div class="body" v-html="renderBody(cur)"></div>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <!-- account dialog-->
    <el-dialog v-model="acctDlg" :title="$t('mail.accountEdit')" width="420px">
      <el-form label-position="top">
        <el-form-item :label="$t('mail.addr')">
          <el-input v-model="acct.addr" />
        </el-form-item>
        <el-form-item :label="$t('mail.imap')">
          <el-input v-model="acct.imap_host" /><el-input v-model="acct.imap_port" class="mt" />
        </el-form-item>
        <el-form-item :label="$t('mail.smtp')">
          <el-input v-model="acct.smtp_host" /><el-input v-model="acct.smtp_port" class="mt" />
        </el-form-item>
        <el-form-item :label="$t('mail.password')">
          <el-input v-model="acct.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="acctDlg = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="onSaveAccount">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- compose dialog-->
    <el-dialog v-model="composeDlg" :title="$t('mail.compose')" width="520px">
      <el-form label-position="top">
        <el-form-item :label="$t('mail.to')">
          <el-input v-model="compose.to" />
        </el-form-item>
        <el-form-item :label="$t('mail.subject')">
          <el-input v-model="compose.subject" />
        </el-form-item>
        <el-form-item :label="$t('mail.body')">
          <el-input v-model="compose.body" type="textarea" :rows="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="composeDlg = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="onSend">{{ $t('mail.send') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '@/i18n'
import {
  apiListMailAccounts, apiCreateMailAccount,
  apiSyncMail, apiListMailMessages, apiMailBody, apiMailSend
} from '@/api/modules'

const { t } = useI18n()
const accounts = ref([])
const curAccount = ref('')
const messages = ref([])
const curMsg = ref('')
const cur = ref(null)
const acctDlg = ref(false)
const composeDlg = ref(false)
const acct = reactive({ addr: '', imap_host: '', imap_port: 993, smtp_host: '', smtp_port: 465, password: '' })
const compose = reactive({ to: '', subject: '', body: '' })

onMounted(loadAccounts)

async function loadAccounts() {
  try {
    const r = await apiListMailAccounts()
    accounts.value = r.accounts || r || []
    if (accounts.value.length) {
      curAccount.value = accounts.value[0].id
      loadMessages()
    }
  } catch (e) {}
}
async function loadMessages() {
  if (!curAccount.value) return
  try {
    const r = await apiListMailMessages({ account_id: curAccount.value })
    messages.value = r.messages || r || []
  } catch (e) {}
}
async function onSync() {
  if (!curAccount.value) { ElMessage.warning(t('mail.noServer')); return }
  try {
    await apiSyncMail(curAccount.value)
    ElMessage.success(t('common.success'))
    loadMessages()
  } catch (e) {}
}
async function openMail(m) {
  curMsg.value = m.id
  cur.value = m
  if (m.body_html || m.body_text) return
  try {
    cur.value = await apiMailBody(m.id)
  } catch (e) {}
}
function renderBody(m) {
  const raw = m.body_html || m.body_text || ''
  // escape plain text simply; html from backend is trusted
  if (m.body_html) return raw
  return String(raw).replace(/</g, '&lt;').replace(/\n/g, '<br>')
}
function openAccount() { acctDlg.value = true }
async function onSaveAccount() {
  try {
    await apiCreateMailAccount({ ...acct })
    ElMessage.success(t('common.success'))
    acctDlg.value = false
    loadAccounts()
  } catch (e) {}
}
function openCompose() {
  if (!accounts.value.length) { ElMessage.warning(t('mail.noServer')); return }
  composeDlg.value = true
}
async function onSend() {
  try {
    await apiMailSend({ ...compose, account_id: curAccount.value })
    ElMessage.success(t('common.success'))
    composeDlg.value = false
  } catch (e) {}
}
</script>

<style scoped>
.page { padding: 16px; max-width: 1200px; margin: 0 auto; }
.page-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.title { font-size: 16px; font-weight: 600; }
.full { width: 100%; }
.mt { margin-top: 8px; }
.mb { margin-bottom: 12px; }
.h-full { min-height: 60vh; }
.tip { color: #909399; font-size: 12px; }
.card-title { font-weight: 600; margin-bottom: 8px; display: flex; justify-content: space-between; align-items: center; }
.mail-item { padding: 8px; border-bottom: 1px solid #f0f0f0; cursor: pointer; border-radius: 4px; }
.mail-item.active, .mail-item:active { background: #ecf5ff; }
.subj { font-size: 13px; }
.subj-lg { font-size: 16px; font-weight: 600; margin-bottom: 4px; }
.body { font-size: 13px; line-height: 1.6; }
</style>
