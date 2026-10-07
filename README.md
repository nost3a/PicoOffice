# PicoOffice

> 私有化、自托管的办公协作套件——文字 / 表格 / 演示三合一，数据完全留在你自己的服务器上。
>
> Self-hosted private office suite — documents / spreadsheets / slides in one, your data stays entirely on your own server.

**PicoOffice** 是一套能塞进一台小 VPS、甚至树莓派里跑起来的在线办公系统。一个单二进制 + 一个数据库文件，就能给你一套带协同、邮件、日历、全文搜索的私有 Office。不依赖任何第三方 SaaS，不上云，数据归你自己管。

**PicoOffice** is an online office suite compact enough to run on a small VPS — even a Raspberry Pi. A single binary plus one database file gives you a private Office with collaboration, mail, calendar, and full-text search. No third-party SaaS, no cloud dependency; your data stays under your control.

> 本项目仓库地址：`github.com/nost3a/PicoOffice`（下文命令中的仓库路径均已填好，可直接复制使用）。
>
> Repository: `github.com/nost3a/PicoOffice` (repo paths in the commands below are already filled in and can be copied directly).

![license](https://img.shields.io/badge/license-MIT-blue.svg)
[![build](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![docker](https://img.shields.io/badge/docker-pull%202.3k-blue.svg)](#)
![version](https://img.shields.io/badge/version-v0.3.0-orange.svg)

---

## 为什么做这个 / Why We Built This

现成的在线办公套件要么是 SaaS，数据在别人手里；要么重得要命，装一步劝退。PicoOffice 想要的是：**装得上、跑得动、数据是自己的**。你拿到的是一个能跑在自己服务器上、带协同编辑的轻量 Office，而不是一个要你天天运维的大家伙。

Off-the-shelf online office suites are either SaaS (your data sits in someone else's hands) or so heavy that setup alone puts you off. What PicoOffice wants is: **easy to install, easy to run, data is yours**. You get a lightweight Office with collaborative editing that runs on your own server — not a behemoth you have to babysit daily.

- 数据不出你的服务器（数据库 + 对象存储都在本地 / 私有 OSS）；
  - Your data never leaves your server (database + object storage are local / private OSS).
- 单二进制部署，SQLite 起步，想上 MySQL 也行；
  - Single-binary deployment, SQLite to start, MySQL if you prefer.
- 三端都给：Web、桌面（Electron）、移动（Capacitor）。
  - All three clients: Web, desktop (Electron), mobile (Capacitor).

---

## 功能清单 / Features

### 文字（Tiptap）/ Documents (Tiptap)

- 页面设置：纸张大小、页边距；
  - Page setup: paper size, margins.
- 页眉 / 页脚 / 页码；
  - Header / footer / page numbers.
- 分栏；
  - Columns.
- 段落样式、行距、缩进；
  - Paragraph styles, line spacing, indentation.
- 图片 7 种环绕方式（嵌入型 / 四周型 / 紧密型 / 穿越型 / 上下型 / 衬于文字下方 / 浮于文字上方）；
  - 7 image wrapping modes (inline / square / tight / through / top-and-bottom / behind text / in front of text).
- 自动目录；
  - Automatic table of contents.
- 表格单元格合并 / 拆分；
  - Table cell merge / split.
- 查找替换；
  - Find and replace.
- 字数统计；
  - Word count.
- 大纲面板；
  - Outline panel.
- 分页符。
  - Page breaks.

### 表格（Luckysheet）/ Spreadsheets (Luckysheet)

- 多 Sheet 页；
  - Multiple sheets.
- 公式：`SUM` / `AVERAGE` / `COUNT` / `VLOOKUP` / `IF`；
  - Formulas: `SUM` / `AVERAGE` / `COUNT` / `VLOOKUP` / `IF`.
- 单元格格式：货币、数字、百分比、日期；
  - Cell formats: currency, number, percentage, date.
- 冻结行列、筛选、排序；
  - Frozen rows/columns, filter, sort.
- 条件格式；
  - Conditional formatting.
- 图表。
  - Charts.

### 演示 / Presentations

- 母版、版式；
  - Master slides, layouts.
- 主题色；
  - Theme colors.
- 演讲者备注；
  - Speaker notes.
- 幻灯片切换动画；
  - Slide transition animations.
- 全屏播放；
  - Full-screen playback.
- 导出 PPTX。
  - Export to PPTX.

### 导入 / 导出 / Import / Export

- 导入：`docx` / `xlsx` / `pptx`（经 LibreOffice 转换）；
  - Import: `docx` / `xlsx` / `pptx` (converted via LibreOffice).
- 导出：`pdf` / `docx` / `xlsx` / `pptx`，PDF 按 A4 排版；
  - Export: `pdf` / `docx` / `xlsx` / `pptx`, PDF laid out to A4.
- 大文件导出走后台队列，异步生成完再下载。
  - Large-file export runs through a background queue and downloads once generation finishes.

### 分享 / Sharing

- 生成公开分享链接；
  - Generate public share links.
- 权限分 `view`（只读）/ `edit`（可编辑）。
  - Permissions: `view` (read-only) / `edit` (editable).

### 协作 / Collaboration

- 文字编辑器：真 OT（操作转换），多人同时编辑不互相覆盖；
  - Document editor: true OT (operational transformation), multiple people editing at once without overwriting each other.
- 表格 / 演示：乐观锁，版本冲突返回 `409` 由前端提示合并。
  - Spreadsheets / presentations: optimistic locking; a version conflict returns `409` and the frontend prompts to merge.

### 邮件 / 日历 / Mail / Calendar

- 邮件：IMAP / SMTP，多账号收发；
  - Mail: IMAP / SMTP, multi-account send/receive.
- 日历：事件创建 + 提醒。
  - Calendar: event creation + reminders.

### 搜索与存储 / Search & Storage

- 全文搜索：bleve 引擎，内置中文分词，跨文档检索；
  - Full-text search: bleve engine with built-in Chinese tokenization, cross-document retrieval.
- 对象存储：Storage 抽象层，本地磁盘 / S3 兼容（OSS、MinIO）可切换。
  - Object storage: a Storage abstraction layer, switchable between local disk / S3-compatible (OSS, MinIO).

### 安全 / Security

- 2FA（TOTP 两步验证）；
  - 2FA (TOTP two-factor authentication).
- JWT + refresh token 双令牌；
  - JWT + refresh token dual tokens.
- 登录限流、失败锁定；
  - Login rate limiting and failure lockout.
- 富文本 HTML 消毒（防 XSS）；
  - Rich-text HTML sanitization (XSS protection).
- 操作审计日志。
  - Operation audit logs.

### 其他 / Misc

- PWA / 离线：service worker + IndexedDB，断网可编辑、联网自动同步；
  - PWA / offline: service worker + IndexedDB, edit while disconnected and auto-sync when back online.
- 回收站：软删除 / 还原 / 超 30 天自动清理；
  - Trash: soft delete / restore / auto-purge after 30 days.
- 配额：按用户 / 全局存储限额；
  - Quota: per-user / global storage limits.
- 文件夹 / 标签 / 星标；
  - Folders / tags / stars.
- i18n：中文 / 英文；
  - i18n: Chinese / English.
- 打印（A4）；
  - Print (A4).
- Docker / docker-compose 一键起。
  - One-click Docker / docker-compose.

---

## 技术栈 / Tech Stack

| 层 / Layer | 选型 / Choice |
| -- | ---- |
| 后端 / Backend | Go / Gin / GORM，SQLite（默认）或 MySQL / SQLite (default) or MySQL |
| 前端 Web / Web Frontend | Vue 3 / Vite / Element Plus / Pinia |
| 编辑器 / Editors | Tiptap（文字）/ Luckysheet（表格）/ Tiptap (docs) / Luckysheet (sheets) |
| 桌面端 / Desktop | Electron |
| 移动端 / Mobile | Capacitor |
| 关键库 / Key Libraries | bleve（搜索）/ bluemonday（消毒）/ pquerna-otp（2FA）/ minio-go（S3）/ go-imap（邮件）/ gorilla/websocket（WS）/ golang-jwt |
| 文档转换 / Document Conversion | LibreOffice（docx/xlsx/pptx ⇄ 内部格式 / PDF）/ LibreOffice (docx/xlsx/pptx ⇄ internal format / PDF) |

---

## 目录结构 / Repository Layout

```
picooffice/
├── backend/               # Go 后端 / Go backend
│   ├── main.go
│   ├── go.mod / go.sum
│   ├── internal/
│   │   ├── handler/       # HTTP 接口 / HTTP handlers
│   │   ├── middleware/     # JWT / 限流 / 消毒等 / JWT / rate-limit / sanitize, etc.
│   │   ├── model/         # 数据模型 / data models
│   │   ├── service/       # 业务逻辑 / business logic
│   │   ├── storage/       # local / s3 存储抽象 / local / s3 storage abstraction
│   │   └── ws/            # WebSocket 协作 / WebSocket collaboration
│   └── cmd/
│       └── picobackup/    # 独立备份工具 / standalone backup tool
├── frontend-web/          # Vue3 Web 前端 / Vue3 web frontend
│   ├── src/
│   │   ├── api/  views/  components/  stores/  router/  i18n/
│   └── vite.config.js
├── desktop-electron/       # Electron 桌面壳 / Electron desktop shell
├── mobile-capacitor/      # Capacitor 移动壳（android/）/ Capacitor mobile shell (android/)
├── scripts/                # 安装脚本（install.sh / install.ps1）/ install scripts
├── storage/                # 本地存储（头像 / 索引等）/ local storage (avatars / index, etc.)
├── Dockerfile
├── docker-compose.yml
├── LICENSE
├── README.md                 # 项目首页 / project landing
├── CONTRIBUTING.md           # 贡献指南 / contributing guide
├── docs/                     # 文档目录 / documentation
│   ├── DETAILS.md           # 系统详情 / system details
│   ├── DEPLOY.md            # 部署手册 / deployment guide
│   ├── USAGE.md             # 使用教程 / user guide
│   ├── CHANGELOG.md         # 变更日志 / changelog
│   └── THANKS.md            # 第三方声明 / third-party notices
```

---

## 快速开始 / Quick Start

最简单的方式：跑起来单二进制，浏览器打开就用。

The simplest way: run the single binary and open it in a browser.

```bash
# 编译好的单二进制直接跑 / Run the prebuilt single binary
./picooffice
```

打开浏览器访问 `http://localhost:8080`。**第一个注册的账号自动成为管理员。**

Open `http://localhost:8080` in a browser. **The first registered account automatically becomes the administrator.**

默认用 SQLite，数据就在当前目录的 `picooffice.db` 里，想换 MySQL 改下配置即可（见 [DEPLOY.md](docs/DEPLOY.md)）。

It uses SQLite by default; data lives in `picooffice.db` in the current directory. Switch to MySQL by changing the config (see [DEPLOY.md](docs/DEPLOY.md)).

---

## 安装命令 / Installation

```bash
# Linux / macOS 一键 / One-liner for Linux / macOS
curl -fsSL https://raw.githubusercontent.com/nost3a/PicoOffice/main/scripts/install.sh | bash

# Windows PowerShell（管理员）/ Windows PowerShell (admin)
irm https://raw.githubusercontent.com/nost3a/PicoOffice/main/scripts/install.ps1 | iex

# Docker
docker run -d -p 8080:8080 -v pico_data:/data picooffice:latest

# Docker Compose
docker compose up -d

# 源码编译 / go install / Build from source via go install
go install github.com/nost3a/PicoOffice/...@latest

# 桌面 deb 安装 / Install desktop deb
sudo dpkg -i picooffice-desktop_1.0.0_amd64.deb

# 移动 apk 安装到设备 / Install mobile apk to device
adb install app-debug.apk
```

> 安装脚本本体在 `scripts/install.sh`（Linux/macOS）和 `scripts/install.ps1`（Windows），随仓库分发。
>
> The install scripts live in `scripts/install.sh` (Linux/macOS) and `scripts/install.ps1` (Windows) and are distributed with the repo.

---

## Docker / Docker Deployment

### 直接跑 / Run Directly

```bash
docker run -d \
  --name picooffice \
  -p 8080:8080 \
  -v pico_data:/data \
  -e PICO_JWT_SECRET=$(openssl rand -hex 32) \
  picooffice:latest
```

### Compose

仓库根目录自带 `docker-compose.yml`：

The repo root ships with `docker-compose.yml`:

```bash
docker compose up -d
```

### 常用环境变量 / Common Environment Variables

| 变量 / Variable | 说明 / Description | 默认 / Default |
| ---- | ---- | ---- |
| `PICO_JWT_SECRET` | JWT 签名密钥，**生产必改** / JWT signing secret, **change in production** | 启动检测到弱值会告警 / warns on weak value at startup |
| `PICO_PORT` | 监听端口 / Listen port | `8080` |
| `PICO_DB` | 数据库 DSN（SQLite 路径或 MySQL 连接串）/ DB DSN (SQLite path or MySQL connstr) | `picooffice.db` |
| `PICO_STORAGE` | 存储后端：`local` / `s3` / Storage backend: `local` / `s3` | `local` |
| `PICO_DATA_DIR` | 本地数据目录 / Local data directory | `./storage` |

完整配置项见 [DEPLOY.md](docs/DEPLOY.md)。

Full configuration options are in [DEPLOY.md](docs/DEPLOY.md).

---

## 文档 / Documentation

| 文档 / Document | 内容 / Content |
| ---- | ---- |
| [DETAILS.md](docs/DETAILS.md) | 详细设计 / 实现细节 / Detailed design / implementation details |
| [DEPLOY.md](docs/DEPLOY.md) | 部署手册（裸机 / Docker / 反代 / MySQL）/ Deployment guide (bare metal / Docker / reverse proxy / MySQL) |
| [USAGE.md](docs/USAGE.md) | 用户使用说明 / User guide |
| [CHANGELOG.md](docs/CHANGELOG.md) | 版本变更记录 / Changelog |
| [CONTRIBUTING.md](CONTRIBUTING.md) | 参与开发指引 / Contributing guide |

---

## 常见问题（FAQ） / Frequently Asked Questions

**Q：第一个注册的账号是什么权限？**
**Q: What permissions does the first registered account get?**

A：自动成为管理员，之后注册的都是普通用户。想关掉开放注册，在管理员设置里改。
A: It automatically becomes an administrator; accounts registered afterward are normal users. To disable open registration, change it in the admin settings.

**Q：数据存在哪？**
**Q: Where is the data stored?**

A：默认 SQLite（`picooffice.db`）+ 本地磁盘 `storage/` 目录。全在你服务器上，想备份就把这俩拷走，或者用 `picobackup` 工具导出。
A: By default SQLite (`picooffice.db`) + the local `storage/` directory. Everything is on your server; to back up, just copy those two, or use the `picobackup` tool to export.

**Q：能换 MySQL 吗？**
**Q: Can I switch to MySQL?**

A：能。GORM 支持，改下数据库连接配置即可，详见 [DEPLOY.md](docs/DEPLOY.md)。
A: Yes. GORM supports it; change the database connection config. See [DEPLOY.md](docs/DEPLOY.md).

**Q：断网还能编辑吗？**
**Q: Can I still edit while offline?**

A：开了 PWA 后可以离线编辑，内容先存 IndexedDB，联网后自动同步。
A: With PWA enabled you can edit offline; content is stored in IndexedDB first and auto-synced when back online.

**Q：导入的 docx 排版会走样吗？**
**Q: Will imported docx formatting be distorted?**

A：经 LibreOffice 转换，常见排版都能保真；特别复杂的宏 / 嵌套对象可能有差异，建议导入后过一遍。
A: Conversion via LibreOffice preserves common layouts faithfully; very complex macros / nested objects may differ, so it's best to review after import.

---

## 路线图（Roadmap） / Roadmap

下面是规划中的方向，均为"规划中"，具体排期以实际开发为准：

The directions below are all planned ("in planning"); actual scheduling follows real development:

- [ ] 把 OT 协同从文字编辑器扩展到表格、演示；
  - [ ] Extend OT collaboration from the document editor to spreadsheets and presentations.
- [ ] 原生移动编辑器（替换当前 WebView 套壳）；
  - [ ] Native mobile editor (replacing the current WebView shell).
- [ ] 支持 CalDAV / CardDAV，日历联系人可对接第三方客户端；
  - [ ] Support CalDAV / CardDAV so calendar and contacts can integrate with third-party clients.
- [ ] 更多存储后端（WebDAV、FTP 等）；
  - [ ] More storage backends (WebDAV, FTP, etc.).
- [ ] 插件 / 扩展机制；
  - [ ] Plugin / extension mechanism.
- [ ] 更细粒度的权限与部门 / 组织架构。
  - [ ] More fine-grained permissions and department / org structure.

欢迎在 Issue 里投票、催更或认领。

Vote, nudge, or claim items in the Issues.

---

## 参与贡献 / Contributing

想一起搞？看 [CONTRIBUTING.md](CONTRIBUTING.md) 把环境跑起来，照着提 PR 就行。

Want to join in? Read [CONTRIBUTING.md](CONTRIBUTING.md) to get the environment running, then open a PR.

## 开源协议 / License

本项目以 **MIT** 协议开源，详见 [LICENSE](LICENSE)。

This project is open-sourced under the **MIT** license, see [LICENSE](LICENSE).

## 致谢 / Acknowledgements

PicoOffice 建立在众多优秀开源项目之上，完整清单与许可证见 [THANKS.md](docs/THANKS.md)。

PicoOffice is built on many excellent open-source projects; the full list and licenses are in [THANKS.md](docs/THANKS.md).
