# PicoOffice 系统详情 / System Details

本文档描述 PicoOffice 的完整系统事实：技术选型、架构、目录结构、数据模型、API、前端路由、协作协议、与 WPS 个人版的功能对标、安全设计、已知差距和开源组件致谢。所有路径、端口、环境变量名均为真实产物，不写占位符。

This document describes the complete system facts of PicoOffice: technology choices, architecture, directory layout, data model, API, frontend routes, collaboration protocol, feature comparison against WPS Personal Edition, security design, known gaps, and open-source acknowledgements. All paths, ports, and environment variable names are real artifacts — no placeholders.

---

## 1. 项目简介 / Introduction

PicoOffice 是一套可私有化部署的在线办公套件，对标 WPS 个人版的核心三件套：文字文档、电子表格、演示文稿。单后端二进制 + 前端静态资源内嵌，开箱即跑；同时提供 Linux deb、Windows 绿色 zip、Android APK 三种客户端壳，浏览器、桌面、移动三端共用同一个后端。

PicoOffice is a self-hostable online office suite, benchmarking against the three core apps of WPS Personal Edition: word documents, spreadsheets, and presentations. A single backend binary with embedded frontend static assets runs out of the box; at the same time it ships three client shells — Linux deb, Windows portable zip, and Android APK — so the browser, desktop, and mobile clients all share the same backend.

定位：

Positioning:

- **私有化**：整个系统跑在你自己的服务器上，文档和附件不出内网，不依赖任何第三方云。
  - **Private deployment**: The entire system runs on your own server; documents and attachments never leave the intranet and it depends on no third-party cloud.
- **轻量**：后端是一个 Go 单二进制（`backend/picooffice`，约 66MB，已 `go:embed` 前端 dist），SQLite 单文件数据库，不需要单独装数据库服务。
  - **Lightweight**: The backend is a single Go binary (`backend/picooffice`, about 66MB, with the frontend dist already `go:embed`ded), using a single-file SQLite database — no separate database service to install.
- **可演进**：默认 SQLite，环境变量一切即可切 MySQL；前端是标准 Vue3 工程，可二次开发。
  - **Evolvable**: SQLite by default, switchable to MySQL with a single environment variable; the frontend is a standard Vue3 project open to further development.

它不是 WPS 的复刻版，也不内置 WPS 的云字体、模板库、OCR、PDF 批注等版权或重工作量功能。它解决的是"我要有一个自己的在线文档站，能写、能算、能放 PPT，能多人同时看，能导出 PDF/docx/xlsx"这个最小闭环。

It is not a clone of WPS, nor does it bundle WPS's copyright-protected or heavy-lift features such as cloud fonts, template library, OCR, or PDF annotation. What it solves is the minimal closed loop of "I want my own online document site where I can write, calculate, show slides, let multiple people view simultaneously, and export PDF/docx/xlsx."

---

## 2. 技术选型表 / Technology Choices

| 层 / Layer | 组件 / Component | 版本 / Version | 为什么选它 / Why chosen | 被淘汰的备选 / Rejected alternative |
|---|---|---|---|---|
| 后端语言 / Backend language | Go | 1.23+ | 编译单二进制，交叉编译方便，goroutine 天然适合 WebSocket 房间广播<br>Compiles to a single binary, easy cross-compilation, goroutines naturally fit WebSocket room broadcasting | Node.js（要带运行时部署）、Python（性能/分发都差）<br>Node.js (requires shipping a runtime), Python (poor on both performance and distribution) |
| Web 框架 / Web framework | Gin | v1.12.0 | 路由轻、中间件模型简单，社区资料多<br>Light routing, simple middleware model, abundant community resources | Echo、Chi、Fiber |
| ORM | GORM | v1.31.2 | AutoMigrate 开箱即用，SQLite/MySQL 驱动可换<br>AutoMigrate works out of the box, SQLite/MySQL drivers interchangeable | sqlx（手写 SQL 太多）、Ent<br>sqlx (too much hand-written SQL), Ent |
| 数据库 / Database | SQLite | via mattn/go-sqlite3 | 零运维单文件，私有化部署默认首选<br>Zero-ops single file, the default first choice for private deployment | PostgreSQL（要单独装服务）<br>PostgreSQL (requires a separate service) |
| 数据库（可选）/ Database (optional) | MySQL | 由 PICO_DB 切<br>Switched via PICO_DB | 已有 MySQL 集群的公司不用再运维一套<br>Companies with an existing MySQL cluster need not run another | —— |
| 鉴权 / Auth | golang-jwt/v5 | v5.3.1 | 标准 JWT，无状态<br>Standard JWT, stateless | session+redis（多一个组件）<br>session+redis (one extra component) |
| WebSocket | gorilla/websocket | v1.5.3 | Go 生态事实标准<br>De-facto standard in the Go ecosystem | nhooyr/websocket |
| 密码哈希 / Password hash | golang.org/x/crypto/bcrypt | DefaultCost | 加盐慢哈希，抗爆破<br>Salted slow hash, resistant to brute force | md5/sha256（裸哈希不安全）、argon2（依赖更重）<br>md5/sha256 (bare hashes are unsafe), argon2 (heavier dependency) |
| 格式转换 / Format conversion | LibreOffice (soffice) | headless | docx/xlsx/pdf 互转的唯一开源方案<br>The only open-source solution for docx/xlsx/pdf interconversion | OnlyOffice（要起一整个服务）、Aspose（收费）<br>OnlyOffice (requires spinning up a whole service), Aspose (paid) |
| 前端框架 / Frontend framework | Vue3 | 3.4.x | 响应式 + 组件化，上手快<br>Reactive + componentized, quick to pick up | React（团队不熟）<br>React (team unfamiliar) |
| 构建工具 / Build tool | Vite | 5.3.x | 冷启动快，HMR 快<br>Fast cold start, fast HMR | webpack |
| UI 库 / UI library | Element Plus | 2.7.x | 后台/工具类界面组件全，默认蓝白符合办公风<br>Complete admin/tool-style components, default blue-white fits office look | Ant Design Vue、Naive UI |
| 状态管理 / State management | Pinia | 2.1.x | Vue3 官方推荐，比 Vuex 简洁<br>Official Vue3 recommendation, simpler than Vuex | Vuex |
| 路由 / Router | Vue Router | 4.4.x | Vue 官方<br>Official Vue | —— |
| HTTP 客户端 / HTTP client | Axios | 1.7.x | 拦截器统一加 token<br>Interceptors uniformly attach token | fetch（要自己包拦截器）<br>fetch (must wrap interceptors yourself) |
| 文字编辑器 / Word editor | Tiptap (@tiptap/vue-3) | 2.6.6 | ProseMirror 封装，扩展示例多，协同底子好<br>ProseMirror wrapper, many extension examples, good collaboration foundation | Quill、CKEditor |
| 表格编辑器 / Spreadsheet editor | Luckysheet | 2.1.13 | 开箱即用的在线表格，API 全<br>Out-of-the-box online spreadsheet, complete API | Univer（npmmirror 上包名/版本对不上，降级是有意选择，见"已知差距"）<br>Univer (package name/version mismatch on npmmirror; downgrade was a deliberate choice, see "Known Gaps") |
| 演示编辑器 / Presentation editor | 自写 slide 编辑器 / Custom slide editor | —— | 左缩略列表 + 中间编辑 + 全屏播放，需求简单不值得引重型库<br>Left thumbnail list + center editing + full-screen playback; needs are simple, not worth pulling in a heavy library | reveal.js（不是编辑器）、impress.js<br>reveal.js (not an editor), impress.js |
| 桌面壳 / Desktop shell | Electron | 31.x | 一套前端代码包 Windows/Linux 桌面<br>One frontend codebase packages Windows/Linux desktop | Tauri（要 Rust，且 Windows 打包链更长）<br>Tauri (needs Rust, and the Windows packaging chain is longer) |
| 桌面打包 / Desktop packaging | electron-builder | 24.13.x | deb/zip/nsis 都出<br>Outputs deb/zip/nsis | electron-forge |
| 移动壳 / Mobile shell | Capacitor | 8.x | Web 工程直接套壳出 Android/iOS<br>Wrap a web project directly into Android/iOS | Cordova（停更边缘）、uni-app（要重写前端）<br>Cordova (near EOL), uni-app (requires rewriting frontend) |
| Android 构建 / Android build | Gradle 8.14.3 + AGP 8.13 + JDK 21 | —— | Capacitor 8 的硬性要求<br>Hard requirement of Capacitor 8 | —— |

---

## 3. 系统架构图 / System Architecture

```
                          ┌─────────────────────────────────────┐
                          │            用户终端                   │
                          └─────────────────────────────────────┘
        ┌────────────┬──────────────────┬──────────────────┬──────────────┐
        │ 浏览器      │ Electron 桌面壳   │ Capacitor Android │ (macOS/iOS   │
        │ Chrome/Edge│ Linux deb/Win zip│ apk (12MB debug)  │  需自行 build)│
        └─────┬──────┴────────┬─────────┴────────┬─────────┴──────┬───────┘
              │   HTTPS/WSS    │                  │                │
              └───────────────┴────────┬─────────┴────────────────┘
                                       │
                            ┌──────────▼──────────┐
                            │   Nginx (80/443)    │  反向代理 + WS upgrade
                            │   certbot HTTPS      │
                            └──────────┬──────────┘
                                       │ 127.0.0.1:8080
                            ┌──────────▼──────────────────────────────┐
                            │        Go 单二进制 backend/picooffice    │
                            │  Gin + GORM + gorilla/websocket         │
                            │  ┌────────────────────────────────────┐  │
                            │  │ HTTP handlers: auth/docs/fs/admin  │  │
                            │  │ WS Hub: 按 room=doc-{id} 广播       │  │
                            │  │ go:embed web/  (前端 dist 已内嵌)   │  │
                            │  └────────────────────────────────────┘  │
                            └───┬──────────────┬──────────────┬─────┘
                                │              │              │
                    ┌───────────▼──┐  ┌───────▼────────┐  ┌──▼─────────────┐
                    │ SQLite 文件   │  │ 本地磁盘目录    │  │ soffice 进程    │
                    │ picooffice.db │  │ storage/uploads│  │ --headless      │
                    │ (默认, 单文件) │  │ (用户上传附件)  │  │ 转 pdf/docx/xlsx│
                    └───────────────┘  └────────────────┘  └─────────────────┘
                                │
                                └── PICO_DB 一切即可换成 MySQL DSN
```

数据流要点：

Key data-flow points:

1. 浏览器/桌面/移动三端只认一个后端地址（HTTPS），其余逻辑完全一致。
2. The browser, desktop, and mobile clients only know one backend address (HTTPS); the rest of the logic is identical.
3. 后端是无状态 HTTP + 有状态 WS Hub。HTTP 全部走 JWT；WS 的 token 走 query string（`/api/ws/docs/:id?token=xxx`）。
4. The backend is stateless HTTP + a stateful WS Hub. All HTTP goes through JWT; the WS token goes via query string (`/api/ws/docs/:id?token=xxx`).
5. 文档正文默认存在 SQLite 的 `docs.content` 字段（longtext）；大附件走 `storage/uploads/` 文件系统。
6. Document body is stored by default in the SQLite `docs.content` field (longtext); large attachments go through the `storage/uploads/` filesystem.
7. 导出 PDF/docx/xlsx 时，后端把当前文档落临时文件，调 `soffice --headless --convert-to`，再把结果流回给前端。
8. When exporting PDF/docx/xlsx, the backend writes the current document to a temp file, invokes `soffice --headless --convert-to`, then streams the result back to the frontend.
9. 前端 dist 在编译期通过 `//go:embed all:web` 打进二进制，运行时不需要再单独部署静态资源。
10. The frontend dist is baked into the binary at compile time via `//go:embed all:web`, so no separate static-asset deployment is needed at runtime.

---

## 4. 目录结构 / Directory Layout

项目根：`/home/user/Doubao/chats/1235147905824770/picooffice/`

Project root: `/home/user/Doubao/chats/1235147905824770/picooffice/`

