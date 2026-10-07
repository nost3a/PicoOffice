# PicoOffice 使用教程 / User Guide

面向最终用户：怎么注册、怎么写文档、怎么算表格、怎么放 PPT、怎么协作、怎么在手机/电脑上用。所有路径和功能描述都对着真实界面写，不贴图。

For end users: how to register, write documents, work with spreadsheets, build presentations, collaborate, and use it on phones and computers. All paths and feature descriptions are written against the real interface — no screenshots.

---

## 1. 首次使用 / First Steps

### 1.1 拿到服务器地址 / Get the Server Address

部署方会给你一个地址，长得像：

The deployment team gives you an address that looks like:

- `https://office.yourcompany.com`（正式环境）
  - `https://office.yourcompany.com` (production)
- `http://192.168.1.10:8080`（内网测试）
  - `http://192.168.1.10:8080` (internal test)

浏览器直接打开这个地址，或在桌面/移动 App 首次启动时填进去。

Open the address directly in a browser, or enter it when the desktop/mobile app starts for the first time.

### 1.2 注册账号 / Register an Account

打开后默认跳登录页，点底部"注册"链接：

After opening, you land on the login page by default. Click the "Register" link at the bottom:

- **用户名**：≥ 3 个字符，登录用，唯一。
  - **Username**: ≥ 3 characters, used to log in, must be unique.
- **密码**：≥ 6 个字符。
  - **Password**: ≥ 6 characters.
- **昵称**：选填，不填就等于用户名。
  - **Nickname**: optional; if left blank it equals the username.

点"注册"按钮，注册成功自动登录并跳到 Dashboard。

Click the "Register" button. On success you are logged in automatically and taken to the Dashboard.

### 1.3 第一个用户自动是 admin / The First User Is Automatically admin

服务器上**第一个注册成功的人**，角色自动变成 `admin`，之后注册的全部是普通 `user`。

The **first person to register successfully** on the server automatically gets the `admin` role; everyone who registers afterward is a normal `user`.

admin 和普通用户的区别：

The difference between admin and a normal user:

| 能力 / Capability | 普通 user / Normal user | admin |
|---|---|---|
| 管理自己的文档 / Manage own docs | ✅ | ✅ |
| 上传附件 / Upload attachments | ✅ | ✅ |
| 看 `/admin` 用户管理页 / View `/admin` user page | ❌ | ✅ |
| 看别人的文档 / View others' docs | ❌（当前版本）<br>❌ (current version) | ❌（当前版本）<br>❌ (current version) |

所以**部署完第一件事就是抢注第一个账号**，把 admin 拿到手。

So **the first thing to do after deployment is to claim the first account** and grab the admin role.

---

## 2. 登录 / 登出 / Login & Logout

### 2.1 登录 / Login

- 打开地址 → 登录页。
  - Open the address → login page.
- 输入用户名 + 密码。
  - Enter username + password.
- 点"登录"，成功后跳 Dashboard。
  - Click "Login"; on success you are taken to the Dashboard.
- 失败提示："wrong username or password"。
  - On failure it shows: "wrong username or password".

### 2.2 token 存哪 / Where the Token Is Stored

前端登录成功后，token 存在浏览器 localStorage 里。下次打开自动带上，不用重复登录。token 失效（比如服务器换了 JWT 密钥）会被踢回登录页，重新登即可。

After a successful login, the frontend stores the token in the browser's localStorage. The next time you open it, the token is sent automatically so you don't have to log in again. If the token becomes invalid (e.g. the server changed its JWT secret) you get kicked back to the login page — just log in again.

### 2.3 登出 / Logout

Dashboard 右上角头像 → 点"退出登录"。

Top-right avatar on the Dashboard → click "Log out".

---

## 3. Dashboard（文件列表页） / Dashboard

登录后默认落地页，路由 `/dashboard`。

The default landing page after login, at route `/dashboard`.

### 3.1 页面布局 / Page Layout

```
┌──────────────────────────────────────────────────────┐
│  PicoOffice          [搜索框...]    [新建▾]  [头像] │
├──────────────────────────────────────────────────────┤
│  类型筛选: [全部] [文档] [表格] [演示]               │
├──────────────────────────────────────────────────────┤
│  📄 周报第30周.doc        2026-10-05   [⋯]          │
│  📊 销售数据-9月.xlsx     2026-10-04   [⋯]          │
│  📽 产品发布会-v2.slide   2026-10-03   [⋯]          │
│  ...                                                │
└──────────────────────────────────────────────────────┘
```

### 3.2 新建文档 / Create a Document

右上角"新建"按钮下拉三种：

The top-right "New" button dropdown offers three options:

| 类型 / Type | 路由 / Route | 编辑器 / Editor |
|---|---|---|
| 文档<br>Document | `/editor/doc/:id` | Tiptap 富文本<br>Tiptap rich text |
| 表格<br>Spreadsheet | `/editor/sheet/:id` | Luckysheet |
| 演示<br>Presentation | `/editor/slide/:id` | 自写 slide<br>Custom slide editor |

点完立即创建并跳进编辑器，默认标题"未命名"。

Clicking immediately creates the item and jumps into the editor, with a default title of "Untitled".

### 3.3 搜索 / Search

顶部搜索框输入关键词，回车。后端按**标题**模糊匹配（`title like %kw%`），不搜正文。清空搜索框恢复全部列表。

Type a keyword in the top search box and press Enter. The backend does a fuzzy match on the **title** (`title like %kw%`); it does not search the body. Clearing the search box restores the full list.

### 3.4 类型筛选 / Type Filter

"全部 / 文档 / 表格 / 演示"三个 tab，对应 `type=doc|sheet|slide`。

The three tabs "All / Document / Spreadsheet / Presentation" correspond to `type=doc|sheet|slide`.

### 3.5 排序 / Sorting

当前按 `updated_at` 倒序，最近改过的排最上面。

Currently sorted by `updated_at` descending — the most recently edited items appear at the top.

### 3.6 文档操作（每行右侧 ⋯ 按钮） / Document Actions (the ⋯ Button on Each Row)

- **重命名**：弹框输入新标题，调 `PUT /api/docs/:id`。
  - **Rename**: a dialog pops up to enter a new title, calling `PUT /api/docs/:id`.
- **删除**：确认后调 `DELETE /api/docs/:id`，不可恢复。
  - **Delete**: after confirmation calls `DELETE /api/docs/:id`; not recoverable.
- **打开**：点文档标题本身就进编辑器。
  - **Open**: clicking the document title itself enters the editor.

### 3.7 分页 / Pagination

列表接口带 `page` 和 `size` 参数，默认 `size=20`。文档多了底部会出现翻页按钮。

The list API takes `page` and `size` parameters, with a default `size=20`. When there are many documents, pagination buttons appear at the bottom.

---

## 4. 文字文档编辑器（Tiptap） / Document Editor

路由 `/editor/doc/:id`。

Route `/editor/doc/:id`.

### 4.1 工具栏按钮 / Toolbar Buttons

| 按钮 / Button | 作用 / Function |
|---|---|
| 段落 / H1 / H2 / H3 | 切换当前行的块级样式<br>Switch the current line's block style |
| **B** | 加粗<br>Bold |
| *I* | 斜体<br>Italic |
| U | 下划线<br>Underline |
| ~S~ | 删除线<br>Strikethrough |
| 彩色 A | 文字颜色<br>Text color |
| 引用 | 把当前行变成引用块<br>Turn the current line into a quote block |
| 代码块 | 把当前行变成等宽代码块<br>Turn the current line into a monospace code block |
| 无序列表 | `- 项`<br>`- item` |
| 有序列表 | `1. 项`<br>`1. item` |
| 插入表格 | 3×3 表格<br>3×3 table |
| 插入图片 | 粘贴图片 URL（当前不支持本地上传到正文）<br>Paste an image URL (local upload into the body not supported yet) |
| 撤销 / 重做 | 撤销 / 重做<br>Undo / Redo |

### 4.2 自动保存 / Auto-save

- 光标停 5 秒无操作，前端自动调 `PUT /api/docs/:id/content` 把整段 HTML 存到后端。
  - After the cursor is idle for 5 seconds, the frontend automatically calls `PUT /api/docs/:id/content` to save the whole HTML to the backend.
