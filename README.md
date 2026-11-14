# <img src="resources/icon.png" width=25> 流记 · FlowLog

一个记录每日学习情况的桌面应用：按科目、章节与要点组织笔记，附带当天待办。

开发者还只是个大学生，该项目仅作为本地小工具，如有不满，请见谅。

使用 DeepSeek Harness 配合 DeepSeek V4.1 Flash 模型开发。

![MyGo](https://img.shields.io/badge/MyGo-v0.9.0-c2410c)
![Go](https://img.shields.io/badge/Go-1.27-00add8)
![License](https://img.shields.io/badge/license-MIT-2f855a)
![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-444)
![Agent](https://img.shields.io/badge/Agent-DeepSeek_Harness-blue)

流记使用 [MyGo](https://mygo.run) 的原生 UI 模式编写：界面由 Go 代码描述，由 MyGo 在 GPU 上绘制，不加载 WebView，也没有 HTML、JavaScript 与前端构建步骤。

## 功能

| 模块   | 说明                                                                                              |
| ------ | ------------------------------------------------------------------------------------------------- |
| 首页   | 每日寄语（背景为必应每日一图）、连续记录天数、今日要点、**待办提醒**（今天与明天到期）、最近 7 天 |
| 日记   | 按日期倒序的记录卡片；一条记录包含科目、章、节、若干要点与备注，可标记要点为已复习；支持全文搜索  |
| 待办   | 全部待办列表；可为每条设置截止日期，按截止日期排序；支持勾选、删除（可撤销）、按状态筛选          |
| 侧边栏 | 今天记录了多少条、还有多少待办；月历中标出有日记的日子，点一天直接跳到那一天的日记                |
| 设置   | 称呼、外观（深色 / 浅色 / 跟随系统）、数据备份与恢复、快捷键、关于与版本更新                      |

日记与待办是两套独立的数据：写待办不会影响日记的统计，删除某天的日记也不会删掉待办。
数据保存在本地单个 SQLite 数据库中，可整体复制或通过内置的备份功能导出。

## 截图

| 首页 · 深色（默认）                | 首页 · 浅色                                  |
| ---------------------------------- | -------------------------------------------- |
| ![首页](docs/screenshots/home.png) | ![首页浅色](docs/screenshots/home-light.png) |

| 日记                                | 待办                               |
| ----------------------------------- | ---------------------------------- |
| ![日记](docs/screenshots/diary.png) | ![待办](docs/screenshots/todo.png) |

| 设置                                   | 写日记                                 |
| -------------------------------------- | -------------------------------------- |
| ![设置](docs/screenshots/settings.png) | ![写日记](docs/screenshots/editor.png) |

## 环境要求

- Go 1.27 或更高版本
- MyGo CLI：项目已在 `go.mod` 中以 `tool` 依赖固定版本，直接用 `go tool mygo` 调用，无需单独安装
- 构建安装包时需要 NSIS（Windows 安装器）等平台工具，详见
  [MyGo 分发文档](https://mygo.run/docs/distribution)

## 开发

```bash
go tool mygo dev      # 编译、启动应用，并在 Go 代码改动后自动重启
go test ./...         # 单元测试与界面交互测试
go vet ./...          # 静态检查
gofmt -l .            # 检查格式
```

界面预览会在测试中渲染成 PNG：

```bash
go test ./internal/view -run TestPreview -v
```

## 构建

```bash
go tool mygo build                 # 当前平台
go tool mygo build -platform windows/amd64,darwin/arm64
```

产物位于 `build/`。Windows 生成安装包，macOS 生成 `.app` 与 `.dmg`，Linux 生成 `.tar.gz` 与安装脚本。

## 数据

数据库文件为 `data/flowlog.db`（SQLite）。位置随运行方式变化：

| 运行方式                        | 数据目录                                                                                                             |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `go tool mygo dev` / `go run .` | 项目下的 `./data`                                                                                                    |
| 安装后的应用                    | Windows `%APPDATA%\FlowLog\data`，macOS `~/Library/Application Support/FlowLog/data`，Linux `~/.config/FlowLog/data` |
| 设置了 `FLOWLOG_DATA`           | 使用该目录（便于放到同步盘）                                                                                         |

表结构：

| 表         | 内容                                           |
| ---------- | ---------------------------------------------- |
| `days`     | 有日记的日期                                   |
| `entries`  | 学习记录，属于某一天                           |
| `points`   | 要点，属于某条记录                             |
| `todos`    | 待办：内容、完成状态、创建日期、可选的截止日期 |
| `settings` | 键值设置（外观、称呼）                         |

日记的写入以事务为单位：一次保存要么全部成功，要么全部回滚。待办逐条读写。

旧版本（0.3 及更早）的数据库会在启动时自动升级，已有数据保留。

**备份与迁移**

- **导出备份**：生成一个 JSON 文件，包含全部日记、待办与设置。
- **从备份恢复**：选择备份文件，覆盖当前全部数据（会二次确认）。

## 自动更新

应用使用 MyGo 的更新机制：更新包由 Ed25519 私钥签名，应用内置对应公钥，启动后自动检查、
下载、校验并替换文件，随后重启生效。

### 发布流程

1. 修改 `mygo.json` 中的 `version`，在 `CHANGELOG.md` 中补一节 `## <版本号>`。
2. 提交并打标签：`git tag v0.4.0 && git push origin v0.4.0`。
3. GitHub Actions 读取该版本的 CHANGELOG 作为发布说明，创建草稿 Release，
   并在三个平台并行构建、上传产物。
4. 全部成功后自动发布；已安装的应用会在下次检查时看到新版本。

工作流见 [.github/workflows/release.yml](.github/workflows/release.yml)。

### 一次性配置

在仓库的 **Settings → Secrets and variables → Actions** 中添加：

| Secret                     | 内容                             |
| -------------------------- | -------------------------------- |
| `MYGO_UPDATER_PRIVATE_KEY` | `mygo-update.key` 文件的完整内容 |

私钥由 `go tool mygo keygen` 生成，默认位于用户配置目录的 `mygo/update-keys/` 下。
私钥丢失将无法再向已安装的用户推送更新，请妥善备份；泄露则任何人都能伪造更新。

可选（正式签名，避免 macOS Gatekeeper 与 Windows SmartScreen 拦截）：
`MACOS_CERTIFICATE`、`MACOS_CERTIFICATE_PASSWORD`、`APPLE_ID`、`APPLE_TEAM_ID`、
`APPLE_APP_PASSWORD`、`WINDOWS_CERTIFICATE`、`WINDOWS_CERTIFICATE_PASSWORD`。

### 更新如何生效

- 已安装的应用每 24 小时检查一次，也可在设置页手动检查。
- 发现新版本后可在设置页直接下载安装，安装完成后应用自动重启。
- 只有用同一私钥签名的更新才会被接受。
- 开发构建（`mygo dev`、`go run`）与包管理器安装的版本不参与自动更新；
  用户也可以从 Releases 手动下载新版本。

## 项目结构

```
main.go                    启动、窗口、数据库打开与关闭、发布说明
internal/model/            数据结构（Day / Entry / Point / Todo / Settings）
internal/store/            SQLite 读写、schema 升级、备份导出与恢复
internal/view/             界面：状态、路由、四个页面、弹窗、主题、背景图、月历、更新
internal/view/*_test.go    交互测试与界面渲染测试
resources/                 应用图标（同时嵌入程序）
docs/                      开发手记与界面截图
.github/workflows/         三平台发布工作流
```

## 已知限制

- 界面文案为中文；内置控件（日历、数字输入）已随中文语言包显示。
- 自动更新依赖 GitHub Releases，需要仓库公开或使用可访问的 HTTPS 地址。
- 目前仅支持单窗口、本地单用户数据。

## 许可

[MIT](LICENSE)
