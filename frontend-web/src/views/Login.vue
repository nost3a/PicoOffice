<template>
  <div class="login-wrap">
    <el-card class="login-card" shadow="never">
      <div class="brand">PicoOffice</div>
      <div class="sub">私有化办公套件</div>
      <el-form :model="form" @submit.prevent="onSubmit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="onSubmit"
          />
        </el-form-item>
        <el-button type="primary" :loading="loading" class="btn" @click="onSubmit">登 录</el-button>
      </el-form>
      <div class="foot">
        还没有账号？<router-link to="/register">去注册</router-link>
      </div>
      <div class="foot-links">
        <router-link to="/privacy" target="_blank">隐私政策</router-link>
        <span class="dot">·</span>
        <router-link to="/terms" target="_blank">用户协议</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({ username: '', password: '' })
const loading = ref(false)

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await auth.login(form)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e) {
    // error already toated by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.login-card {
  width: 360px;
}
.brand {
  font-size: 22px;
  font-weight: 600;
  text-align: center;
}
.sub {
  color: #909399;
  font-size: 12px;
  text-align: center;
  margin: 4px 0 20px;
}
.btn {
  width: 100%;
}
.foot {
  margin-top: 14px;
  text-align: center;
  font-size: 13px;
  color: #606266;
}
.foot-links {
  margin-top: 8px;
  text-align: center;
  font-size: 12px;
  color: #909399;
}
.foot-links .dot { margin: 0 6px; }
</style>
