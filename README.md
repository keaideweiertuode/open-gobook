# 🪙 GoBook

> 终身级纯本地 AI 记账与财务资产复盘系统

[![Go Version](https://img.shields.io/github/go-mod/go-version?filename=go.mod)]()
[![SQLite](https://img.shields.io/badge/storage-SQLite-blue)]()
[![License](https://img.shields.io/badge/license-MIT-green)]()
![Platform](https://img.shields.io/badge/platform-linux%20|%20macOS-lightgrey)

GoBook 是一个面向 **20~30 年长周期运行设计** 的个人财务基础设施，彻底推翻传统云端记账软件的"隐私泄露、依赖腐化、服务关停"等长期痛点。

- **写入端** — 极致无感。通过 CLI 或配合 OpenClaw Agent 使用大白话完成模糊语义录入
- **展现端** — 多维穿透。基于原生 Web UI 提供日/月/年三级联动的收支看板、双轨对比图表及消费排行

---

## 设计哲学

### SQLite 物理留痕

数据资产完全托管在单文件 SQLite 中。官方向下兼容承诺长达数十年，备份仅需 `cp my_money.db`。字段中特设 `raw_text`，终身保留用户输入原文，供未来数十年随时洗数或训练个人财务模型。

### Go 100% 纯净全编译

使用 `modernc.org/sqlite` 纯 Go 驱动，**彻底移除 CGO 依赖**。实现跨平台单文件静态编译，绝不因操作系统 C 库变更或 `npm/pip` 依赖树腐化而崩溃。编译产物约 **15MB** 单文件二进制。

### 前端零构建、零死角

Web 视图层摒弃 Vite/Webpack 等构建工具，采用原生三剑客直连轻量 CDN（Vue 3 + Tailwind CSS + ECharts）。任意设备双击 `static/index.html` 即可无感渲染（需后端 API）。

### 年份无限制生长

年份选择器由数据库物理留痕的数据驱动（Data-Driven），通过 SQL 自动盘点账单边界并兜底当前年份，彻底解绑硬编码。

---

## 前提条件

- Go 1.25+（编译用，运行时只需要二进制）
- 浏览器（Chrome / Firefox / Edge 等现代浏览器）

---

## 快速开始

```bash
# 编译
go build -o bin/gobook ./cmd/gobook

# 直接运行
./bin/gobook              # Web 模式 → http://localhost:8080
./bin/gobook 吃饭 25      # CLI 记账模式

### 自动化测试保障

本项目内建了严格的单元与集成测试，覆盖了核心 API 及数据库逻辑，全程使用 `:memory:` 纯内存数据库，**无痕测试不脏账本**。

```bash
go test -v ./internal/...
```

### Systemd 服务（推荐，开机自启）

```bash
sudo cp gobook /usr/local/bin/gobook

# 创建 /etc/systemd/system/gobook.service
[Unit]
Description=GoBook - Local AI Bookkeeping Service
After=network.target

[Service]
Type=simple
User=ian
WorkingDirectory=/home/ian/vscode/Go/gobook
ExecStart=/home/ian/vscode/Go/gobook/bin/gobook
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target

# 启动并启用开机自启
sudo systemctl daemon-reload
sudo systemctl enable --now gobook

# 管理命令
sudo systemctl start|stop|restart gobook
journalctl -u gobook -f
```

---

## 使用示例

### 终端 CLI 记账

```bash
# 记录一笔支出
./bin/gobook 吃饭 25

# 记录一笔收入（自动识别为收入类目）
./bin/gobook 发工资 8000

# 记录带备注的支出
./bin/gobook 打车去机场 120
```

### Web 看板

启动服务后访问 `http://localhost:8080`，提供：

- **日视图** — 当日收支明细与时段分布
- **月视图** — 月度汇总与趋势图
- **年视图** — 年度总览与对比
- **流水明细** — 带 ID 的完整记录列表，支持删除操作

### HTTP API

| 方法 | 路由 | 说明 |
|------|------|------|
| `POST` | `/api/transactions` | 创建账单（body: amount, category, remark, raw_text） |
| `POST`/`DELETE` | `/api/transactions/delete?id=N` | 物理删除账单 (双路由支持) |
| `GET` | `/api/transactions/list?period=2026-05-27` | 流水明细列表 |
| `GET` | `/api/stats/daily?period=2026-05` | 双轨收支统计 |
| `GET` | `/api/stats/category?period=2026-05` | 支出分类排行 |
| `GET` | `/api/stats/years` | 可用年份选项 |
| `GET` | `/api/export?period=2026` | 导出全量/指定区间的财务明细 (CSV格式，带BOM防乱码) |

---

## 项目结构

```
gobook/
├── cmd/
│   └── gobook/
│       └── main.go         # 系统入口（CLI 记账 + HTTP 服务）
├── internal/
│   ├── api/
│   │   └── api.go          # HTTP 服务与接口路由
│   ├── db/
│   │   └── db.go           # SQLite 核心层（DDL、事务、查询）
│   └── models/
│       └── models.go       # 统一强类型数据模型
├── web/
│   └── static/
│       └── index.html      # 纯前端 UI（Vue 3 + Tailwind CSS + ECharts）
├── scripts/
│   ├── import_csv.py       # CSV 账单导入工具
│   ├── seed_2025.py        # 2025 全年测试数据生成器
│   └── auto_backup.sh      # Git 自动备份脚本
├── gemini.md               # 备忘文档
├── data/
│   ├── my_money.db         # SQLite 数据库（运行后自动生成）
│   ├── my_money.db-wal     # SQLite 预写式日志（保持超高读写性能的关键文件）
│   ├── my_money.db-shm     # SQLite 共享内存映射索引文件
│   └── *.csv               # CSV 数据文件
├── go.mod                  # Go 模块依赖
├── go.sum                  # 依赖校验和
├── bin/                    
│   └── gobook              # 编译产物（15MB 静态二进制）
└── .git/                   # Git 仓库
```

---

## 数据模型

### 物理表结构

| 表名 | 说明 |
|------|------|
| `accounts` | 资产账户（默认 id=1 默认账本，预留多账户扩展） |
| `categories` | 财务类目（支持 `expense` / `income` 类型） |
| `transactions` | 财务流水主表（含 `raw_text` 原始文本留痕，建有日期 + 分类复合索引） |

### 内置分类

**支出类目（23 个）：**

三餐 · 零食 · 衣服 · 交通 · 旅行 · 孩子 · 宠物 · 话费网费 · 烟酒 · 学习 · 日用品 · 住房 · 美妆 · 医疗 · 发红包 · 汽车 · 娱乐 · 请客送礼 · 电器数码 · 运动 · 其它 · 水电煤 · 办公

**收入类目（5 个）：**

工资 · 生活费 · 收红包 · 股票基金 · 其他

---

## 工具脚本

### CSV 导入（`import_csv.py`）

从其他记账 App（如"钱迹"）导出的 CSV 文件导入历史数据。自动匹配分类，未知分类兜底到"其它/其他"并保留原分类名在备注中。

```bash
python3 scripts/import_csv.py
# 默认读取 data 目录下的 Qian_Ji_*.csv 文件
```

### 测试数据生成（`seed_2025.py`）

为 Web 看板注入 2025 全年模拟消费数据，方便调试和体验完整功能。

```bash
python3 scripts/seed_2025.py
```

### Git 自动备份（`auto_backup.sh`）

配合 crontab 实现每日自动备份到 GitHub：

```bash
crontab -e
# 每天 23:00 自动备份
0 23 * * * /home/ian/vscode/Go/gobook/scripts/auto_backup.sh
```

---

## 集成：OpenClaw Agent

GoBook 配有完整的 OpenClaw Skill（位于项目根目录的 `openclaw_skill/` 文件夹），安装后可通过自然语言记账：

| 场景 | 用户说 | Agent 行为 |
|------|--------|-----------|
| 支出 | "刚才打车花了 32.5" | 调用 `record_financial_transaction` |
| 收入 | "股票基金收益到账 1500" | 调用 `record_financial_transaction` |
| 删除 | "把 #15 那笔删了" | 调用 `delete_financial_transaction_by_id` |
| 查账 | "这个月花了多少" | 调用 `query_financial_summary` |
| 明细 | "看看今天的流水" | 调用 `list_recent_transactions` |
| 导出 | "把本月账单导出成 CSV" | 调用 `export_financial_data` |

**安装方法：**
将本项目中的 `openclaw_skill/` 目录复制到您本机的 OpenClaw Skills 目录（通常为 `~/.openclaw/workspace/skills/gobook`）下即可。

- `bookkeeping.py` — 核心工具函数（记账、删除、查账汇总、查明细、导出）
- `SKILL.md` — 完整触发规则与使用文档

---

## 安全与异地灾备

本系统为 **100% 纯本地私有化部署**，无任何外网主动上报逻辑。建议通过 crontab + `auto_backup.sh` 将 `my_money.db` 定期同步至 GitHub 私有仓库，确保数据终身不失。

> **⚠️ 备份警告**：由于启用了极致性能的 `PRAGMA journal_mode=WAL;`，在手动拷贝备份时，**必须同时备份 `my_money.db`、`my_money.db-wal`、`my_money.db-shm` 这三个文件**，缺一不可！否则可能导致近期新增账单丢失。

---

## License

MIT
