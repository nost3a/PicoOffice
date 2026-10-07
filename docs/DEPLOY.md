# PicoOffice 部署教程 / Deployment Guide

从一台裸 Linux 开始，到浏览器、桌面、手机三端都能连上私有化服务器。所有命令在 Ubuntu 22.04 / Debian 12 上验证过，CentOS 8+ 同理，包名按发行版换。

From a bare Linux machine to having the browser, desktop, and mobile clients all connect to your self-hosted server. All commands are verified on Ubuntu 22.04 / Debian 12; the same applies to CentOS 8+, just swap package names per distribution.

---

## 1. 硬件要求 / Hardware Requirements

| 用途 / Use case | CPU | 内存 / Memory | 磁盘 / Disk |
|---|---|---|---|
| 最低（个人/小团队 < 10 人）/ Minimum (personal / small team < 10) | 1 核 / 1 core | 1 GB | 10 GB |
| 推荐（10–50 人）/ Recommended (10–50 people) | 2 核 / 2 cores | 2 GB | 50 GB |
| 重度（>50 人，多人协作）/ Heavy (>50 people, multi-user collaboration) | 4 核 / 4 cores | 4 GB | 200 GB SSD |

soffice 转 PDF 时单次会吃 200–500MB 内存，并发导出多了要加内存。

When soffice converts to PDF, a single conversion consumes 200–500 MB of memory; if you run many concurrent exports, add more RAM.

---

## 2. 系统要求 / System Requirements

- Ubuntu 22.04 / Debian 12 / CentOS 8+，其他现代 Linux 也能跑。
  - Ubuntu 22.04 / Debian 12 / CentOS 8+, other modern Linux also works.
- 内核 ≥ 4.15（跑 SQLite + Nginx 足够）。
  - Kernel ≥ 4.15 (sufficient to run SQLite + Nginx).
- 出网：装依赖、下 Go/Node 时需要；跑起来后内网即可。
  - Outbound network: needed when installing dependencies and downloading Go/Node; after it's running, internal network is enough.
- 域名：要上 HTTPS 才需要；纯内网用 IP 也行。
  - Domain name: only needed for HTTPS; a plain internal network can use an IP.

---

## 3. 从源码编译（裸机第一步） / Build from Source

### 3.1 装基础依赖 / Install Basic Dependencies

```bash
# Ubuntu / Debian
sudo apt update
sudo apt install -y git curl wget build-essential libssl-dev

# LibreOffice（导出 pdf/docx/xlsx 必须）
sudo apt install -y libreoffice

# 验证 soffice 在 PATH 里
soffice --version
# 期望输出：LibreOffice 7.x.x ...
```

```bash
# CentOS 8+
sudo dnf install -y git curl wget gcc gcc-c++ make openssl-devel
sudo dnf install -y libreoffice
soffice --version
```

### 3.2 装 Go 1.23+ / Install Go 1.23+

```bash
# 下官方 tarball（以 1.23.4 为例，新版自己换号）
cd /tmp
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# 写进 PATH
echo 'export PATH=/usr/local/go/bin:$PATH' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh

go version
# 期望：go version go1.23.4 linux/amd64
```

### 3.3 装 Node 20+ / Install Node 20+

```bash
# 用 nodesource 源
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

node -v   # 期望 v20.x
npm -v    # 期望 10.x
```

### 3.4 拉代码 / Fetch the Code

```bash
sudo mkdir -p /opt
sudo chown $USER:$USER /opt
cd /opt

# 如果代码在 git 上：
# git clone https://github.com/nost3a/PicoOffice picooffice

# 如果是从别处拷过来的 picooffice/ 目录，直接放进来即可
ls /opt/picooffice/
# 期望看到 backend/  frontend-web/  desktop-electron/  mobile-capacitor/  storage/
```

### 3.5 编译前端 / Build the Frontend

```bash
cd /opt/picooffice/frontend-web
npm ci            # 按 lock 文件装，不要用 npm install
npm run build

# 产物在 frontend-web/dist/
ls dist/
# 期望：index.html  assets/  luckysheet/
```

### 3.6 把前端 dist 拷进后端 web 目录 / Copy the Frontend dist into the Backend web Directory

后端用 `//go:embed all:web` 内嵌前端，所以编译后端前必须把 dist 拷到 `backend/web/`：

The backend embeds the frontend via `//go:embed all:web`, so before building the backend you must copy dist into `backend/web/`:

```bash
cd /opt/picooffice
rm -rf backend/web
mkdir -p backend/web
cp -r frontend-web/dist/* backend/web/

ls backend/web/
# 期望：index.html  assets/  luckysheet/
```

### 3.7 编译后端单二进制 / Build the Single Backend Binary

```bash
cd /opt/picooffice/backend
go build -o picooffice .

ls -lh picooffice
# 期望：约 66MB 的可执行文件
```

### 3.8 首次手动启动验证 / First Manual Startup Verification

```bash
cd /opt/picooffice/backend

# 用环境变量跑一次
PICO_PORT=8080 \
PICO_DB=./picooffice.db \
PICO_STORAGE=../storage \
PICO_JWT_SECRET='change-me-in-production-please-32chars' \
./picooffice
```

看到日志：

You'll see the log:

```
picooffice listen :8080, db=./picooffice.db, storage=../storage
```

另开一个终端：

In a separate terminal:

```bash
curl http://127.0.0.1:8080/api/health
# 期望：{"status":"ok"}
```

浏览器开 `http://你的服务器IP:8080`，能看到登录页就成。`Ctrl+C` 停掉，继续下面配 systemd。

Open `http://your-server-IP:8080` in a browser; if you see the login page, you're good. Press `Ctrl+C` to stop, then continue to configure systemd below.

---

## 4. systemd 服务配置 / systemd Service

### 4.1 建一个专用用户 / Create a Dedicated User

```bash
sudo useradd -r -m -d /opt/picooffice -s /usr/sbin/nologin pico
sudo chown -R pico:pico /opt/picooffice
```

### 4.2 写 service 文件 / Write the Service File

```bash
sudo tee /etc/systemd/system/picooffice.service > /dev/null <<'EOF'
[Unit]
Description=PicoOffice server
After=network.target

[Service]
Type=simple
User=pico
Group=pico
WorkingDirectory=/opt/picooffice/backend
Environment=PICO_PORT=8080
Environment=PICO_DB=/opt/picooffice/backend/picooffice.db
Environment=PICO_STORAGE=/opt/picooffice/storage
Environment=PICO_JWT_SECRET='把这里换成你自己的长随机串_至少32字符'
ExecStart=/opt/picooffice/backend/picooffice
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

# 安全加固（可选）
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
```

生成一个随机 JWT 密钥：

Generate a random JWT secret:

```bash
openssl rand -hex 32
# 把输出粘贴到 PICO_JWT_SECRET 那一行
```

### 4.3 启动并设开机自启 / Start and Enable on Boot

```bash
sudo systemctl daemon-reload
sudo systemctl enable picooffice
sudo systemctl start picooffice

sudo systemctl status picooffice
# 期望：active (running)

# 看日志
journalctl -u picooffice -f
```

---

## 5. Nginx 反向代理 / Nginx Reverse Proxy

裸 `:8080` 只适合内网测试。对外服务前面挂 Nginx，统一 80/443，顺带处理 WebSocket upgrade。

The bare `:8080` is only suitable for internal testing. For external service, put Nginx in front to unify 80/443 and handle the WebSocket upgrade.

### 5.1 装 Nginx / Install Nginx

```bash
sudo apt install -y nginx
```

### 5.2 写站点配置 / Write the Site Config

把 `your-domain.com` 换成你的域名（或用 IP）：

Replace `your-domain.com` with your domain (or use an IP):

```bash
sudo tee /etc/nginx/sites-available/picooffice > /dev/null <<'EOF'
server {
    listen 80;
    server_name your-domain.com;

    # 上传文件大小限制（按需调大）
    client_max_body_size 100m;

    # WebSocket upgrade map 在 http 块里更好，这里放 server 块也能用
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 必须的两行
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # WS 心跳别超时
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
EOF
```

启用站点：

Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/picooffice /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t          # 检查语法
sudo systemctl reload nginx
```

浏览器开 `http://your-domain.com`，能看到登录页就 OK。

Open `http://your-domain.com` in a browser; if you see the login page, it's OK.

---

## 6. HTTPS（Let's Encrypt） / HTTPS with Let's Encrypt

### 6.1 装 certbot / Install certbot

```bash
sudo apt install -y certbot python3-certbot-nginx
```

### 6.2 申请证书 / Request a Certificate

```bash
sudo certbot --nginx -d your-domain.com
```

交互里：

During the interaction:

1. 填邮箱（过期提醒用）。
  - Enter an email (for expiry reminders).
2. 同意条款。
  - Agree to the terms.
3. 是否分享邮箱给 EFF（随便）。
  - Whether to share your email with EFF (your choice).
