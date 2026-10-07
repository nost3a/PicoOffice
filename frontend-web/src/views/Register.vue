<template>
  <div class="login-wrap">
    <el-card class="login-card" shadow="never">
      <div class="brand">注册 PicoOffice</div>
      <el-form :model="form">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.nickname" placeholder="昵称（选填）" prefix-icon="Star" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="agree">
            我已阅读并同意
            <router-link to="/privacy" target="_blank">《隐私政策》</router-link>与
            <router-link to="/terms" target="_blank">《用户协议》</router-link>
          </el-checkbox>
        </el-form-item>
        <el-button type="primary" :loading="loading" class="btn" @click="onSubmit">注 册</el-button>
      </el-form>
      <div class="foot">
        已有账号？<router-link to="/login">去登录</router-link>
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
const form = reactive({ username: '', password: '', nickname: '' })
const loading = ref(false)
const agree = ref(false)

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  if (!agree.value) {
    ElMessage.warning('请先勾选并同意隐私政策与用户协议')
    return
  }
  loading.value = true
  try {
    await auth.register(form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (e) {
    // interceptor already warned
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
  font-size: 20px;
  font-weight: 600;
  text-align: center;
  margin-bottom: 20px;
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
</style>