- 编辑器左下角会显示"已保存 HH:MM:SS"或"保存中…"。
  - The bottom-left of the editor shows "Saved HH:MM:SS" or "Saving…".
- 关掉页面再打开，正文还在。
  - Close and reopen the page — the content is still there.

### 4.3 手动保存 / Manual Save

工具栏右侧"保存"按钮，立即触发一次 `PUT`。

The "Save" button on the right of the toolbar triggers a `PUT` immediately.

### 4.4 导出 / Export

工具栏右侧"导出"下拉：

The "Export" dropdown on the right of the toolbar:

- **导出 PDF**：后端把 HTML 落临时文件，调 `soffice --headless --convert-to pdf`，流回浏览器下载。
  - **Export PDF**: the backend writes the HTML to a temp file, calls `soffice --headless --convert-to pdf`, and streams it back for download in the browser.
- **导出 DOCX**：同上，转成 docx。
  - **Export DOCX**: same as above, converted to docx.

导出失败会弹错"soffice not available"——说明服务器没装 LibreOffice，找运维。

If export fails with "soffice not available" — the server does not have LibreOffice installed; contact ops.

### 4.5 协作 / Collaboration

右上角"在线用户"头像组：

The top-right "Online users" avatar group:

- 有人打开这篇文档，头像组里多一个他的昵称。
  - When someone opens this document, their nickname is added to the avatar group.
- 他的光标位置会以彩色竖线出现在你这边的编辑器里。
  - Their cursor position appears in your editor as a colored vertical line.
- 他改了内容，**不会实时同步到你这边**——你这边看到的还是自己的版本，直到你刷新或重新进文档。
  - When they edit, the content is **not synced to you in real time** — you still see your own version until you refresh or re-enter the document.

> 重要：两个人同时编辑，谁后点保存谁覆盖对方。看到"别人正在编辑"的提示时，要么等他存完你再改，要么改完尽快保存。
>
> Important: when two people edit at the same time, whoever saves later overwrites the other. When you see the "someone else is editing" prompt, either wait for them to finish before you edit, or save quickly after editing.

---

## 5. 电子表格编辑器（Luckysheet） / Spreadsheet Editor

路由 `/editor/sheet/:id`。

Route `/editor/sheet/:id`.

### 5.1 界面 / Interface

标准表格布局：

A standard spreadsheet layout:

- 顶部菜单栏：文件、开始、插入、公式、数据、审阅。
  - Top menu bar: File, Home, Insert, Formula, Data, Review.
- 工具栏：字体、字号、加粗、斜体、边框、填充色、对齐。
  - Toolbar: font, font size, bold, italic, borders, fill color, alignment.
- 中间网格：行列单元格。
  - Center grid: row/column cells.
- 底部 sheet 标签：Sheet1 / Sheet2 …
  - Bottom sheet tabs: Sheet1 / Sheet2 …

### 5.2 基本操作 / Basic Operations

| 操作 / Operation | 怎么做 / How to do it |
|---|---|
| 输内容 | 点单元格，直接打字，回车<br>Enter content: click a cell, type, press Enter |
| 公式 | 在单元格输 `=SUM(A1:A10)`<br>Formula: enter `=SUM(A1:A10)` in a cell |
| 合并单元格 | 选中区域 → 工具栏"合并"<br>Merge cells: select a range → toolbar "Merge" |
| 插入行/列 | 右键行列号 → 插入<br>Insert row/column: right-click the row/column header → Insert |
| 删行/列 | 右键 → 删除<br>Delete row/column: right-click → Delete |
| 新增 sheet | 底部"+"号<br>New sheet: the "+" at the bottom |
| 重命名 sheet | 双击 sheet 标签<br>Rename sheet: double-click the sheet tab |
| 复制格式 | 选中 → 格式刷<br>Copy format: select → format painter |

### 5.3 保存 / Save

- 编辑后点右上角"保存"按钮，把整个 Luckysheet JSON 序列化后 `PUT /api/docs/:id/content`。
  - After editing, click the top-right "Save" button; the entire Luckysheet JSON is serialized and sent via `PUT /api/docs/:id/content`.
- 当前版本**没有自动保存**，记得手动存。
  - The current version has **no auto-save** — remember to save manually.
- 退出页面前如果没保存，会丢。
  - If you leave the page without saving, you lose the changes.

### 5.4 导出 XLSX / Export XLSX

右上角"导出"→"导出 xlsx"。后端调 soffice 转 xlsx，浏览器下载。

Top-right "Export" → "Export xlsx". The backend calls soffice to convert to xlsx and the browser downloads it.

### 5.5 协作 / Collaboration

和文字文档一样：右上角在线列表 + 光标广播。多人同时改一格，最后保存的人赢。

Same as the document editor: top-right online list + cursor broadcast. When multiple people edit the same cell, the last person to save wins.

---

## 6. 演示文稿编辑器（自写 slide） / Slides Editor

路由 `/editor/slide/:id`。

Route `/editor/slide/:id`.

### 6.1 布局 / Layout

三栏：

Three columns:

```
┌──────────┬──────────────────────────┬─────────┐
│ 缩略列表 │      中间编辑区            │ 属性面板│
│          │                           │         │
│ [Slide1] │   ┌─────────────────┐    │ 标题    │
│ [Slide2] │   │                 │    │ 正文    │
│ [Slide3] │   │   当前 slide     │    │ 背景色  │
│  [+]     │   │                 │    │         │
│          │   └─────────────────┘    │         │
└──────────┴──────────────────────────┴─────────┘
```

### 6.2 操作 / Operations

| 操作 / Operation | 怎么做 / How to do it |
|---|---|
| 新增 slide | 左侧列表底部"+"号<br>Add slide: the "+" at the bottom of the left list |
| 删 slide | 选中 → 缩略图上右键 → 删除<br>Delete slide: select → right-click the thumbnail → Delete |
| 改顺序 | 拖拽缩略图上下移动<br>Reorder: drag the thumbnail up/down |
| 编辑文字 | 中间编辑区直接点文字改<br>Edit text: click the text directly in the center editing area |
| 改背景色 | 右侧属性面板选颜色<br>Change background: pick a color in the right properties panel |
| 加备注 | 右侧属性面板"备注"栏<br>Add notes: the "Notes" field in the right properties panel |

### 6.3 保存 / Save

右上角"保存"按钮，把整个 slides 数组 JSON 存到 `content` 字段。

The top-right "Save" button stores the entire slides array JSON into the `content` field.

### 6.4 全屏播放 / Full-screen Playback

右上角"播放"按钮，进入全屏：

The top-right "Play" button enters full screen:

- 空格 / 右方向键：下一页。
  - Space / Right arrow: next slide.
- 左方向键：上一页。
  - Left arrow: previous slide.
- Esc：退出全屏。
  - Esc: exit full screen.

---

## 7. 协作 / Collaboration

### 7.1 邀请别人 / Inviting Others

当前版本没有"分享给指定用户"的按钮，邀请方式：

The current version has no "share with a specific user" button; the way to invite someone is:

1. 把浏览器地址栏那串 `/editor/doc/123` 复制给同事。
   - Copy the `/editor/doc/123` string from the browser address bar and send it to a colleague.
2. 同事登录后打开这个链接。
   - The colleague logs in and opens the link.
3. 只要他的账号是同一个服务器上注册的，就能看到这篇文档。
   - As long as their account is registered on the same server, they can see this document.

> 注意：当前版本后端列表接口只返回 `owner_id = 当前 uid` 的文档。也就是说同事即使打开了 `/editor/doc/123`，如果这篇文档不是他建的，`GET /api/docs/123` 会返回他无权访问。**真正的协作需要文档属于同一个 owner，或者后端加分享逻辑。**
>
> Note: in the current version, the backend list API only returns documents where `owner_id = current uid`. That means even if a colleague opens `/editor/doc/123`, if they did not create it, `GET /api/docs/123` returns "no access". **Real collaboration requires the document to belong to the same owner, or the backend to add sharing logic.**

> 实际用法：admin 建文档，把链接发给大家，大家登 admin 的账号一起改。或在后端加"协作者"字段（后续版本）。
>
> Practical approach: the admin creates the document and sends the link to everyone; they all log in with the admin account and edit together. Or add a "collaborator" field in the backend (a future version).