```
picooffice/
├── backend/                    # Go 后端
│   ├── go.mod
│   ├── go.sum
│   ├── main.go                 # 入口，路由 + embed + AutoMigrate
│   ├── picooffice              # 编译产物单二进制 (~66MB, 已 embed 前端)
│   ├── picooffice.db           # SQLite 数据库文件 (运行时生成)
│   ├── picooffice.log          # 运行日志
│   ├── web/                    # 前端 dist 拷进来的 (被 go:embed)
│   │   ├── index.html
│   │   ├── assets/
│   │   └── luckysheet/
│   └── internal/
│       ├── handler/
│       │   ├── auth.go         # register/login/me
│       │   ├── doc.go          # docs CRUD + content + export
│       │   ├── fs.go           # 文件上传/下载/列表
│       │   ├── admin.go        # 用户列表
│       │   ├── ws.go           # WS 升级
│       │   └── common.go
│       ├── middleware/
│       │   ├── cors.go
│       │   └── jwt.go
│       ├── model/
│       │   ├── user.go
│       │   ├── doc.go
│       │   └── file.go
│       └── ws/
│           └── hub.go          # 房间管理器
├── frontend-web/               # Vue3 前端源码
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── public/
│   ├── dist/                   # vite build 产物
│   └── src/
│       ├── main.js
│       ├── App.vue
│       ├── api/
│       │   ├── http.js         # axios 实例 + token 拦截器
│       │   └── index.js
│       ├── router/index.js     # 路由表 + 守卫
│       ├── stores/auth.js     # Pinia: 当前用户/token
│       ├── components/
│       ├── assets/
│       └── views/
│           ├── Login.vue
│           ├── Register.vue
│           ├── Dashboard.vue   # 文件列表 + 新建 + 搜索
│           ├── Admin.vue       # 用户管理 (仅 admin)
│           └── editor/
│               ├── DocEditor.vue    # Tiptap 文字文档
│               ├── SheetEditor.vue  # Luckysheet 表格
│               └── SlideEditor.vue # 自写 slide
├── desktop-electron/           # Electron 桌面壳
│   ├── package.json
│   ├── main.js                 # Electron 主进程
│   ├── app-dist/               # 前端 dist 拷进来的
│   └── dist/
│       ├── picooffice-desktop_1.0.0_amd64.deb   # ~75MB
│       ├── PicoOffice-1.0.0-win.zip            # ~108MB (解压即用)
│       ├── linux-unpacked/
│       └── win-unpacked/
│           └── PicoOffice.exe                  # 双击即跑
├── mobile-capacitor/           # Capacitor 移动壳
│   ├── package.json
│   ├── capacitor.config.json   # server.url = http://10.0.2.2:8080
│   ├── README.md                # 出 apk/ipa 的完整命令
│   ├── www/                    # 前端 dist 拷进来的
│   └── android/                # 已生成的 Android 工程
│       └── app/build/outputs/apk/debug/app-debug.apk   # ~12MB debug 签名
└── storage/
    └── uploads/                # 用户上传附件落盘处 (PICO_STORAGE 默认 ../storage)
```

注意几个运行时约定：

A few runtime conventions to note:

- 后端默认 `PICO_STORAGE=../storage`，意思是"二进制所在目录的上一级 `storage/`"。所以把 `picooffice` 二进制放在 `backend/` 里跑时，附件正好落到项目根的 `storage/uploads/`。
  - The backend defaults `PICO_STORAGE=../storage`, meaning "the `storage/` directory one level above where the binary lives". So when the `picooffice` binary runs inside `backend/`, attachments land exactly in the project root's `storage/uploads/`.
- `backend/web/` 是编译前端后拷进来的，不是手写的；重新 build 前端要重新拷一遍再 `go build`。
  - `backend/web/` is copied in after building the frontend, not hand-written; to rebuild the frontend you must copy it in again and then run `go build`.
- `desktop-electron/dist/` 里的 deb/zip 是打包产物，不是源码；改前端要重新 build 前端再 `npm run dist:deb` / `dist:win`。
  - The deb/zip inside `desktop-electron/dist/` are packaging artifacts, not source; to change the frontend you must rebuild it and then run `npm run dist:deb` / `dist:win`.

---

## 5. 数据模型 / Data Model

三张表，由 GORM AutoMigrate 自动建表。字段名与 struct tag 一一对应。

Three tables, auto-created by GORM AutoMigrate. Field names correspond one-to-one with the struct tags.

### 5.1 users 表 / users Table

| 字段 / Field | 类型 / Type | 约束 / Constraint | JSON | 说明 / Notes |
|---|---|---|---|---|
| id | uint | primarykey | id | 自增主键<br>Auto-increment primary key |
| username | string(64) | uniqueIndex | username | 登录名，唯一<br>Login name, unique |
| password_hash | string(128) | — | 不输出 (`json:"-"`)<br>Not output (`json:"-"`) | bcrypt 哈希，绝不回传前端<br>bcrypt hash, never returned to the frontend |
| nickname | string(64) | — | nickname | 显示名，空则等于 username<br>Display name; if empty, equals username |
| role | string(16) | default 'user' | role | `user` 或 `admin`<br>`user` or `admin` |
| created_at | time.Time | — | created_at | 注册时间<br>Registration time |
| updated_at | time.Time | — | updated_at | 更新时间<br>Update time |

规则：注册时 `db.Model(&User{}).Count(&cnt)`，`cnt == 0` 的第一个用户自动 `role=admin`，之后全部 `role=user`。

Rule: at registration `db.Model(&User{}).Count(&cnt)` is called; the first user with `cnt == 0` is automatically set `role=admin`, and everyone after that is `role=user`.

### 5.2 docs 表 / docs Table

| 字段 / Field | 类型 / Type | 约束 / Constraint | JSON | 说明 / Notes |
|---|---|---|---|---|
| id | uint | primarykey | id | 自增主键<br>Auto-increment primary key |
| owner_id | uint | index | owner_id | 属主 user.id<br>Owning user.id |
| title | string(255) | — | title | 文档标题<br>Document title |
| doc_kind | string(16) | — | doc_kind | `doc` / `sheet` / `slide` |
| file_ext | string(16) | — | file_ext | 导出/导入时的扩展名<br>Extension used on export/import |
| size | int64 | — | size | 正文大小（字节）<br>Body size (bytes) |
| content | longtext | — | 不输出 (`json:"-"`)<br>Not output (`json:"-"`) | 文档正文（Tiptap HTML / Luckysheet JSON / slide JSON）<br>Document body (Tiptap HTML / Luckysheet JSON / slide JSON) |
| content_path | string(255) | — | 不输出<br>Not output | 大正文可落盘时的路径（当前默认空，正文在 content 字段）<br>Path for offloading large bodies to disk (currently empty by default; body lives in the content field) |
| created_at | time.Time | — | created_at | 创建时间<br>Creation time |
| updated_at | time.Time | — | updated_at | 最近修改时间<br>Last modified time |

### 5.3 files 表 / files Table

| 字段 / Field | 类型 / Type | 约束 / Constraint | JSON | 说明 / Notes |
|---|---|---|---|---|
| id | uint | primarykey | id | 自增主键<br>Auto-increment primary key |
| uploader_id | uint | index | uploader_id | 上传者 user.id<br>Uploader user.id |
| original_name | string(255) | — | original_name | 用户上传时的原始文件名<br>Original filename at upload time |
| stored_path | string(512) | — | 不输出<br>Not output | 服务器上的落盘相对路径<br>Relative on-disk path on the server |
| size | int64 | — | size | 文件大小<br>File size |
| mime | string(128) | — | mime | MIME 类型<br>MIME type |
| created_at | time.Time | — | created_at | 上传时间<br>Upload time |

---

## 6. API 全表 / API Reference

全部已冒烟通过。Base URL：`http://your-host:8080`（私有化后通常是 `https://your-domain`）。除 `/api/health`、`/api/auth/register`、`/api/auth/login`、WS 升级接口外，都要在 Header 带 `Authorization: Bearer <token>`。

All endpoints have passed smoke testing. Base URL: `http://your-host:8080` (after private deployment typically `https://your-domain`). Except for `/api/health`, `/api/auth/register`, `/api/auth/login`, and the WS upgrade interface, all requests must carry `Authorization: Bearer <token>` in the Header.

### 6.1 健康检查 / Health Check

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 请求 / Request | 响应 / Response |
|---|---|---|---|---|
| GET | `/api/health` | 无 / None | — | `200 {"status":"ok"}` |

### 6.2 认证 / Authentication

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 请求体 / Body | 成功响应 / Success Response |
|---|---|---|---|---|
| POST | `/api/auth/register` | 无 / None | `{"username":"alice","password":"secret123","nickname":"Alice"}` | `200 {"token":"...","user":{"id":1,"username":"alice","role":"admin","nickname":"Alice"}}` |
| POST | `/api/auth/login` | 无 / None | `{"username":"alice","password":"secret123"}` | `200 {"token":"...","user":{...}}` |
| GET | `/api/auth/me` | JWT | — | `200 {"id":1,"username":"alice","role":"admin","nickname":"Alice"}` |

注册校验：username ≥ 3 字符，password ≥ 6 字符；username 重复返回 409。

Registration validation: username ≥ 3 characters, password ≥ 6 characters; a duplicate username returns 409.

### 6.3 文档 / Documents

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 请求 / Request | 响应 / Response |
|---|---|---|---|---|
| GET | `/api/docs?type=doc&keyword=&page=1&size=20` | JWT | query: type=doc\|sheet\|slide，keyword，page，size<br>query: type=doc\|sheet\|slide, keyword, page, size | `200 {"list":[...],"total":N}` |
| POST | `/api/docs` | JWT | `{"title":"未命名","type":"doc"}`<br>`{"title":"Untitled","type":"doc"}` | `200` 新建的 doc 对象<br>`200` newly created doc object |
| GET | `/api/docs/:id` | JWT | — | `200` doc 元信息<br>`200` doc metadata |
| PUT | `/api/docs/:id` | JWT | `{"title":"新标题"}`<br>`{"title":"New Title"}` | `200` 更新后的 doc<br>`200` updated doc |
| DELETE | `/api/docs/:id` | JWT | — | `204` |
| GET | `/api/docs/:id/content` | JWT | — | `200 {"content":"..."}` 正文 HTML/JSON<br>`200 {"content":"..."}` body HTML/JSON |
| PUT | `/api/docs/:id/content` | JWT | `{"content":"..."}` | `200 {"ok":true}` 覆盖正文（最后写入获胜）<br>`200 {"ok":true}` overwrites body (last write wins) |
| POST | `/api/docs/:id/export?type=pdf` | JWT | query: type=pdf\|docx\|xlsx | 直接流回文件；soffice 不可用时 501<br>Streams file back directly; 501 when soffice is unavailable |

列表接口只返回当前用户自己的文档（`owner_id = 当前 uid`），不返回别人的。

The list endpoint only returns the current user's own documents (`owner_id = current uid`), not others'.

### 6.4 文件系统 / File System

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 请求 / Request | 响应 / Response |
|---|---|---|---|---|
| GET | `/api/fs/files` | JWT | — | `200 {"list":[...]}` 当前用户上传过的附件列表<br>`200 {"list":[...]}` list of attachments the current user has uploaded |
| POST | `/api/fs/upload` | JWT | multipart/form-data, field 名 `file`<br>multipart/form-data, field name `file` | `200` 新建的 file 对象（含 id/original_name/size/mime）<br>`200` newly created file object (with id/original_name/size/mime) |
| GET | `/api/fs/download/:id` | JWT | — | 文件流，带 `Content-Disposition`<br>File stream with `Content-Disposition` |

### 6.5 WebSocket

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 说明 / Notes |
|---|---|---|---|
| GET (upgrade) | `/api/ws/docs/:id?token=xxx` | query token | 升级为 WS，room = `doc-{id}`<br>Upgrade to WS, room = `doc-{id}` |

WS 不是 REST，消息格式见第 8 节。

WS is not REST; the message format is in Section 8.

### 6.6 管理 / Administration

| 方法 / Method | 路径 / Path | 鉴权 / Auth | 说明 / Notes |
|---|---|---|---|
| GET | `/api/admin/users` | JWT + admin | 返回全部用户列表（不含 password_hash）<br>Returns the full user list (excluding password_hash) |

非 admin 访问 `/api/admin/*` 返回 403。

Non-admin access to `/api/admin/*` returns 403.

### 6.7 错误响应格式 / Error Response Format

所有 4xx/5xx 统一 JSON：`{"error":"描述"}`。常见码：

All 4xx/5xx return a uniform JSON: `{"error":"description"}`. Common codes:

- 400 bad body / 参数错
  - 400 bad body / invalid parameter
- 401 wrong username or password / missing token / invalid token
- 403 非 admin 访问管理接口
  - 403 non-admin accessing admin interface
- 404 not found
- 409 username already exists
- 501 soffice not available（导出时服务器没装 LibreOffice）
  - 501 soffice not available (LibreOffice not installed on the server at export time)

---

## 7. 前端路由表 / Frontend Routes

定义在 `frontend-web/src/router/index.js`。

Defined in `frontend-web/src/router/index.js`.