4. 是否把 HTTP 自动跳 HTTPS —— 选 **2（重定向）**。
  - Whether to auto-redirect HTTP to HTTPS — choose **2 (redirect)**.

certbot 会自动改 Nginx 配置，加上 443 server block 和证书路径。完成后：

certbot automatically edits the Nginx config, adding the 443 server block and certificate paths. After that:

```bash
sudo systemctl reload nginx
```

浏览器开 `https://your-domain.com`，锁标志出来就成。

Open `https://your-domain.com` in a browser; when the lock icon appears, you're done.

### 6.3 自动续期 / Auto-Renewal

certbot 装完自带 systemd timer，验证一下：

After installing certbot, a systemd timer comes built in; verify it:

```bash
systemctl list-timers | grep certbot
# 期望看到 certbot.timer
```

手动 dry-run 测试续期：

Manually test renewal with a dry-run:

```bash
sudo certbot renew --dry-run
```

### 6.4 客户端侧注意 / Client-Side Notes

上 HTTPS 后：

After enabling HTTPS:

- 浏览器直接访问 `https://your-domain.com`。
  - Access `https://your-domain.com` directly in the browser.
- Electron 桌面壳、Android APK 里填服务器地址也要写 `https://your-domain.com`。
  - The Electron desktop shell and Android APK must also use `https://your-domain.com` as the server address.
- Android 9+ 默认禁明文 HTTP，上 HTTPS 后正好把 `mobile-capacitor/capacitor.config.json` 里的 `cleartext: true` 关掉。
  - Android 9+ blocks plain HTTP by default; once on HTTPS you can turn off `cleartext: true` in `mobile-capacitor/capacitor.config.json`.

---

## 7. SQLite vs MySQL 切换 / Switching SQLite to MySQL

默认 SQLite，零运维。要切 MySQL 走两步。

Defaults to SQLite with zero maintenance. To switch to MySQL, follow two steps.

### 7.1 改环境变量 / Change Environment Variables

```bash
# 原来
# Environment=PICO_DB=/opt/picooffice/backend/picooffice.db

# 改成 MySQL DSN（GORM 格式）
Environment=PICO_DB='root:密码@tcp(127.0.0.1:3306)/picooffice?charset=utf8mb4&parseTime=True&loc=Local'
```

### 7.2 改 main.go 里的驱动 / Change the Driver in main.go

当前 `backend/main.go` 第 12 行 import 的是 sqlite，第 31 行用的是 `sqlite.Open`。切 MySQL 需要改两处：

In the current `backend/main.go`, line 12 imports sqlite and line 31 uses `sqlite.Open`. Switching to MySQL requires two changes:

**diff：**

**diff:**

```diff
 import (
     "embed"
     "io/fs"
     "log"
     "net/http"
     "os"
     "strings"

     "github.com/gin-gonic/gin"
-    "gorm.io/driver/sqlite"
+    "gorm.io/driver/mysql"
     "gorm.io/gorm"

     "picooffice/internal/handler"
     "picooffice/internal/middleware"
     "picooffice/internal/model"
     "picooffice/internal/ws"
 )
```

```diff
-    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
+    db, err := gorm.Open(mysql.Open(dbPath), &gorm.Config{})
     if err != nil {
         log.Fatalf("open db: %v", err)
     }
```

同时把 `backend/go.mod` 里加上 mysql 驱动：

Also add the MySQL driver to `backend/go.mod`:

```bash
cd /opt/picooffice/backend
go get gorm.io/driver/mysql
go mod tidy
go build -o picooffice .
```

GORM AutoMigrate 会自动在 MySQL 里建表。第一次跑前先在 MySQL 里 `CREATE DATABASE picooffice CHARACTER SET utf8mb4;`。

GORM AutoMigrate will automatically create the tables in MySQL. Before the first run, execute `CREATE DATABASE picooffice CHARACTER SET utf8mb4;` in MySQL.

### 7.3 数据迁移（从 SQLite 搬过来）/ Data Migration (from SQLite)

```bash
# 用 sqlite3 命令行导出（需先 sudo apt install sqlite3）
sqlite3 /opt/picooffice/backend/picooffice.db .dump > /tmp/pico.sql

# 再导入 MySQL
mysql -u root -p picooffice < /tmp/pico.sql
```

迁移完重启：

After migration, restart:

```bash
sudo systemctl restart picooffice
```

---

## 8. 各客户端如何连私有化服务器 / Connecting Clients

### 8.1 浏览器 / Browser

直接开 `https://your-domain.com`。第一个注册的用户自动是 admin。

Open `https://your-domain.com` directly. The first registered user automatically becomes admin.

### 8.2 Linux deb 桌面端 / Linux deb Desktop

产物：`desktop-electron/dist/picooffice-desktop_1.0.0_amd64.deb`（约 75MB）。

Artifact: `desktop-electron/dist/picooffice-desktop_1.0.0_amd64.deb` (~75 MB).

```bash
sudo dpkg -i picooffice-desktop_1.0.0_amd64.deb
# 缺依赖的话
sudo apt install -f
```

启动方式：

How to launch:

```bash
# 命令行
picooffice-desktop

# 或在 GNOME/KDE 应用菜单里找 "PicoOffice"
```

首次启动会弹一个服务器地址输入框，填 `https://your-domain.com`，回车，进入登录页。

On first launch a server-address input pops up; enter `https://your-domain.com`, press Enter, and you'll reach the login page.

### 8.3 Windows 绿色版 / Windows Portable Version

产物：`desktop-electron/dist/PicoOffice-1.0.0-win.zip`（约 108MB）。

Artifact: `desktop-electron/dist/PicoOffice-1.0.0-win.zip` (~108 MB).

1. 把 zip 拷到 Windows 机器上，解压。
  - Copy the zip to a Windows machine and extract it.
2. 进 `win-unpacked/` 目录，双击 `PicoOffice.exe`。
  - Enter the `win-unpacked/` directory and double-click `PicoOffice.exe`.
3. 首次启动弹服务器地址框，填 `https://your-domain.com`。
  - On first launch a server-address box appears; enter `https://your-domain.com`.

不需要安装、不需要管理员权限，解压即用。

No installation, no admin rights needed; just extract and run.

### 8.4 Windows 真 nsis 安装包（需在 Windows/Mac 上打）/ Windows NSIS Installer (build on Windows/Mac)

沙箱 Linux 没 wine，打不出 nsis exe。在 Windows 或 Mac 上：

The sandbox Linux has no wine, so it can't build the nsis exe. On Windows or Mac:

```bash
cd desktop-electron/
npm install
npm run dist:win
```

产物：`desktop-electron/dist/PicoOffice Setup 1.0.0.exe`。双击安装，可选安装目录。

Artifact: `desktop-electron/dist/PicoOffice Setup 1.0.0.exe`. Double-click to install; choose the install directory.

### 8.5 Android APK / Android APK

产物：`mobile-capacitor/android/app/build/outputs/apk/debug/app-debug.apk`（约 12MB，debug 签名）。

Artifact: `mobile-capacitor/android/app/build/outputs/apk/debug/app-debug.apk` (~12 MB, debug-signed).

1. 把 apk 拷到 Android 手机上。
  - Copy the apk to an Android phone.
2. 允许"安装未知来源应用"。
  - Allow "install from unknown sources".
3. 点开 apk 安装。
  - Tap the apk to install.
4. 启动后弹服务器地址框，填 `https://your-domain.com`。
  - After launch a server-address box appears; enter `https://your-domain.com`.

注意：

Note:

- 模拟器默认填的是 `http://10.0.2.2:8080`（10.0.2.2 = 模拟器宿主机），真机不要用这个。
  - The emulator defaults to `http://10.0.2.2:8080` (10.0.2.2 = the emulator host); do not use this on a real device.
- 如果服务器是 `http://`（没上 HTTPS），Android 9+ 会拦，必须在 `capacitor.config.json` 里开 `cleartext: true` 并重新打包。
  - If the server is `http://` (no HTTPS), Android 9+ will block it; you must enable `cleartext: true` in `capacitor.config.json` and rebuild the package.

### 8.6 macOS dmg（需在 Mac 上打）/ macOS dmg (build on Mac)

沙箱是 Linux，出不了 dmg。在 Mac 上：

The sandbox is Linux, so it can't produce a dmg. On a Mac:

```bash
cd desktop-electron/
npm install
npm run dist:mac
# 或手动调 electron-builder
npx electron-builder --mac dmg
```

产物：`desktop-electron/dist/PicoOffice-1.0.0.dmg`。双击 dmg，把 PicoOffice 拖进 Applications。

Artifact: `desktop-electron/dist/PicoOffice-1.0.0.dmg`. Double-click the dmg and drag PicoOffice into Applications.

### 8.7 iOS ipa（需在 Mac 上打）/ iOS ipa (build on Mac)

按 `mobile-capacitor/README.md`：

Per `mobile-capacitor/README.md`:

