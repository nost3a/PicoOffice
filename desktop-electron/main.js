const { app, BrowserWindow, ipcMain, Menu } = require('electron');
const path = require('path');
const fs = require('fs');

let win = null;

function configPath() {
  return path.join(app.getPath('userData'), 'server.json');
}

function readSavedUrl() {
  try {
    const obj = JSON.parse(fs.readFileSync(configPath(), 'utf-8'));
    if (obj && typeof obj.url === 'string') return obj.url;
  } catch (_) {}
  return 'http://localhost:8080';
}

function saveUrl(url) {
  try { fs.writeFileSync(configPath(), JSON.stringify({ url }), 'utf-8'); } catch (_) {}
}

function askServerUrl(defaultUrl, cb) {
  const dlg = new BrowserWindow({
    width: 480,
    height: 240,
    resizable: false,
    title: '服务器设置',
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false,
    },
  });

  const html = `<!doctype html><html><head><meta charset="utf-8"><style>
    body{font-family:system-ui,sans-serif;padding:24px;margin:0;}
    h3{margin:0 0 12px;font-weight:600;}
    input{width:100%;padding:8px;box-sizing:border-box;margin-bottom:12px;}
    button{padding:8px 20px;margin-right:8px;cursor:pointer;}
  </style></head><body>
    <h3>PicoOffice 服务器地址</h3>
    <input id="u" value="${defaultUrl}" />
    <button onclick="ok()">连接</button>
    <button onclick="offline()">离线打开本地页</button>
    <script>
      const { ipcRenderer } = require('electron');
      function ok(){ ipcRenderer.send('server-url', document.getElementById('u').value.trim()); }
      function offline(){ ipcRenderer.send('server-url', 'offline'); }
    </script>
  </body></html>`;

  dlg.loadURL('data:text/html;charset=utf-8,' + encodeURIComponent(html));

  const onMsg = (event, url) => {
    ipcMain.removeListener('server-url', onMsg);
    dlg.close();
    cb(url);
  };
  ipcMain.on('server-url', onMsg);
}

function createWindow(targetUrl) {
  win = new BrowserWindow({
    width: 1280,
    height: 800,
    title: 'PicoOffice',
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
    },
  });

  Menu.setApplicationMenu(Menu.buildFromTemplate([
    { role: 'fileMenu' },
    { role: 'editMenu' },
    { role: 'viewMenu' },
    { role: 'helpMenu' },
  ]));

  if (targetUrl === 'offline') {
    win.loadFile(path.join(__dirname, 'app-dist', 'index.html'));
  } else {
    saveUrl(targetUrl);
    win.loadURL(targetUrl);
  }
}

app.whenReady().then(() => {
  askServerUrl(readSavedUrl(), (url) => createWindow(url));

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow(readSavedUrl());
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