| 路径 / Path | 组件 / Component | 鉴权 / Auth | 说明 / Notes |
|---|---|---|---|
| `/login` | Login.vue | 未登录可访问；已登录跳 dashboard<br>Accessible when logged out; redirects to dashboard if logged in | 登录页<br>Login page |
| `/register` | Register.vue | 未登录可访问；已登录跳 dashboard<br>Accessible when logged out; redirects to dashboard if logged in | 注册页<br>Register page |
| `/dashboard` | Dashboard.vue | 必须登录<br>Must be logged in | 文件列表 + 新建 doc/sheet/slide + 搜索框<br>File list + create doc/sheet/slide + search box |
| `/editor/doc/:id` | editor/DocEditor.vue | 必须登录<br>Must be logged in | Tiptap 文字文档编辑器<br>Tiptap word document editor |
| `/editor/sheet/:id` | editor/SheetEditor.vue | 必须登录<br>Must be logged in | Luckysheet 表格编辑器<br>Luckysheet spreadsheet editor |
| `/editor/slide/:id` | editor/SlideEditor.vue | 必须登录<br>Must be logged in | 自写 slide 编辑器<br>Custom slide editor |
| `/admin` | Admin.vue | 必须登录 + admin<br>Must be logged in + admin | 用户管理页<br>User management page |
| `/:pathMatch(.*)*` | — | — | 一律重定向到 `/dashboard`<br>Always redirects to `/dashboard` |

路由守卫逻辑：

Route guard logic:

1. 非 login/register 且未登录 → 踢回 `/login`。
2. Not login/register and not logged in → kick back to `/login`.
3. 在 login/register 但已登录 → 跳 `/dashboard`。
4. On login/register but already logged in → redirect to `/dashboard`.
5. `meta.requireAdmin` 且当前用户不是 admin → 跳 `/dashboard`。
6. `meta.requireAdmin` and the current user is not admin → redirect to `/dashboard`.

---

## 8. 协作协议（WebSocket） / Collaboration Protocol

### 8.1 连接 / Connection

```
wss://your-domain/api/ws/docs/12?token=<JWT>
```