```bash
# Mac 上装 Xcode + CocoaPods
sudo gem install cocoapods

cd mobile-capacitor/
npx cap add ios
npx cap sync ios
open ios/App/App.xcodeproj
```

在 Xcode 里：

In Xcode:

1. 选 Team（开发者账号）。
  - Select a Team (developer account).
2. Bundle ID 已是 `com.picooffice.app`。
  - The Bundle ID is already `com.picooffice.app`.
3. Product → Archive。
  - Product → Archive.
4. 导出 ipa 或上传 App Store。
  - Export the ipa or upload to the App Store.

---

## 9. 备份 / Backup

要备两样：SQLite 数据库文件 + 上传附件目录。

Back up two things: the SQLite database file + the uploaded-attachments directory.

### 9.1 手动备份 / Manual Backup

```bash
BACKUP_DIR=/backup/picooffice
DATE=$(date +%Y%m%d-%H%M)

mkdir -p $BACKUP_DIR
sudo systemctl stop picooffice

cp /opt/picooffice/backend/picooffice.db $BACKUP_DIR/picooffice-$DATE.db
tar czf $BACKUP_DIR/storage-$DATE.tar.gz -C /opt/picooffice storage/

sudo systemctl start picooffice
```

### 9.2 crontab 自动备份（不停服务）/ crontab Auto-Backup (without stopping service)

SQLite 用 `.backup` 命令做热备，不锁库：

Use SQLite's `.backup` command for a hot backup that doesn't lock the database:

```bash
sudo crontab -e -u root
```

加一行（每天凌晨 3 点备份，保留 14 天）：

Add a line (backup at 3 AM daily, keep 14 days):

```cron
0 3 * * * DATE=$(date +\%Y\%m\%d-\%H\%M); sqlite3 /opt/picooffice/backend/picooffice.db ".backup '/backup/picooffice/picooffice-$DATE.db'" && tar czf /backup/picooffice/storage-$DATE.tar.gz -C /opt/picooffice storage/ && find /backup/picooffice -mtime +14 -delete
```

如果切了 MySQL，把 sqlite3 那行换成 `mysqldump -u root -p密码 picooffice > /backup/picooffice/picooffice-$DATE.sql`。

If you switched to MySQL, replace the sqlite3 line with `mysqldump -u root -p密码 picooffice > /backup/picooffice/picooffice-$DATE.sql`.

### 9.3 恢复 / Restore

```bash
sudo systemctl stop picooffice

# 恢复数据库
cp /backup/picooffice/picooffice-20261006-0300.db /opt/picooffice/backend/picooffice.db

# 恢复附件
rm -rf /opt/picooffice/storage
tar xzf /backup/picooffice/storage-20261006-0300.tar.gz -C /opt/picooffice/

sudo chown -R pico:pico /opt/picooffice
sudo systemctl start picooffice
```

---

## 10. 升级 / Upgrading

### 10.1 只换二进制（前端没动）/ Binary-only Swap (frontend unchanged)

```bash
# 1. 备份
sudo cp /opt/picooffice/backend/picooffice /opt/picooffice/backend/picooffice.bak
sudo cp /opt/picooffice/backend/picooffice.db /opt/picooffice/backend/picooffice.db.bak

# 2. 编译新二进制（在开发机上）
cd backend && go build -o picooffice .

# 3. 拷到服务器，覆盖
scp picooffice pico@your-server:/opt/picooffice/backend/

# 4. 重启
sudo systemctl restart picooffice
sudo systemctl status picooffice
```

### 10.2 前端也改了 / Frontend Also Changed

```bash
# 开发机
cd /opt/picooffice/frontend-web   # 或本地
npm ci && npm run build
rm -rf ../backend/web && mkdir -p ../backend/web
cp -r dist/* ../backend/web/
cd ../backend && go build -o picooffice .

# 拷到服务器
scp picooffice pico@your-server:/opt/picooffice/backend/
sudo systemctl restart picooffice
```

### 10.3 回滚 / Rollback

```bash
sudo systemctl stop picooffice
sudo cp /opt/picooffice/backend/picooffice.bak /opt/picooffice/backend/picooffice
sudo cp /opt/picooffice/backend/picooffice.db.bak /opt/picooffice/backend/picooffice.db
sudo systemctl start picooffice
```

---

## 11. 常见问题（FAQ） / FAQ

### 11.1 端口被占 / Port Occupied

```bash
# 看 8080 被谁占了
sudo ss -ltnp | grep :8080

# 要么杀掉占用进程，要么换端口
sudo systemctl edit picooffice
# 在里面加：
# [Service]
# Environment=PICO_PORT=9090
```

同时改 Nginx 配置里 `proxy_pass http://127.0.0.1:9090;`，reload nginx。

Also change `proxy_pass http://127.0.0.1:9090;` in the Nginx config, then reload nginx.

### 11.2 soffice 找不到，导出 PDF 501 / soffice Not Found, PDF Export 501

```bash
# 验证
which soffice
# 如果空，说明没装
sudo apt install -y libreoffice

# 装完看路径
which soffice
# 期望：/usr/bin/soffice

# 重启服务
sudo systemctl restart picooffice
```

soffice 第一次启动慢（要建 profile），第一次导出等 5–10 秒是正常的。

soffice is slow on first start (it builds a profile); waiting 5–10 seconds on the first export is normal.

### 11.3 上传文件失败 / 413 Request Entity Too Large / Upload Failed / 413 Request Entity Too Large

Nginx 默认限制 1MB。改 `/etc/nginx/sites-available/picooffice`，在 `server` 块里加：

Nginx limits to 1 MB by default. Edit `/etc/nginx/sites-available/picooffice` and add inside the `server` block:

```nginx
client_max_body_size 100m;
```

然后 `sudo nginx -t && sudo systemctl reload nginx`。

Then `sudo nginx -t && sudo systemctl reload nginx`.

### 11.4 WebSocket 连不上（协作文档看不到在线用户）/ WebSocket Won't Connect (collaborative doc看不到 online users)

按顺序查：

Check in order:

1. 浏览器 F12 → Network → WS，看 `wss://` 请求是不是 101 Switching Protocols。
  - In the browser press F12 → Network → WS, and check whether the `wss://` request returns 101 Switching Protocols.
2. 不是 101 就是 Nginx 没配 upgrade，确认 server 块里有：
  - If it's not 101, Nginx isn't configured for upgrade; confirm the server block contains:

    ```nginx
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    ```
3. 看后端日志 `journalctl -u picooffice -f`，有没有 `invalid token`。token 过期就重新登录。
  - Check the backend log `journalctl -u picooffice -f` for `invalid token`. If the token expired, log in again.
4. 上 HTTPS 后前端代码里 WS 地址要从 `ws://` 换成 `wss://`（当前前端已经按 `location.protocol` 自动判断，不需要手动改）。
  - After HTTPS, the WS address in the frontend code must change from `ws://` to `wss://` (the current frontend already auto-detects via `location.protocol`, no manual change needed).

### 11.5 第一个注册用户不是 admin / First Registered User Is Not admin

检查 `users` 表里有没有残留数据：

Check whether there's leftover data in the `users` table:

```bash
sqlite3 /opt/picooffice/backend/picooffice.db "SELECT id,username,role FROM users;"
```

如果表里已经有用户，第一个注册的人就不是 admin 了。把现有用户手动改成 admin：

If the table already has users, the first person to register won't be admin. Manually change an existing user to admin:

```bash
sqlite3 /opt/picooffice/backend/picooffice.db "UPDATE users SET role='admin' WHERE username='alice';"
```

### 11.6 JWT 密钥忘了，所有人被踢下线 / Forgot JWT Secret, Everyone Logged Out

这是正常的 —— 密钥一变，老 token 全部失效。所有人重新登录即可。

This is normal — once the secret changes, all old tokens become invalid. Everyone just needs to log in again.

### 11.7 磁盘满了 / Disk Full

```bash
# 看谁占得多
sudo du -sh /opt/picooffice/*
sudo du -sh /opt/picooffice/storage/uploads/*

# SQLite 跑久了会有碎片，VACUUM 一下
sudo systemctl stop picooffice
sqlite3 /opt/picooffice/backend/picooffice.db "VACUUM;"
sudo systemctl start picooffice
```

### 11.8 Android 真机连不上服务器 / Android Device Can't Reach Server

- 确认手机和服务器在同一网络，或服务器有公网 IP。
  - Confirm the phone and server are on the same network, or the server has a public IP.
- `http://` 明文被 Android 9+ 拦，要么上 HTTPS，要么在 `capacitor.config.json` 开 `cleartext: true` 重打 apk。
  - Plain `http://` is blocked by Android 9+; either use HTTPS or enable `cleartext: true` in `capacitor.config.json` and rebuild the apk.
- 服务器防火墙放开 80/443：
  - Open ports 80/443 on the server firewall:

  ```bash
  sudo ufw allow 80/tcp
  sudo ufw allow 443/tcp
  ```

### 11.9 Windows 双击 exe 闪退 / Windows exe Crashes on Double-Click

