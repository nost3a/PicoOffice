
## 能贡献什么 / What You Can Contribute

- 修 Bug、补测试、改文档；
  - Fix bugs, add tests, improve docs.
- 新增功能（建议先开 Issue 聊清楚再动手，避免白干）；
  - Add features (recommend opening an Issue to discuss first, to avoid wasted effort).
- 翻译、国际化补全；
  - Translation and i18n completion.
- 反馈使用体验、提建议。
  - Give feedback on the experience and suggest improvements.

## 开发环境 / Development Setup

基础工具：

Basic tools:

- **Go 1.21+**（本仓库当前 `go.mod` 要求 1.26，建议直接装新版）；
  - **Go 1.21+** (this repo's current `go.mod` requires 1.26; we recommend installing the newer version directly).
- **Node.js 20+**（自带 npm）；
  - **Node.js 20+** (ships with npm).
- **Git**；
  - **Git**.
- **LibreOffice**（导入/导出转换用，不装也能跑主流程，但导入导出相关测试会跳过）；
  - **LibreOffice** (used for import/export conversion; the main flow runs without it, but import/export-related tests will be skipped).
- 可选：Docker（跑容器版）、electron-builder / Android SDK（打桌面 / 移动包）。
  - Optional: Docker (for the container version), electron-builder / Android SDK (for building desktop / mobile packages).

国内镜像建议先配上：

It's recommended to configure domestic mirrors first:

```bash
# Go 代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOFLAGS=-mod=mod

# npm 镜像
npm config set registry https://registry.npmmirror.com
```

## 拉代码并跑起来 / Clone & Run

```bash
git clone https://github.com/nost3a/PicoOffice.git
cd picooffice
```

### 后端 / Backend

```bash
cd backend
go mod download
go run main.go
# 默认监听 :8080，浏览器打开 http://localhost:8080
# 第一个注册的账号自动成为管理员
```

### 前端 / Frontend

```bash
cd frontend-web
npm install
npm run dev      # 开发模式，带热更新，通常 :5173 代理到后端
npm run build    # 产出 dist/，给后端静态托管
```

### 桌面端（Electron）/ Desktop (Electron)

```bash
cd desktop-electron
npm install
npm run dist   # electron-builder 打 deb / zip
```

### 移动端（Capacitor）/ Mobile (Capacitor)

```bash
cd mobile-capacitor
npm install
npx cap sync
cd android && ./gradlew assembleDebug   # 出 debug apk
```

## 怎么 Build 出单二进制 / Building a Single Binary

前端先 build 出 `dist/`，后端把它 embed 进去（或静态托管），再编译：

The frontend builds `dist/` first, the backend embeds it (or serves it statically), then compile:

```bash
cd frontend-web && npm run build && cd ..
cd backend && go build -o picooffice .
./picooffice
```

## 代码规范 / Code Style

本项目追求"真人商业项目"的手感，不要 AI 味：

This project aims for the feel of a "real commercial project" — avoid an AI-ish tone:

- **注释简洁实用**：写"为什么这么做"和"坑在哪"，别写"这个函数做了个函数"这种废话。
  - **Concise, practical comments**: write "why this is done" and "where the pitfalls are", not nonsense like "this function does a function".
- **命名接地气**：用 `biz_type`、`doc_kind`、`file_ext`、`owner_id` 这类一眼能懂的，别整过度抽象的 `dataProcessorV2Manager`。
  - **Down-to-earth naming**: use ones like `biz_type`, `doc_kind`, `file_ext`, `owner_id` that are understandable at a glance, not over-abstract `dataProcessorV2Manager`.
- **UI 克制**：Element Plus 默认蓝白风格，别堆大圆角、渐变、花哨阴影。头像框 / 聊天气泡这类标识逻辑放在后端，前端只负责渲染。
  - **Restrained UI**: Element Plus's default blue-white style; don't pile on big rounded corners, gradients, or fancy shadows. Identity logic like avatar frames / chat bubbles belongs on the backend; the frontend only renders.
- **别为了通用而通用**：先满足当前需求，要扩展时再重构，别上来就抽一堆接口。
  - **Don't generalize for the sake of generalizing**: satisfy the current need first, then refactor when you need to extend; don't extract a bunch of interfaces up front.
- 提交前自己先跑一遍：后端 `go vet ./...`、前端 `npm run build`。
  - Run it yourself before committing: backend `go vet ./...`, frontend `npm run build`.

## Commit 约定 / Commit Convention

用简单的前缀说明这次改动干嘛，别写废话：

Use a simple prefix to say what this change does — don't write nonsense:

| 前缀 / Prefix | 用途 / Purpose | 例子 / Example |
| ---- | ---- | ---- |
| feat | 新功能<br>New feature | `feat: 表格支持 VLOOKUP`<br>`feat: spreadsheets support VLOOKUP` |
| fix | 修 bug<br>Bug fix | `fix: 回收站还原后权限未继承`<br>`fix: permissions not inherited after trash restore` |
| docs | 文档<br>Documentation | `docs: 补 Docker 部署说明`<br>`docs: add Docker deployment notes` |
| refactor | 重构（不改行为）<br>Refactor (no behavior change) | `refactor: 抽出 storage 抽象`<br>`refactor: extract storage abstraction` |
| test | 测试<br>Tests | `test: 补 OT 冲突用例`<br>`test: add OT conflict cases` |
| chore | 杂项 / 构建<br>Chore / build | `chore: 升级 bleve 依赖`<br>`chore: upgrade bleve dependency` |

一个提交干一件事，别把功能和格式化混在一起。

One commit does one thing; don't mix features and formatting together.

## PR 流程 / Pull Request Process

1. Fork 本仓库，切出特性分支：`git checkout -b feat/xxx`；
   - Fork the repo and cut a feature branch: `git checkout -b feat/xxx`.
2. 本地改完，跑测试（见下）；
   - After local changes, run the tests (see below).
3. commit 按上面的约定写清楚；
   - Write commits clearly per the convention above.
4. push 到你的 fork，在 GitHub 上发 PR 到 `main`；
   - Push to your fork and open a PR to `main` on GitHub.
5. PR 描述里写清楚：改了什么、为什么、怎么验证的；有截图 / 录屏更好；
   - In the PR description, state clearly: what changed, why, and how to verify; screenshots / screencasts are even better.
6. 等 CI 绿了、维护者 review 过就会合入。
   - Wait for CI to go green and for a maintainer to review, then it will be merged.

## 测试要求 / Testing Requirements

提 PR 前至少跑通：

Before submitting a PR, at least get these passing:

```bash
# 后端单测
cd backend && go test ./...

# 前端 build 必须过
cd frontend-web && npm run build
```

涉及导入导出 / 转换的改动，本地装了 LibreOffice 的话手动过一遍：导个 `docx` 进去再导出来看看排版。

For changes involving import/export / conversion, if LibreOffice is installed locally, manually run through it: import a `docx` then export it and check the layout.

## 常见坑 / Common Pitfalls

- **前端 build 报内存不够**：`NODE_OPTIONS=--max-old-space-size=4096 npm run build`。
  - **Frontend build runs out of memory**: `NODE_OPTIONS=--max-old-space-size=4096 npm run build`.
- **go mod 拉不下来**：先确认 `GOPROXY` 配了国内代理，见开头。
  - **go mod can't be fetched**: first confirm `GOPROXY` is set to a domestic proxy (see the beginning).
- **导入导出相关用例挂了**：多半是没装 LibreOffice，或 `soffice` 不在 PATH；装完重开终端再跑。
  - **Import/export tests fail**: most likely LibreOffice isn't installed, or `soffice` isn't on PATH; reinstall and reopen the terminal before running.
- **Web 端调后端跨域**：开发模式走 Vite 代理，别自己在后端乱加 CORS 头。
  - **Web calling backend hits CORS**: in dev mode it goes through the Vite proxy; don't add CORS headers yourself on the backend.
- **改了后端 embed 的前端资源**：要重新 `npm run build` 再 `go build`，否则看到的还是旧页面。
  - **Changed the frontend assets embedded in the backend**: you must re-run `npm run build` then `go build`, otherwise you'll still see the old page.

## 分支与发布 / Branching & Release

- `main`：随时可发布的主线；
  - `main`: the always-releasable mainline.
- 特性分支：`feat/xxx`、`fix/xxx`，从 `main` 切出；
  - Feature branches: `feat/xxx`, `fix/xxx`, cut from `main`.
- 合入走 PR，别直接往 `main` 推。
  - Merges go through PRs; don't push directly to `main`.

## 发版 / Releasing

发版是维护者的事，普通贡献者不用管，但提功能时心里有数：合并进 `main` 后会随下个版本发布。

Releasing is the maintainers' job; ordinary contributors don't need to worry about it, but keep in mind when proposing features: once merged into `main`, they ship with the next release.

有问题随时在 Issue 里问，别客气。

Feel free to ask any questions in an Issue anytime — no need to be shy.
