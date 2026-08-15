<div align="center">

# DeepSeek Harness Desktop

**让 [DeepSeek Harness](https://github.com/deepseek-ai/deepseek-harness) 在 Windows 桌面开箱即用**

[![Release](https://img.shields.io/github/v/release/Salbt/my-dsh-desktop?label=release&color=4D6BFE)](https://github.com/Salbt/my-dsh-desktop/releases)
[![License](https://img.shields.io/github/license/Salbt/my-dsh-desktop)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](https://github.com/Salbt/my-dsh-desktop/releases)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8)](https://go.dev/)

</div>

---

## 设计理念

| 特点 | 说明 |
|---|---|
| 开箱即用 | 发布包内置 Node 24 与 Harness 本体，解压或安装后无需联网、无需任何依赖；首启自动完成准备，后续启动约 6 秒 |
| 轻量级 | 程序本体为 6.7MB 单文件 exe，复用系统 WebView2，无框架运行时；只做干净的外壳，Harness 能力原样呈现 |
| 一键更新 | 托盘"更新 Harness"拉取最新版并自动重启生效 |

## 下载

[Releases](https://github.com/Salbt/my-dsh-desktop/releases) 提供两种形态：

| 产物 | 适合 |
|---|---|
| `my-dsh-desktop-portable-v<版本>.zip`（约 117MB） | 免安装、绿色携带、数据自包含 |
| `my-dsh-desktop-setup-v<版本>.exe` | 标准安装：开始菜单、桌面快捷方式、命令行入口、卸载器 |

> 发布物未签名，SmartScreen 提示时选择"仍要运行"。

## 快速开始

1. 双击 `my-dsh-desktop.exe`
2. 在窗口中打开 **Settings → Models**，填入 DeepSeek API Key（保存即时生效）
3. 点击 **Choose workspace** 选择工作目录

之后就是原汁原味的 DeepSeek Harness：模型配置、插件、审批策略……所有文档均与官方一致。

## 命令行与插件

安装版会把 `dsh` 命令加入系统 `PATH`。安装完成后打开一个新的 PowerShell，即可执行：

```powershell
dsh --version
dsh plugin --profile web add <插件包名>
```

便携版不会修改系统环境变量，请从解压目录调用随包提供的命令入口：

```powershell
.\bin\dsh.cmd --version
.\bin\dsh.cmd plugin --profile web add <插件包名>
```

插件安装在当前发行形态的数据目录下，安装后重启 Desktop 即可加载。命令入口使用发布包内置的 Node.js 与 pnpm，不要求系统另行安装。

## 数据布局

```
便携版 — 全部在解压目录（整体删除即卸载）:
├── my-dsh-desktop.exe   # 程序
├── bin\dsh.cmd          # 命令行入口
├── runtime\             # 内置 Node 24 + Harness
├── home\                # 配置 / 插件 / 密钥
├── npm-cache\           # 更新用的私有缓存
├── logs\app.log         # 运行日志
└── config.json          # 应用配置

安装版 — 程序在 Program Files，数据在 %LOCALAPPDATA%\my-dsh-desktop\
```

## 配置

`config.json`（首次运行生成）：

```json
{
  "dsh_version": "0.1.0-rc.6",
  "registry": ""
}
```

- `dsh_version` — 内置运行时缺失时安装的 Harness 版本
- `registry` — npm 镜像源，留空用官方源（国内网络可填 `https://registry.npmmirror.com`）

## 构建

构建要求：Go 1.25+、gcc/g++（MinGW-w64，webview_go 为 CGo 项目）。`build.ps1` 会做预检并给出安装提示。

```powershell
# 开发构建（无内置运行时，首启自动下载）
powershell -ExecutionPolicy Bypass -File build.ps1

# 发布打包（便携 zip + NSIS 安装包，需 makensis）
powershell -ExecutionPolicy Bypass -File scripts\package.ps1 `
  -AppVersion 0.1.0 -NodeVersion v24.19.0 -DshVersion 0.1.0-rc.6 `
  -PnpmVersion 10.34.5 -RequireInstaller
```

推送 `v*` tag 触发 GitHub Actions 自动构建并发布 Release。

## 已知限制

- Harness 本身处于 developer preview，上游接口可能变动（本应用通过"更新 Harness"按钮跟进）
- 仅支持 Windows 10/11 x64
- 发布物未代码签名

## License

[MIT](LICENSE) · 打包的 `@deepseek-ai/dsh` 为 MIT © DeepSeek · 鲸鱼图标商标归 DeepSeek 所有
