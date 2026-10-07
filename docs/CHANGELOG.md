# 更新日志 / Changelog

本项目遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 与语义化版本（SemVer）。
版本号格式 `v主版本.次版本.修订号`。

This project follows [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) and Semantic Versioning (SemVer).
Version format `vMajor.Minor.Patch`.

## [未发布] / Unreleased

- 计划中的改动会先列在这里，发布时再归入对应版本。
  - Planned changes are listed here first and moved into the relevant release when published.

## [v0.3.0] - 2026-10-07 / Third Release

第三轮迭代，补齐导入导出、分享协作与一整套自托管运维能力。

Third iteration, filling in import/export, sharing & collaboration, and a full set of self-hosting operations capabilities.

### 新增 / Added

- **导入**：支持 `docx` / `xlsx` / `pptx` 导入，经 LibreOffice 转换为内部可编辑格式。
  - **Import**: supports `docx` / `xlsx` / `pptx` import, converted to the internal editable format via LibreOffice.
- **导出**：支持导出 `pdf` / `docx` / `xlsx` / `pptx`，PDF 按 A4 排版；导出走后台队列，异步生成后通知下载。
  - **Export**: supports exporting `pdf` / `docx` / `xlsx` / `pptx`, with PDF laid out to A4; exports run through a background queue and notify for download once generated asynchronously.
- **分享**：生成公开分享链接，可设置 `view`（只读）/ `edit`（可编辑）权限。
  - **Sharing**: generates public share links with `view` (read-only) / `edit` (editable) permissions.
- **富文本消毒**：富文本内容经 `bluemonday` 白名单消毒，防 XSS。
  - **Rich-text sanitization**: rich-text content is sanitized via the `bluemonday` whitelist to prevent XSS.
- **回收站**：文档软删除，支持手动还原；超 30 天自动清理。
  - **Trash**: documents are soft-deleted, with manual restore supported; auto-purged after 30 days.
- **协作**：
  - **Collaboration**:
  - 文字编辑器接入真 OT（操作转换），多人同时编辑不互相覆盖；
    - The document editor integrates true OT (operational transformation), so multiple people editing at once don't overwrite each other.
  - 表格 / 演示采用乐观锁，版本冲突时返回 `409` 由前端合并提示。
    - Spreadsheets / presentations use optimistic locking; on a version conflict it returns `409` and the frontend prompts to merge.
- **打印**：文档打印走 A4 样式，浏览器直接打印预览。
  - **Printing**: documents print with A4 styling, with a direct print preview from the browser.
- **认证增强**：JWT  access token + refresh token 双令牌续期；敏感操作支持基于 TOTP 的二次授权（含备用恢复设备/码管理）。
  - **Enhanced auth**: JWT access token + refresh token dual-token renewal; sensitive operations support TOTP-based step-up authorization (including backup recovery device/code management).
- **邮件**：内置邮件客户端，支持 IMAP/SMTP、多账号收发。
  - **Mail**: a built-in mail client supporting IMAP/SMTP and multi-account send/receive.
- **日历**：事件创建与提醒。
  - **Calendar**: event creation and reminders.
- **全文搜索**：基于 bleve，内置中文分词，跨文档检索。
  - **Full-text search**: based on bleve with built-in Chinese tokenization for cross-document retrieval.
- **对象存储**：Storage 抽象层，本地磁盘 / S3 兼容（OSS、MinIO 等）可切换。
  - **Object storage**: a Storage abstraction layer, switchable between local disk / S3-compatible (OSS, MinIO, etc.).
- **健康检查**：`/healthz` 探活接口，便于容器编排。
  - **Health check**: a `/healthz` liveness endpoint for easier container orchestration.
- **容器化**：提供 `Dockerfile` 与 `docker-compose.yml`，一键起服务。
  - **Containerization**: provides a `Dockerfile` and `docker-compose.yml` for one-click service startup.
- **备份**：独立 `picobackup` 工具，一键导出数据库与存储。
  - **Backup**: a standalone `picobackup` tool that exports the database and storage in one click.
- **触摸适配**：移动端编辑器手势与触摸交互优化。
  - **Touch adaptation**: gesture and touch-interaction optimizations for the mobile editor.
