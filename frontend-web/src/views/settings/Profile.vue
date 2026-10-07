<template>
  <div class="page">
    <div class="page-head">
      <span class="title">{{ $t('profile.title') }}</span>
      <el-radio-group v-model="lang" size="small" class="lang-switch" @change="onLang">
        <el-radio-button value="zh">{{ $t('lang.zh') }}</el-radio-button>
        <el-radio-button value="en">{{ $t('lang.en') }}</el-radio-button>
      </el-radio-group>
    </div>

    <el-row :gutter="16">
      <!-- profile card-->
      <el-col :xs="24" :sm="12">
        <el-card shadow="never">
          <div class="avatar-row">
            <el-avatar :size="64" :src="profile.avatar">
              {{ (profile.nickname || '?').slice(0, 1) }}
            </el-avatar>
            <div class="avatar-ops">
              <el-button size="small" @click="fileInput.click()">
                {{ $t('profile.uploadAvatar') }}
              </el-button>
              <div class="tip">{{ $t('profile.avatarTip') }}</div>
              <input
                ref="fileInput"
                type="file"
                accept="image/*"
                style="display:none"
                @change="onAvatar"
              />
            </div>
          </div>
          <el-form label-position="top" class="mt">
            <el-form-item :label="$t('common.nickname')">
              <el-input v-model="profile.nickname" />
            </el-form-item>
            <el-form-item :label="$t('common.bio')">
              <el-input v-model="profile.bio" type="textarea" :rows="3" />
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="onSave">
              {{ $t('common.save') }}
            </el-button>
          </el-form>
        </el-card>

        <!-- forgot password-->
        <el-card shadow="never" class="mt">
          <div class="card-title">{{ $t('auth.forgotPassword') }}</div>
          <div class="tip">{{ $t('auth.forgotTip') }}</div>
          <el-input v-model="forgotUser" :placeholder="$t('auth.username')" class="mt" />
          <el-button size="small" class="mt" @click="onForgot">{{ $t('auth.sendLink') }}</el-button>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12">
        <!-- 2FA-->
        <el-card shadow="never">
          <div class="card-title">
            {{ $t('profile.totp') }}
            <el-tag size="small" :type="totpOn ? 'success' : 'info'">
              {{ totpOn ? $t('profile.totpOn') : $t('profile.totpOff') }}
            </el-tag>
          </div>

          <template v-if="!totpOn">
            <el-button size="small" @click="onSetup">{{ $t('profile.totpSetup') }}</el-button>
            <template v-if="setup">
              <div class="tip mt">{{ $t('profile.totpScanTip') }}</div>
              <div class="mono box">{{ setup.secret || setup.otpauth_url }}</div>
              <div class="otp mono box">{{ setup.otpauth_url }}</div>
              <el-input v-model="totpCode" :placeholder="$t('profile.totpCode')" class="mt" maxlength="6" />
              <el-button size="small" type="primary" class="mt" @click="onEnable">
                {{ $t('profile.totpEnable') }}
              </el-button>
            </template>
          </template>
          <template v-else>
            <el-input v-model="disableCode" :placeholder="$t('profile.totpCode')" />
            <el-button size="small" type="danger" class="mt" @click="onDisable">{{ $t('profile.totpDisable') }}</el-button>
          </template>
        </el-card>

        <!-- devices-->
        <el-card shadow="never" class="mt">
          <div class="card-title">{{ $t('profile.devices') }}</div>
          <div v-for="d in devices" :key="d.id" class="dev-row">
            <div>
              <span>{{ d.ua || d.name || 'device' }}</span>
              <el-tag v-if="d.current" size="small" type="success">{{ $t('profile.deviceCurrent') }}</el-tag>
            </div>
            <div class="dev-right">
              <span class="tip">{{ $t('profile.deviceLast') }}: {{ d.last_active || '-' }}</span>
              <el-button size="small" link type="danger" :disabled="d.current" @click="onDelDevice(d)">
                {{ $t('profile.deviceLogout') }}
              </el-button>
            </div>
          </div>
        </el-card>

        <el-button type="danger" class="mt" @click="onLogout">{{ $t('auth.logout') }}</el-button>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from '@/i18n'
import { useAuthStore } from '@/stores/auth'
import {
  apiGetProfile, apiPutProfile, apiUploadAvatar,
  apiTotpSetup, apiTotpEnable, apiTotpDisable,
  apiListDevices, apiDeleteDevice,
  apiLogout, apiForgotPassword
} from '@/api/modules'

const { t, state, setLocale } = useI18n()
const router = useRouter()
const auth = useAuthStore()

const profile = reactive({ nickname: '', bio: '', avatar: '' })
const saving = ref(false)
const devices = ref([])
const setup = ref(null)
const totpOn = ref(false)
const totpCode = ref('')
const disableCode = ref('')
const forgotUser = ref('')
const fileInput = ref(null)
const lang = ref(state.lang)

function onLang(v) { setLocale(v) }

onMounted(async () => {
  try {
    const r = await apiGetProfile()
    Object.assign(profile, r.user || r)
  } catch (e) {}
  loadDevices()
})

async function loadDevices() {
  try {
    const r = await apiListDevices()
    devices.value = r.devices || r || []
  } catch (e) {}
}

async function onSave() {
  saving.value = true
  try {
    await apiPutProfile({ nickname: profile.nickname, bio: profile.bio })
    ElMessage.success(t('profile.saved'))
  } catch (e) {} finally { saving.value = false }
}

async function onAvatar(e) {
  const f = e.target.files[0]
  if (!f) return
  try {
    const r = await apiUploadAvatar(f)
    profile.avatar = r.url || r.avatar || (profile.avatar + '?t=' + Date.now())
    ElMessage.success(t('common.success'))
  } catch (e) {}
}

async function onSetup() {
  try {
    setup.value = await apiTotpSetup()
  } catch (e) {}
}
async function onEnable() {
  try {
    await apiTotpEnable(totpCode.value)
    totpOn.value = true
    setup.value = null
    ElMessage.success(t('common.success'))
  } catch (e) {}
}
async function onDisable() {
  try {
    await apiTotpDisable(disableCode.value)
    totpOn.value = false
    disableCode.value = ''
    ElMessage.success(t('common.success'))
  } catch (e) {}
}

async function onDelDevice(d) {
  try {
    await ElMessageBox.confirm(t('common.delete') + '?', t('common.confirm'), { type: 'warning' })
    await apiDeleteDevice(d.id)
    loadDevices()
  } catch (e) {}
}

async function onForgot() {
  if (!forgotUser.value) return
  try {
    await apiForgotPassword({ username: forgotUser.value })
    ElMessage.success(t('common.success'))
  } catch (e) {}
}

async function onLogout() {
  try {
    // refresh_token not stored client-side; backend clears on 401; send anyway as fallback
    await apiLogout(localStorage.getItem('pico_refresh') || '')
  } catch (e) {}
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.page { padding: 16px; max-width: 1100px; margin: 0 auto; }
.page-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.title { font-size: 16px; font-weight: 600; }
.mt { margin-top: 12px; }
.avatar-row { display: flex; gap: 12px; align-items: center; }
.tip { color: #909399; font-size: 12px; }
.card-title { font-weight: 600; margin-bottom: 8px; display: flex; align-items: center; gap: 8px; }
.box { background: #f5f7fa; padding: 6px 8px; border-radius: 4px; word-break: break-all; font-size: 12px; }
.otp { margin-top: 6px; }
.dev-row { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid #f0f0f0; font-size: 13px; }
.dev-right { display: flex; align-items: center; gap: 8px; }
</style>
