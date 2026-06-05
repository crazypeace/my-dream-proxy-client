# my-dream-proxy-client
极简翻墙客户端(壳)

# 开发故事

https://zelikk.blogspot.com/2026/06/xray-reality-mdpc-my-dream-proxy-client.html

# my-dream-proxy-client 使用手册 (配合Xray内核)

<details >
    <summary>点击展开</summary>
  
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

</details>

# my-dream-proxy-client 使用手册 (配合Hysteria内核)

<details>
    <summary>点击展开</summary>

## 1. 下载

前往 [GitHub Releases](https://github.com/crazypeace/my-dream-proxy-client/releases) 页面，根据你的系统下载对应的 zip 包：

- `my-dream-proxy-client-windows-amd64.zip` — Windows 64位
- `my-dream-proxy-client-linux-amd64.zip` — Linux x86_64
- `my-dream-proxy-client-linux-arm64.zip` — Linux ARM64

## 2. 解压

将 zip 包解压到任意目录，例如：

```bash
unzip my-dream-proxy-client-linux-*.zip -d ~/mdpc
cd ~/mdpc
```

解压后应包含：

```
my-dream-proxy-client          # 主程序
mdpc-config-hy2.yaml.default   # 配置文件模板
web     
└── hy2                        # Hysteria 配置页面
    ├── common.css
    ├── common.js
    └── config.html
bin
└── hy2                       # Hysteria 内核目录
    └── put_hy2_bin_config_here
```

## 3. 下载 Hysteria 翻墙内核

前往 [Hysteria Releases](https://github.com/apernet/hysteria/releases) 下载对应平台的 Hysteria 二进制文件。

将 `hysteria`（Linux）或 `hysteria.exe`（Windows）放到 `bin/hy2/` 目录下

## 4 创建 MDPC 配置文件
```bash
cp mdpc-config-hy2.yaml.default mdpc-config.yaml
```

用文本编辑器打开 mdpc-config.yaml，内容如下：

```yaml
listen: "127.0.0.1"
port: "18180"
files-dir: "./bin/hy2/"
core-start: "bin/hy2/hysteria client -c bin/hy2/config.yaml"
core-test: ""
log: ""
```

字段说明：

- listen — 监听地址。127.0.0.1 仅本机访问；0.0.0.0 允许局域网/公网访问
- port — API 服务端口
- files-dir — Hysteria 配置文件存放目录，对应 bin/hy2/
- core-start — 启动 Hysteria 的命令，使用 hysteria client 模式加载 config.yaml
- core-test — Hysteria 没有检查配置文件是否合法的命令，留空
- log — MDPC 日志文件路径，留空则输出到终端

## 5. 启动 MDPC 后端

在解压目录下执行：

```bash
./my-dream-proxy-client
```

看到以下输出表示启动成功：

```
my-dream-proxy-client listening on 127.0.0.1:18180
```

> 18180 是 MDPC 的 API 管理端口，MDPC前端页面通过该端口与MDPC后端通信。保持此终端窗口开启，或使用 `nohup` / `systemd` 托管进程。

## 6. 打开 MDPC 前端配置页面


可以使用浏览器直接打开HTML文件 `file:///config.html`

也可以启动一个本地的HTTP服务, 如 `python -m http.server 8000` 再用浏览器访问 `http://127.0.0.1:8000/config.html`

<img  alt="image" src="https://github.com/user-attachments/assets/d52f54b9-6d6e-4d7f-a9b2-982f50b2223c" />


页面顶部显示 API 地址，确认是 `http://127.0.0.1:18180`。

右侧状态区显示 **"已停止"** 或 **"运行中 (pid xxxx)"**。

## 7. 配置 Hysteria2 节点

> 本项目只是为了简单演示基本原理， 所以只支持自签证书的 Hysteria2 节点

有两种方式填写节点信息：

### 方式 A：粘贴节点链接（推荐）

如果你的代理服务商提供了 `hysteria2://` 或 `hy2://` 分享链接：

1. 将完整链接粘贴到 **"节点链接解析"** 输入框。
2. 点击 **"解析"**。
3. 表单会自动填充：地址、端口、密码、证书指纹。

### 方式 B：手动填写

1. 展开 **"表单"** 区块。
2. 填写以下字段：

| 字段 | 说明 |
|------|------|
| 地址 | 节点服务器 IP 或域名 |
| 端口 | 服务器端口，默认 443 |
| 密码 | Hysteria2 认证密码 |
| 证书指纹 | TLS 证书的 `pinSHA256`，可选 |
| SOCKS 端口 | 本地 SOCKS5 监听端口，默认 10808 |
| HTTP 端口 | 本地 HTTP 监听端口，默认 10809 |


**关于证书指纹：**
- 页面中 `pinSHA256` 使用 **hex 格式**（64 位十六进制字符）。
- 部分工具输出的指纹是 **base64 格式**，可直接粘贴到表单，页面会自动转换为 hex。

## 8. 保存配置

填写完成后，点击 **"直接保存"**。

该操作会：
1. 将表单内容生成为 YAML 配置文件。
2. 写入 `bin/hy2/config.yaml`。
3. 控制台提示 **"保存成功"**。

你也可以点击 **"转YAML"** 先预览生成的配置内容，确认无误后再点 **"直接保存"**。

## 9. 启动 / 停止

- 点击 **"▶ 启动"** → Hysteria 内核进程启动，页面状态切换为 **"运行中"**。
- 点击 **"■ 停止"** → 内核进程终止，状态切换为 **"已停止"**。

启动后，本地代理服务已就绪：

- **SOCKS5 代理**：`127.0.0.1:10808`
- **HTTP 代理**：`127.0.0.1:10809`

在系统或浏览器中配置代理指向上述地址即可使用。

## 10. 读取已有配置

如需修改已保存的配置，点击 **"读取配置"**，页面会从 `bin/hy2/config.yaml` 加载当前配置并填入表单。

---

**端口一览：**

| 端口 | 用途 |
|------|------|
| 18180 | 主程序管理 API（不要占用） |
| 10808 | Hysteria2 SOCKS5 本地代理 |
| 10809 | Hysteria2 HTTP 本地代理 |


</details>

# my-dream-proxy-client 使用手册 (配合sing-box内核)

<details>
    <summary>点击展开</summary>


## 第一步：下载 my-dream-proxy-client(下称MDPC)

从[GitHub Releases](https://github.com/crazypeace/my-dream-proxy-client/releases) 页面下载对应架构的 zip 包：


- `my-dream-proxy-client-windows-amd64.zip` — Windows 64位
- `my-dream-proxy-client-linux-amd64.zip` — Linux x86_64
- `my-dream-proxy-client-linux-arm64.zip` — Linux ARM64

解压：

```bash
unzip my-dream-proxy-client-linux-*.zip -d mdpc
cd mdpc
```

解压后目录结构：

```
mdpc/
├── my-dream-proxy-client          # 主程序
├── mdpc-config-sing-box.yaml.default   # MDPC配置模板（sing-box）
├── bin/
│   └── sing-box/
│       └── put_sing-box_bin_config_here  # 占位符
└── web/
    └── sing-box/                   # sing-box 前端页面
        ├── 01-log.html
        ├── 02-dns.html
        ├── 03-route.html
        ├── 04-inbounds.html
        ├── 05-outbounds.html
        ├── common.css
        └── common.js
```

## 第二步：下载 sing-box

从 [sing-box 官方仓库](https://github.com/SagerNet/sing-box/releases)下载对应架构的二进制文件：

将 `sing-box`（Linux）或 `sing-box.exe`（Windows）放到 `bin/sing-box/` 目录下



## 第三步：创建 MDPC 配置文件
```bash
cp mdpc-config-singbox.yaml.default mdpc-config.yaml
```

用文本编辑器打开 mdpc-config.yaml，内容如下：

```yaml
listen: "127.0.0.1"
port: "18280"
files-dir: "bin/sing-box/"
core-start: "bin/sing-box/sing-box run -c bin/sing-box/config.json"
core-test: "bin/sing-box/sing-box check -c bin/sing-box/config.json"
log: ""
```


## 第四步：启动 MDPC 后端

```bash
./my-dream-proxy-client
```


看到以下输出表示启动成功：

```
my-dream-proxy-client listening on 127.0.0.1:18280
```

## 第五步：打开 MDPC 前端配置页面

可以使用浏览器直接打开HTML文件 `file:///05-outbounds.html`

也可以启动一个本地的HTTP服务, 如 `python -m http.server 8000` 再用浏览器访问 `http://127.0.0.1:8000/05-outbounds.html`

### 5.1 Inbound 配置

打开 `file:///04-inbounds.html`

点击 **「填充建议」** 按钮 → 点击 **「保存」**


### 5.2 Outbound 配置

打开 `file:///05-outbounds.html`

展开 **「🔗 AnyTLS 分享链接解析」** 面板，粘贴 anytls 链接，点击 **「解析并添加」** → 点击 **「保存」**



### 5.3 DNS 配置 (可选)

打开 `file:///02-dns.html`

点击 **「预设-1」** 按钮 → 点击 **「保存」**


### 5.4 Route 配置 (可选)

打开 `file:///03-route.html`

点击 **「GFW 黑名单」** 按钮 → 点击 **「保存」**


## 第六步：启动代理内核

1. 配置好 04-inbounds、05-outbounds (及你希望配置的内容)后
2. 点击页面顶部的 **▶ 启动** 按钮
3. 状态指示器变为 "运行中 (pid XXXX)" 即表示成功

默认inbounds是开启代理：
   - **SOCKS 代理：** `127.0.0.1:10808`
   - **HTTP 代理：** `127.0.0.1:10809`
    
</details>