- 先开 `cmd`，把 exe 拖进去回车，看报错。
  - First open `cmd`, drag the exe in and press Enter to see the error.
- 常见原因：服务器地址填错（比如带了末尾 `/`），改成不带斜杠的 `https://your-domain.com`。
  - Common cause: wrong server address (e.g. with a trailing `/`); change it to `https://your-domain.com` without the slash.
- 杀软误报：Electron 打包的 exe 经常被误报，加白名单。
  - Antivirus false positive: Electron-packed exes are often falsely flagged; add to the allowlist.

### 11.10 升级后前端 404 / 白屏 / Frontend 404 / Blank Screen After Upgrade

八成是 `backend/web/` 没更新就编译后端了。重新走一遍：

Most likely `backend/web/` wasn't updated before building the backend. Run through it again:

```bash
cd frontend-web && npm run build
rm -rf ../backend/web && mkdir -p ../backend/web
cp -r dist/* ../backend/web/
cd ../backend && go build -o picooffice .
```

---

## 12. 部署检查清单（上线前过一遍） / Pre-launch Checklist

- [ ] `PICO_JWT_SECRET` 已改成随机长串（不是默认的 `picooffice-dev-secret`）。
  - [ ] `PICO_JWT_SECRET` has been changed to a random long string (not the default `picooffice-dev-secret`).
- [ ] `PICO_PORT` 不是 80/443，只监听 127.0.0.1，外面走 Nginx。
  - [ ] `PICO_PORT` is not 80/443, only listens on 127.0.0.1, with Nginx in front.
- [ ] Nginx 配了 HTTPS，certbot 自动续期正常。
  - [ ] Nginx is configured with HTTPS and certbot auto-renewal works.
- [ ] Nginx 配了 WebSocket upgrade header。
  - [ ] Nginx is configured with the WebSocket upgrade header.
- [ ] Nginx `client_max_body_size` 调到业务需要的值。
  - [ ] Nginx `client_max_body_size` is tuned to the business requirement.
- [ ] `soffice --version` 能跑，导出 PDF 测试通过。
  - [ ] `soffice --version` runs and a PDF export test passes.
- [ ] systemd 服务 `enable` 了，开机自启。
  - [ ] The systemd service is `enable`d and starts on boot.
- [ ] crontab 备份在跑，手动恢复演练过一次。
  - [ ] The crontab backup is running and a manual restore has been rehearsed once.
- [ ] 防火墙只开 22/80/443，8080 不对外。
  - [ ] The firewall only opens 22/80/443; 8080 is not exposed.
- [ ] 第一个 admin 账号已注册，密码不是弱密码。
  - [ ] The first admin account is registered with a non-weak password.
- [ ] 日志路径确认：`journalctl -u picooffice` 能看到访问日志。
  - [ ] Log path confirmed: `journalctl -u picooffice` shows access logs.

---

## 13. 公网部署：家里服务器，全国员工访问 / Public Deployment

前面 1~11 节讲的都是"在内网跑起来"。这一节解决最后一公里：把家里/办公室那台 8080 端口的小服务器暴露给全国各地员工手机和电脑访问。

Sections 1–11 all cover "getting it running on the internal network". This section solves the last mile: exposing that little 8080-port server at home / in the office to employees' phones and computers across the country.

按推荐度从高到低排：Cloudflare Tunnel（零成本零公网 IP）→ frp + 公网 VPS → Tailscale 组网（内部团队）→ DDNS + 端口映射（有公网 IP 才用）。

Ranked from most to least recommended: Cloudflare Tunnel (zero cost, zero public IP) → frp + public VPS → Tailscale mesh (internal team) → DDNS + port forwarding (only if you have a public IP).

后端统一监听 `127.0.0.1:8080`，不要直接监听 0.0.0.0。

The backend should uniformly listen on `127.0.0.1:8080`; do not listen directly on 0.0.0.0.

### 13.1 Cloudflare Tunnel（推荐）/ Cloudflare Tunnel (Recommended)

家里没公网 IP、不想买 VPS、想白嫖 HTTPS 和 DDoS 防护，就用这个。原理：家里服务器主动往外建一条到 Cloudflare 的加密隧道，员工访问 `office.yourdomain.com`，Cloudflare 从隧道把请求送回来。**路由器不用做任何端口映射。**

Use this if you have no public IP at home, don't want to buy a VPS, and want free HTTPS and DDoS protection. How it works: the home server actively opens an encrypted tunnel to Cloudflare; employees visit `office.yourdomain.com` and Cloudflare sends the request back through the tunnel. **No port forwarding needed on the router.**

前提：域名 NS 已经托管到 Cloudflare。

Prerequisite: the domain's NS must already be delegated to Cloudflare.

**装 cloudflared（Debian/Ubuntu）：**

**Install cloudflared (Debian/Ubuntu):**

```bash
curl -L --output cloudflared.deb https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared.deb
```

**登录授权（浏览器会弹一下，选你的域名）：**

**Log in to authorize (a browser prompt appears; pick your domain):**

```bash
cloudflared tunnel login
```

**建隧道：**

**Create the tunnel:**

```bash
cloudflared tunnel create picooffice
# 输出一行 Tunnel ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
# 凭证文件落到 ~/.cloudflared/<TUNNEL_ID>.json
```

**写配置 `~/.cloudflared/config.yml`（或 `/etc/cloudflared/config.yml`）：**

**Write the config `~/.cloudflared/config.yml` (or `/etc/cloudflared/config.yml`):**

```yaml
tunnel: picooffice
credentials-file: /home/USER/.cloudflared/<TUNNEL_ID>.json

ingress:
  - hostname: office.yourdomain.com
    service: http://127.0.0.1:8080
  - service: http_status:404
```

**绑定 DNS（把 office.yourdomain.com CNAME 到隧道）：**

**Bind DNS (CNAME office.yourdomain.com to the tunnel):**

```bash
cloudflared tunnel route dns picooffice office.yourdomain.com
```

**注册 systemd 服务：**

**Register the systemd service:**

```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
sudo systemctl status cloudflared
```

完事。员工访问 `https://office.yourdomain.com` 就是 PicoOffice，HTTPS 由 Cloudflare 自动签，证书自动续，不用 certbot。

Done. Employees visiting `https://office.yourdomain.com` get PicoOffice; HTTPS is auto-signed by Cloudflare, certificates auto-renew, no certbot needed.

WebSocket 也走同一条隧道，不用额外配。

WebSocket also goes through the same tunnel, no extra config.

**优点：** 不用公网 IP、不用开路由器端口、自动 HTTPS、Cloudflare 挡一层 DDoS、免费。

**Pros:** No public IP needed, no router port opening, automatic HTTPS, Cloudflare absorbs a layer of DDoS, free.

**缺点：** 域名要在 Cloudflare；国内访问 Cloudflare 偶尔慢，员工大量在国内时测一下延迟。

**Cons:** The domain must be on Cloudflare; access to Cloudflare from within China is occasionally slow, so test latency if most employees are in China.

### 13.2 frp + 公网 VPS / frp + Public VPS

已经有一台阿里云/腾讯云 VPS（有公网 IP），就用 frp。家里服务器主动连 VPS 的 7000 端口，员工访问 VPS 上的 Nginx。

If you already have an Alibaba Cloud / Tencent Cloud VPS (with a public IP), use frp. The home server actively connects to the VPS's port 7000, and employees access the Nginx on the VPS.

**VPS 上：frps**

**On the VPS: frps**

`/etc/frp/frps.ini`：

`/etc/frp/frps.ini`:

```ini
[common]
bind_port = 7000
vhost_http_port = 8080
token = 换一串长一点的随机密码_jk38sdJHkqWEr983
# dashboard（可选，看连接状态）
dashboard_port = 7500
dashboard_user = admin
dashboard_pwd = 换个密码
```

`/etc/systemd/system/frps.service`：

`/etc/systemd/system/frps.service`:

```ini
[Unit]
Description=frps
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/frps -c /etc/frp/frps.ini
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now frps
sudo ufw allow 7000/tcp
sudo ufw allow 8080/tcp
sudo ufw allow 7500/tcp
```

**家里服务器上：frpc**

**On the home server: frpc**

`/etc/frp/frpc.ini`：

`/etc/frp/frpc.ini`:

```ini
[common]
server_addr = 你的VPS公网IP
server_port = 7000
token = 和 frps.ini 里一模一样

[picooffice-web]
type = http
local_port = 8080
custom_domains = office.yourdomain.com
```

`/etc/systemd/system/frpc.service`：

`/etc/systemd/system/frpc.service`:

```ini
[Unit]
Description=frpc
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/frpc -c /etc/frp/frpc.ini
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now frpc
```

**VPS 上再套一层 Nginx + certbot：**

**Add another layer of Nginx + certbot on the VPS:**

员工最终访问的是 `https://office.yourdomain.com`（443），Nginx 在 VPS 上把请求转发到本地 8080（也就是 frp 吐出来的端口）：

