# PicoOffice 移动端壳（Capacitor）

## 目录结构
- `www/` — 前端静态文件，从 `frontend-web/dist/` 拷贝
- `capacitor.config.json` — Capacitor 配置
- `android/` — 安卓原生工程（已生成，可直接打开 Android Studio）

## 服务器地址
`capacitor.config.json` 里写死了 `server.url = http://10.0.2.2:8080`（10.0.2.2 = 安卓模拟器宿主机）。
真机调试改成你电脑局域网 IP，比如 `http://192.168.1.10:8080`，然后 `npx cap sync android` 重新同步。

上线正式包时把 `server.url` 改成正式域名（https），并把 `cleartext` 关掉。

## 本机出 APK（Linux / macOS）

### 依赖
- JDK 21（Capacitor 8 / AGP 8.13 要求）
- Android SDK：platform-tools、platforms;android-36、build-tools;36.0.0

### 一键命令
```bash
export JAVA_HOME=/path/to/jdk-21
export ANDROID_HOME=/path/to/android-sdk
cd android
./gradlew assembleDebug
```
产物：`android/app/build/outputs/apk/debug/app-debug.apk`

出 release 包：
```bash
./gradlew assembleRelease
```
release 包需要签名，在 `android/app/build.gradle` 的 `signingConfigs` 里配 keystore。

### 更新前端后重新打包
```bash
cp -r ../frontend-web/dist/* www/
npx cap sync android
cd android && ./gradlew assembleDebug
```

## iOS（必须在 Mac 上做）

沙箱是 Linux，出不了 ipa。在 Mac 上：

```bash
# 1. 装 Xcode（App Store）
# 2. 装 CocoaPods
sudo gem install cocoapods

# 3. 初始化 iOS 工程
npx cap add ios

# 4. 同步
npx cap sync ios

# 5. 用 Xcode 打开工程，配签名
open ios/App/App.xcodeproj
```

在 Xcode 里：
- 选 Team（开发者账号）
- Bundle ID 已经是 `com.picooffice.app`
- 改版本号
- Product → Archive → 上传 App Store / Ad Hoc 出 ipa

命令行出 ipa（已有证书）：
```bash
cd ios/App
xcodebuild -workspace App.xcworkspace -configuration Release -archivePath build/App.xcarchive archive
xcodebuild -exportArchive -archivePath build/App.xcarchive -exportOptionsPlist ExportOptions.plist -exportPath build/
```

## 注意
- 沙箱里已经预出过一个 debug apk：`android/app/build/outputs/apk/debug/app-debug.apk`
- 安卓 9+ 默认禁明文 HTTP，`capacitor.config.json` 里 `cleartext: true` 已打开；正式环境用 https 后关掉
