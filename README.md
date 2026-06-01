# my-dream-proxy-client
极简翻墙客户端(壳)
# my-dream-proxy-client 使用手册 (配合Xray内核)

## 一、简介

**my-dream-proxy-client** (下简称MDPC) 是一个极简翻墙客户端(壳)，用 Go 编写的后端提供API服务, 基于HTML文件的Web 管理界面，用于管理 Xray-core 代理核心的配置文件和进程。

---

## 二、下载与安装

### 2.1 下载

前往 [GitHub Releases](https://github.com/crazypeace/my-dream-proxy-client/releases) 页面，根据你的系统下载对应的 zip 包：

- `my-dream-proxy-client-windows-amd64.zip` — Windows 64位
- `my-dream-proxy-client-linux-amd64.zip` — Linux x86_64
- `my-dream-proxy-client-linux-arm64.zip` — Linux ARM64

### 2.2 解压

解压后目录结构如下：

```
my-dream-proxy-client(.exe)   ← 主程序
mdpc-config.yaml.default      ← 配置文件模板
bin/xray/put_xray_bin_config_here  ← 占位文件，提示你把 xray 的执行文件和配置文件放这里
web/xray/                     ← Web 界面文件
  ├── common.css
  ├── common.js
  ├── 01-log.html
  ├── 02-dns.html
  ├── 03-router.html
  ├── 04-inbounds.html
  ├── 05-outbounds.html
  └── 06-api.html
```

### 2.3 下载 Xray-core

前往 [Xray-core Releases](https://github.com/XTLS/Xray-core/releases) 下载对应平台的 Xray 二进制文件。

将 `xray`（Linux）或 `xray.exe`（Windows）放到 `bin/xray/` 目录下，同时将 `geoip.dat` 和 `geosite.dat` 也放进去。

### 2.4 创建MDPC配置文件

```bash
cp mdpc-config.yaml.default mdpc-config.yaml
```

用文本编辑器打开 `mdpc-config.yaml`，内容如下：

```yaml
listen: "127.0.0.1"
port: "18080"
files-dir: "./bin/core/"
core-start: "bin/xray/xray run -confdir bin/xray/"
core-test: "bin/xray/xray run -confdir bin/xray/ -test"
log: ""
```

**字段说明：**

- **listen** — 监听地址。`127.0.0.1` 仅本机访问；`0.0.0.0` 允许局域网/公网访问
- **port** — API 服务端口
- **files-dir** — Xray 配置 JSON 文件存放目录
- **core-start** — 启动 Xray 的命令
- **core-test** — 测试 Xray 配置的命令
- **log** — MDPC 日志文件路径，留空则输出到终端



---

## 三、启动程序

### Linux

```bash
chmod +x my-dream-proxy-client
./my-dream-proxy-client
```

### Windows

双击 `my-dream-proxy-client.exe`，或在命令行中运行：

```
my-dream-proxy-client.exe
```

启动后终端会显示类似：

```
Server starting on 127.0.0.1:18080
```

---

## 四、使用 Web 界面

### 4.1 打开Web界面

可以使用浏览器直接打开HTML文件 `file:///05-outbounds.html`

也可以启动一个本地的HTTP服务, 如 `python -m http.server 8000` 再用浏览器访问 `http://127.0.0.1:8000/05-outbounds.html`

### 4.2 界面布局

<img alt="image" src="https://github.com/user-attachments/assets/0671df3a-a5b6-42b8-a117-20b81ff811cb" />


每个页面顶部都有：

- **API URL 输入框** — 管理后端地址（默认 `http://127.0.0.1:18080`，可以保存在浏览器 localStorage 中）
- **▶ 启动** 按钮 — 启动 Xray
- **■ 停止** 按钮 — 停止 Xray
- **状态指示器** — 显示 "运行中 (pid XXXX)" 或 "已停止"
- **导航栏** — 6 个标签页，对应 6 个配置文件

### 4.3 配置出站代理（目前只实现Reality协议作为演示）

进入 **05-outbounds** 页面，有两种方式配置：

**方式一：粘贴分享链接**

1. 点击顶部 "分享链接解析" 展开
2. 粘贴 `vless://` 格式的分享链接
3. 点击 **解析**，表单会自动填充
4. 检查参数后点击 **直接保存**

**方式二：手动填写表单**

1. 展开 "表单" 区域
2. 填写地址、端口、UUID、传输协议等参数
3. 点击 **转JSON** 预览，确认后 **直接保存**
> 也可以直接点击 **直接保存**

**方式三：直接编辑 JSON**
1. 展开 "JSON" 区域
2. 在 JSON 文本框中直接编辑 JSON 内容，点 **保存**。

### 4.4 配置其他模块

| 页面 | 文件 | 说明 |
|------|------|------|
| 01-log | 01-log.json | 日志配置, 默认为空 |
| 02-dns | 02-dns.json | DNS 设置，有 **填充建议** 按钮可快速填入默认 DNS 8.8.8.8 和 1.1.1.1 |
| 03-router | 03-router.json | 路由规则, 默认为空 |
| 04-inbounds | 04-inbounds.json | 入站设置，有 **填充建议**（默认 SOCKS:10808 + HTTP:10809） |
| 05-outbounds | 05-outbounds.json | 出站代理服务器, 默认为空 |
| 06-api | 06-api.json | Xray API 设置, 默认为空 |

### 4.5 通用操作

- **读取配置** — 从服务端读取对应的 JSON 文件内容
- **保存** — 将编辑器内容保存到服务端。**如果编辑器为空则删除该文件**
- **填充建议** — 用预设内容填充编辑器（仅 02-dns 和 04-inbounds 有此功能）

---

## 五、启动代理内核

1. 配置好 04-inbounds、05-outbounds (及你希望配置的内容)后
2. 点击页面顶部的 **▶ 启动** 按钮
3. 状态指示器变为 "运行中 (pid XXXX)" 即表示成功

默认inbounds是开启代理：
   - **SOCKS 代理：** `127.0.0.1:10808`
   - **HTTP 代理：** `127.0.0.1:10809`