Employees ultimately visit `https://office.yourdomain.com` (443); the Nginx on the VPS forwards the request to local 8080 (the port frp exposes):

```nginx
server {
    listen 80;
    server_name office.yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name office.yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/office.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/office.yourdomain.com/privkey.pem;

    client_max_body_size 100m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 必须
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
    }
}
```

```bash
sudo certbot --nginx -d office.yourdomain.com
```

### 13.3 Tailscale / ZeroTier 组网（内部团队，不暴露公网）/ Tailscale / ZeroTier Mesh (internal team, not exposed)

员工都在公司或家里，但不想把服务暴露给整个公网，就用 VPN 组网。所有员工设备和服务器装 Tailscale，互相之间直接加密通信，PicoOffice 完全不开放公网端口。

Use a VPN mesh when employees are all in the office or at home but you don't want to expose the service to the whole public internet. Install Tailscale on all employee devices and the server; they communicate directly over encrypted channels, and PicoOffice never opens a public port.

**服务器上：**

**On the server:**

```bash
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up
# 输出一个登录链接，浏览器登录后这台机器拿到 100.x.x.x 地址
```

**员工手机/电脑：**

**Employee phones / computers:**

- iOS / Android：应用商店装 Tailscale，登录同一账号。
  - iOS / Android: install Tailscale from the app store and log into the same account.
- Windows / macOS：官网下客户端。
  - Windows / macOS: download the client from the official site.

然后员工浏览器访问 `http://100.x.x.x:8080`（服务器那台 Tailscale 地址）就是 PicoOffice。