### 7.2 看在线用户 / See Online Users

- 编辑器右上角头像组：谁现在开着这篇文档，一目了然。
  - The editor's top-right avatar group shows at a glance who currently has this document open.
- 头像上悬停显示昵称。
  - Hover over an avatar to show the nickname.

### 7.3 看别人光标 / See Others' Cursors

- Tiptap 文档里，别人的光标是一根彩色竖线，颜色按用户名 hash。
  - In Tiptap documents, others' cursors are colored vertical lines, colored by a hash of the username.
- Luckysheet 和 slide 编辑器当前只广播"有人在线"，不广播光标。
  - The Luckysheet and slide editors currently only broadcast "someone is online", not cursors.

### 7.4 冲突怎么办 / Handling Conflicts

两个人同时改了同一段，后保存的覆盖先保存的。避免方式：

When two people edit the same section, the later save overwrites the earlier one. Ways to avoid this:

- 大块内容改动前，先看右上角有没有别人在线。
  - Before making large changes, check the top-right to see if anyone else is online.
- 真撞上了，让后保存的人重新改一遍（因为没有版本历史，找不回被覆盖的内容）。
  - If a collision happens, have the later-saver redo their edits (there is no version history, so the overwritten content cannot be recovered).

---

## 8. 权限模型 / Permission Model

### 8.1 普通 user / Normal user

- 只能看到自己建的文档（`GET /api/docs` 只返 `owner_id = 自己`）。
  - Can only see documents they created (`GET /api/docs` returns only `owner_id = self`).
- 只能看到自己上传的附件（`GET /api/fs/files` 只返 `uploader_id = 自己`）。
  - Can only see attachments they uploaded (`GET /api/fs/files` returns only `uploader_id = self`).
- 访问 `/admin` 会被路由守卫踢回 `/dashboard`。
  - Accessing `/admin` is bounced back to `/dashboard` by the route guard.
- 直接访问 `/api/admin/users` 返回 403。
  - Directly accessing `/api/admin/users` returns 403.

### 8.2 admin