- **国际化**：i18n 框架落地，中文 / 英文双语。
  - **Internationalization**: the i18n framework is in place, with Chinese / English bilingual support.
- 表格函数补齐 `SUM` / `AVERAGE` / `COUNT` / `VLOOKUP` / `IF`。
  - Spreadsheet functions expanded to `SUM` / `AVERAGE` / `COUNT` / `VLOOKUP` / `IF`.
- 后台测试体系补齐：单元测试 + 冒烟 / 端到端测试脚本。
  - Backend test suite completed: unit tests + smoke / end-to-end test scripts.
- 工程目录与文件组织规整化（backend / frontend-web / desktop-electron / mobile-capacitor / scripts / docs 分层）。
  - Project directory and file organization restructured (layered as backend / frontend-web / desktop-electron / mobile-capacitor / scripts / docs).

### 修复 / 优化 / Fixed / Changed

- 登录接口限流与失败锁定收紧。
  - Login endpoint rate limiting and failure lockout tightened.
- 导出大文件 OOM 风险缓解，改队列 + 流式。
  - Mitigated OOM risk on large-file export by switching to queue + streaming.

## [v0.2.0] - 2026-09 / Second Release

第二轮迭代，专注排版专业性与离线可用。

Second iteration, focused on professional typography and offline usability.

### 新增 / Added

- **专业排版**：文字编辑器补齐页面设置、页眉页脚页码、分栏、样式、行距缩进、图片 7 种环绕方式、目录、表格合并、查找替换、字数统计、大纲、分页符。
  - **Professional typography**: the document editor adds page setup, header/footer/page numbers, columns, styles, line spacing & indentation, 7 image wrapping modes, table of contents, table merge, find & replace, word count, outline, and page breaks.
- **PWA / 离线**：service worker 缓存壳资源，IndexedDB 暂存编辑内容，断网可编辑、联网后自动同步。
  - **PWA / offline**: service worker caches shell resources and IndexedDB stages edits, allowing offline editing that auto-syncs when back online.
- **配额**：按用户 / 全局配置存储配额，超限拦截上传。
  - **Quota**: per-user / global storage quota configuration, blocking uploads beyond the limit.
- **分块上传**：大文件分块上传 + 断点续传。
  - **Chunked upload**: large files upload in chunks with resume support.
- **登录限流**：基础失败计数与临时锁定。
  - **Login rate limiting**: basic failure counting and temporary lockout.
- **WebSocket 心跳**：WS 保活心跳，断线自动重连。
  - **WebSocket heartbeat**: WS keep-alive heartbeat with automatic reconnect on disconnect.
- **版本历史**：文档历史版本记录与回滚。
  - **Version history**: document version history recording and rollback.
- **A4 PDF**：导出 PDF 走 A4 页面排版。
  - **A4 PDF**: exported PDF follows A4 page layout.

## [v0.1.0] - 2026-08 / First Release

首个可用版本，搭起三端骨架与编辑器。

First usable release, setting up the three-client skeleton and editors.

### 新增 / Added

- **后端**：Go + Gin + GORM，SQLite 起步、兼容 MySQL。
  - **Backend**: Go + Gin + GORM, starting with SQLite and compatible with MySQL.
- **前端**：Vue3 + Vite + Element Plus + Pinia Web 端。
  - **Frontend**: Vue3 + Vite + Element Plus + Pinia web client.
- **三大编辑器**：文字（Tiptap）、表格（Luckysheet）、演示（自研幻灯片）。
  - **Three editors**: documents (Tiptap), spreadsheets (Luckysheet), presentations (in-house slides).
- **桌面端**：Electron 套壳，产出 deb 安装包。
  - **Desktop**: Electron shell producing a deb installer.
- **移动端**：Capacitor 套壳，产出 debug apk。
  - **Mobile**: Capacitor shell producing a debug apk.

[未发布]: https://github.com/nost3a/PicoOffice/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/nost3a/PicoOffice/releases/tag/v0.3.0
[v0.2.0]: https://github.com/nost3a/PicoOffice/releases/tag/v0.2.0
[v0.1.0]: https://github.com/nost3a/PicoOffice/releases/tag/v0.1.0
