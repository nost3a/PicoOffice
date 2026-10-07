# 安全政策 / Security Policy

本项目重视安全。如果你发现了安全漏洞，请**不要**直接发公开 Issue，按下面的方式私下报告。

This project takes security seriously. If you discover a security vulnerability, please **do not** open a public Issue; report it privately using the methods below.

> 安全联系邮箱：`chengsigzs@foxmail.com`
>
> Security contact email: `chengsigzs@foxmail.com`

## 支持的版本 / Supported Versions

| 版本 / Version | 是否提供安全更新 / Security Updates Provided |
| ---- | ---------------- |
| v0.3.x | ✅ 维护中<br>✅ Maintained |
| v0.2.x | ⚠️ 仅严重问题，建议尽快升级<br>⚠️ Critical issues only, upgrade recommended |
| v0.1.x | ❌ 已停止维护<br>❌ No longer maintained |
| main 分支 / main branch | ✅ 开发主线<br>✅ Development mainline |

当前正式发布版本为 **v0.3.0**。

The current officially released version is **v0.3.0**.

## 上报漏洞 / Reporting a Vulnerability

推荐两种方式，任选其一：

We recommend two methods; choose either one:

1. **GitHub Security Advisories**（推荐，加密且私有）
   - **GitHub Security Advisories** (recommended, encrypted and private)
   在仓库页面 `Security` → `Report a vulnerability` 提交私有报告。
   - On the repo page `Security` → `Report a vulnerability`, submit a private report.
2. **私有邮件**
   - **Private email**
   发送到 `chengsigzs@foxmail.com`，邮件主题建议带 `[picooffice-security]` 前缀。
   - Send to `chengsigzs@foxmail.com`, with a subject prefixed with `[picooffice-security]` recommended.

报告里尽量包含：

Please include the following in your report where possible:

- 受影响的版本 / commit；
  - Affected version / commit.
- 复现步骤（越简单越好）；
  - Steps to reproduce (the simpler the better).
- 影响范围与严重程度评估；
  - Impact scope and severity assessment.
- 你希望如何署名（公开披露时是否保留匿名）。
  - How you would like to be credited (whether to remain anonymous at public disclosure).

## 响应策略 / Response Policy

| 阶段 / Stage | 预期时间 / Expected Time |
| ---- | -------- |
| 确认收到 / Acknowledgment | 3 个工作日内回复确认<br>Reply to confirm within 3 business days |
| 风险评估 / Risk assessment | 确认后 7 个工作日内给出初判<br>Initial assessment within 7 business days after confirmation |
| 修复与验证 / Fix & verification | 高危尽快出补丁，普通问题随下个版本发布<br>Critical issues patched ASAP; normal issues shipped with the next release |
| 公开披露 / Public disclosure | 修复发布并给用户留出升级窗口后再披露<br>Disclosed after the fix is released and users have an upgrade window |

我们会在修复版本的 CHANGELOG 中致谢报告者（除非你要求匿名）。

We will credit the reporter in the CHANGELOG of the fix release (unless you request anonymity).

## 安全相关注意事项（部署方必读） / Security Notes for Deployers

PicoOffice 是自托管软件，很多安全基线取决于你自己的部署方式：

PicoOffice is self-hosted software; many security baselines depend on how you deploy it:

- **JWT 密钥**：生产环境务必通过环境变量 `PICO_JWT_SECRET` 设置一串足够长的随机值，**不要**用默认/空值启动。启动时若检测到弱密钥会打警告日志。
  - **JWT secret**: In production, always set a sufficiently long random value via the `PICO_JWT_SECRET` environment variable, and **do not** start with the default/empty value. A warning log is emitted at startup if a weak secret is detected.
- **2FA（TOTP）**：管理员应优先开启两步验证。首次开启后请妥善保存备用恢复码。
  - **2FA (TOTP)**: Administrators should enable two-factor authentication first. After first enabling it, store the backup recovery codes safely.
- **上传与导出文件消毒**：富文本内容经 `bluemonday` 做 HTML 消毒后再渲染；导入的 `docx/xlsx/pptx` 走 LibreOffice 转换，转换目录与运行用户权限要收紧。
  - **Upload & export file sanitization**: Rich-text content is sanitized with `bluemonday` before rendering; imported `docx/xlsx/pptx` go through LibreOffice conversion, and the conversion directory and runtime user permissions should be tightened.
- **登录限流**：登录接口有失败次数限制与锁定策略，避免暴力破解。若前面还有反向代理，请把真实客户端 IP 正确透传（`X-Forwarded-For`）。
  - **Login rate limiting**: The login endpoint has failure-count limits and a lockout policy to prevent brute force. If a reverse proxy sits in front, correctly forward the real client IP (`X-Forwarded-For`).
- **传输加密**：公网部署务必套 HTTPS（反向代理或自带 TLS），别把明文接口暴露出去。
  - **Transport encryption**: For public deployments, always put HTTPS in front (reverse proxy or built-in TLS); do not expose plaintext endpoints.
- **数据库与备份**：SQLite 数据库文件、备份文件（`picobackup` 产物）包含全部用户数据，文件权限建议限制到运行用户可读，不要放进 Web 可直接访问的目录。
  - **Database & backups**: The SQLite database file and backup files (`picobackup` output) contain all user data. Restrict file permissions to the runtime user only, and do not place them in a directory directly accessible by the web.
- **依赖更新**：定期 `go get -u` / `npm update` 跟进上游安全补丁。
  - **Dependency updates**: Periodically run `go get -u` / `npm update` to pick up upstream security patches.

发现疑似被利用的情况，请第一时间轮换 JWT 密钥、强制全员重新登录，并清理异常会话。

If you suspect exploitation, immediately rotate the JWT secret, force all users to re-login, and purge anomalous sessions.