- 文档/附件权限和普通 user 一样（当前版本 admin 也看不到别人的文档）。
  - Document/attachment permissions are the same as a normal user (in the current version admin also cannot see others' documents).
- 额外能访问 `/admin` 页面，看全部用户列表。
  - Additionally can access the `/admin` page and view the full user list.

### 8.3 修改别人的文档 / Editing Others' Documents

当前版本不支持。要改别人的文档，让 owner 把账号借给你，或让 admin 在数据库里把 `owner_id` 改成你。

Not supported in the current version. To edit someone else's document, have the owner lend you their account, or ask admin to change the `owner_id` to you in the database.

---

## 9. 文件上传（附件） / File Upload

### 9.1 在 Dashboard 上传 / Upload on the Dashboard

Dashboard 顶部"上传附件"按钮：

The top "Upload attachment" button on the Dashboard:

1. 点按钮，选本地文件。
   - Click the button and select a local file.
2. 前端以 multipart/form-data 调 `POST /api/fs/upload`。
   - The frontend calls `POST /api/fs/upload` with multipart/form-data.
3. 文件落到服务器 `storage/uploads/`。
   - The file lands in the server's `storage/uploads/`.
4. 列表里多出一条记录：原文件名、大小、上传时间。
   - A new record appears in the list: original filename, size, upload time.

### 9.2 下载 / Download

附件列表里点"下载"，后端调 `GET /api/fs/download/:id`，带 `Content-Disposition: attachment`，浏览器弹保存框。

Click "Download" in the attachment list; the backend calls `GET /api/fs/download/:id` with `Content-Disposition: attachment`, and the browser shows a save dialog.

### 9.3 删除附件 / Delete Attachments

当前版本前端没有删附件按钮，要删直接在服务器上：

The current version has no delete-attachment button in the frontend; to delete, go directly on the server:

```bash
rm /opt/picooffice/storage/uploads/<文件名>
sqlite3 /opt/picooffice/backend/picooffice.db "DELETE FROM files WHERE id=<id>;"
```

---

## 10. 管理员后台（/admin） / Admin Panel

只有 admin 能进。

Only admin can enter.

### 10.1 能看到什么 / What You Can See

全部用户列表：

The full user list:

| 字段 / Field | 说明 / Description |
|---|---|
| id | 用户 ID<br>User ID |
| username | 登录名<br>Login name |
| nickname | 昵称<br>Nickname |
| role | user / admin |
| created_at | 注册时间<br>Registration time |

不显示密码哈希（后端就不返这个字段）。

Password hashes are not shown (the backend simply does not return that field).

### 10.2 能做什么 / What You Can Do

当前版本**只能看**，不能在前端改用户角色、不能删用户、不能重置密码。要改只能直接动数据库：

The current version is **view-only** — you cannot change roles, delete users, or reset passwords from the frontend. To change anything you must edit the database directly:

```bash
# 把 alice 提成 admin
# Promote alice to admin
sqlite3 /opt/picooffice/backend/picooffice.db "UPDATE users SET role='admin' WHERE username='alice';"

# 删一个用户（小心，他的文档会变成孤儿）
# Delete a user (caution: their documents become orphans)
sqlite3 /opt/picooffice/backend/picooffice.db "DELETE FROM users WHERE username='bob';"
```

---

## 11. 移动端（Android） / Mobile (Android)

### 11.1 装 APK / Install the APK

1. 拿到 `app-debug.apk` 文件，发到手机（微信/QQ/USB 都行）。
   - Get the `app-debug.apk` file and send it to your phone (WeChat/QQ/USB all work).
2. 点开 apk，系统提示"未知来源"，允许。
   - Open the apk; the system prompts "unknown source", allow it.
3. 安装完成，桌面出现 PicoOffice 图标。
   - Once installed, the PicoOffice icon appears on the home screen.

### 11.2 首次启动 / First Launch

1. 点图标启动。
   - Tap the icon to launch.
2. 弹服务器地址框，填 `https://your-domain.com`。
   - A server address box pops up; enter `https://your-domain.com`.
3. 进入登录页，输账号密码。
   - Enter the login page and type your account and password.
4. 后续操作和浏览器版一致，Dashboard → 新建/打开文档。
   - Subsequent operations are the same as the browser version: Dashboard → create/open documents.

### 11.3 触控操作 / Touch Operations

- 文字文档：手指点光标位置，弹系统输入法。
  - Document: tap the cursor position and the system keyboard pops up.
- 表格：手指点单元格，双击进入编辑。
  - Spreadsheet: tap a cell, double-tap to edit.
- 演示：左右滑动翻页（全屏模式下）。
  - Presentation: swipe left/right to change slides (in full-screen mode).
- 协作在线列表在右上角，屏幕小的时候会挤，可横屏。
  - The collaboration online list is top-right; on a small screen it gets crowded — switch to landscape.

### 11.4 真机连不上 / Can't Connect on a Real Device

- 服务器是 `http://` 明文：Android 9+ 拦，找运维上 HTTPS，或让重新打包时开 `cleartext`。
  - Server is plain `http://`: Android 9+ blocks it; ask ops for HTTPS, or have `cleartext` enabled when repackaging.
- 服务器是内网 IP：手机要和服务器同一 WiFi。
  - Server is an internal IP: the phone must be on the same WiFi as the server.
- 服务器是公网域名：确认域名解析正确，防火墙放通 443。
  - Server is a public domain: confirm the DNS resolves correctly and the firewall allows 443.

---

## 12. 桌面端（Linux / Windows） / Desktop Client

### 12.1 Linux deb

```bash
sudo dpkg -i picooffice-desktop_1.0.0_amd64.deb
```

启动：应用菜单找 PicoOffice，或命令行 `picooffice-desktop`。

Launch: find PicoOffice in the application menu, or run `picooffice-desktop` from the command line.

首次启动弹服务器地址框，填 `https://your-domain.com`。

On first launch a server-address box pops up; enter `https://your-domain.com`.

### 12.2 Windows 绿色版 / Windows Portable

1. 解压 `PicoOffice-1.0.0-win.zip`。
   - Unzip `PicoOffice-1.0.0-win.zip`.
2. 进 `win-unpacked/`，双击 `PicoOffice.exe`。
   - Enter `win-unpacked/` and double-click `PicoOffice.exe`.
3. 首次启动弹服务器地址框，填 `https://your-domain.com`。
   - On first launch a server-address box pops up; enter `https://your-domain.com`.

不需要安装、不写注册表、不进系统目录。要卸载就删文件夹。

No installation, no registry writes, no system-directory changes. To uninstall, just delete the folder.

### 12.3 Windows nsis 安装包 / Windows nsis Installer

需要在 Windows/Mac 上自己打（沙箱打不出），打完双击 `PicoOffice Setup 1.0.0.exe`，可选安装目录，装完桌面有快捷方式。

You build it yourself on Windows/Mac (the sandbox can't produce it). After building, double-click `PicoOffice Setup 1.0.0.exe`, choose an install directory, and a shortcut appears on the desktop after installation.

### 12.4 桌面端 vs 浏览器端 / Desktop vs Browser

功能完全一致，只是套了个窗口。区别：

Functionally identical — it just wraps a window. The differences:

- 桌面端关掉再开，直接上次的服务器地址，不用重填。
  - The desktop client reopens with the last server address, no need to re-enter it.
- 桌面端不会被浏览器的标签页管理误关。
  - The desktop client won't be accidentally closed by the browser's tab management.
- 其他和浏览器版一模一样。
  - Everything else is exactly the same as the browser version.

---

## 13. FAQ / FAQ

### 13.1 忘记密码 / Forgot Password

当前版本**没有"忘记密码"按钮**。两条路：

The current version has **no "Forgot password" button**. Two routes:

**路 1：找 admin 直接改数据库。**

**Route 1: Have admin edit the database directly.**

```bash
# 生成一个新的 bcrypt 哈希（在任何有 go/python 的机器上）
# Generate a new bcrypt hash (on any machine with go/python)
python3 -c "import bcrypt; print(bcrypt.hashpw(b'newpassword123', bcrypt.gensalt()).decode())"

# 假设输出是 $2a$10$xxxx...
# Suppose the output is $2a$10$xxxx...
sqlite3 /opt/picooffice/backend/picooffice.db "UPDATE users SET password_hash='\$2a\$10\$xxxx...' WHERE username='alice';"
```

注意 sqlite3 shell 里 `$` 要转义成 `\$`，或把 SQL 写进文件再 `.read`。

Note: in the sqlite3 shell, `$` must be escaped as `\$`, or write the SQL into a file and `.read` it.

**路 2：自己注册一个新账号。**

**Route 2: Register a new account yourself.**

如果第一个 admin 就是你自己但你把密码忘了，没救，只能动数据库。

If the first admin is yourself but you forgot the password, there's no recovery — you must touch the database.

### 13.2 导出 PDF / docx 失败 / PDF / docx Export Fails

报错"soffice not available"：

Error "soffice not available":

- 服务器没装 LibreOffice。找运维 `sudo apt install libreoffice`。
  - LibreOffice is not installed on the server. Ask ops to `sudo apt install libreoffice`.
- 装完重启服务 `sudo systemctl restart picooffice`。
  - After installing, restart the service with `sudo systemctl restart picooffice`.

报错"convert failed"：

Error "convert failed":

- 文档内容太复杂，soffice 解析不了。试着精简文档再导。
  - The document is too complex for soffice to parse. Try simplifying it and exporting again.
- 看服务器日志 `journalctl -u picooffice -f`，找 soffice 的 stderr。
  - Check the server log `journalctl -u picooffice -f` for soffice's stderr.

### 13.3 协作文档冲突了怎么办 / Collaborative Document Conflicts

当前版本没有版本历史，被覆盖的内容找不回来。下次：

The current version has no version history, so overwritten content cannot be recovered. Next time:

- 大块修改前先看右上角有没有别人在线。
  - Before large edits, check the top-right to see if anyone else is online.
- 改完立即点"保存"。
  - Click "Save" immediately after editing.
- 重要文档改之前，先在本地另存一份备份。
  - Before editing an important document, keep a local backup copy.

### 13.4 上传大文件失败 / Large-file Upload Fails

- 报错 413：Nginx 限了大小。找运维把 `client_max_body_size` 调大。
  - Error 413: Nginx limits the size. Ask ops to increase `client_max_body_size`.
- 上传到一半断网：文件会留在服务器 `storage/uploads/` 里一条记录，重新传一次即可。
  - Network drops mid-upload: the file leaves a record in the server's `storage/uploads/`; just re-upload once.

### 13.5 文档列表里看不到自己刚建的文档 / Can't See a Document You Just Created

- 刷新一下页面。
  - Refresh the page.
- 确认没在"类型筛选"里选了别的 tab。
  - Make sure you didn't select another tab in "Type filter".
- 清搜索框。
  - Clear the search box.

### 13.6 浏览器 F12 控制台报 401 / Browser F12 Console Shows 401

token 过期了。右上角退出登录，重新登。

The token expired. Log out via the top-right menu and log in again.

### 13.7 WebSocket 连不上，看不到在线用户 / WebSocket Won't Connect, No Online Users

- 看浏览器 F12 → Network → WS，请求是不是 101。
  - In browser F12 → Network → WS, check whether the request returns 101.
- 不是 101 就是服务器 Nginx 没配 upgrade，找运维。
  - If it's not 101, the server's Nginx has no upgrade configured — ask ops.
- 是 101 但没消息进来，看 token 是不是过期，重新登录。
  - If it's 101 but no messages arrive, check whether the token expired and log in again.

### 13.8 手机上看不了表格 / Can't View Spreadsheet on Phone

- Luckysheet 在小屏上布局会挤，横屏试试。
  - Luckysheet gets cramped on small screens — try landscape.
- 真不行就用浏览器版。
  - If it really won't work, use the browser version.

### 13.9 怎么改自己的昵称 / How to Change Your Own Nickname

当前版本前端没有改昵称入口。找 admin 动数据库：

The current version has no frontend entry to change the nickname. Have admin edit the database:

```bash
sqlite3 /opt/picooffice/backend/picooffice.db "UPDATE users SET nickname='新昵称' WHERE username='alice';"
```

### 13.10 怎么彻底删除一篇文档 / How to Permanently Delete a Document

Dashboard 里文档右侧 ⋯ → 删除。确认后调 `DELETE /api/docs/:id`，正文从数据库里删掉，不可恢复。

In the Dashboard, the ⋯ on the right of a document → Delete. After confirmation, `DELETE /api/docs/:id` is called; the content is removed from the database and cannot be recovered.

### 13.11 导出的 PDF 排版乱 / Exported PDF Layout Is Messy

soffice 转 HTML→PDF 的排版和浏览器里看到的不完全一致。这是 LibreOffice 的已知行为。要求高的话，浏览器里直接 `Ctrl+P` → 另存为 PDF，效果更好。

The HTML→PDF layout from soffice is not exactly the same as what you see in the browser. This is known LibreOffice behavior. If precision matters, in the browser press `Ctrl+P` → Save as PDF for better results.

### 13.12 多人同时在线时，谁是"主" / Who Is the "Master" When Multiple People Are Online

没有主从概念，所有人平等。谁后保存谁赢。

There is no master/slave concept; everyone is equal. Whoever saves later wins.

---

## 14. 快捷键 / Keyboard Shortcuts

### 14.1 全局 / Global

| 键 / Key | 作用 / Function |
|---|---|
| `Ctrl+S` | 保存文档（编辑器内）<br>Save document (within editor) |
| `Ctrl+Z` / `Ctrl+Shift+Z` | 撤销 / 重做<br>Undo / Redo |

### 14.2 文字文档 / Document

| 键 / Key | 作用 / Function |
|---|---|
| `Ctrl+B` | 加粗<br>Bold |
| `Ctrl+I` | 斜体<br>Italic |
| `Ctrl+U` | 下划线<br>Underline |
| `Ctrl+P` | 浏览器打印（可另存 PDF）<br>Browser print (can save as PDF) |

### 14.3 演示播放 / Presentation Playback

| 键 / Key | 作用 / Function |
|---|---|
| 空格 / → | 下一页<br>Next slide |
| ← | 上一页<br>Previous slide |
| Esc | 退出全屏<br>Exit full screen |

---

## 15. 支持与反馈 / Support & Feedback

- 找到 bug：记下来复现步骤（什么浏览器、什么账号、点了什么按钮、报什么错），发给运维。
  - Found a bug: write down the reproduction steps (which browser, which account, what button was clicked, what error appeared) and send it to ops.
- 要新功能：列需求给 admin，admin 动后端加。
  - Want a feature: list the requirement for admin, and admin adds it on the backend.
- 数据丢失：先停服务 `sudo systemctl stop picooffice`，找运维从 `/backup/picooffice/` 恢复。
  - Data loss: first stop the service with `sudo systemctl stop picooffice`, then ask ops to restore from `/backup/picooffice/`.

---

## 16. 离线使用（PWA） / Offline Mode (PWA)

PicoOffice 做了 PWA，手机和电脑 Chrome/Edge 上都能"装"到桌面，断网也能打开看过的文档。

PicoOffice is a PWA; on phone and desktop Chrome/Edge it can be "installed" to the home screen, and you can open previously viewed documents even offline.

### 16.1 怎么添加到主屏幕 / How to Add to Home Screen

- **Android Chrome：** 打开 `https://office.yourdomain.com` → 右上角 ⋮ → "添加到主屏幕" / "安装应用"。桌面会出现一个蓝色圆角方块白字 P 的图标，点开就是全屏 PicoOffice，没有浏览器地址栏。
  - **Android Chrome:** open `https://office.yourdomain.com` → top-right ⋮ → "Add to Home screen" / "Install app". A blue rounded square with a white "P" appears on the home screen; tapping it opens full-screen PicoOffice with no browser address bar.
- **iPhone Safari：** 地址栏中间那个分享箭头 → "添加到主屏幕"。注意：iOS 上 PWA 必须用 Safari 装，Chrome for iOS 不支持。
  - **iPhone Safari:** the share arrow in the middle of the address bar → "Add to Home Screen". Note: on iOS a PWA must be installed via Safari; Chrome for iOS does not support it.
- **桌面 Chrome/Edge：** 地址栏右边会出现一个"安装"小图标（显示器带向下箭头），点一下就装成桌面应用。
  - **Desktop Chrome/Edge:** an "Install" icon (a monitor with a down arrow) appears to the right of the address bar; click it to install as a desktop app.

### 16.2 离线能做什么 / What You Can Do Offline

- 打开**之前联网时看过的**文档列表和正文（这些被 service worker 缓存了）。
  - Open document lists and bodies **viewed while online before** (these are cached by the service worker).
- 新建文档、编辑现有文档——编辑内容先存在浏览器 IndexedDB 里。
  - Create documents and edit existing ones — edits are first stored in the browser's IndexedDB.
- PDF 批注——高亮和批注完全存在本地。
  - PDF annotations — highlights and notes are stored entirely locally.
- 不能做：注册新账号、看同事刚新建的文档（没缓存过）、上传新附件。
  - Cannot do: register a new account, view a colleague's just-created document (not cached), or upload new attachments.

顶栏右侧有个状态 Tag：

There is a status Tag on the right of the top bar:

- **在线**（绿）：一切正常。
  - **Online** (green): everything is normal.
- **离线**（灰）：断网了，编辑内容会攒着。
  - **Offline** (gray): disconnected; edits accumulate.
- **待同步 N 条**（橙）：离线期间改了 N 次，联网后会自动重放。
  - **N pending sync** (orange): edited N times while offline; they replay automatically once back online.

### 16.3 联网后怎么同步 / How Sync Works After Reconnecting

不用手动操作。网络一恢复，前端自动按顺序把离线期间攒的修改发回服务器，每条成功就从队列里删掉，"待同步 N 条"数字往下减，归零变绿。

No manual action needed. As soon as the network returns, the frontend automatically sends the accumulated offline edits back to the server in order; each successful one is removed from the queue, the "N pending sync" count goes down, and it turns green at zero.

**注意：** 如果你在 A 电脑离线改了文档，又在 B 电脑离线改了同一篇，联网后谁的请求后到谁覆盖先到，先到的修改会丢。离线编辑尽量在一台设备上做完。

**Note:** if you edit a document offline on computer A and also offline on computer B, after reconnecting whoever's request arrives later overwrites the earlier one, and the earlier edit is lost. Do offline editing on a single device whenever possible.

---

## 17. 配额 / Quota

### 17.1 在哪看 / Where to View

Dashboard 右上角有一条蓝色进度条，鼠标悬停显示 `已用 xxx MB / 总 1 GB`。所有上传的文件（附件、图片）都算在这个配额里，文档正文（存在数据库）不算。

There is a blue progress bar at the top-right of the Dashboard; hovering shows `used xxx MB / total 1 GB`. All uploaded files (attachments, images) count toward this quota, but document bodies (stored in the database) do not.

### 17.2 超了怎么办 / What If You Exceed It

上传文件时如果超过配额，会直接报错："配额不足，请联系管理员扩容"。这时候：

If an upload exceeds the quota, it errors directly: "Insufficient quota, please contact the admin to expand." In that case:

1. 去"我的文件"（Dashboard → 文件 tab）删掉不要的旧附件、图片、视频，`used_bytes` 会自动降下来。
   - Go to "My Files" (Dashboard → Files tab) and delete unwanted old attachments, images, and videos; `used_bytes` drops automatically.
2. 删完还不够，找 admin，在 Admin → 用户列表里你那行点"改配额"，让 admin 给你调到 5GB / 10GB。
   - If that's not enough, ask admin to open Admin → User list, click "Change quota" on your row, and raise it to 5GB / 10GB.

admin 自己默认 10GB，一般够用。

Admin's own default is 10GB, which is generally enough.

### 17.3 单文件大小 / Single-file Size

单个文件不能超过 100MB。超过会被后端拒。视频大的话压一压再传。允许的后缀：`.pdf .doc .docx .xls .xlsx .ppt .pptx .txt .md .png .jpg .jpeg .gif .webp .csv .zip .mp4`，其他后缀传不上去。

A single file cannot exceed 100MB; larger ones are rejected by the backend. Compress large videos before uploading. Allowed extensions: `.pdf .doc .docx .xls .xlsx .ppt .pptx .txt .md .png .jpg .jpeg .gif .webp .csv .zip .mp4`; other extensions cannot be uploaded.

---

## 18. 版本历史 / Version History

每保存一次文档（Ctrl+S 或自动保存），旧版本会自动留一个快照，最多留 20 版。

Each time a document is saved (Ctrl+S or auto-save), the old version is automatically kept as a snapshot, up to 20 versions.

### 18.1 怎么看历史版本 / How to View History

打开任意文档 → 右上角 🕐 图标（或菜单 → "版本历史"），右侧滑出一个抽屉，按时间倒序列出最近 20 版，每版显示时间和保存人。点某一版，正文区切到那一版的内容（只读）。

Open any document → the 🕐 icon at top-right (or menu → "Version history"); a drawer slides out from the right listing the most recent 20 versions in reverse chronological order, each showing time and the saver. Click a version and the body switches to that version's content (read-only).

### 18.2 怎么回滚 / How to Roll Back

在版本历史抽屉里选中想恢复的那版 → 点顶部"回滚到此版本"。当前文档内容会被这一版覆盖，同时这一版回滚的动作本身又会留一个新快照（不会丢历史）。

In the version-history drawer, select the version to restore → click "Roll back to this version" at the top. The current document content is overwritten by that version, and the rollback action itself leaves a new snapshot (history is not lost).

回滚前最好看一眼时间，别回滚到几小时前把刚写的东西盖了。

Before rolling back, glance at the time — don't roll back to a few hours ago and overwrite what you just wrote.

---

## 19. PDF 批注 / PDF Annotation

在 Dashboard 点开一个 PDF 附件，会进到 PDF 批注查看器 `/viewer/pdf/:id`。

Open a PDF attachment in the Dashboard and you enter the PDF annotation viewer at `/viewer/pdf/:id`.

### 19.1 怎么高亮 / How to Highlight

- 鼠标拖选一段区域 → 自动弹出"高亮"按钮 → 点一下，那块变成黄色。
  - Drag to select a region with the mouse → a "Highlight" button pops up automatically → click it and the area turns yellow.
- 想去掉高亮：右键那块黄色 → 删除。
  - To remove the highlight: right-click the yellow area → Delete.

### 19.2 怎么加文本批注 / How to Add a Text Annotation

- 在 PDF 上**双击**某个位置 → 弹出一个输入框 → 写批注内容 → 确认。该位置会出现一个小图标，鼠标悬停能看到批注全文。
  - **Double-click** a spot on the PDF → an input box pops up → write the annotation → confirm. A small icon appears at that spot; hover to see the full annotation.

### 19.3 重要提醒 / Important Reminder

批注和高亮**只存在你自己的浏览器里**（IndexedDB），不上传服务器。换台电脑、换个浏览器、清缓存，批注就没了。原 PDF 文件本身不会被改动。要真正把批注写进 PDF 发给别人，用 Adobe Reader 或 Foxit 在本地做。

Annotations and highlights exist **only in your own browser** (IndexedDB) and are not uploaded to the server. Switch computers, switch browsers, or clear the cache and the annotations are gone. The original PDF itself is not modified. To truly write annotations into a PDF to send to others, use Adobe Reader or Foxit locally.

---

## 20. 多员工协作流程 / Team Collaboration Workflow

### 20.1 admin 这边 / On the admin side

1. 部署好 PicoOffice 后，第一个注册的账号自动是 admin。
   - After deploying PicoOffice, the first registered account is automatically admin.
2. admin 登录 → Admin → 用户管理 → "新建用户"，填员工用户名和初始密码。
   - Admin logs in → Admin → User management → "New user", fill in the employee's username and initial password.
3. 把访问地址 `https://office.yourdomain.com` 发给员工。
   - Send the access address `https://office.yourdomain.com` to the employees.

### 20.2 员工这边 / On the employee side

1. 打开链接，用 admin 给的用户名密码登录。
   - Open the link and log in with the username/password given by admin.
2. 右上角头像 → 修改密码，把初始密码改成自己的。
   - Top-right avatar → Change password, set your own instead of the initial one.
3. 回到 Dashboard，就能看到 admin 分享出来的文档（或者自己新建）。
   - Back on the Dashboard, you can see the documents admin shared (or create your own).
4. 多人同时打开同一篇文档，右上角会显示在线用户头像列表。谁在编辑，他输入的内容会实时推给所有人（WebSocket）。
   - When multiple people open the same document, the top-right shows the online-user avatar list. Whoever is editing — their input is pushed live to everyone (WebSocket).

### 20.3 冲突怎么算 / How Conflicts Are Resolved

两个人同时改同一段，谁后按保存谁赢，先保存的修改会被覆盖。没有智能合并。所以多人协作时养成习惯：**改之前看一眼右上角在线人数，超过一个人就打个招呼**。

When two people edit the same section, whoever saves later wins and the earlier edit is overwritten. There is no smart merge. So when collaborating, build the habit: **before editing, glance at the online count top-right, and say hi if more than one person is there**.

---

## 21. 查找替换 / Find & Replace

文字文档里：

In a document:

- `Ctrl+F`：调出查找框，输入关键词，文档里所有匹配处黄色高亮。
  - `Ctrl+F`: brings up the find box; type a keyword and all matches in the document are highlighted yellow.
- `Ctrl+H`：调出替换框，输入"查找内容"和"替换为"，可以"替换一处"或"全部替换"。
  - `Ctrl+H`: brings up the replace box; enter "find" and "replace with", then "replace one" or "replace all".
- `Esc` 关闭查找替换框。
  - `Esc` closes the find/replace box.

表格和演示文档里暂不支持查找替换，自己 `Ctrl+F` 在浏览器里搜。

Find & replace is not yet supported in spreadsheets and presentations; use `Ctrl+F` in the browser to search there.

---

## 22. 常见操作速查 / Common Actions Cheat Sheet

| 想干嘛 / Want to | 去哪 / Where to go |
|---|---|
| 新建文档 / 表格 / 演示 | Dashboard 顶部 "+" 按钮<br>Dashboard top "+" button |
| 改自己密码 | 右上角头像 → 修改密码<br>Top-right avatar → Change password |
| 看自己上传了多少 | Dashboard 右上角进度条<br>Dashboard top-right progress bar |
| 删自己上传的文件 | Dashboard → 文件 tab → 每行删除按钮<br>Dashboard → Files tab → per-row delete button |
| 看文档历史版本 | 打开文档 → 右上角 🕐<br>Open document → top-right 🕐 |
| 回滚到旧版本 | 版本历史抽屉 → 选中那一版 → 回滚<br>Version-history drawer → select version → roll back |
| PDF 高亮批注 | 点开 PDF → 拖选高亮 / 双击加批注<br>Open PDF → drag to highlight / double-click to annotate |
| 手机装成 App | Chrome/Safari → 添加到主屏幕<br>Chrome/Safari → Add to Home screen |
| 看离线攒了多少修改 | 顶栏"待同步 N 条"橙色 Tag<br>Top bar "N pending sync" orange Tag |
| 找 admin 加配额 | Admin → 用户列表 → 改配额<br>Admin → User list → Change quota |
| 全局搜索文档 | Dashboard 顶部搜索框（按标题模糊搜，不搜正文）<br>Dashboard top search box (fuzzy title search, not body) |

如果以上都找不到对应入口，直接问 admin。PicoOffice 目前是内部工具形态，没有用户手册站点，所有用法都在这一篇里。

If none of the above has the right entry, just ask admin. PicoOffice is currently an internal-tool form with no user-manual site; all usage is in this one document.

---

## 23. 反馈 / Feedback

用着别扭、缺功能、出 bug，直接告诉 admin：

If something is awkward, missing a feature, or buggy, tell admin directly:

- bug 描述里带上：浏览器版本、登录账号、点了什么按钮、界面上显示什么报错。
  - In the bug description include: browser version, login account, what button was clicked, and what error shows on the interface.
- 功能建议列清楚"现在怎么做 → 希望怎么做"，admin 评估后安排进下一轮迭代。
  - For feature suggestions, clearly state "how it works now → how you'd like it"; admin evaluates and schedules it for the next iteration.

---

# 第三轮新增用法 / Features Added in Round 3

下面这些是第三轮加上的功能，原来 1~23 节没覆盖。按场景挑着看。

The following features were added in round 3 and were not covered in sections 1–23. Pick by scenario.

---

## 24. 导入 Office 文件 / Import Office Files

以前只能手动新建文档/表格/演示。现在能直接把电脑上的 Word/Excel/PPT 传上来：

Before, you could only manually create documents/spreadsheets/presentations. Now you can directly upload Word/Excel/PPT from your computer:

1. Dashboard 顶部 "+" → 选"导入文件"。
   - Dashboard top "+" → choose "Import file".
2. 选一个 `.docx` / `.xlsx` / `.pptx`（老版本 `.doc` / `.xls` / `.ppt` 也行，会自动转）。
   - Select a `.docx` / `.xlsx` / `.pptx` (old `.doc` / `.xls` / `.ppt` also work and are converted automatically).
3. 等几秒，服务端会用 LibreOffice 把它转成网页格式，自动建好一篇对应类型的文档，直接跳进编辑器。
   - Wait a few seconds; the server uses LibreOffice to convert it to a web format, automatically creates a document of the matching type, and jumps straight into the editor.
4. 导入完跟你新建的文档一样，能继续改、能分享、能导出。
   - Once imported it's just like a freshly created document — you can keep editing, share it, and export it.

注意：

Note:

- 复杂排版（文本框嵌套、宏、图表）转过来可能丢东西，导入后扫一眼重要内容。
  - Complex layouts (nested text boxes, macros, charts) may lose something in conversion; glance over important content after importing.
- 有密码的 office 文件转不了，先解密再传。
  - Password-protected Office files can't be converted — decrypt them first, then upload.

## 25. 导出 PPTX / Export to PPTX

以前演示文档只能导出 PDF。现在：打开演示文档 → 右上角"导出" → 选 PPTX，等队列转好就能下。下载下来的 pptx 能直接用 PowerPoint / WPS 打开接着改。导出走后台队列，文件多了会排队，右上角有进度提示。

Before, presentations could only be exported to PDF. Now: open a presentation → top-right "Export" → choose PPTX, wait for the queue to finish, then download. The downloaded pptx opens directly in PowerPoint / WPS for further editing. Export runs through a background queue; many files queue up and the top-right shows progress.

## 26. 分享链接 / Share Links

### 26.1 生成分享链接 / Generate a Share Link

打开一篇文档 → 右上角"分享"：

Open a document → top-right "Share":

- **私有（默认）**：只有自己和 admin 能看。
  - **Private (default)**: only you and admin can view.
- **共享（shared）**：开了之后下面出一个链接，形如 `https://你的域名/api/share/xxxxxx`。
  - **Shared**: once enabled, a link appears below, of the form `https://your-domain/api/share/xxxxxx`.
- 权限选 **view**（对方只能看，不能改）或 **edit**（理论上对方能编辑）。
  - Choose permission **view** (the other party can only see, not edit) or **edit** (in theory the other party can edit).

把链接复制发给别人就行。

Just copy the link and send it to others.

### 26.2 匿名查看 / Anonymous Viewing

没登录的人打开 view 链接，能直接读文档内容，不需要注册账号。这就是把文档发给公司外的客户/合作方看的方式。

A person who isn't logged in can open a view link and read the document directly, no registration needed. This is how you send a document to external clients/partners.

注意：edit 权限的链接，**匿名（没登录）打开目前仍是只读的**，要真编辑必须登录后访问。这个限制见文末"已知没做的"。

Note: an edit-permission link, **when opened anonymously (not logged in), is still read-only for now** — to actually edit you must log in first. This limitation is listed under "Known Not Done" at the end.

### 26.3 取消分享 / Cancel Sharing

回到分享弹窗，把共享关掉（visibility 改回 private），链接立刻失效，再打开报 404。"分享"页（Dashboard → 分享 tab）能看到你分享出去的所有链接，随时收。

Go back to the share dialog and turn off sharing (set visibility back to private); the link becomes invalid immediately and reopening it returns 404. The "Share" page (Dashboard → Share tab) shows all links you've shared, and you can retract them anytime.

## 27. 回收站 / Trash

### 27.1 还原 / Restore

删了的文档不会立刻没，进回收站：

Deleted documents aren't gone immediately — they go to the trash:

1. Dashboard → 左侧"回收站"。
   - Dashboard → "Trash" on the left.
2. 找到误删的那篇 → 点"还原"。
   - Find the mistakenly deleted one → click "Restore".
3. 文档回到原来的位置，历史版本也一起回来。
   - The document returns to its original location, and its version history comes back too.

### 27.2 彻底删除 / Permanent Deletion

回收站里的文档会保留 **30 天**，超过自动清掉。想立刻清：回收站里点"彻底删除"，不可恢复，连带历史快照一起删。彻底删之前再确认一眼。

Documents in the trash are kept for **30 days**, then auto-purged. To clear immediately: click "Delete permanently" in the trash — unrecoverable, and the version snapshots are deleted too. Double-check before permanently deleting.

## 28. 文件夹、标签、星标、批量 / Folders, Tags, Stars, Batch

文档多了以后用这些整理：

Use these to organize once you have many documents:

- **文件夹**：Dashboard 左侧"文件夹"→ 新建文件夹，能建子文件夹。拖拽文档进文件夹，或右键"移动到"。
  - **Folders**: Dashboard left "Folders" → create a folder, subfolders supported. Drag documents into a folder, or right-click "Move to".
- **标签**：给文档打标签（比如"合同""产品""2026"），标签像滤镜一样，点一下标签列出所有打过这个标的文档。
  - **Tags**: tag documents (e.g. "contract", "product", "2026"); tags act like filters — click one to list all documents with that tag.
- **星标**：文档右边点 ☆ 变 ★，Dashboard 顶部"星标"筛出你常用的那几篇。
  - **Stars**: click ☆ on the right of a document to make it ★; Dashboard top "Starred" filters to your frequently used ones.
- **批量**：Dashboard 文档列表左上角勾选多篇，底部浮出操作条，能批量移动 / 删 / 打标签 / 星标。
  - **Batch**: check multiple documents at the top-left of the Dashboard list; an action bar floats at the bottom to batch move / delete / tag / star.
- **最近**：Dashboard 默认按最近打开排序，不用自己记哪个是刚改的。
  - **Recent**: the Dashboard sorts by most recently opened by default, so you don't have to remember which was just edited.

## 29. 两步验证（2FA） / Two-Factor Authentication

建议每个人都开，特别是 admin。

Recommended for everyone, especially admin.

### 29.1 开通 / Enable

1. 右上角头像 → 个人资料 → "两步验证" → 开通。
   - Top-right avatar → Profile → "Two-factor auth" → Enable.
2. 页面弹出一个二维码，下面还有一串手写密钥。
   - The page shows a QR code, with a handwritten secret key below it.
3. 用手机上的认证器 App（Google Authenticator、微软 Authenticator、豆包认证器都行）扫二维码；扫不了就手动输入那串密钥。
   - Use an authenticator app on your phone (Google Authenticator, Microsoft Authenticator, Doubao Authenticator all work) to scan the QR code; if you can't scan, manually enter that secret key.
4. App 里会出一个每 30 秒变一次的 6 位数字，把当前数字填回网页 → 确认开通。
   - The app shows a 6-digit number that changes every 30 seconds; enter the current number back into the page → confirm to enable.
5. **备份好恢复密钥**：页面会再给你一串备用码，存到密码管理器里。手机丢了靠它找回。
   - **Back up the recovery key**: the page gives you a set of backup codes — store them in a password manager. If you lose your phone, these get you back in.

### 29.2 之后登录 / Subsequent Logins

登录要两步：先输密码，再输认证器里的 6 位数字。

Login takes two steps: first enter the password, then the 6-digit number from the authenticator.

### 29.3 手机丢了 / 密钥没了 / Lost Phone / Lost Key

找 admin，在后台用户列表里把你的 2FA 重置掉，你下次登录重新绑一个。

Ask admin to reset your 2FA in the backend user list; you re-bind one on your next login.

## 30. 登录设备管理 / Logged-in Devices

头像 → "登录设备"，能看到你现在在哪些设备上登着：每台的浏览器、大概 IP、上次活跃时间。

Avatar → "Logged-in devices" shows which devices you're currently logged in on: each one's browser, approximate IP, and last active time.

- 看到不认识的设备（比如上次在网吧登了没退）→ 点"吊销"，那台设备立刻被踢下线。
  - See an unrecognized device (e.g. you logged in at an internet café and didn't log out last time) → click "Revoke" and that device is immediately kicked off.
- 自己手机换了 / 电脑卖了，记得来这清掉旧设备。
  - If you switched phones / sold a computer, remember to clear the old device here.

## 31. 自动续登（refresh）和忘记密码 / Refresh Sessions & Forgot Password

### 31.1 免重复登录 / No Repeated Login

登录一次后，access 凭证 2 小时过期，但系统会在后台自动用 refresh 换新的，正常用不用反复登录。一周不操作才会要求重新登。

After one login, the access credential expires in 2 hours, but the system automatically exchanges it for a new one via refresh in the background, so normal use doesn't require repeated login. You're only asked to log in again after a week of inactivity.

### 31.2 忘记密码 / Forgot Password

登录页点"忘记密码" → 输注册邮箱 → 系统发重置链接。目前邮件还没接自动发信，**重置链接会打在后端日志里，找 admin 要**（admin 看 `docker compose logs` 或 systemd 日志）。拿到链接进去设新密码。

On the login page click "Forgot password" → enter the registered email → the system sends a reset link. Email auto-sending isn't wired up yet, so **the reset link is printed in the backend log — ask admin for it** (admin checks `docker compose logs` or the systemd log). Use the link to set a new password.

## 32. 邮件 / Mail

左侧导航多了"邮件"：

A "Mail" item is added to the left navigation:

1. 第一次用 → 添加账号 → 填邮箱地址、IMAP/SMTP 参数（问 admin 或看 DEPLOY 14.4 那张表）。**注意填的是邮箱授权码，不是登录密码**。
   - First use → Add account → fill in the email address, IMAP/SMTP parameters (ask admin or see the DEPLOY 14.4 table). **Note: enter the email authorization code, not the login password.**
2. 加完自动拉收件箱，能看历史邮件。
   - After adding, it auto-pulls the inbox and you can view historical mail.
3. 点"写信" → 填收件人/主题/正文 → 发送。
   - Click "Compose" → fill in recipient/subject/body → send.
4. 支持加多个邮件账号，左上角切换。
   - Multiple mail accounts are supported; switch via the top-left.

目前只是基础收发，没有附件预览、没有已发送文件夹同步、没有草稿。复杂需求先用电脑邮箱客户端。

Currently it's basic send/receive only — no attachment preview, no sent-folder sync, no drafts. For complex needs, use a desktop mail client for now.

## 33. 日历 / Calendar

左侧"日历"：

The left "Calendar":

- 看到月视图，点某天加事件：标题、开始/结束时间、提醒时间。
  - View the month; click a day to add an event: title, start/end time, reminder time.
- 到了提醒时间，下次登录 PicoOffice 会弹一下。注意：App 关着不会主动推送，得你打开网页才看得到。
  - At reminder time, PicoOffice pops a notice next time you log in. Note: with the app closed it won't push proactively — you must open the page to see it.
- 点事件能改时间或删。
  - Click an event to change its time or delete it.

## 34. 全文搜索 / Full-text Search

Dashboard 顶部搜索框现在搜的是**全文**，不只是标题：

The Dashboard top search box now searches the **full text**, not just the title:

- 输入关键词，所有正文里提到这个词的文档都会列出来，按相关度排。
  - Type a keyword and every document mentioning it in the body is listed, ranked by relevance.
- 中文搜两个字以上效果好，搜一个字不准。
  - For Chinese, two or more characters search well; a single character is imprecise.
- 搜不到刚改的东西？等一两秒，索引是异步更新的。
  - Can't find something you just changed? Wait a second or two — the index updates asynchronously.
- 索引坏了 / 换了分词，找 admin 在搜索框加 `?rebuild=1` 全量重建一次。
  - Index broken / tokenizer changed? Ask admin to append `?rebuild=1` in the search box to do a full rebuild.

## 35. 真·多人协作（文字文档） / Real-time Collab (Documents)

以前两人同时改一篇，后保存的把先保存的覆盖了。现在文字文档是真协作：

Before, when two people edited one document, the later save overwrote the earlier. Now documents have true collaboration:

- 两个人同时打开同一篇文档，右上角能看到对方头像。
  - Two people open the same document at once; you can see each other's avatar top-right.
- A 打字，B 那边光标旁边实时看到字一个一个冒出来，不会互相覆盖。
  - When A types, B sees the characters appear live next to the cursor — no mutual overwriting.
- 断网再连上，会自动把你离线打的内容合并进去，不用手动冲突解决。
  - Reconnect after a dropout and your offline typing is auto-merged — no manual conflict resolution.
- **注意**：这个真协作只对**文字文档**有效。表格和演示文档还是老规矩——同时改会冲突，谁后保存谁赢，改之前打个招呼。
  - **Note**: this true collaboration applies only to **documents**. Spreadsheets and presentations follow the old rule — simultaneous edits conflict, later save wins, so say hi before editing.

## 36. 手机触摸优化版 / Touch-optimized Mobile

手机浏览器直接打开网址就是"触摸优化版"：

Opening the URL in a phone browser gives you the "touch-optimized version":

- 工具栏按钮变大，手指点不疼。
  - Toolbar buttons are larger, easier on the fingers.
- 双指捏合缩放页面。
  - Pinch to zoom the page.
- 调起手机软键盘时编辑器自动顶上去，不被挡住。
  - When the phone keyboard pops up, the editor auto-scrolls up so it isn't covered.
- 想当 App 用：Safari/Chrome → 添加到主屏幕（见 16.1）。
  - To use as an app: Safari/Chrome → Add to Home screen (see 16.1).

## 37. 中英文切换 / Language Switch

右上角头像 → 语言 → 中文 / English。整个界面立刻切换，不用刷新。新账号默认跟随浏览器语言。

Top-right avatar → Language → 中文 / English. The whole interface switches instantly, no refresh needed. New accounts default to following the browser language.

## 38. 第三轮已知没做的（别等） / Known Not Done in v0.3

几个你可能以为有、其实还没做的：

A few things you might assume exist but don't yet:

1. 分享链接 edit 权限下，没登录的人**不能**真编辑，只能看。要协作编辑必须登录。
   - Under an edit-permission share link, a logged-out person **cannot** actually edit, only view. Collaborative editing requires login.
2. 表格和演示文档没有实时协作，同时改会互相覆盖。
   - Spreadsheets and presentations have no real-time collaboration; simultaneous edits overwrite each other.
3. 邮件没有附件收发、没有草稿、没有已发送同步。
   - Mail has no attachment send/receive, no drafts, no sent-folder sync.
4. 日历不主动推送，关着网页不提醒。
   - Calendar doesn't push proactively; with the page closed it gives no reminder.
5. 忘记密码的重置链接目前找 admin 要，不自动发邮件。
   - The forgot-password reset link currently comes from admin, not auto-emailed.
6. 搜索是简单中文分词，不是百度那种语义搜索，长句搜不准。
   - Search is simple Chinese tokenization, not Baidu-style semantic search; long sentences are imprecise.

有强烈需求就找 admin 排下一轮。

If you have a strong need, ask admin to schedule it for the next round.

---

## 39. 打印 / Printing

文字文档右上角"打印"按钮：浏览器弹打印对话框，选打印机或另存 PDF。走的是浏览器打印样式（@media print），屏幕上什么样打出来基本什么样。注意：

The "Print" button at the top-right of a document opens the browser print dialog to choose a printer or save as PDF. It uses the browser print style (@media print), so what you see on screen is basically what prints. Note:

- 演示文档打印会按每页一张幻灯片出。
  - Presentations print one slide per page.
- 表格打印建议先在表格里调好打印区域和纸张方向，不然列太多打不全。
  - For spreadsheets, first set the print area and paper orientation in the sheet, or too many columns won't fit.
- 想拿到排版精美的 PDF 还是用右上角"导出 PDF"，比浏览器打印可控。
  - For a nicely laid-out PDF, use the top-right "Export PDF" instead — it's more controllable than browser printing.

---

## 40. 第三轮新操作速查 / v0.3 Quick Reference

| 想干嘛 / Want to | 去哪 / Where to go |
|---|---|
| 上传 Word/Excel/PPT | Dashboard "+" → 导入文件<br>Dashboard "+" → Import file |
| 演示导出 pptx | 演示文档 → 导出 → PPTX<br>Presentation → Export → PPTX |
| 发只读链接给外人 | 文档 → 分享 → 开共享 → 复制链接<br>Document → Share → enable shared → copy link |
| 收回来链接 | 分享弹窗关掉共享，或 Dashboard → 分享 tab<br>Share dialog turn off shared, or Dashboard → Share tab |
| 找回删错的文档 | Dashboard → 回收站 → 还原<br>Dashboard → Trash → Restore |
| 彻底删掉不要的 | 回收站 → 彻底删除<br>Trash → Delete permanently |
| 建文件夹整理 | Dashboard 左侧 → 文件夹<br>Dashboard left → Folders |
| 给文档打标签/星标 | 文档列表行尾按钮<br>Document list end-of-row buttons |
| 批量移动/删除 | Dashboard 勾选多篇 → 底部操作条<br>Dashboard check multiple → bottom action bar |
| 开 2FA | 头像 → 个人资料 → 两步验证<br>Avatar → Profile → Two-factor auth |
| 踢掉别的设备登录 | 头像 → 登录设备 → 吊销<br>Avatar → Logged-in devices → Revoke |
| 改密码忘了 | 登录页 → 忘记密码 → 找 admin 要链接<br>Login page → Forgot password → ask admin for link |
| 收发邮件 | 左侧"邮件" → 添加账号<br>Left "Mail" → Add account |
| 加日程提醒 | 左侧"日历" → 点某天<br>Left "Calendar" → click a day |
| 搜文档正文 | Dashboard 顶部搜索框<br>Dashboard top search box |
| 切中英文 | 头像 → 语言<br>Avatar → Language |
| 手机上用 | 手机浏览器直接开网址<br>Open the URL directly in phone browser |