- token 必须放 query（WS 握手时浏览器原生 API 不能自定义 Header）。
  - The token must go in the query (the browser's native API cannot set custom Headers during the WS handshake).
- 后端升级成功后，把当前用户加入 `room = "doc-{id}"`。
  - After a successful upgrade, the backend adds the current user to `room = "doc-{id}"`.
- 升级失败返回 401 JSON：`missing token` / `invalid token`。
  - A failed upgrade returns 401 JSON: `missing token` / `invalid token`.

### 8.2 服务端 → 客户端消息 / Server → Client Messages

后端主动推两类消息：

The backend proactively pushes two kinds of messages:

**join**（有人进房）/ **join** (someone enters the room):

```json
{
  "type": "join",
  "user": "alice",
  "users": ["alice", "bob", "carol"]
}
```

**leave**（有人退房）/ **leave** (someone leaves the room):

```json
{
  "type": "leave",
  "user": "bob",
  "users": ["alice", "carol"]
}
```

`users` 是房间内当前所有去重后的用户名列表，前端用来渲染右上角"在线用户头像组"。

`users` is the current deduplicated list of usernames in the room; the frontend uses it to render the "online user avatar group" in the top-right corner.

### 8.3 客户端 → 服务端 → 其他客户端消息 / Client → Server → Other Clients Messages

前端发任意 JSON，后端在消息体里注入 `user` 和 `user_id` 后，广播给房间内**除发送者本人外**的所有人。当前前端用到两类：

The frontend sends arbitrary JSON; after injecting `user` and `user_id` into the message body, the backend broadcasts it to **everyone in the room except the sender**. The current frontend uses two kinds:

**光标位置**（DocEditor.vue 里 Tiptap 选区变化时发）/ **Cursor position** (sent by DocEditor.vue when the Tiptap selection changes):

```json
{"type": "cursor", "pos": 123, "anchor": "alice"}
```

广播后收到：

After broadcast, the received message:

```json
{"type": "cursor", "pos": 123, "anchor": "alice", "user": "alice", "user_id": 2}
```

**操作通知**（内容修改后发，仅通知"我改了"，不携带增量）/ **Operation notice** (sent after content changes; only notifies "I changed it", carries no delta):

```json
{"type": "op", "ts": 1728000000}
```

### 8.4 协作语义（重要）/ Collaboration Semantics (Important)

- 正文保存走 HTTP `PUT /api/docs/:id/content`，**整段覆盖**。
  - Body saving goes through HTTP `PUT /api/docs/:id/content`, **overwriting the whole body**.
- 两个用户同时编辑同一篇文档，最后一个 `PUT` 成功的人覆盖前者 —— **最后写入获胜（Last Write Wins）**。
  - When two users edit the same document simultaneously, the last one whose `PUT` succeeds overwrites the former — **Last Write Wins**.
- WS 只做"在线列表 + 光标广播 + 有人动了"的通知，**不传增量操作，没有 OT/CRDT**。
  - WS only does "online list + cursor broadcast + someone moved" notifications; it **does not transmit delta operations and has no OT/CRDT**.
- 多人同时改一格表格或一段文字，会互相覆盖，不会合并。这是有意为之的简化，见"已知差距"第 4 条。
  - When multiple people edit the same spreadsheet cell or text passage simultaneously, they overwrite each other and are not merged. This is a deliberate simplification; see "Known Gaps" item 4.

---

## 9. 与 WPS 个人版功能对标表 / Feature Comparison vs WPS

| 功能项 / Feature | WPS 个人版 / WPS Personal | PicoOffice 实现 / PicoOffice Implementation | 差距 / Gap |
|---|---|---|---|
| 文字文档编辑 / Word editing | 完整 Word 兼容<br>Full Word compatibility | Tiptap：加粗/斜体/下划线/标题/列表/引用/代码/表格/图片/颜色<br>Tiptap: bold/italic/underline/heading/list/quote/code/table/image/color | 没有复杂排版（分栏、文本框、艺术字）<br>No complex layout (columns, text boxes, WordArt) |
| 电子表格 / Spreadsheet | 完整 Excel 兼容<br>Full Excel compatibility | Luckysheet 2.1.13：单元格编辑、公式、多 sheet、基本格式化<br>Luckysheet 2.1.13: cell editing, formulas, multiple sheets, basic formatting | 没有宏、没有 VBA、没有数据透视表高级项<br>No macros, no VBA, no advanced pivot tables |
| 演示文稿 / Presentation | 完整 PPT 兼容<br>Full PPT compatibility | 自写 slide 编辑器：加/删 slide、编辑文字、全屏播放<br>Custom slide editor: add/delete slides, edit text, full-screen playback | 没有动画、没有切换效果、没有母版<br>No animations, no transitions, no master slides |
| 导出 PDF / Export PDF | 有 / Yes | soffice --headless 转<br>Converted via soffice --headless | 依赖服务器装 LibreOffice<br>Depends on LibreOffice installed on the server |
| 导出 docx/xlsx / Export docx/xlsx | 有 / Yes | soffice 转<br>Converted via soffice | 同上<br>Same as above |
| 多人实时协作 / Real-time collaboration | 有（云端）<br>Yes (cloud) | 最后写入获胜 + 在线列表 + 光标广播<br>Last-write-wins + online list + cursor broadcast | 无 OT/CRDT，会互相覆盖<br>No OT/CRDT, can overwrite each other |
| 文档版本历史 / Version history | 有 / Yes | 无<br>None | 不可回滚<br>No rollback |
| 文件上传附件 / File attachments | 有（云文档）<br>Yes (cloud docs) | storage/uploads 本地磁盘<br>storage/uploads on local disk | 无 OSS/S3<br>No OSS/S3 |
| 模板库 / Template library | 有 / Yes | 无<br>None | 版权/工作量原因<br>Copyright/effort reasons |
| 云字体 / Cloud fonts | 有 / Yes | 无<br>None | 版权原因<br>Copyright reasons |
| OCR 文字识别 / OCR | 有 / Yes | 无<br>None | 工作量大<br>Heavy effort |
| PDF 批注 / PDF annotation | 有 / Yes | 无<br>None | 工作量大<br>Heavy effort |
| 邮件客户端 / Mail client | 有 / Yes | 无<br>None | 不在本套件范围<br>Out of scope for this suite |
| 日历 / Calendar | 有 / Yes | 无<br>None | 不在本套件范围<br>Out of scope for this suite |
| 网盘同步 / Cloud sync | 有 / Yes | 无<br>None | 本地磁盘即"网盘"<br>Local disk is the "cloud drive" |
| 全文搜索 / Full-text search | 有 / Yes | 仅按 title like<br>Only `title` LIKE | 无倒排索引<br>No inverted index |
| 离线编辑 / Offline editing | 有（客户端缓存）<br>Yes (client cache) | 无（前端必须连后端）<br>None (frontend must connect to backend) | 无 Service Worker 缓存策略<br>No Service Worker caching strategy |
| 用户体系 / User system | 有（WPS 账号）<br>Yes (WPS account) | 本地注册 + 第一个用户 admin<br>Local registration + first user admin | 无 SSO/OAuth<br>No SSO/OAuth |
| 权限管理 / Permission management | 有（分享链接/协作者）<br>Yes (share links/collaborators) | 仅 owner 可见，admin 看全部<br>Only owner visible, admin sees all | 无细粒度分享<br>No fine-grained sharing |

---

## 10. 安全设计 / Security Design

### 10.1 密码存储 / Password Storage

- 注册时 `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`。
  - At registration: `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`.
- 登录时 `bcrypt.CompareHashAndPassword`。
  - At login: `bcrypt.CompareHashAndPassword`.
- `password_hash` 字段在 struct 上标 `json:"-"`，任何 API 响应都不会带回。
  - The `password_hash` field is tagged `json:"-"` on the struct, so no API response ever returns it.

### 10.2 JWT

- 签发：`middleware.IssueToken(uid, username, role)`，claims 里带 `uid`、`username`、`role`。
  - Issuance: `middleware.IssueToken(uid, username, role)`, with `uid`, `username`, `role` in the claims.
- 密钥来源：环境变量 `PICO_JWT_SECRET`，默认 `"picooffice-dev-secret"`。
  - Key source: the environment variable `PICO_JWT_SECRET`, default `"picooffice-dev-secret"`.
- **生产部署必须改**，否则任何人都能伪造 token。
  - **Must be changed for production deployment**, otherwise anyone can forge tokens.
- 中间件 `middleware.JWT()` 解析后把 `uid`/`username`/`role` 塞进 gin context。
  - The `middleware.JWT()` middleware parses it and stuffs `uid`/`username`/`role` into the gin context.
- WS 不走 Header，token 走 query `?token=`，在 `handler.WSDocs` 里手动 `ParseToken`。
  - WS does not use Header; the token goes via query `?token=`, manually `ParseToken`'d in `handler.WSDocs`.

### 10.3 角色与权限 / Roles and Permissions

- 普通 `user`：只能 CRUD 自己的 docs（`owner_id = 当前 uid`），只能看自己上传的 files。
  - A normal `user` can only CRUD their own docs (`owner_id = current uid`) and only see files they uploaded.
- `admin`：额外能访问 `/api/admin/users` 看全部用户列表。
  - An `admin` can additionally access `/api/admin/users` to see the full user list.
- 当前**没有**"把文档分享给指定用户"的能力，文档默认私域。
  - There is currently **no** ability to "share a document with a specific user"; documents are private by default.

### 10.4 CORS

- `internal/middleware/cors.go` 中间件允许跨域，方便开发期前端 dev server (Vite :5173) 直连后端 (:8080)。
  - The `internal/middleware/cors.go` middleware allows cross-origin requests, so the frontend dev server (Vite :5173) can connect directly to the backend (:8080) during development.
- 生产环境前端由后端 `go:embed` 直接吐，同源，CORS 不影响。
  - In production the frontend is served directly by the backend via `go:embed`, same-origin, so CORS is a non-issue.

### 10.5 文件上传 / File Upload

- multipart/form-data，field 名 `file`。
  - multipart/form-data, field name `file`.
- 落盘到 `storage/uploads/`，文件名服务端重命名（避免路径穿越）。
  - Saved to `storage/uploads/`, with the server renaming the file (to avoid path traversal).
- `original_name` 只用于展示和下载时的 `Content-Disposition`。
  - `original_name` is only used for display and the `Content-Disposition` at download time.
- 未做 MIME 白名单校验（当前接受任意类型），私有化内网使用可接受；公网部署建议在 Nginx 层加 `client_max_body_size` 并加 MIME 校验。
  - No MIME whitelist validation is done (any type is currently accepted), which is acceptable for a private intranet; for public deployment it is recommended to add `client_max_body_size` at the Nginx layer and enforce MIME validation.

### 10.6 WebSocket Origin

- `upgrader.CheckOrigin = func(r *http.Request) bool { return true }`，即不校验 Origin。
  - `upgrader.CheckOrigin = func(r *http.Request) bool { return true }`, i.e. the Origin is not validated.
- 私有化部署建议在 Nginx 层限制来源；公网部署前应改回白名单校验。
  - For private deployment it is recommended to restrict the source at the Nginx layer; before public deployment it should be switched back to whitelist validation.

### 10.7 传输层 / Transport Layer

- 裸 HTTP 监听 `:8080` 不加密。
  - Bare HTTP listens on `:8080` with no encryption.
- 生产必须前面挂 Nginx + certbot 上 HTTPS/WSS，见 DEPLOY.md。
  - In production you must put Nginx + certbot in front for HTTPS/WSS; see DEPLOY.md.

---

## 11. 已知差距（诚实清单） / Known Gaps

以下是当前版本明确没做或做了简化的事，部署前请知悉：

The following are things the current version explicitly did not do or simplified; please be aware before deploying:

1. **Windows 真 nsis 安装包需要在 Windows/Mac 上跑** `npm run dist:win`。沙箱是 Linux 且无 wine，打不出来。当前 `desktop-electron/dist/PicoOffice-1.0.0-win.zip` 是绿色版，解压后 `win-unpacked/PicoOffice.exe` 双击即用；nsis 真安装包需要用户在 Windows/Mac 上进 `desktop-electron/` 目录执行 `npm install && npm run dist:win`。
   - **A real Windows nsis installer must be built on Windows/Mac** via `npm run dist:win`. The sandbox is Linux without wine, so it cannot be produced. The current `desktop-electron/dist/PicoOffice-1.0.0-win.zip` is a portable build; after extraction `win-unpacked/PicoOffice.exe` runs on double-click. A real nsis installer requires the user to enter the `desktop-electron/` directory on Windows/Mac and run `npm install && npm run dist:win`.
2. **macOS .dmg 和 iOS .ipa 必须在 Mac 上构建**（沙箱无 Xcode）。工程和脚本已留好，`mobile-capacitor/README.md` 里写了完整命令。
   - **macOS .dmg and iOS .ipa must be built on a Mac** (the sandbox has no Xcode). The project and scripts are already prepared, and the full commands are written in `mobile-capacitor/README.md`.
3. **表格用 Luckysheet 而非 Univer**。原计划 Univer，但 npmmirror 上包名/版本对不上，装不上，降级到 Luckysheet 2.1.13 是有意选择，不是 bug。
   - **Spreadsheets use Luckysheet rather than Univer**. Univer was the original plan, but the package name/version on npmmirror did not match and could not be installed; downgrading to Luckysheet 2.1.13 was a deliberate choice, not a bug.
4. **协作是"最后写入获胜 + 在线用户列表 + 光标广播"**，没有 OT/CRDT。多人同时改一格会互相覆盖。
   - **Collaboration is "last-write-wins + online user list + cursor broadcast"**, with no OT/CRDT. Multiple people editing the same cell simultaneously overwrite each other.
5. **没有文档版本历史 / 回滚**。`PUT /api/docs/:id/content` 直接覆盖，旧正文不可恢复。
   - **No document version history / rollback**. `PUT /api/docs/:id/content` overwrites directly and the old body cannot be recovered.
6. **没有 WPS 专有模板、云字体、OCR、PDF 批注**。这些要么有版权要么工作量大，本版本不做。
   - **No WPS-specific templates, cloud fonts, OCR, or PDF annotation**. These are either copyrighted or heavy-effort, and are not done in this version.
7. **离线编辑不支持**。前端必须连后端，没有 Service Worker 缓存草稿。
   - **Offline editing is not supported**. The frontend must connect to the backend; there is no Service Worker caching drafts.
8. **没有 OSS/S3 存储后端**。文件只存本地磁盘 `storage/uploads/`。
   - **No OSS/S3 storage backend**. Files are stored only on the local disk `storage/uploads/`.
9. **没有邮件、日历、网盘同步等 WPS 套件里的周边模块**。本版本只做三件套。
   - **No mail, calendar, cloud-sync, or other peripheral modules from the WPS suite**. This version only does the three core apps.
10. **没有全文搜索**。Dashboard 搜索只对 `title` 做 `like`，不搜正文。
   - **No full-text search**. Dashboard search only does `like` on `title`, not on the body.

---

## 12. 开源组件致谢 / Open Source Acknowledgements

PicoOffice 站在以下开源项目的肩膀上：

PicoOffice stands on the shoulders of the following open-source projects:

| 组件 / Component | License | 用途 / Purpose |
|---|---|---|
| Go | BSD-3-Clause | 编程语言<br>Programming language |
| Gin | MIT | Web 框架<br>Web framework |
| GORM | MIT | ORM |
| gorilla/websocket | BSD-2-Clause | WebSocket |
| golang-jwt/jwt | MIT | JWT |
| golang.org/x/crypto | BSD-3-Clause | bcrypt |
| mattn/go-sqlite3 | MIT | SQLite 驱动<br>SQLite driver |
| Vue | MIT | 前端框架<br>Frontend framework |
| Vite | MIT | 构建工具<br>Build tool |
| Element Plus | MIT | UI 组件库<br>UI component library |
| Pinia | MIT | 状态管理<br>State management |
| Vue Router | MIT | 路由<br>Router |
| Axios | MIT | HTTP 客户端<br>HTTP client |
| Tiptap | MIT | 文字编辑器<br>Word editor |
| Luckysheet | MIT | 表格编辑器<br>Spreadsheet editor |
| Electron | MIT | 桌面壳<br>Desktop shell |
| electron-builder | MIT | 桌面打包<br>Desktop packaging |
| Capacitor | MIT | 移动壳<br>Mobile shell |
| LibreOffice | MPL-2.0 / LGPL | 文档格式转换 (soffice)<br>Document format conversion (soffice) |

PicoOffice 本身按 MIT 协议发布。使用 LibreOffice 进行格式转换时，请遵守 LibreOffice 的许可证。

PicoOffice itself is released under the MIT license. When using LibreOffice for format conversion, please comply with LibreOffice's license.

---

## 13. 环境变量速查 / Environment Variables

| 变量 / Variable | 默认值 / Default | 说明 / Notes |
|---|---|---|
| `PICO_PORT` | `8080` | HTTP 监听端口<br>HTTP listen port |
| `PICO_DB` | `./picooffice.db` | SQLite 文件路径；要切 MySQL 填 MySQL DSN<br>SQLite file path; fill in a MySQL DSN to switch to MySQL |
| `PICO_STORAGE` | `../storage` | 上传附件落盘根目录<br>Root directory where uploaded attachments are written |
| `PICO_JWT_SECRET` | `picooffice-dev-secret` | JWT 签名密钥，生产必须改<br>JWT signing secret, must be changed in production |

启动日志第一行会打印当前生效的 `addr`、`db`、`storage` 三个值，部署后先看这一行确认环境变量没设错。

The first line of the startup log prints the currently effective `addr`, `db`, and `storage` values; after deployment, check this line first to confirm the environment variables are set correctly.

---

## 14. 第二轮新增功能 / v0.2 New Features

第一轮（前面章节）跑通了三件套的基本骨架。第二轮在不推翻既有表结构和 API 的前提下，把"能自己用"补成"能给一个小团队用"：加了配额、限流、分块上传、版本历史、PWA 离线、WS 心跳、PDF 批注。

The first round (the previous chapters) got the basic skeleton of the three core apps working. The second round, without overturning the existing table structures and APIs, upgrades "usable by yourself" into "usable by a small team": it adds quota, rate limiting, chunked upload, version history, PWA offline, WS heartbeat, and PDF annotation.

### 14.1 后端新增 / Backend Additions

| 模块 / Module | 改动 / Change | 说明 / Notes |
|---|---|---|
| SQLite | WAL 模式 + `busy_timeout=5000` | 读写并发不锁库；连接池 10 open / 5 idle / 1h life<br>Read/write concurrency without locking the DB; connection pool 10 open / 5 idle / 1h life |
| User 表 / User table | 新增 `quota_bytes`、`used_bytes`<br>Added `quota_bytes`, `used_bytes` | 默认 1GB；第一个注册的 admin 自动给 10GB<br>Default 1GB; the first registered admin automatically gets 10GB |
| 登录/注册 / Login/Register | IP 限流<br>IP rate limiting | 每 IP 每分钟 10 次，超过返回 429<br>10 times per IP per minute; exceeding returns 429 |
| 文件上传 / File upload | 白名单 + 大小限制<br>Whitelist + size limit | 见 14.5 配额策略表<br>See the quota policy table in 14.5 |
| 分块上传 / Chunked upload | init → chunk → complete → status | 支持断点续传，临时目录 `storage/tmp/{upload_id}/`<br>Supports resumable upload; temp dir `storage/tmp/{upload_id}/` |
| 文件管理 / File management | `DELETE /api/fs/files/:id` | 用户删自己上传的文件，扣 `used_bytes`<br>User deletes their own uploaded file, deducting `used_bytes` |
| 配额查询 / Quota query | `GET /api/me/quota` | 前端 Dashboard 进度条用<br>For the Dashboard progress bar on the frontend |
| 配额调整 / Quota adjustment | `PUT /api/admin/users/:id/quota` | 仅 admin，Admin 列表每行一个"改配额"按钮<br>Admin only; one "change quota" button per row in the Admin list |
| WebSocket | 服务端心跳<br>Server heartbeat | 30s 发 ping；60s 收不到任何消息就踢<br>Sends ping every 30s; kicks if no message received within 60s |
| 版本历史 / Version history | DocVersion 表<br>DocVersion table | 每次 PUT content 留快照，每 doc 最多 20 版（超出删最老）<br>Snapshots on every PUT content; up to 20 versions per doc (oldest deleted when exceeded) |
| 版本 API / Version API | 列表 / 详情 / 回滚<br>List / detail / rollback | 见 14.3<br>See 14.3 |
| PDF 导出 / PDF export | 改用 Flat ODT 包 A4<br>Switched to Flat ODT for A4 | `@page` CSS 会被 soffice 忽略，改成生成 `.fodt` 再转 PDF，固定 595×842pt A4<br>`@page` CSS is ignored by soffice; instead generate `.fodt` then convert to PDF, fixed 595×842pt A4 |

### 14.2 前端新增 / Frontend Additions

| 编辑器 / Editor | 新增能力 / New Capabilities |
|---|---|
| DocEditor | 页面设置（A4/A5/Letter、横竖、边距、分栏、行距、首行缩进）；样式下拉 H1-H3/正文/引用；样式管理改字号颜色字重；页眉页脚页码；自动目录；图片环绕 7 种（inline / float-left / float-right / 四周 / 紧密 / 衬于下方 / 浮于上方）；表格合并拆分 / 底色 / 边框三态；分页符 / 换行符；查找替换 Ctrl+F / Ctrl+H；字数统计；大纲视图；撤销重做；版本历史抽屉<br>Page setup (A4/A5/Letter, landscape/portrait, margins, columns, line spacing, first-line indent); style dropdown H1-H3/body/quote; style management for font size/color/weight; header/footer/page numbers; auto TOC; 7 image wrapping modes (inline / float-left / float-right / square / tight / behind / in front); table merge/split / background color / 3-state borders; page break / line break; find-replace Ctrl+F / Ctrl+H; word count; outline view; undo-redo; version history drawer |
| SheetEditor | 冻结首行 + 冻结到选中；筛选排序；色阶条件格式；canvas 柱状 / 折线 / 饼图；多 sheet 标签；公式引擎 SUM / AVERAGE / COUNT / VLOOKUP / IF；数字格式（常规 / 千分位 / 货币 / 百分比 / 日期）<br>Freeze first row + freeze to selection; filter/sort; color-scale conditional formatting; canvas bar / line / pie charts; multiple sheet tabs; formula engine SUM / AVERAGE / COUNT / VLOOKUP / IF; number formats (general / thousands / currency / percent / date) |
| SlideEditor | 母版模式；5 套主题色（蓝/绿/紫/橙/灰）；演讲者备注；fade / slide 切换动画；全屏播放<br>Master mode; 5 theme colors (blue/green/purple/orange/gray); speaker notes; fade / slide transition animations; full-screen playback |
| PWA | 手写 sw.js（app shell 预缓存 + 文档 GET 网络优先 + POST/PUT 不缓存）；manifest.webmanifest；自绘图标 icon-192/512（蓝色圆角方块白字 P）；IndexedDB 存草稿 + 同步队列；离线状态 Tag（在线 / 离线 / 待同步 N 条）；联网自动同步<br>Hand-written sw.js (app shell precache + doc GET network-first + POST/PUT not cached); manifest.webmanifest; custom icons icon-192/512 (blue rounded square, white letter P); IndexedDB for drafts + sync queue; offline status Tag (online / offline / N pending); auto-sync when online |
| PDF 批注查看器 / PDF annotation viewer | `/viewer/pdf/:id`，embed PDF + 透明叠加层框选黄色高亮 + 双击加文本批注，批注存 IndexedDB<br>`/viewer/pdf/:id`, embed PDF + transparent overlay marquee yellow highlight + double-click to add text annotation; annotations stored in IndexedDB |
| Dashboard | 右上角配额用量条 `el-progress`<br>Quota usage bar `el-progress` in the top-right |
| Admin | 用户列表每行"改配额"按钮<br>"Change quota" button on each row of the user list |
| 登录页 / Login page | 429 提示文案<br>429 hint text |
| WS 客户端 / WS client | 30s 回 pong；断线指数退避重连（3s 起，最多 5 次）<br>Replies pong every 30s; exponential backoff reconnect on disconnect (starts at 3s, max 5 times) |
| 上传 / Upload | 大文件（>5MB）自动走分块上传<br>Large files (>5MB) automatically use chunked upload |

### 14.3 新增 API / New API

| 方法 / Method | 路径 / Path | 说明 / Notes |
|---|---|---|
| GET | `/api/docs/:id/versions` | 列出该文档全部版本（最多 20 条）<br>List all versions of the document (up to 20) |
| GET | `/api/docs/:id/versions/:vid` | 取某一版本正文<br>Get a specific version's body |
| POST | `/api/docs/:id/rollback` | body `{"version_id": <vid>}`，把该版本内容写成新的一份快照并覆盖当前 content<br>body `{"version_id": <vid>}`, writes that version's content as a new snapshot and overwrites the current content |
| GET | `/api/me/quota` | 返回 `{quota_bytes, used_bytes}`<br>Returns `{quota_bytes, used_bytes}` |
| PUT | `/api/admin/users/:id/quota` | body `{"quota_bytes": 10737418240}`，仅 admin<br>body `{"quota_bytes": 10737418240}`, admin only |
| POST | `/api/fs/upload/init` | body `{filename, size, mime}`，返回 `upload_id`，创建 `storage/tmp/{upload_id}/`<br>body `{filename, size, mime}`, returns `upload_id`, creates `storage/tmp/{upload_id}/` |
| POST | `/api/fs/upload/chunk` | form：`upload_id`、`index`、`chunk` 文件；落盘为 `storage/tmp/{upload_id}/{index}`<br>form: `upload_id`, `index`, `chunk` file; written to `storage/tmp/{upload_id}/{index}` |
| POST | `/api/fs/upload/complete` | body `{upload_id, total_chunks}`，校验分片齐全后合并到 `storage/files/`，写 file 表，扣配额<br>body `{upload_id, total_chunks}`, after verifying all chunks merges into `storage/files/`, writes the file table, deducts quota |
| GET | `/api/fs/upload/status?upload_id=xxx` | 返回已收到的 chunk index 列表，断点续传用<br>Returns the list of received chunk indexes, for resumable upload |
| DELETE | `/api/fs/files/:id` | 用户删自己上传的文件，删磁盘文件 + 删行 + 退 `used_bytes`<br>User deletes their own uploaded file: deletes the disk file + deletes the row + returns `used_bytes` |

### 14.4 数据模型：DocVersion 表 / Data Model: DocVersion Table

```sql
CREATE TABLE doc_versions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    doc_id      INTEGER NOT NULL,
    doc_type    TEXT    NOT NULL,           -- doc / sheet / slide
    content     TEXT    NOT NULL,           -- 快照正文（HTML / JSON）
    title       TEXT,
    created_by  INTEGER,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_doc_versions_doc ON doc_versions(doc_id);
```

写入规则：每次 `PUT /api/docs/:id` 成功后，把旧 content 作为一条 DocVersion 插入；同一 `doc_id` 超过 20 条时，DELETE 最老的一条。回滚不删历史，只是把目标版本再写一份快照覆盖当前 content。

Write rule: after each successful `PUT /api/docs/:id`, the old content is inserted as a DocVersion; when the same `doc_id` exceeds 20 rows, the oldest one is DELETE'd. Rollback does not delete history — it just writes the target version as another snapshot overwriting the current content.

User 表新增字段：

New fields on the User table:

```sql
ALTER TABLE users ADD COLUMN quota_bytes INTEGER DEFAULT 1073741824;   -- 1GB
ALTER TABLE users ADD COLUMN used_bytes  INTEGER DEFAULT 0;
-- 第一个 admin 在初始化时单独 UPDATE 为 10737418240（10GB）
```

### 14.5 配额策略 / Quota Policy

| 项 / Item | 值 / Value |
|---|---|
| 默认用户配额 / Default user quota | 1 GiB（`1073741824` 字节）<br>1 GiB (`1073741824` bytes) |
| 第一个 admin / First admin | 10 GiB |
| 单文件上限 / Single-file limit | 100 MiB |
| 允许后缀 / Allowed extensions | `.pdf .doc .docx .xls .xlsx .ppt .pptx .txt .md .png .jpg .jpeg .gif .webp .csv .zip .mp4` |
| 超限行为 / Over-limit behavior | 上传时预校验 `used_bytes + file_size > quota_bytes`，超了直接 413；白名单不匹配 415<br>On upload, pre-check `used_bytes + file_size > quota_bytes`; over the limit returns 413 directly; whitelist mismatch returns 415 |

`used_bytes` 在上传 complete 时加上文件大小，在 DELETE file 时减去。不实时统计目录，避免每次列表都扫盘。

`used_bytes` is increased by the file size on upload complete and decreased on DELETE file. The directory is not scanned in real time, to avoid disk scans on every list.

### 14.6 PWA 离线机制 / PWA Offline Mechanism

**sw.js 策略（手写，不 workbox）/ sw.js Strategy (hand-written, no workbox):**

- 预缓存：`app shell` 四个文件——`/`、`/index.html`、`/assets/index.js`、`/assets/index.css`，install 阶段 `cache.addAll`。
  - Precaching: the four `app shell` files — `/`, `/index.html`, `/assets/index.js`, `/assets/index.css` — via `cache.addAll` during the install phase.
- 运行时 / Runtime:
  - `GET /docs/:id` 类文档请求 → **网络优先**，成功就回写 cache，失败回退 cache。
    - `GET /docs/:id`-style document requests → **network-first**; on success write back to cache, on failure fall back to cache.
  - `GET /`、静态资源 → **缓存优先**。
    - `GET /`, static assets → **cache-first**.
  - `POST /api/*`、`PUT /api/*`、`DELETE /api/*` → **完全不缓存**，离线时直接 reject。
    - `POST /api/*`, `PUT /api/*`, `DELETE /api/*` → **never cached**; rejected directly when offline.
- activate 阶段清旧 cache（`CACHE_VERSION` 字面量改一下就走全套升级）。
  - The activate phase clears old caches (change the `CACHE_VERSION` literal to trigger a full upgrade).

**IndexedDB 结构（库名 `picooffice`）/ IndexedDB Structure (database name `picooffice`):**

| store | keyPath | 用途 / Purpose |
|---|---|---|
| `drafts` | `doc_id` | 离线时未提交的正文草稿，联网后自动 PUT 回去<br>Unsubmitted body drafts while offline, auto-PUT back when online |
| `queue` | `id` (自增)<br>`id` (auto-increment) | 离线期间所有 POST/PUT/DELETE 请求体，联网后按顺序重放<br>All POST/PUT/DELETE request bodies during offline, replayed in order when online |
| `pdf_annotations` | `annotation_id` | PDF 批注本地存储，不上云（见 14.9）<br>Local PDF annotation storage, not uploaded to cloud (see 14.9) |

**同步流程 / Sync Flow:**

1. 前端 `fetch` 包一层，请求失败且 `navigator.onLine === false` 时把 body 序列化塞进 `queue`。
2. The frontend wraps `fetch`; when a request fails and `navigator.onLine === false`, it serializes the body and pushes it into `queue`.
3. `online` 事件触发后，按时间顺序逐条重放 `queue`，成功一条删一条。
4. After the `online` event fires, it replays `queue` entries in time order, deleting each one after success.
5. 顶栏 Tag 显示：在线（绿）/ 离线（灰）/ 待同步 N 条（橙，N = `queue.size`）。
6. The top-bar Tag shows: online (green) / offline (gray) / N pending (orange, N = `queue.size`).

### 14.7 WebSocket 心跳协议 / WebSocket Heartbeat Protocol

| 方向 / Direction | 事件 / Event | 周期 / Period | 说明 / Notes |
|---|---|---|---|
| 服务端 → 客户端 / Server → Client | `ping` 帧（opcode 0x9）<br>`ping` frame (opcode 0x9) | 30s | 到点就发，不管客户端上一条是什么<br>Sent on schedule regardless of the client's last message |
| 客户端 → 服务端 / Client → Server | `pong` 帧（opcode 0xA）<br>`pong` frame (opcode 0xA) | 收到 ping 立即回<br>Replied immediately on ping | 前端 ws 原生自动回 pong，不用手写<br>The frontend's ws auto-replies pong natively, no hand-writing needed |
| 服务端 / Server | 读超时 / Read timeout | 60s | 60s 内一个字节都没收到，主动 close（断线重连交给前端）<br>If not a single byte is received within 60s, actively close (reconnection is left to the frontend) |
| 前端 / Frontend | 重连 / Reconnect | 指数退避<br>Exponential backoff | 3s → 6s → 12s → 24s → 48s，最多 5 次，再不行提示"连接已断开，请刷新"<br>3s → 6s → 12s → 24s → 48s, max 5 times, then prompt "connection lost, please refresh" |

### 14.8 限流策略 / Rate Limiting Strategy

只做了一层粗粒度 IP 限流，没上 Redis：

Only one coarse-grained IP rate limiter was implemented, without Redis:

- 内存里维护 `map[ip]window[10]timestamp`。
  - In-memory `map[ip]window[10]timestamp`.
- 命中路由：`POST /api/auth/login`、`POST /api/auth/register`。
  - Targeted routes: `POST /api/auth/login`, `POST /api/auth/register`.
- 阈值：1 分钟 10 次。
  - Threshold: 10 times per minute.
- 超了返回 `429 Too Many Requests`，body `{"error": "too many attempts, slow down"}`。
  - Exceeding returns `429 Too Many Requests`, body `{"error": "too many attempts, slow down"}`.
- 重启服务清零。小规模自用足够；真要暴露公网给几百人，前面挂 Cloudflare / Nginx limit_req 更稳。
  - Cleared on restart. Enough for small-scale personal use; if you really must expose it publicly to hundreds of people, put Cloudflare / Nginx limit_req in front for stability.

### 14.9 已知差距（补充）/ Known Gaps (Supplementary)

11 节列过一轮，第二轮又新发现/新接受了这些：

Section 11 listed a round already; the second round newly discovered/accepted these:

1. **PDF 批注是前端叠加层，不改原 PDF 文件**。高亮框和文本批注都画在透明 overlay div 上，存 IndexedDB，不上云。换个浏览器/设备就看不到别人的批注。要真改 PDF 得后端跑 `pdf-lib` 重写 PDF，没做。
   - **PDF annotation is a frontend overlay and does not modify the original PDF file**. Highlight boxes and text annotations are both drawn on a transparent overlay div, stored in IndexedDB, and not uploaded to cloud. Switch browsers/devices and you won't see others' annotations. Truly modifying the PDF would require the backend to run `pdf-lib` to rewrite the PDF — not done.
2. **协作仍是 LWW（Last Write Wins），没有 OT / CRDT**。两个人同时改一段，后保存的覆盖先保存的，不会智能合并。
   - **Collaboration is still LWW (Last Write Wins), with no OT / CRDT**. Two people editing a passage simultaneously — the later save overwrites the earlier one, no intelligent merge.
3. **PWA 离线编辑多端同时改会冲突**。A 电脑离线改了文档，B 电脑离线也改了同一篇，联网后两条 PUT 谁后到谁赢，A 的修改被吞掉。离线只适合单人单端临时断网的场景。
   - **PWA offline editing conflicts across multiple devices editing simultaneously**. If A's computer edits a document offline and B's computer also edits the same one offline, after reconnecting whichever PUT arrives later wins and A's changes are swallowed. Offline is only suitable for a single person on a single device temporarily losing connection.
4. **Windows nsis exe 必须在 Windows 上打**；macOS dmg 和 iOS ipa 必须在 Mac 上打。Linux 上交叉打包不了签名链。
   - **The Windows nsis exe must be built on Windows**; macOS dmg and iOS ipa must be built on a Mac. The signing chain cannot be cross-packaged on Linux.
5. **PDF 导出走 Flat ODT** 解决了 A4 页面尺寸，但字体回退仍依赖服务器装的字体。中文服务器要装 `fonts-noto-cjk`，否则导出 PDF 中文变方块。
   - **PDF export via Flat ODT** solves the A4 page size, but font fallback still depends on the fonts installed on the server. A Chinese-language server must install `fonts-noto-cjk`, otherwise Chinese characters become boxes in the exported PDF.

### 14.10 环境变量补充 / Environment Variable Supplement

第二轮没新增环境变量，但有两个隐含约定写在这里，避免新人踩坑：

The second round added no new environment variables, but two implicit conventions are written here to keep newcomers from stepping on traps:

| 项 / Item | 约定 / Convention |
|---|---|
| SQLite 文件 / SQLite file | 必须在同一台磁盘上，WAL 会生成 `picooffice.db-wal` 和 `picooffice.db-shm` 两个伴随文件，备份时一起拷<br>Must be on the same disk; WAL generates two companion files `picooffice.db-wal` and `picooffice.db-shm`, copy them together when backing up |
| `PICO_STORAGE` | 必须是持久盘，分块上传的临时目录 `storage/tmp/{upload_id}/` 在重启时要能续传；tmp 目录每小时清理一次超过 24h 的残留<br>Must be a persistent disk; the chunked-upload temp dir `storage/tmp/{upload_id}/` must support resumption after restart; the tmp dir is cleaned hourly of residue older than 24h |
| 上传白名单 / Upload whitelist | 后端在 `/api/fs/upload/init` 校验后缀，前端也做一遍过滤，两道防线<br>The backend validates extensions in `/api/fs/upload/init`, and the frontend also filters — two lines of defense |
| DocVersion 快照 / DocVersion snapshot | 只存 content 文本，不存附件。附件始终以最新版为准<br>Only stores content text, not attachments. Attachments always use the latest version |

### 14.11 第二轮工作量估算（回顾用）/ Second-Round Effort Estimate (for review)

| 模块 / Module | 大致改动 / Approximate Changes |
|---|---|
| 后端 / Backend | GORM 加字段、连接池初始化、中间件限流、分块上传状态机、DocVersion CRUD、WS ping ticker、Flat ODT 模板生成<br>GORM field additions, connection pool init, middleware rate limiting, chunked-upload state machine, DocVersion CRUD, WS ping ticker, Flat ODT template generation |
| 前端 / Frontend | Tiptap 扩展（页面设置/图片环绕/查找替换/大纲）、Luckysheet 接入公式与条件格式、SlideEditor 母版与主题色、PWA 手写 sw.js + IndexedDB 封装、PDF viewer 叠加层、Dashboard 配额条、Admin 改配额弹窗<br>Tiptap extensions (page setup/image wrap/find-replace/outline), Luckysheet formula & conditional formatting integration, SlideEditor master & theme colors, hand-written PWA sw.js + IndexedDB wrapper, PDF viewer overlay, Dashboard quota bar, Admin change-quota dialog |
| 部署 / Deployment | cloudflared / frp / Tailscale / DDNS 四种公网暴露方式的文档化<br>Documentation of four public-exposure methods: cloudflared / frp / Tailscale / DDNS |

### 14.12 已知不做的事 / Known Non-Goals

下一轮也不打算做，先写下来免得反复讨论：

Not planned for the next round either; written down to avoid repeated discussion:

- 不做 OT / CRDT 协同编辑，LWW 够用。
  - No OT / CRDT collaborative editing; LWW is sufficient.
- 不做全文搜索，sqlite FTS5 留到真有搜索需求再说。
  - No full-text search; sqlite FTS5 is deferred until there is a real search need.
- 不做邮件 / 日历 / IM，PicoOffice 就是三件套。
  - No mail / calendar / IM; PicoOffice is just the three core apps.
- 不做权限细粒度（文档级 read/write 分享链接），现在只有"登录就能看全部"，小团队够。
  - No fine-grained permissions (document-level read/write share links); currently only "log in to see everything", which is enough for small teams.
- 不做多租户 / SaaS 化，一套部署服务一个团队。
  - No multi-tenancy / SaaS-ification; one deployment serves one team.

### 14.13 小结 / Summary

第二轮目标是把"单机 demo"补成"小团队真能用"：配额管资源、限流防爆破、分块上传解决大文件、版本历史给后悔药、PWA 给断网兜底、公网部署章节让运维能照着抄。所有改动都走增量，老用户升级只需重新编译后端 + 重新 build 前端，SQLite 用 GORM AutoMigrate 自动加列，不丢数据。

The second-round goal was to turn the "single-machine demo" into "genuinely usable by a small team": quota manages resources, rate limiting prevents brute force, chunked upload solves large files, version history provides a regret pill, PWA covers disconnection, and the public-deployment chapter lets ops copy from it. All changes are incremental; existing users only need to recompile the backend + rebuild the frontend to upgrade, and SQLite uses GORM AutoMigrate to add columns automatically without data loss.

---

## 15. 第三轮新增功能 / v0.3 New Features

第三轮目标：从"内部三件套"补成"能对外分享、能收邮件日历、有真协作"的工作台。依然是增量改，老库 AutoMigrate 加列加表，老前端 build 完直接覆盖。

Third-round goal: upgrade from "internal three core apps" to a "workbench that can share externally, receive mail/calendar, and has real collaboration". Still incremental changes — the old DB gets columns/tables via AutoMigrate, and the old frontend is directly overwritten after build.

### 15.1 新增功能总表（按阶段）/ New Feature Summary (by Phase)

| 阶段 / Phase | 优先级 / Priority | 功能 / Feature | 一句话说明 / One-line description |
|---|---|---|---|
| 1 | P0 | Office 文件导入 / Office file import | docx/xlsx/pptx（兼容老 doc/xls/ppt）上传后经 soffice 转 html，自动建对应类型文档<br>docx/xlsx/pptx (compat with old doc/xls/ppt) converted to html via soffice after upload, auto-creating the matching doc type |
| 1 | P0 | PPTX 导出 / PPTX export | slide 走 soffice 直接转 pptx，不再 odp 中转<br>slides converted directly to pptx via soffice, no longer via odp |
| 1 | P0 | 文档分享 / Document sharing | private/shared 两态，share_token + view/edit 权限，匿名只读公开链接<br>private/shared two states, share_token + view/edit permission, anonymous read-only public link |
| 1 | P0 | HTML 消毒 / HTML sanitization | bluemonday 写入前过滤，放行排版标签与安全 style，剥 script/iframe/on*/javascript:<br>bluemonday filters before write, allowing layout tags and safe style, stripping script/iframe/on*/javascript: |
| 1 | P0 | 软删除 + 回收站 / Soft delete + trash | gorm.DeletedAt，30 天自动清理<br>gorm.DeletedAt, auto-purge after 30 days |
| 1 | P0 | 配额口径统一 / Unified quota basis | 正文 + 历史快照 + 上传文件统一 RecalcUsedBytes<br>body + history snapshots + uploaded files unified via RecalcUsedBytes |
| 2 | P1 | 导出 worker pool / Export worker pool | 默认 2 worker，120s 超时，PICO_EXPORT_WORKERS 调<br>Default 2 workers, 120s timeout, tuned by PICO_EXPORT_WORKERS |
| 2 | P1 | Docker 部署 / Docker deployment | 多阶段 Dockerfile + docker-compose，带 libreoffice 和 fonts-noto-cjk<br>Multi-stage Dockerfile + docker-compose, with libreoffice and fonts-noto-cjk |
| 2 | P1 | 表格函数与格式 / Spreadsheet functions & formats | SUM/AVERAGE/COUNT/VLOOKUP/IF，数字/货币/百分比/日期格式，多 sheet<br>SUM/AVERAGE/COUNT/VLOOKUP/IF, number/currency/percent/date formats, multiple sheets |
| 2 | P1 | 乐观锁 / Optimistic locking | DocVersion，base_version 不符返回 409<br>DocVersion, base_version mismatch returns 409 |
| 2 | P1 | 打印 / Print | @media print + window.print |  |
| 2 | P1 | refresh token + 密码找回 / Refresh token + password recovery | access 2h + refresh 30d 轮转，登出吊销，reset token 找回<br>access 2h + refresh 30d rotation, revoked on logout, reset token for recovery |
| 2 | P1 | 文件组织 / File organization | 文件夹树、标签、星标、批量操作、最近排序<br>Folder tree, tags, stars, batch operations, recent sort |
| 3 | 新模块 / New module | 个人资料 / 头像 / 2FA / Profile / avatar / 2FA | TOTP(pquerna/otp)，登录设备管理<br>TOTP (pquerna/otp), login device management |
| 3 | 新模块 / New module | 隐私条款 / 服务条款 / Privacy / Terms | 隐私政策 / 服务条款，注册强制勾选<br>Privacy Policy / Terms of Service, mandatory checkbox at registration |
| 3 | 新模块 / New module | 真 OT 协作 / Real OT collaboration | 仅文字，ProseMirror collab step，表格/演示仍 LWW<br>Documents only, ProseMirror collab step, spreadsheets/presentations still LWW |
| 3 | 新模块 / New module | 移动触摸适配 / Mobile touch adaptation | 触摸友好工具栏 / 手势缩放 / 软键盘<br>Touch-friendly toolbar / gesture zoom / soft keyboard |
| 3 | 新模块 / New module | 邮件 / Mail | go-imap/v2 拉取 + net/smtp 发送，多账号<br>go-imap/v2 fetch + net/smtp send, multi-account |
| 3 | 新模块 / New module | 日历 / Calendar | 事件 CRUD + 提醒<br>Event CRUD + reminders |
| 3 | 新模块 / New module | 全文搜索 / Full-text search | bleve + cjk bigram 中文分词<br>bleve + cjk bigram Chinese tokenization |
| 3 | 新模块 / New module | S3/OSS 存储 / S3/OSS storage | Storage 抽象 local/s3，minio-go<br>Storage abstraction local/s3, minio-go |
| 4 | P2 | 健康探针 / Health probes | /healthz liveness、/readyz readiness |  |
| 4 | P2 | 测试 / 日志 / 审计 / Tests / logs / audit | service/storage 单测、slog 结构化、审计日志<br>service/storage unit tests, slog structured, audit log |
| 4 | P2 | i18n | 中 / 英切换<br>Chinese / English switch |  |

### 15.2 新增 API 全表 / New API Full Table

认证中间件照旧：除 `/api/share/:token`、`/api/healthz`、`/api/readyz`、`/api/auth/*` 外都要登录。

The auth middleware is unchanged: everything except `/api/share/:token`, `/api/healthz`, `/api/readyz`, `/api/auth/*` requires login.

| 方法 / Method | 路径 / Path | 说明 / Notes | 备注 / Remark |
|---|---|---|---|
| POST | /api/docs/import | 导入 office 文件<br>Import office file | multipart `file` 字段；docx/xlsx/pptx（兼容 doc/xls/ppt）；服务端 soffice 转 html 建 doc_kind=doc/sheet/slide，返回新 doc，前端自动跳编辑器<br>multipart `file` field; docx/xlsx/pptx (compat doc/xls/ppt); server converts to html via soffice, creates doc_kind=doc/sheet/slide, returns new doc, frontend auto-redirects to editor |
| POST | /api/docs/:id/share | 生成/更新分享<br>Generate/update share | body `{visibility, share_perm}`；返回 share_token<br>body `{visibility, share_perm}`; returns share_token |
| DELETE | /api/docs/:id/share | 取消分享<br>Cancel share | visibility 回 private<br>visibility reverts to private |
| GET | /api/shared | 列出我分享出去的 / 别人分享给我的<br>List what I shared / what others shared with me |  |
| GET | /api/share/:token | 匿名只读访问<br>Anonymous read-only access | 不需要登录；share_perm=view；edit 权限下匿名保存**未做**，见 15.10<br>No login needed; share_perm=view; anonymous saving under edit permission **not done**, see 15.10 |
| GET | /api/trash | 回收站列表<br>Trash list | 软删文档<br>Soft-deleted documents |
| POST | /api/trash/:id/restore | 还原<br>Restore | 从回收站捞回<br>Recover from trash |
| DELETE | /api/trash/:id | 彻底删除<br>Permanently delete | 不可恢复，连带清快照<br>Irrecoverable, snapshots cleared too |
| POST | /api/folders | 新建文件夹<br>Create folder | body `{name, parent_id}` |
| GET/PUT/DELETE | /api/folders/:id | 文件夹 CRUD<br>Folder CRUD | 树形<br>Tree-shaped |
| POST/DELETE | /api/docs/:id/tags | 打标签 / 摘标签<br>Tag / untag | body `{tag_id}` |
| POST/DELETE | /api/docs/:id/star | 星标 / 取消星标<br>Star / unstar | toggle |
| POST | /api/docs/batch | 批量操作<br>Batch operation | body `{doc_ids:[], action:"move\|delete\|star\|tag", ...}` |
| PUT | /api/me/profile | 改昵称/简介<br>Change nickname/bio |  |
| POST | /api/me/avatar | 上传头像<br>Upload avatar | multipart，走 Storage<br>multipart, via Storage |
| POST | /api/me/2fa/setup | 生成 TOTP 密钥<br>Generate TOTP secret | 返回 otpauth:// URI + secret<br>Returns otpauth:// URI + secret |
| POST | /api/me/2fa/enable | 开通 2FA<br>Enable 2FA | body `{code}` 校验后才置 enabled<br>body `{code}` validated before enabling |
| DELETE | /api/me/2fa | 关闭 2FA<br>Disable 2FA | 要验当前密码 + code<br>Requires verifying current password + code |
| GET | /api/me/devices | 登录设备列表<br>Login device list | UA / IP / 最近活跃<br>UA / IP / last active |
| DELETE | /api/me/devices/:id | 吊销某设备<br>Revoke a device | 该 refresh token 作废<br>That refresh token is invalidated |
| POST | /api/auth/refresh | 刷新 access token<br>Refresh access token | 轮转 refresh token，旧的立即吊销<br>Rotate refresh token, old one revoked immediately |
| POST | /api/auth/logout | 登出<br>Logout | 吊销当前 refresh token<br>Revoke current refresh token |
| POST | /api/auth/password-reset/request | 申请密码找回<br>Request password reset | 发 reset 链接（邮件未接，日志输出）<br>Sends reset link (mail not wired, logged) |
| POST | /api/auth/password-reset/confirm | 重置密码<br>Confirm password reset | body `{token, new_password}` |
| GET | /api/mail/accounts | 邮件账号列表<br>Mail account list |  |
| POST | /api/mail/accounts | 添加账号<br>Add account | imap/smtp 参数<br>imap/smtp parameters |
| GET | /api/mail/:account_id/inbox | 收件箱<br>Inbox | 拉取后落库<br>Fetched then stored in DB |
| POST | /api/mail/:account_id/send | 写信发送<br>Compose & send | net/smtp |
| GET/POST | /api/calendar/events | 日历事件列表 / 新建<br>Calendar event list / create |  |
| PUT/DELETE | /api/calendar/events/:id | 改 / 删事件<br>Edit / delete event | 带 remind_at 提醒<br>With remind_at reminder |
| GET | /api/search | 全文搜索<br>Full-text search | `?q=关键词`；`?rebuild=1` 重建索引<br>`?q=keyword`; `?rebuild=1` rebuild index |
| POST | /api/docs/:id/collab/step | 提交 OT step<br>Submit OT step | body `{version, step}`；冲突返回 steps_conflict 区间<br>body `{version, step}`; conflict returns steps_conflict range |
| GET | /api/docs/:id/collab/steps | 拉区间 steps<br>Pull range steps | `?after=version` |
| GET | /healthz | liveness | 永远 200，进程活着就行<br>Always 200, process alive is enough |
| GET | /readyz | readiness | db ping + soffice LookPath + 磁盘 Statfs，任一失败 503<br>db ping + soffice LookPath + disk Statfs, any failure returns 503 |
| GET | /api/audit | 审计日志<br>Audit log | 仅 admin，登录等敏感操作<br>Admin only, sensitive ops like login |

### 15.3 新数据模型 / New Data Model

GORM AutoMigrate 自动建表。下面字段名即 struct 字段（snake_case 落库）。

GORM AutoMigrate auto-creates the tables. The field names below are the struct fields (stored in snake_case).

**Folder（文件夹）/ Folder:**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| owner_id | uint | 谁的文件夹<br>Whose folder |
| parent_id | uint | 0 = 根<br>0 = root |
| name | string |  |
| sort | int | 同级排序<br>Sibling sort order |
| CreatedAt/UpdatedAt/DeletedAt | time |  |

**Tag / DocTag（标签）/ Tag / DocTag:**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| Tag.id | uint PK |  |
| Tag.owner_id | uint |  |
| Tag.name | string |  |
| DocTag.doc_id | uint |  |
| DocTag.tag_id | uint |  |

DocTag 联合主键 (doc_id, tag_id)。

DocTag has a composite primary key (doc_id, tag_id).

**RefreshToken（refresh 会话）/ RefreshToken (refresh session):**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| user_id | uint |  |
| token_hash | string | 存哈希不存明文<br>Store hash, not plaintext |
| user_agent | string |  |
| ip | string |  |
| expires_at | time | 30 天<br>30 days |
| revoked_at | *time | 登出/吊销置位<br>Set on logout/revoke |
| CreatedAt | time |  |

access token 2h 过期，前端拿 refresh 调 /api/auth/refresh 换新 access，同时 refresh 本身轮转（旧的 revoked，发新的）。

The access token expires in 2h; the frontend uses the refresh to call /api/auth/refresh for a new access, while the refresh itself rotates (old revoked, new issued).

**LoginDevice（登录设备）/ LoginDevice (login device):**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| user_id | uint |  |
| refresh_token_id | uint | 对应会话<br>Corresponding session |
| ua | string |  |
| ip | string |  |
| last_seen_at | time |  |

**AuditLog（审计）/ AuditLog (audit):**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| user_id | uint | 谁<br>Who |
| action | string | login / logout / 2fa_enable / share / delete_perm 等<br>login / logout / 2fa_enable / share / delete_perm, etc. |
| ip | string |  |
| ua | string |  |
| detail | string | JSON 字符串<br>JSON string |
| created_at | time | 只增不改<br>Append-only |

**MailAccount / MailMessage**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| MailAccount.id | uint PK |  |
| owner_id | uint |  |
| address | string | 邮件地址<br>Email address |
| imap_host/imap_port | string/int |  |
| smtp_host/smtp_port | string/int |  |
| password_enc | string | 仅混淆（见 15.10）<br>Only obfuscated (see 15.10) |
| CreatedAt | time |  |
| MailMessage.id | uint PK |  |
| account_id | uint |  |
| from_addr | string |  |
| to_addr | string |  |
| subject | string |  |
| body | text |  |
| uid | uint | IMAP UID，去重用<br>IMAP UID, for dedup |
| seen | bool |  |
| received_at | time |  |

**CalendarEvent**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| owner_id | uint |  |
| title | string |  |
| start_at / end_at | time |  |
| remind_at | *time | 提醒时间<br>Reminder time |
| note | string |  |

**CollabStep（OT 有序 step）/ CollabStep (ordered OT step):**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| id | uint PK |  |
| doc_id | uint |  |
| seq | uint | 单调递增版本号<br>Monotonically increasing version number |
| user_id | uint | 谁提交的<br>Who submitted |
| step | blob | ProseMirror step 序列化 JSON<br>ProseMirror step serialized JSON |
| client_id | string | 前端随机 id<br>Frontend random id |
| created_at | time |  |

**Doc 新增字段 / New Doc Fields**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| visibility | string | private / shared |
| share_token | string | 随机串，公开链接用<br>Random string for public link |
| share_perm | string | view / edit |
| folder_id | uint | 所属文件夹<br>Belonging folder |
| starred | bool | 星标<br>Starred |
| base_version | uint | 乐观锁基准版本<br>Optimistic-lock base version |
| DeletedAt | gorm.DeletedAt | 软删除<br>Soft delete |

**User 新增字段 / New User Fields**

| 字段 / Field | 类型 / Type | 说明 / Notes |
|---|---|---|
| nickname | string |  |
| avatar_key | string | Storage 里的 key<br>Key in Storage |
| totp_secret | string | 加密存，未开通为空<br>Encrypted; empty if not enabled |
| totp_enabled | bool |  |
| lang | string | zh / en |
| password_reset_token | string |  |
| password_reset_exp | time |  |

### 15.4 OT 真协作协议（仅文字文档）/ Real OT Collaboration Protocol (documents only)

第二轮是 LWW：谁后保存谁赢。第三轮文字文档换成 ProseMirror collab step，表格/演示没做（仍 LWW，见 15.10）。

The second round was LWW: whoever saves last wins. The third round switches documents to ProseMirror collab steps; spreadsheets/presentations are not done (still LWW, see 15.10).

协议基于"服务端存有序 steps + 版本号"：

The protocol is based on "the server stores ordered steps + version numbers":

1. 客户端打开文档，GET `/api/docs/:id/collab/steps?after=0`，拿到从版本 0 到当前的全部 step，逐条 apply 到本地 ProseMirror state。
2. The client opens the document and GETs `/api/docs/:id/collab/steps?after=0`, obtaining all steps from version 0 to the current one, applying each to the local ProseMirror state.
3. 本地编辑产生一个 step，带上自己当前看到的 `version`（即本地已 apply 到的最新 seq），POST `/api/docs/:id/collab/step`，body `{version, step}`。
4. Local editing produces a step, carrying the `version` it currently sees (i.e. the latest seq it has applied locally), and POSTs `/api/docs/:id/collab/step` with body `{version, step}`.
5. 服务端校验：
   - 若 `version == 服务端当前 seq`：接受，seq+1，落 CollabStep，通过 WebSocket 广播给其他在线客户端。
   - 若 `version < 服务端当前 seq`：说明你落后了，返回 `409 steps_conflict`，body 里带 `{server_version, steps: [version+1 .. server_version]}`。客户端先把这些别人的 step apply/transform 到本地，再把自己的 step transform 到最新版本，重新 POST。
6. The server validates:
   - If `version == server's current seq`: accept, seq+1, persist CollabStep, broadcast to other online clients via WebSocket.
   - If `version < server's current seq`: you are behind, return `409 steps_conflict` with body `{server_version, steps: [version+1 .. server_version]}`. The client first applies/transforms these other people's steps locally, then transforms its own step to the latest version and POSTs again.
7. 其他客户端收到广播 step，直接 apply 到本地 state（已经是基于最新版本的，无需 transform）。
8. Other clients receiving the broadcast step apply it directly to local state (it is already based on the latest version, no transform needed).
9. 离线 / 断网回来同理：打开时 after=本地最新 seq，拉中间区间，自行 transform。
10. Offline / reconnect works the same: on open, after=local latest seq, pull the middle range, transform on its own.

关键点：

Key points:

- **服务端不做 transform**，只做有序存储和转发，transform 全在客户端（ProseMirror 自带 transform）。服务端只保证 seq 单调递增、区间可取。
  - **The server does no transform**; it only stores and forwards in order. Transform is entirely on the client (ProseMirror's built-in transform). The server only guarantees monotonically increasing seq and retrievable ranges.
- step 是增量，不是全量 content。历史回看可以从 step 0 重放还原（调试用，正式历史快照仍走 DocVersion 表）。
  - A step is a delta, not the full content. History playback can replay from step 0 to reconstruct (for debugging; the official history snapshot still goes through the DocVersion table).
- 表格 sheet 和演示 slide **不接 OT**，并发编辑走 base_version 乐观锁：PUT 时带 base_version，对不上返回 409，前端提示"别人刚改了，刷新再试"。
  - Spreadsheet sheet and presentation slide **do not use OT**; concurrent edits go through base_version optimistic locking: PUT carries base_version, mismatch returns 409, and the frontend prompts "someone just changed it, refresh and retry".

### 15.5 Storage 抽象与 S3/OSS / Storage Abstraction and S3/OSS

第二轮文件全落本地磁盘 `storage/files/`。第三轮抽了一层 Storage 接口，本地和 S3/OSS 可切。

In the second round all files landed on the local disk `storage/files/`. The third round abstracts a Storage interface, switchable between local and S3/OSS.

**接口方法（storage.Storage）/ Interface Methods (storage.Storage):**

```
Put(ctx, key string, r io.Reader, size int64) error
Get(ctx, key string) (io.ReadCloser, error)
Delete(ctx, key string) error
URL(ctx, key string) (string, error)   // 预签名或本地 /files/ 路径
Stat(ctx, key string) (size int64, err error)
```

**两种实现 / Two Implementations:**

- `local`：落 `PICO_STORAGE_DIR`（默认 `./storage`），URL 走后端 `/files/:key` 静态路由。
  - `local`: lands in `PICO_STORAGE_DIR` (default `./storage`), URL goes through the backend's `/files/:key` static route.
- `s3`：minio-go SDK，兼容 AWS S3、阿里云 OSS、腾讯 COS、MinIO 自建。
  - `s3`: minio-go SDK, compatible with AWS S3, Alibaba Cloud OSS, Tencent COS, and self-hosted MinIO.

**key 规则 / Key Rules:**

- 上传文件：`files/{yyyy}/{mm}/{uuid}{ext}`
  - Uploaded files: `files/{yyyy}/{mm}/{uuid}{ext}`
- 分块临时：`tmp/{upload_id}/{part_no}`（合并后重命名到 files/）
  - Chunk temp: `tmp/{upload_id}/{part_no}` (renamed to files/ after merge)
- 头像：`avatars/{user_id}.{ext}`
  - Avatars: `avatars/{user_id}.{ext}`
- 备份包不走 Storage，仍本地 tar.gz（见 15.10）。
  - Backup packages do not go through Storage; they remain local tar.gz (see 15.10).

**环境变量 / Environment Variables:**

| 变量 / Variable | 说明 / Description | local | s3 |
|---|---|---|---|
| PICO_STORAGE_KIND | local / s3 | 必填 / Required | 必填 / Required |
| PICO_STORAGE | 本地根目录<br>Local root dir | 默认 ../storage（compose 里挂 /data/storage）<br>Default ../storage (mounted /data/storage in compose) | 忽略 / Ignored |
| PICO_S3_ENDPOINT | s3 地址<br>s3 address | 忽略 / Ignored | 如 oss-cn-hangzhou.aliyuncs.com<br>e.g. oss-cn-hangzhou.aliyuncs.com |
| PICO_S3_BUCKET | 桶名<br>Bucket name | 忽略 / Ignored | 必填 / Required |
| PICO_S3_ACCESS_KEY |  | 忽略 / Ignored | 必填 / Required |
| PICO_S3_SECRET_KEY |  | 忽略 / Ignored | 必填 / Required |
| PICO_S3_SSL | true/false | 忽略 / Ignored | 默认 false，公网 OSS 设 true<br>Default false, set true for public OSS |
| PICO_S3_REGION |  | 忽略 / Ignored | 可选 / Optional |

切换时把老 `storage/files/` 同步到桶里（rclone / mc mirror），再改 env 重启即可，代码无感。

When switching, sync the old `storage/files/` into the bucket (rclone / mc mirror), then change the env and restart — the code is agnostic.

### 15.6 全文搜索（bleve）/ Full-Text Search (bleve)

- 引擎：`bleve`，索引落 `./storage/bleve/`（local 时）或独立目录。
  - Engine: `bleve`, index lands in `./storage/bleve/` (when local) or a separate directory.
- 中文分词：cjk bigram，把中文按相邻两字切，"常斟清茶" 切成 常斟/斟清/清茶。英文按空格词。比 jieba 轻，不用维护词典。
  - Chinese tokenization: cjk bigram, splits Chinese into adjacent two-character pairs, e.g. "常斟清茶" becomes 常斟/斟清/清茶. English splits by spaces. Lighter than jieba, no dictionary maintenance.
- 索引内容：文档标题 + 正文 html 去标签后的纯文本 + 文件名。
  - Indexed content: document title + body html after stripping tags into plain text + file name.
- 更新钩子：文档 POST/PUT/DELETE（含软删）、导入完成、彻底删除，都触发对应 doc 的索引 upsert/delete。不做实时增量队列，直接同步写，量小够用。
  - Update hooks: document POST/PUT/DELETE (including soft delete), import completion, and permanent deletion all trigger the corresponding doc's index upsert/delete. No real-time incremental queue — direct synchronous writes, fine for small volumes.
- 接口：GET `/api/search?q=关键词`，返回命中文档按相关度排序。`?rebuild=1` 全量重建（索引损坏或改分词后跑一次）。
  - Interface: GET `/api/search?q=keyword`, returns matched documents sorted by relevance. `?rebuild=1` does a full rebuild (run once after index corruption or tokenizer change).
- 软删文档从索引里剔掉；回收站还原要重新加回索引。
  - Soft-deleted documents are removed from the index; trash restore must re-add them to the index.

### 15.7 配额、限流、导出队列、refresh 会话 / Quota, Rate Limiting, Export Queue, Refresh Session

**配额口径（统一）/ Quota Basis (Unified):**

第二轮配额只算上传文件。第三轮改成 `RecalcUsedBytes(doc/user)` 一次算清：

The second-round quota only counted uploaded files. The third round changes to `RecalcUsedBytes(doc/user)` computing it all at once:

```
used = 正文html字节数 + 该doc所有DocVersion快照字节数 + 该doc关联上传文件字节数
```

任何写操作（保存正文、打快照、传附件、导入）后触发 Recalc。管理员后台改配额上限后，下次写时校验。超了返回 413/402。

Recalc is triggered after any write operation (saving body, taking snapshot, uploading attachment, importing). After an admin changes the quota cap in the backend, it is validated on the next write. Exceeding returns 413/402.

**限流 / Rate Limiting:**

沿用第二轮 IP 内存限流，第三轮把 forgot-password、share 公开链接读取也加进限流名单，防刷。

Reusing the second-round IP in-memory rate limiting, the third round also adds forgot-password and share public-link reads to the rate-limit list, to prevent abuse.

**导出 worker pool / Export Worker Pool:**

- 导出（PDF/PPTX/ODT）丢进带缓冲 channel，默认 2 个 worker 跑。
  - Export (PDF/PPTX/ODT) is dropped into a buffered channel, run by 2 workers by default.
- 单任务 120s 超时，超时杀 soffice 子进程。
  - Single-task 120s timeout; on timeout the soffice child process is killed.
- worker 数 env：`PICO_EXPORT_WORKERS`，默认 2，机器多核可调到 4。
  - Worker count env: `PICO_EXPORT_WORKERS`, default 2, tunable to 4 on multi-core machines.
- 前端轮询导出任务状态（排队中 / 转换中 / 完成给下载链接 / 失败）。
  - The frontend polls export task status (queued / converting / done with download link / failed).

**refresh 会话 / Refresh Session:**

- access JWT 2h，refresh token 30d。
  - access JWT 2h, refresh token 30d.
- 每次 refresh 轮转：发新 refresh、旧 refresh 置 revoked。一个设备一个会话，可在"登录设备"里看到并吊销。
  - Each refresh rotates: issue a new refresh, set the old refresh to revoked. One session per device, viewable and revocable in "Login Devices".
- 登出吊销当前 refresh。改密码后吊销全部 refresh（强制重登）。
  - Logout revokes the current refresh. Changing the password revokes all refreshes (forces re-login).
- JWT secret 务必改强随机（见 DEPLOY）。
  - The JWT secret must be changed to a strong random value (see DEPLOY).

### 15.8 健康探针 / Health Probes

- `GET /healthz`：liveness，进程活着就 200，不查依赖。容器编排用来决定要不要重启进程。
  - `GET /healthz`: liveness, 200 as long as the process is alive, no dependency checks. Container orchestration uses it to decide whether to restart the process.
- `GET /readyz`：readiness，依次 db ping、`exec.LookPath("soffice")`、磁盘 `syscall.Statfs` 看可用字节。任一失败返回 503，带 JSON 说明哪一项挂了。负载均衡 / K8s 用它决定要不要摘流量。
  - `GET /readyz`: readiness, in turn db ping, `exec.LookPath("soffice")`, and disk `syscall.Statfs` to check available bytes. Any failure returns 503 with JSON indicating which item is down. Load balancers / K8s use it to decide whether to drain traffic.

### 15.9 i18n 与审计日志 / i18n and Audit Logs

- i18n：前端按 `User.lang`（zh/en）切，字典放前端 `locales/`。后端错误消息也带 i18n key。
  - i18n: the frontend switches by `User.lang` (zh/en), with dictionaries in the frontend `locales/`. Backend error messages also carry i18n keys.
- 审计：slog 结构化日志写文件 + stdout。登录、登出、2FA 开通/关闭、分享、彻底删除、管理员改配额，都落 AuditLog 表，admin 可在后台翻。
  - Audit: slog structured logs written to file + stdout. Login, logout, 2FA enable/disable, sharing, permanent deletion, and admin quota changes all land in the AuditLog table, reviewable by admin in the backend.

### 15.10 已知差距（第三轮诚实清单）/ Known Gaps (Third-Round Honest List)

1. **公开分享 edit 权限下，匿名用户不能保存**。`/api/share/:token` 只实现了匿名只读 view；share_perm=edit 的公开链接，匿名访问打开是只读的，要协作编辑必须登录。匿名保存改动没做（身份归属没法记）。
   - **Under public share edit permission, anonymous users cannot save**. `/api/share/:token` only implements anonymous read-only view; a public link with share_perm=edit opens read-only for anonymous visitors — collaborative editing requires login. Anonymous saving of changes is not done (identity attribution cannot be recorded).
2. **表格 sheet 和演示 slide 没有 OT**。这两类仍走 base_version 乐观锁 LWW，并发改冲突返回 409，手动刷新。只有文字文档接了 ProseMirror OT。
   - **Spreadsheet sheet and presentation slide have no OT**. These two still go through base_version optimistic-lock LWW; concurrent-edit conflicts return 409 and require manual refresh. Only documents are wired to ProseMirror OT.
3. **邮件密码仅混淆存储**。MailAccount.password_enc 是可逆混淆（base64 + 固定盐异或），不是真加密。服务进程被拿到后理论上能还原。真要安全得接 KMS / 用户主密码派生加密，没做。
   - **Mail passwords are only obfuscated**. MailAccount.password_enc is reversible obfuscation (base64 + fixed-salt XOR), not real encryption. If the service process is obtained, it can theoretically be reversed. Real security would require KMS / user-master-password-derived encryption — not done.
4. **S3 备份仍本地打包**。picobackup 用 `VACUUM INTO` 出 sqlite 再 tar.gz，包仍落在本地磁盘，没有自动推到 S3。用 S3 当主存储的部署要自己再 rclone 推一份。
   - **S3 backups are still packaged locally**. picobackup uses `VACUUM INTO` to export sqlite then tar.gz; the package still lands on local disk, with no automatic push to S3. Deployments using S3 as primary storage must rclone a copy themselves.
5. **IMAP/SMTP 未做真实往返联调**。代码接了 go-imap/v2 和 net/smtp，但只在自家测试邮箱跑通过收发；企业 Exchange / 自建邮箱的特殊鉴权（OAuth2、NTLM）没适配。
   - **IMAP/SMTP has no real round-trip integration testing**. The code is wired to go-imap/v2 and net/smtp, but was only tested against our own test mailbox; enterprise Exchange / self-hosted mailboxes' special auth (OAuth2, NTLM) is not adapted.
6. **bleve 中文是 bigram，不是分词**。搜长词会偏噪，搜短语还行。别指望它达到 Elasticsearch 水准。
   - **bleve Chinese is bigram, not tokenization**. Searching long words is noisy, phrase search is okay. Don't expect it to reach Elasticsearch levels.
7. **OT step 历史会无限长**。没做 step 压缩/快照截断，文档编辑几年后 CollabStep 表会很大。目前小团队量还无所谓。
   - **OT step history grows unbounded**. No step compression / snapshot truncation was done, so after a few years of document editing the CollabStep table will be huge. For current small-team volumes it doesn't matter.
8. **2FA 是 TOTP，没有短信/邮件二次验证**。丢了密钥只能 admin 在后台重置该用户 2FA。
   - **2FA is TOTP, with no SMS/email second factor**. If the secret is lost, only an admin can reset that user's 2FA in the backend.
9. **日历提醒靠后端轮询**，没有推送（APNs/FCM）。到点了只是下次登录时提示，关着 App 不弹。
   - **Calendar reminders rely on backend polling**, with no push (APNs/FCM). When the time comes it only prompts on next login; with the App closed nothing pops up.

### 15.11 第三轮小结 / Third-Round Summary

这轮把"能不能对外给人看"（分享/导入/导出）、"能不能长期挂着跑"（Docker/备份/探针/导出队列）、"能不能多人真协作"（OT）、"能不能当日常工作台"（邮件/日历/搜索/文件夹标签）补齐了。所有新表新列 AutoMigrate，老数据不动。坦白说几处是"打通了链路但没压测"：OT 百人大文档、S3 大文件、邮件多账号，小团队内部用够，公网大流量要再压。

This round fills in "can it be shown to outsiders" (sharing/import/export), "can it run long-term" (Docker/backup/probes/export queue), "can multiple people really collaborate" (OT), and "can it be a daily workbench" (mail/calendar/search/folders-tags). All new tables/columns are AutoMigrate, old data untouched. Frankly, a few spots are "the pipeline is connected but not load-tested": OT on hundred-person large docs, S3 large files, mail multi-account — fine for internal small-team use, but public high-traffic needs more testing.
