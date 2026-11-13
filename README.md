# <img src="resources/icon.png" width=25> 流记 · FlowLog

一个用 [MyGo](https://mygo.run/zh) v0.9.0 **原生模式**（非 WebView、无前端构建）写的桌面日记应用：
把每天的学习记成「科目 → 章节 → 要点」，顺手记几条待办。

![首页](docs/screenshots/home.png)

## 它是什么

- **首页**：每天的寄语、连续记录天数、今天学了什么、今天要做什么、最近 7 天。
- **日记**：按日期倒序的卡片流。右上角「＋ 写日记」弹出模态框；
  一条记录 = 一个科目 + 一条章节路径（`第 4 章 >> 第 8 节`）+ 若干要点。
  要点可以勾「已复习」，支持搜索（科目/章节/要点/备注/待办）。
- **待办**：输入回车即添加、打勾、删除（可撤销）、按「全部 / 进行中 / 已完成」筛选、
  进度条、一键清除已完成。
- **关于**：版本、作者、仓库、许可证、数据目录（可复制路径）、快捷键。

界面是 MyGo 自己画的（Windows 用 Direct3D 11），没有 HTML、没有 JS、没有 node_modules。

## 运行

需要先装好 MyGo 并用 `mygo doctor` 检查环境。

```bash
go tool mygo dev      # 开发：编译 + 启动 + 改 .go 自动重启
go tool mygo build    # 打包，产物在 build/
go test ./...         # 测试
```

## 数据放在哪

```
data/
├── 2026-11-13.json      # 一天一个文件，文件名就是日期
└── settings.json        # 外观：跟随系统 / 浅色 / 深色 + 强调色
```

具体位置随运行方式变（「关于」页永远显示当前生效的绝对路径，可一键复制）：

| 运行方式 | 数据目录 |
| --- | --- |
| `go tool mygo dev` / `go run .` | 项目里的 `./data` |
| `go tool mygo build` 打出的安装包 | Windows 在 `%APPDATA%\FlowLog\data`，macOS 在 `~/Library/Application Support/FlowLog/data`，Linux 在 `~/.config/FlowLog/data` |
| 设了 `FLOWLOG_DATA` 环境变量 | 用你指定的目录（例如放到网盘同步） |

一天的文件长这样（人可读，可以直接丢进 Git 或网盘）：

```json
{
  "date": "2026-11-13",
  "entries": [
    {
      "id": "33cb09671503581b",
      "subject": "计算机组成原理",
      "topic": "第 4 章 存储系统 >> 第 8 节 虚拟存储器",
      "points": [
        { "id": "93d4ac5be35e16b2", "text": "4.8.2 页式虚拟存储器 -（2）地址映射和页表" },
        { "id": "8f2c1a0b7d3e4f56", "text": "4.8.2 页式虚拟存储器 -（3）访问流程", "done": true }
      ],
      "minutes": 45,
      "at": "17:26"
    }
  ],
  "todos": [
    { "id": "1", "text": "复习 B 树的分裂与合并", "done": false, "at": "09:10" }
  ]
}
```

写入是原子的（临时文件 → fsync → rename），写到一半崩溃不会留下坏文件。
内置对早期版本 `{"diary": "一段文字"}` 格式的兼容：读取时会自动变成一条「随手记」记录。

## 快捷键

| 键 | 作用 |
| --- | --- |
| `Ctrl/⌘ + N` | 写今天的日记 |
| `Ctrl/⌘ + 1 / 2 / 3 / 4` | 首页 / 日记 / 待办 / 关于 |
| `Ctrl/⌘ + F` | 在日记里搜索 |
| `Ctrl/⌘ + Enter` | 保存正在写的日记 |
| `Esc` | 关闭弹窗 |

## 目录结构

```
main.go                    启动、窗口、应用菜单
internal/model/            数据结构（Day / Entry / Point / Todo）
internal/store/            JSON 读写、原子写、迁移、设置
internal/view/             界面：状态、路由、四个页面、弹窗、主题
internal/view/*_test.go    交互测试 + 把页面渲染成 PNG 的「肉眼测试」
resources/                 图标
docs/开发手记.md            开发思路、踩坑记录、视觉验证方法
```

## 想继续改的话

- 作者、仓库地址、许可证在 `internal/view/meta.go` 顶部，改那几行就行。
- `docs/开发手记.md` 是本项目的「说明书 + 经验帖」，包括架构决策、MyGo 的三个核心概念、
  六个真实的坑，以及怎么把界面渲染成图来检查。建议先读它再动代码。

## 说明

- 界面文案是中文；`DateInput`、`NumberInput` 等内置控件的按钮文案跟随系统语言
  （没有 `locales/` 目录时是英文）。想本地化见开发手记第 13 节。
- `e2e.go` 只在 `-tags e2e` 时编入，用来配合 `mygo-automation` 做真窗口截图与端到端操作；
  发布版不会包含它。
