<template>
  <div class="doc">
    <h1>隐私政策</h1>
    <p class="upd">更新日期：2026-10-07</p>

    <h2>一、我们收集的信息</h2>
    <p>您注册时提供的用户名、昵称；您主动上传的文档与文件；为提供同步服务而产生的登录设备记录。</p>

    <h2>二、信息的使用</h2>
    <p>仅用于提供文档存储、协同编辑、邮件同步与日历提醒等核心功能，不用于向第三方广告投放。</p>

    <h2>三、信息存储与安全</h2>
    <p>数据保存在您私有化部署的服务器中，传输使用 HTTPS；敏感配置（如对象存储密钥）以服务端环境变量方式管理，不前端落库。</p>

    <h2>四、Cookie 与本地存储</h2>
    <p>我们在浏览器本地存储登录令牌（pico_token）与语言偏好，用于维持登录态与界面语言选择，不做跨站追踪。</p>

    <h2>五、您的权利</h2>
    <p>您可在「个人资料」页查看与修改信息、管理登录设备、开启两步验证或退出登录。如需删除账号，请联系管理员。</p>

    <h2>六、联系我们</h2>
    <p>如有隐私相关问题，请联系系统管理员。</p>

    <div class="back"><router-link to="/login">返回登录</router-link></div>
  </div>
</template>

<style scoped>
.doc { max-width: 760px; margin: 0 auto; padding: 32px 20px; line-height: 1.8; font-size: 14px; }
.upd { color: #909399; font-size: 12px; }
h1 { font-size: 22px; }
h2 { font-size: 16px; margin-top: 24px; }
.back { margin-top: 32px; }
</style>