Then employees visit `http://100.x.x.x:8080` (the server's Tailscale address) in the browser to reach PicoOffice.

**优点：** 零配置加密、不暴露公网、不用域名、不用 HTTPS（Tailscale 自己加密）。

**Pros:** Zero-config encryption, no public exposure, no domain needed, no HTTPS needed (Tailscale encrypts itself).

**缺点：** 外部客户不能访问；员工手机必须装 Tailscale 才能用。适合纯内部小团队。

**Cons:** External clients can't access it; employees' phones must install Tailscale to use it. Best for purely internal small teams.

ZeroTier 用法类似，把 `tailscale up` 换成 `zerotier-cli join <network_id>`，这里不展开。

ZeroTier works similarly — replace `tailscale up` with `zerotier-cli join <network_id>`; we won't go into detail here.

### 13.4 DDNS + 路由器端口映射（有真公网 IP 才用）/ DDNS + Router Port Forwarding (only with real public IP)

先确认家里宽带有没有真公网 IP。很多联通/电信给的是 CGNAT（运营商 NAT），这种方法用不了。查法：登录路由器看 WAN 口 IP，和百度搜"我的 IP"对比，一致就是真公网。

First confirm whether your home broadband has a real public IP. Many China Unicom / China Telecom connections are CGNAT (carrier NAT), making this method unusable. How to check: log into the router to see the WAN IP and compare it with what Baidu shows for 'my IP'; if they match, it's a real public IP.

**真公网 IP 情况下：**

**With a real public IP:**

1. 域名解析到家里公网 IP。家庭宽带 IP 会变，用 DDNS：
  - Point the domain to your home public IP. The home broadband IP changes, so use DDNS:
   - 阿里云域名 → 阿里云 DDNS 客户端（路由器里通常自带，选阿里云，填 AccessKey）。
     - Alibaba Cloud domain → Alibaba Cloud DDNS client (usually built into the router; pick Alibaba Cloud and fill in the AccessKey).
   - 不想用阿里云 → no-ip.com 客户端，路由器一般也支持。
     - Prefer not Alibaba Cloud → no-ip.com client, which routers generally also support.
2. 路由器管理页 → 端口映射（虚拟服务器）：
  - Router admin page → port forwarding (virtual server):
   - 外部 80 → 内网服务器 80
     - External 80 → internal server 80
   - 外部 443 → 内网服务器 443
     - External 443 → internal server 443
3. 家里服务器跑 Nginx 反代到 127.0.0.1:8080（配置见 13.5）。
  - The home server runs Nginx reverse proxy to 127.0.0.1:8080 (config in 13.5).
4. certbot 申请证书：
  - Request a certificate with certbot:

  ```bash
  sudo certbot --nginx -d office.yourdomain.com
  ```

**坑：**

**Gotchas:**

- 运营商封 80/443 的话（家宽经常封 80），员工只能访问 `https://office.yourdomain.com:8443`，Nginx listen 改成 8443。
  - If the carrier blocks 80/443 (home broadband often blocks 80), employees can only reach `https://office.yourdomain.com:8443`; change Nginx's listen to 8443.
- CGNAT 没真公网 IP → 这条路走不通，回到 13.1 Cloudflare Tunnel。
  - No real public IP under CGNAT → this path won't work; go back to 13.1 Cloudflare Tunnel.
- 家宽上行带宽通常只有 30~50Mbps，多人同时用会卡，别指望当正经生产服务器。
  - Home broadband upstream is usually only 30–50 Mbps; many concurrent users will lag, so don't expect it to serve as a real production server.

### 13.5 Nginx 反代完整样例（家里服务器本机）/ Complete Nginx Reverse Proxy Example (home server localhost)

如果走 13.4 或者你自己已经有 Nginx，完整配置长这样：

If you're following 13.4 or already have Nginx, the complete config looks like this:

```nginx
server {
    listen 80;
    server_name office.yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name office.yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/office.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/office.yourdomain.com/privkey.pem;
    ssl_protocols      TLSv1.2 TLSv1.3;

    # 后端单文件 100MB 上限，必须比后端大或等于
    client_max_body_size 100m;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 心跳长连接
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

关键点：

Key points:

- `client_max_body_size 100m` 必须配，否则分块上传虽然过得了，但小文件直接上传会被 Nginx 在 1MB 默认值那里拦下。
  - `client_max_body_size 100m` must be set, otherwise chunked uploads may pass but small direct uploads get blocked by Nginx's 1 MB default.
- `Upgrade` / `Connection` 两个 header 只在 `/api/` 下配也行（WS 走 `/api/ws`），全局配也无害。
  - The `Upgrade` / `Connection` headers can be set only under `/api/` (WS goes through `/api/ws`), and setting them globally is harmless too.
- `proxy_read_timeout 3600s` 一定要调，默认 60s 会把 WebSocket 心跳干掉，60s 后服务端还没回包就被 Nginx 掐了。
  - `proxy_read_timeout 3600s` must be adjusted; the default 60s kills the WebSocket heartbeat, and after 60s with no response Nginx drops the connection.

### 13.6 公网部署前检查清单 / Pre-Public-Deployment Checklist

走公网之前再做一遍，比第 12 节那份更针对"暴露在公网"这件事：

Do this again before going public; it targets "exposure to the public internet" more specifically than the list in section 12:

- [ ] `PICO_JWT_SECRET` 改成 `openssl rand -hex 32` 生成的强随机串，不是默认的 `picooffice-dev-secret`。
  - [ ] `PICO_JWT_SECRET` is a strong random string from `openssl rand -hex 32`, not the default `picooffice-dev-secret`.
- [ ] 第一个 admin 密码已经改成长密码，不要 `admin123`。
  - [ ] The first admin password is already a long password, not `admin123`.
- [ ] 登录限流生效：故意连续输错密码 11 次，第 11 次看到 429。
  - [ ] Login rate limiting works: intentionally enter the wrong password 11 times in a row and see 429 on the 11th.
- [ ] 备份 crontab 在跑：`crontab -l` 能看到 sqlite 备份条目；手动跑一次 `/backup/picooffice/` 有新文件。
  - [ ] The backup crontab is running: `crontab -l` shows the sqlite backup entry; running it once produces a new file in `/backup/picooffice/`.
- [ ] HTTPS 证书自动续期：`sudo certbot renew --dry-run` 没报错。
  - [ ] HTTPS certificate auto-renewal: `sudo certbot renew --dry-run` reports no error.
- [ ] 后端只监听 127.0.0.1：`ss -tlnp | grep 8080` 看到的是 `127.0.0.1:8080`，不是 `0.0.0.0:8080`。
  - [ ] The backend only listens on 127.0.0.1: `ss -tlnp | grep 8080` shows `127.0.0.1:8080`, not `0.0.0.0:8080`.
- [ ] 防火墙 8080 不对外：`sudo ufw status` 里没有 `8080` 这条 allow。
  - [ ] Firewall doesn't expose 8080: `sudo ufw status` has no allow rule for `8080`.
- [ ] 上传大文件测过：传一个 50MB 的 mp4，看 `storage/files/` 落盘、Dashboard 配额条数字涨上去。
  - [ ] Large-file upload tested: upload a 50 MB mp4 and confirm it lands in `storage/files/` and the Dashboard quota bar goes up.
- [ ] WebSocket 在线协作测过：两个浏览器登录不同账号打开同一文档，A 打字 B 那边实时看到。
  - [ ] WebSocket online collaboration tested: two browsers logged into different accounts open the same doc and A's typing shows in real time on B's side.
- [ ] PWA 测过：手机 Chrome 打开 → 添加到主屏幕 → 关网 → 能打开已看过的文档 → 开网 → 顶栏"待同步 N 条"数字归零。
  - [ ] PWA tested: open in mobile Chrome → add to home screen → go offline → can open previously viewed docs → go online → the top bar's "N items pending sync" count returns to zero.

---

## 14. 第三轮部署补充 / v0.3 Deployment Notes

第二轮讲的是 systemd + 手动编译。第三轮加了 Docker 一键部署、S3 存储切换、备份子命令、健康探针、邮件/2FA 这些新东西，运维要知道怎么接。

Round 2 covered systemd + manual compilation. Round 3 adds one-click Docker deployment, S3 storage switching, a backup subcommand, health probes, and mail/2FA — ops needs to know how to wire these up.

### 14.1 Docker 一键部署 / One-Click Docker Deployment

项目根已经带了 `Dockerfile` 和 `docker-compose.yml`。Dockerfile 是多阶段：builder 阶段用 golang:1.26-bookworm 编译后端（CGO sqlite 要 gcc），runtime 阶段用 debian:bookworm-slim 装 libreoffice + fonts-noto-cjk。

The repo root already ships with `Dockerfile` and `docker-compose.yml`. The Dockerfile is multi-stage: the builder stage uses golang:1.26-bookworm to compile the backend (CGO sqlite needs gcc), and the runtime stage uses debian:bookworm-slim with libreoffice + fonts-noto-cjk.

**起服务**：

**Start the service**:

```bash
cd /home/user/Doubao/chats/1235147905824770/picooffice   # 即项目根，Dockerfile 所在目录
docker compose up -d --build
docker compose logs -f picooffice
```

起来后访问 `http://服务器IP:8080`。数据落在宿主机 `./data/db/`（sqlite）和 `./data/storage/`（上传文件/头像/bleve 索引），compose 里挂了 volume，容器删了数据还在。

After it's up, visit `http://server-IP:8080`. Data lands on the host in `./data/db/` (sqlite) and `./data/storage/` (uploaded files / avatars / bleve index); the compose file mounts a volume so data survives even if the container is removed.

**关键点**：

**Key points**:

- 镜像里已经装了 `fonts-noto-cjk` 和 `fonts-noto-cjk-extra`，导出 PDF/PPTX 中文不会变方块，不用再像第二轮那样手动 apt 装字体。
  - The image already installs `fonts-noto-cjk` and `fonts-noto-cjk-extra`, so Chinese in exported PDF/PPTX won't turn into boxes; no need to manually `apt install` fonts like in round 2.
- soffice 就在镜像里，导出 worker 直接调容器内 `/usr/bin/soffice`。
  - soffice is in the image; the export worker calls the in-container `/usr/bin/soffice` directly.
- 改前端要重新 build：前端产物已经 `//go:embed` 打进二进制了，改了前端先在宿主机 `cd frontend-web && npm run build`，把产物拷到 `backend/web/`，再 `docker compose up -d --build`。Dockerfile 里注释掉了一个可选的 node 构建阶段，要在镜像里编前端就放开那一段。
  - Changing the frontend requires a rebuild: the frontend artifact is already `//go:embed`ded into the binary. After editing the frontend, first run `cd frontend-web && npm run build` on the host, copy the artifact into `backend/web/`, then `docker compose up -d --build`. The Dockerfile has an optional node build stage commented out; uncomment it if you want to build the frontend inside the image.
- 别在容器里把 8080 直接暴露公网，前面还是挂 Nginx + HTTPS（见 13.5）。compose 里 `"8080:8080"` 改成 `"127.0.0.1:8080:8080"` 更稳。
  - Don't expose 8080 directly to the public internet from the container; still put Nginx + HTTPS in front (see 13.5). In compose, changing `"8080:8080"` to `"127.0.0.1:8080:8080"` is safer.

### 14.2 环境变量完整样例（含 S3）/ Complete Environment Variable Example (with S3)

`docker-compose.yml` 的 environment 段照下面填：

Fill the environment section of `docker-compose.yml` as below:

```yaml
environment:
  - PICO_PORT=8080
  - PICO_DB=/data/db/picooffice.db
  - PICO_STORAGE=/data/storage
  - PICO_EXPORT_WORKERS=2

  # 必须改！不然后果见 14.5
  - PICO_JWT_SECRET=换成openssl_rand_hex_32的输出

  # ---- S3/OSS 切换（不用 S3 就整段注释掉，默认 local）----
  - PICO_STORAGE_KIND=s3
  - PICO_S3_ENDPOINT=oss-cn-hangzhou.aliyuncs.com
  - PICO_S3_BUCKET=picooffice
  - PICO_S3_ACCESS_KEY=LTAIxxxxxxxx
  - PICO_S3_SECRET_KEY=yyyyyyyy
  - PICO_S3_REGION=oss-cn-hangzhou
  - PICO_S3_SSL=true
```

切 S3 前先把老本地数据同步上去：

Before switching to S3, first sync the old local data up:

```bash
# 用 rclone 或 mc，把老 storage/files、storage/avatars 整个桶 mirror 上去
rclone copy /path/to/storage/files/  myoss:picooffice/files/
rclone copy /path/to/storage/avatars/ myoss:picooffice/avatars/
```

同步完改 `PICO_STORAGE_KIND=s3` 重启。注意 bleve 索引（`storage/bleve/`）不用上 S3，留在本地盘就行，它可以重建（启动后访问一次 `/api/search?q=x&rebuild=1`）。

After syncing, change `PICO_STORAGE_KIND=s3` and restart. Note that the bleve index (`storage/bleve/`) need not go to S3; keep it on local disk — it can be rebuilt (after startup, hit `/api/search?q=x&rebuild=1` once).

### 14.3 备份：picobackup 子命令 + crontab / Backup: picobackup Subcommand + crontab

第三轮把备份做成了独立子命令 `picobackup`（Dockerfile 里一起编出来了，宿主机编译也在 `backend/cmd/picobackup/`）。原理：sqlite `VACUUM INTO` 出一个干净库，连同 storage 目录一起打 tar.gz。

Round 3 turned backup into a standalone subcommand `picobackup` (built together in the Dockerfile; on the host it's compiled from `backend/cmd/picobackup/`). How it works: SQLite `VACUUM INTO` produces a clean database, then the storage directory is packaged together into a tar.gz.

**手动跑一次**：

**Run it manually once**:

```bash
# 容器里
docker compose exec picooffice /app/picobackup -db /data/db/picooffice.db -storage /data/storage -out /data/backup/

# 宿主机直接编译跑
./picobackup -db ./picooffice.db -storage ./storage -out ./backup/
```

产物类似 `/data/backup/picobackup-2026-10-07_0900.tar.gz`。

The artifact looks like `/data/backup/picobackup-2026-10-07_0900.tar.gz`.

**配 crontab 每天凌晨 3 点备一份**：

**Configure crontab to back up daily at 3 AM**:

```bash
crontab -e
```

加一行（Docker 部署用 exec 方式）：

Add a line (Docker deployment uses the exec approach):

```
0 3 * * * cd /home/user/Doubao/chats/1235147905824770/picooffice && docker compose exec -T picooffice /app/picobackup -db /data/db/picooffice.db -storage /data/storage -out /data/backup/ >> /var/log/picobackup.log 2>&1
```

再配个清理：保留最近 14 天：

Also configure cleanup: keep the last 14 days:

```
0 4 * * * find /home/user/Doubao/chats/1235147905824770/picooffice/data/backup/ -name 'picobackup-*.tar.gz' -mtime +14 -delete
```

**注意**：备份包目前仍落在服务器本地磁盘（`data/backup/`）。如果你把主存储切到了 S3，记得再挂个 `rclone copy` 把 tar.gz 也推一份到对象存储或另一台机器，别只在本机——本机盘坏了备份一起没。

**Note**: the backup package still lands on the server's local disk (`data/backup/`). If you switched the main storage to S3, remember to add an `rclone copy` to also push the tar.gz to object storage or another machine — don't keep it only on the local host, or a dead disk takes the backup with it.

### 14.4 邮件账号接入参数 / Mail Account Connection Parameters

用户在前端"邮件"页加账号时要填 IMAP/SMTP。常见厂商端口：

When users add an account on the frontend 'Mail' page, they must fill in IMAP/SMTP. Common providers' ports:

| 厂商 / Provider | IMAP 服务器 / IMAP Server | IMAP SSL 端口 / IMAP SSL Port | SMTP 服务器 / SMTP Server | SMTP SSL/STARTTLS 端口 / SMTP SSL/STARTTLS Port |
|---|---|---|---|---|
| QQ 邮箱 / QQ Mail | imap.qq.com | 993 | smtp.qq.com | 465（SSL）/ 587（STARTTLS） |
| 163 网易 / 163 NetEase | imap.163.com | 993 | smtp.163.com | 465 / 25（加密用 465） |
| Gmail | imap.gmail.com | 993 | smtp.gmail.com | 465 / 587 |
| 企业微信/腾讯企业邮 / WeCom / Tencent Enterprise Mail | imap.exmail.qq.com | 993 | smtp.exmail.qq.com | 465 / 587 |
| 阿里企业邮 / Alibaba Enterprise Mail | imap.mxhichina.com | 993 | smtp.mxhichina.com | 465 / 587 |
| 自建 Postfix/Dovecot / Self-hosted Postfix/Dovecot | 自己的域名 / your own domain | 993 | 自己的域名 / your own domain | 465 / 587 |

**密码不是登录密码**：

**The password is not the login password**:

- QQ/163/Gmail 都要去邮箱设置里开"IMAP/SMTP 服务"，拿到**授权码/应用专用密码**，填那个，不是网页登录密码。
  - QQ/163/Gmail all require you to enable "IMAP/SMTP service" in mail settings and obtain an **authorization code / app-specific password**; enter that, not the web login password.
- Gmail 要开两步验证后建应用密码。
  - Gmail requires two-step verification before you can create an app password.
- 密码在 PicoOffice 里是可逆混淆存的（见 DETAILS 15.10），别填完就当它是银行级加密。
  - The password is stored reversibly obfuscated in PicoOffice (see DETAILS 15.10); don't treat it as bank-grade encryption once filled in.

服务器出站要能连到上面这些 993/465/587 端口。很多云厂商默认封 25 端口防垃圾邮件，发邮件走 465 或 587。

The server's outbound must be able to reach those 993/465/587 ports. Many cloud providers block port 25 by default to prevent spam, so send mail over 465 or 587.

### 14.5 refresh token / 2FA 部署注意 / refresh token / 2FA Deployment Notes

- **`PICO_JWT_SECRET` 必须改成强随机**。生成：
  - **`PICO_JWT_SECRET` must be changed to a strong random value**. Generate it:

  ```bash
  openssl rand -hex 32
  ```

  把输出填到 compose 的 `PICO_JWT_SECRET`。为什么现在特别强调：第三轮加了 refresh token 轮转，JWT secret 一旦泄露，攻击者能伪造任意用户的 access + refresh，直接登进任何账号。第二轮泄露最多是 access 2h 内有效，第三轮 refresh 30 天，泄露窗口长得多。
  - Then fill the output into the compose `PICO_JWT_SECRET`. Why emphasize this now: round 3 added refresh token rotation; once the JWT secret leaks, an attacker can forge any user's access + refresh and log into any account directly. In round 2 a leak only exposed access valid for up to 2h; in round 3 refresh lasts 30 days, so the exposure window is far longer.
- **改 JWT secret = 所有人被踢下线**。老 access/refresh 全作废，用户重新登录。所以上线前就定好，别上线后随便换。要换选低峰期，发通知。
  - **Changing the JWT secret = everyone logged out**. All old access/refresh are voided and users must log in again. So decide it before going live; don't change it casually afterward. If you must change, pick off-peak hours and send a notice.
- **2FA 是 TOTP**（Google Authenticator / 微软 Authenticator / 豆包认证器都行）。用户开通时扫 otpauth:// 二维码或手输 secret。服务器时间要准，TOTP 靠本机时间算 30s 窗口，时间差超过 30s 就验证失败：
  - **2FA is TOTP** (Google Authenticator / Microsoft Authenticator / Doubao Authenticator all work). When enabling, users scan the otpauth:// QR code or type the secret manually. The server clock must be accurate; TOTP computes a 30s window from the local time, and a drift over 30s causes verification to fail:

  ```bash
  timedatectl status   # 看 System clock sync 是不是 yes
  ```
- 用户丢了 2FA 密钥：admin 在后台用户列表里点"重置 2FA"，把该用户 `totp_enabled` 置回 false，他下次登录重新绑。
  - If a user loses the 2FA key: the admin clicks "Reset 2FA" in the backend user list, setting that user's `totp_enabled` back to false, and they re-bind on next login.
- 邮件找回密码：第三轮还没接真发信，reset 链接打在后端日志里。**部署时 `docker compose logs` 或 stdout 要能看到**，否则用户点"忘记密码"你找不到链接给他。要接真邮件 SMTP 才能自动发。
  - Mail-based password recovery: round 3 hasn't wired up real sending yet; the reset link is printed in the backend log. **Make sure `docker compose logs` or stdout is visible at deployment**, otherwise when a user clicks "forgot password" you can't find the link to give them. You need a real mail SMTP to send automatically.

### 14.6 健康探针接负载均衡 / 监控 / Health Probes for Load Balancer / Monitoring

第三轮加了两个端点：

Round 3 added two endpoints:

- `GET /healthz`：liveness，进程活着就 200。K8s livenessProbe 打它，挂了就重启容器。
  - `GET /healthz`: liveness — returns 200 as long as the process is alive. K8s livenessProbe hits it and restarts the container if it's down.
- `GET /readyz`：readiness，依次查 db 能不能 ping、`soffice` 在不在 PATH、磁盘还有没有空间。任何一项失败返回 503。
  - `GET /readyz`: readiness — checks in turn whether the db pings, `soffice` is on PATH, and disk has space. Any failure returns 503.

**怎么用**：

**How to use**:

- Docker /compose 起的话，compose 里可以加 healthcheck：
  - If started via Docker / compose, you can add a healthcheck to compose:

  ```yaml
  healthcheck:
    test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8080/readyz"]
    interval: 30s
    timeout: 5s
    retries: 3
  ```
- 前面挂 Nginx / 云负载均衡时，把被动健康检查指向 `/readyz`。soffice 崩了或磁盘满了，负载均衡自动把这个实例摘下来，别让用户打到一个导出会 500 的节点。
  - When Nginx / a cloud load balancer is in front, point passive health checks at `/readyz`. If soffice crashes or the disk fills, the load balancer automatically removes that instance, so users don't hit a node that returns 500 on export.
- 监控告警也打 `/readyz`：返回 503 就告警。`/healthz` 别拿来告警，它不查依赖，查了也白查。
  - Monitoring alerts should also hit `/readyz`: alert on a 503. Don't alert on `/healthz` — it doesn't check dependencies, so checking it is pointless.

容器里默认没装 curl，healthcheck 改用 wget 或直接在镜像里加 `apt install curl`，或者用后端自己加一个 TCP 探针端口。

Containers don't have curl by default; change the healthcheck to use wget, add `apt install curl` in the image, or have the backend expose a TCP probe port.

### 14.7 员工注册与 2FA 开通流程 / Employee Registration and 2FA Setup Flow

**admin 侧**：

**Admin side**:

1. 部署好后第一个注册的用户自动是 admin（见 11.5）。admin 登录后在"管理 → 用户"里建员工账号，或开注册让员工自己注册。
  - After deployment the first registered user automatically becomes admin (see 11.5). The admin logs in and creates employee accounts under "Admin → Users", or opens registration for employees to self-register.
2. 注册页强制勾选"已阅读并同意隐私政策与服务条款"，不勾注册按钮是灰的。
  - The registration page forces checking "have read and agree to the Privacy Policy and Terms of Service"; the register button stays grey until checked.
3. admin 在用户列表给员工分配配额（见第二轮 Dashboard）。
  - The admin assigns quotas to employees in the user list (see the round 2 Dashboard).

**员工侧**：

**Employee side**:

1. 打开站点注册/登录。
  - Open the site to register / log in.
2. 右上角头像 → 个人资料，改昵称、传头像。
  - Top-right avatar → profile; change nickname, upload avatar.
3. 头像 → "两步验证 2FA" → 点开通 → 弹出二维码 + 一串 secret。
  - Avatar → "Two-step verification 2FA" → click enable → a QR code + a secret string pops up.
4. 用认证器 App 扫二维码（扫不了就手输 secret），App 出 6 位数字。
  - Use an authenticator app to scan the QR (or type the secret if scanning fails); the app shows a 6-digit number.
5. 把 6 位数字填回网页 → 开通成功。下次登录要输密码 + 6 位码。
  - Enter the 6-digit number back into the page → enablement succeeds. Next login requires password + 6-digit code.
6. 头像 → "登录设备"能看到自己在哪些设备登着，不认识的设备直接吊销。
  - Avatar → "Login devices" shows which devices you're logged into; revoke any unrecognized device directly.

**建议**：admin 自己务必先开 2FA，再要求全员开。员工丢密钥走 admin 重置，别把"全局关闭 2FA"当后门。

**Suggestion**: The admin must enable 2FA first, then require everyone to enable it. For lost employee keys, use admin reset; don't treat "globally disable 2FA" as a backdoor.

### 14.8 第三轮部署检查清单（追加到 13.6 后面过一遍）/ v0.3 Deployment Checklist (review after 13.6)

- [ ] `docker compose up -d --build` 起来，`/readyz` 返回 200（容器里 curl 或宿主机 `curl localhost:8080/readyz`）。
  - [ ] `docker compose up -d --build` is up and `/readyz` returns 200 (curl inside the container or `curl localhost:8080/readyz` on the host).
- [ ] 容器里 `which soffice` 有输出，`fc-list :lang=zh | head` 能看到 Noto CJK 字体。
  - [ ] Inside the container `which soffice` has output and `fc-list :lang=zh | head` shows the Noto CJK fonts.
- [ ] `PICO_JWT_SECRET` 是 `openssl rand -hex 32` 的输出，不是默认值。
  - [ ] `PICO_JWT_SECRET` is the output of `openssl rand -hex 32`, not the default value.
- [ ] 手动跑一次 picobackup，`data/backup/` 下有 tar.gz，解压能看到 db 文件。
  - [ ] Ran picobackup once manually; `data/backup/` has a tar.gz and extracting it shows the db file.
- [ ] crontab 里有备份条目，`crontab -l` 能看到。
  - [ ] The crontab has a backup entry visible via `crontab -l`.
- [ ] 加一个测试邮件账号，发一封给自己，收件箱能收到。
  - [ ] Added a test mail account, sent a message to yourself, and it arrived in the inbox.
- [ ] 两个浏览器登录同一账号，A 打字 B 实时看到（OT 协作），断网再连不丢字。
  - [ ] Two browsers logged into the same account: A's typing shows on B in real time (OT collaboration), and reconnecting after going offline loses no characters.
- [ ] S3（如启用）：传一个附件，去 S3 控制台桶里能看到对应 key。
  - [ ] S3 (if enabled): uploaded an attachment and can see the corresponding key in the S3 console bucket.
- [ ] 2FA：admin 自己开通一次，关掉浏览器重开要输 6 位码。
  - [ ] 2FA: the admin enabled it once, and reopening the browser requires the 6-digit code.
- [ ] 手机 Safari 打开，工具栏按钮够大能点（触摸优化版）。
  - [ ] Opened in mobile Safari; the toolbar buttons are large enough to tap (touch-optimized).

### 14.9 第三轮新出的坑 / New v0.3 Pitfalls

**导入 office 文件报错 / 转出来是空白**：

**Import of office files errors / output is blank**:

- 先看容器里 soffice 能不能跑：`docker compose exec picooffice soffice --version`。空镜像或 soffice 崩了就重启容器。
  - First check whether soffice runs in the container: `docker compose exec picooffice soffice --version`. If the image is empty or soffice crashed, restart the container.
- 老版本 `.doc/.xls/.ppt`（97-2003）要靠 soffice 转，文件太老（95 以前）可能转失败，让用户另存成 docx 再传。
  - Old `.doc/.xls/.ppt` (97-2003) rely on soffice to convert; files that are too old (pre-95) may fail, so ask the user to re-save as docx before uploading.
- 转出来中文乱码 = 字体没装，确认镜像里有 fonts-noto-cjk（14.1 已带）。
  - Chinese garbled in output = fonts missing; confirm the image has fonts-noto-cjk (already included in 14.1).

**分享链接别人打不开**：

**Others can't open the share link**:

- 公开链接走 `/api/share/:token`，不要登录。如果前面 Nginx 配了全站 basic auth 或 SSO，要把 `/api/share/` 路径放行，不然匿名访问被拦在登录页。
  - Public links go through `/api/share/:token` without login. If Nginx in front has site-wide basic auth or SSO, you must allow the `/api/share/` path, otherwise anonymous access is blocked at the login page.
- 容器部署确认端口映射和反代没把 `/api/share/` 吃掉。
  - For container deployments, confirm port mapping and the reverse proxy don't swallow `/api/share/`.

**OT 协作时双方都看到"冲突"红字**：

**Both sides see "conflict" in red during OT collaboration**:

- 多半是时钟不同步或 WebSocket 被中间层断了。先看 `docker compose logs` 有没有 WS 断开重连。
  - Usually it's clock desync or the WebSocket being severed by a middle layer. First check `docker compose logs` for WS disconnect/reconnect.
- Nginx 反代记得配 `proxy_read_timeout 3600s`（13.5 已写），短了 WS 心跳被掐，step 推送不出去。
  - Remember to set `proxy_read_timeout 3600s` in the Nginx reverse proxy (written in 13.5); too short kills the WS heartbeat and step pushes can't get through.

**搜索搜不到刚改的文档**：

**Search can't find a just-edited document**:

- 写操作会自动更新索引，但量大时有几百 ms 延迟。等一下再搜。
  - Writes auto-update the index, but under heavy load there's a few-hundred-ms delay. Wait a moment before searching.
- 索引坏了（bleve 文件锁/磁盘写满导致损坏），跑一次 `GET /api/search?q=x&rebuild=1` 全量重建。
  - If the index is corrupted (bleve file lock / disk full), run `GET /api/search?q=x&rebuild=1` once for a full rebuild.
- 中文搜不到词但英文能搜 = cjk bigram 对超短词（1 个字）效果差，搜两个字以上。
  - Chinese words can't be found but English can = cjk bigram works poorly on ultra-short terms (1 character); search for two or more characters.

**邮件收不到 / 发不出**：

**Mail not received / not sent**:

- 993/465/587 出站被云厂商拦了。在服务器上 `telnet imap.qq.com 993` 测，连不通找云厂商开出站。
  - Outbound 993/465/587 is blocked by the cloud provider. Test on the server with `telnet imap.qq.com 993`; if it can't connect, ask the provider to open outbound.
- 密码填的是登录密码而不是授权码 → IMAP 报 LOGIN rejected。去邮箱设置拿授权码。
  - The password entered is the login password instead of the authorization code → IMAP reports LOGIN rejected. Go to mail settings to get the authorization code.
- 发件进了垃圾箱：PicoOffice 发信没配 DKIM/SPF，收件方可能判垃圾。正式用要在你域名 DNS 上加 SPF/TXT 记录。
  - Mail lands in spam: PicoOffice sending isn't configured with DKIM/SPF, so the receiver may flag it. For production, add SPF/TXT records to your domain's DNS.

**2FA 一直校验失败**：

**2FA verification keeps failing**:

- 服务器时间不对。`timedatectl` 校时，容器里时间跟宿主机走。
  - The server clock is wrong. Use `timedatectl` to correct it; inside a container the time follows the host.
- 用户手机时间不对。让他手机开自动对时。
  - The user's phone clock is wrong. Tell them to enable automatic time sync.

**Docker 磁盘越用越大**：

**Docker disk keeps growing**:

- sqlite WAL + 上传文件 + bleve 索引 + 导出临时文件都会涨。定期 `docker system prune`（别 prune volume！），跑 picobackup 后 `VACUUM` 过的库不会再膨胀。
  - sqlite WAL + uploaded files + bleve index + export temp files all grow. Periodically run `docker system prune` (don't prune volumes!), and a `VACUUM`ed database after picobackup won't bloat again.
- 导出 worker 的临时 soffice profile 在容器 `/tmp`，容器重启即清，不用管。
  - The export worker's temporary soffice profile is in the container's `/tmp` and cleared on restart, so no action needed.

### 14.10 从第二轮升级到第三轮 / Upgrade from v0.2 to v0.3

老用户（第二轮 systemd 部署）升级：

For existing users (round 2 systemd deployment) upgrading:

```bash
git pull
cd backend && go build -o ../picooffice . && go build -o ../picobackup ./cmd/picobackup
cd ../frontend-web && npm ci && npm run build
# 重启 systemd
sudo systemctl restart picooffice
```

- 启动时 GORM AutoMigrate 自动建新表（folders/tags/collab_steps/mail_accounts/...）和新列，老数据不动。
  - At startup GORM AutoMigrate automatically creates new tables (folders/tags/collab_steps/mail_accounts/...) and new columns; existing data is untouched.
- 第一次起来后访问一次 `GET /api/search?q=x&rebuild=1`，把老文档灌进 bleve 索引，不然搜索是空的。
  - After the first startup, hit `GET /api/search?q=x&rebuild=1` once to load old documents into the bleve index; otherwise search is empty.
- 老的本地 `storage/` 目录结构没变，直接沿用。要换 S3 按 14.2 切。
  - The old local `storage/` directory structure is unchanged and carries over directly. To switch to S3 follow 14.2.
- 老用户的 JWT secret 没动的话不用重新登录；但第三轮新加了 refresh，老 access 过期后会自动走 refresh，无感。
  - If the old users' JWT secret wasn't changed, no re-login is needed; but round 3 adds refresh, so after old access expires it auto-uses refresh seamlessly.
- 想切 Docker 部署：先按 14.3 跑一次 picobackup，把 tar.gz 解开的 db 和 storage 拷进新 compose 的 `./data/`，再 `docker compose up -d`。
  - To switch to Docker deployment: first run picobackup per 14.3, extract the tar.gz, copy its db and storage into the new compose's `./data/`, then `docker compose up -d`.
