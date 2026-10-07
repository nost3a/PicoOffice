# 致谢 / 第三方声明（THANKS） / Acknowledgements & Third-party Notices

PicoOffice 站在一堆优秀开源项目的肩膀上。下面列出主要第三方依赖及其许可证，按此分发。我们对这些项目的维护者表示感谢。

PicoOffice stands on the shoulders of many excellent open-source projects. The main third-party dependencies and their licenses are listed below and distributed accordingly. We thank the maintainers of these projects.

> 完整、逐版本锁定的依赖清单以 `backend/go.sum`、`frontend-web/package-lock.json` 为准；本表为面向使用者的说明。
>
> The complete, version-pinned dependency list is authoritative from `backend/go.sum` and `frontend-web/package-lock.json`; this table is an explanation for users.

## 后端（Go） / Backend (Go)

| 依赖 / Dependency | 用途 / Purpose | 许可证 / License |
| ---- | ---- | ------ |
| Gin | HTTP Web 框架<br>HTTP web framework | MIT |
| GORM | ORM | MIT |
| SQLite | 嵌入式数据库（默认存储）<br>Embedded database (default storage) | 公有领域（Public Domain）<br>Public Domain |
| gorilla/websocket | WebSocket 实时协作通道<br>WebSocket real-time collaboration channel | BSD-2-Clause |
| golang-jwt | JWT 签发与校验<br>JWT issuance and verification | MIT |
| bleve | 全文搜索（含中文分词）<br>Full-text search (with Chinese tokenization) | Apache-2.0 |
| bluemonday | 富文本 HTML 消毒<br>Rich-text HTML sanitization | BSD / MIT（见上游）<br>BSD / MIT (see upstream) |
| go-imap | 邮件 IMAP 收信<br>Mail IMAP receiving | MIT |
| pquerna/otp | TOTP 两步验证<br>TOTP two-factor authentication | Apache-2.0 |
| minio-go | S3 / OSS 对象存储客户端<br>S3 / OSS object storage client | **AGPL-3.0 / Apache-2.0 双许可**（按上游选择，见下）<br>**AGPL-3.0 / Apache-2.0 dual license** (choose per upstream, see below) |

## 前端（Web / 移动端） / Frontend (Web / Mobile)

| 依赖 / Dependency | 用途 / Purpose | 许可证 / License |
| ---- | ---- | ------ |
| Vue 3 | 前端框架<br>Frontend framework | MIT |
| Vite | 构建工具<br>Build tool | MIT |
| Element Plus | UI 组件库<br>UI component library | MIT |
| Pinia | 状态管理<br>State management | MIT |
| Tiptap | 富文本（文字）编辑器<br>Rich-text (document) editor | MIT |
| Luckysheet | 在线表格引擎<br>Online spreadsheet engine | MIT |
| Electron | 桌面端壳<br>Desktop shell | MIT |
| Capacitor | 移动端壳<br>Mobile shell | MIT |

## 运行时外部依赖 / Runtime External Dependencies

| 依赖 / Dependency | 用途 / Purpose | 许可证 / License |
| ---- | ---- | ------ |
| LibreOffice | `docx/xlsx/pptx` 导入转换、PDF 渲染导出<br>`docx/xlsx/pptx` import conversion, PDF render & export | MPL-2.0 |

## 特别说明：minio-go 双许可 / Note: minio-go Dual License

`minio-go` 上游采用 **AGPL-3.0 与 Apache-2.0 双许可**。若你的部署场景对许可证有严格要求（例如闭源分发），请按上游仓库说明确认所选许可并遵守相应义务；也可将存储后端切换回本地磁盘（`storage: local`）以规避该依赖。

`minio-go` upstream uses a **dual license of AGPL-3.0 and Apache-2.0**. If your deployment scenario has strict license requirements (e.g., closed-source distribution), please confirm your chosen license per the upstream repo's notes and comply with its obligations; you can also switch the storage backend back to local disk (`storage: local`) to avoid this dependency.

## 数据来源与素材 / Data Sources & Assets

- 文档默认图标、占位图来自自绘 / 开源图标集，随本项目 MIT 许可分发。
  - Default document icons and placeholder images come from self-drawn / open-source icon sets, distributed under this project's MIT license.
- 截图中出现的示例文档内容均为虚构。
  - Example document content shown in screenshots is entirely fictional.

如果你认为某个依赖被遗漏或许可标注有误，欢迎提 Issue 或 PR 指正。

If you believe a dependency is missing or a license label is incorrect, please open an Issue or PR to point it out.
