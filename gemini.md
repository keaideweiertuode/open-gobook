# GoBook — 项目智能文档 (Gemini)

> 本文件为 AI 编码助手提供项目级上下文，确保在代码生成、重构、Debug 时与项目设计意图一致。

---

## 项目概览

GoBook 是一个**纯本地、零云端**的个人记账与财务资产复盘系统，面向 20~30 年长周期运行设计。采用 **Go + SQLite + 纯前端三剑客 (Vue 3 + Tailwind CSS + ECharts)** 的技术栈，编译产物为约 15MB 的单文件静态二进制。

### 核心设计哲学

| 原则 | 实现 |
|------|------|
| **数据终身可控** | 单文件 SQLite (`my_money.db`)，备份 = `cp` |
| **零 CGO** | 使用 `modernc.org/sqlite` 纯 Go 驱动，彻底消除 C 依赖 |
| **前端零构建** | 原生 HTML + CDN 加载 Vue 3 / Tailwind / ECharts，无 Vite/Webpack |
| **年份数据驱动** | 年份选择器由 SQL `DISTINCT strftime('%Y')` 自动盘点，零硬编码 |
| **原文留痕** | `raw_text` 字段终身保留用户输入原文，便于未来洗数或训练模型 |

---

## 技术栈

- **语言**: Go 1.25+
- **数据库**: SQLite (via `modernc.org/sqlite` v1.50.1，纯 Go，无 CGO)
- **前端框架**: Vue 3 (CDN)
- **样式**: Tailwind CSS (CDN)
- **图表**: ECharts 5.5 (CDN)
- **HTTP 服务器**: Go 标准库 `net/http`
- **辅助脚本**: Python 3 (CSV 导入 / 测试数据生成), Bash (自动备份)

---

## 架构总览

```
┌─────────────────────────────────────────────────────────┐
│                     用户入口                              │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐ │
│  │  CLI 终端    │  │  Web 浏览器  │  │ OpenClaw Agent   │ │
│  │ (大白话录入)  │  │ (财务看板)   │  │ (自然语言记账)    │ │
│  └──────┬──────┘  └──────┬──────┘  └────────┬─────────┘ │
└─────────┼────────────────┼──────────────────┼───────────┘
          │                │                  │
          ▼                ▼                  ▼
┌─────────────────────────────────────────────────────────┐
│                 cmd/gobook/main.go (入口)                 │
│  • CLI 模式: args > 0 → handleCLIRecord / rm|del        │
│  • Web 模式: args = 0 → api.StartWebServer()            │
└────────────────────────┬────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
   ┌─────────────┐┌─────────────┐┌─────────────────┐
   │internal/api/││internal/    ││   web/static/   │
   │   api.go    ││ models/     ││   index.html    │
   │ 6 个端点    ││  models.go  ││                 │
   └─────┬───────┘└─────────────┘└─────────────────┘
         │
         ▼
   ┌─────────────┐
   │internal/db/ │
   │   db.go     │
   │ SQLite      │
   │ 核心层      │
   └─────┬───────┘
         │
         ▼
   ┌─────────────┐
   │    data/    │
   │ my_money.db │
   └─────────────┘
```

---

## 文件职责速查

| 文件 | 职责 | 行数 | 关键函数/结构 |
|------|------|------|-------------|
| `cmd/gobook/main.go` | 程序入口，CLI 模式分发与 Web 服务启动 | ~91 | `main()`, `handleCLIRecord()`, `mockCallAI()` |
| `internal/api/api.go` | HTTP 路由注册与 6 个 API handler | ~137 | `StartWebServer()`, `handleCreateTransaction()`, `handleDailyStats()`, `handleCategoryStats()`, `handleListTransactions()`, `handleDeleteTransaction()`, `handleAvailableYears()` |
| `internal/db/db.go` | SQLite DDL、种子数据、CRUD 和统计查询 | ~221 | `InitDB()`, `SaveTransaction()`, `DeleteTransaction()`, `QueryTrendStats()`, `QueryCategoryStats()`, `QueryTransactions()`, `QueryAvailableYears()` |
| `internal/db/db_test.go` | SQLite 内存模式自动化测试 | ~86 | `TestGetCategoryDetailsByName()`, `TestTransactionCRUD()` |
| `internal/models/models.go` | 统一的数据模型 | ~47 | `Category`, `Transaction`, `TrendStat`, `CategoryStat`, `TransactionDetail` |
| `internal/api/api_test.go` | 核心路由自动化集成测试 (httptest) | ~73 | `TestHandleCreateTransaction()`, `TestHandleExport()` |
| `web/static/index.html` | 单文件 SPA 前端 (Vue 3 + Tailwind + ECharts) | ~326 | 日/月/年视图切换、双轨趋势图、分类饼图、流水明细表 |
| `scripts/import_csv.py` | 从钱迹等 App 导入 CSV 历史数据 | ~113 | `import_csv_to_sqlite()`, `parse_time()` |
| `scripts/seed_2025.py` | 生成 2025 全年模拟消费数据 | ~113 | 测试/演示用 |
| `scripts/auto_backup.sh` | Git 自动备份脚本 (配合 crontab) | ~16 | 每日 23:00 自动 commit + push |

---

## 数据模型

### 三张物理表

```sql
-- 1. 资产账户（预留多账本扩展，当前固定 id=1 "默认账本"）
accounts (id, name, balance, is_active)

-- 2. 财务类目（23 支出 + 5 收入 = 28 个内置分类）
categories (id, name, type['expense'|'income'], is_active)

-- 3. 财务流水主表（核心，含 raw_text 原文留痕）
transactions (id, amount, type, account_id, category_id,
              transaction_date, raw_text, remark, created_at)
  -- 索引: idx_tx_date (transaction_date), idx_tx_category (category_id)
```

### Go 结构体映射

| 结构体 | 用途 |
|--------|------|
| `Category` | 分类元数据 (id, name, type, is_active) |
| `Transaction` | 账单写入模型 (含 `RawText` 原文字段) |
| `TrendStat` | 双轨收支趋势统计 (date, expense, income) |
| `CategoryStat` | 分类支出排行 (category_name, total_amount) |
| `TransactionDetail` | 流水明细展示模型 (定义在 models.go 中) |

---

## API 端点

| 方法 | 路由 | Handler | 说明 |
|------|------|---------|------|
| `POST` | `/api/transactions` | `handleCreateTransaction` | 创建账单 (body: amount, category, remark, raw_text) |
| `POST`/`DELETE` | `/api/transactions/delete?id=N` | `handleDeleteTransaction` | 物理删除账单 (双路由注册) |
| `GET` | `/api/transactions/list?period=YYYY-MM-DD` | `handleListTransactions` | 流水明细列表 (支持年/月/日粒度) |
| `GET` | `/api/stats/daily?period=YYYY-MM` | `handleDailyStats` | 双轨收支趋势统计 |
| `GET` | `/api/stats/category?period=YYYY-MM` | `handleCategoryStats` | 支出分类排行 |
| `GET` | `/api/stats/years` | `handleAvailableYears` | 可用年份列表 |
| `GET` | `/api/export?period=YYYY-MM` | `handleExport` | 导出指定时间跨度的明细 (CSV，带 UTF-8 BOM) |

> **注**: 所有 API 均包裹了 `corsMiddleware` 允许全量跨域跨源访问，失败响应体格式统一为 JSON `{"status": "error", "message": "..."}`。

### period 参数约定

查询的 `period` 参数通过字符串长度自动识别粒度：

| 长度 | 格式 | 粒度 | 示例 |
|------|------|------|------|
| 10 | `YYYY-MM-DD` | 日 | `2026-05-31` |
| 7 | `YYYY-MM` | 月 | `2026-05` |
| 4 | `YYYY` | 年 | `2026` |

---

## CLI 模式

```bash
# 记账（大白话录入）
./bin/gobook 吃饭 25          # → 分类"三餐"，金额 25
./bin/gobook 发工资 8000      # → 分类"工资"，自动识别为收入

# 删除
./bin/gobook rm 15            # 物理删除 ID=15 的账单
./bin/gobook del 15           # 同上
```

### 分类识别逻辑 (`mockCallAI`)

当前为关键词匹配的 Mock 实现：
- 包含 `饭`/`面`/`餐` → 三餐
- 包含 `工资` → 工资
- 其他 → 其它 (兜底)

> **设计意图**: `mockCallAI` 函数是 AI 解析的占位实现，未来可替换为真实 LLM 调用或 OpenClaw Agent 集成。

---

## 关键编码约定

### Go 代码风格
1. **包结构**: 遵循 Go 标准项目布局 (`cmd/`, `internal/`) 进行包隔离，防止循环依赖
2. **错误处理**: 内部使用 `fmt.Errorf("...: %w", err)` 包裹，外部 API 返回统一 `jsonError` 结构
3. **数据库连接**: 全局变量 `var db *sql.DB` 位于 `internal/db/db.go` 中，由 `InitDB()` 初始化，进程退出前须调用 `db.CloseDB()` 安全关闭
4. **JSON 序列化**: 使用 `json` struct tag，所有字段小写下划线命名
5. **时间格式**: Go 标准 `2006-01-02` 布局字符串
6. **API 路由**: 全面拥抱 Go 1.22+ 原生 HTTP 方法路由 (`http.HandleFunc("POST /xxx", ...)`)
7. **分类查找兜底**: `GetCategoryDetailsByName` 找不到时自动降级到 "其它"，并在 HTTP 响应中明确返回 `matched_category_id`

### 前端约定
1. **单文件 SPA**: 所有前端逻辑集中在 `web/static/index.html`
2. **CDN 依赖**: Vue 3 (unpkg), Tailwind CSS (cdn.tailwindcss.com), ECharts 5.5
3. **状态管理**: Vue 3 Options API + `data()` 响应式状态
4. **视图模式**: 三档切换 (day / month / year)，联动年/月/日选择器
5. **静态文件服务**: `http.FileServer(http.Dir("./web/static"))` 挂载到根路径

### 数据库约定
1. **自动建表**: `InitDB()` 使用 `CREATE TABLE IF NOT EXISTS` 幂等建表
2. **种子数据**: `INSERT OR IGNORE` 确保分类数据只初始化一次
3. **物理删除**: 无软删除机制，`DELETE FROM` 直接物理擦除
4. **日期存储**: `transaction_date` 存储为 `YYYY-MM-DD` 纯文本格式
5. **时间查询**: 使用 SQLite `strftime()` 函数进行日期范围过滤

---

## 已知设计决策

1. **单账本模式**: `account_id` 固定为 1，但表结构已预留多账本扩展
2. **无认证机制**: 纯本地部署，无用户认证/鉴权
3. **网络与安全**: 仅监听 `localhost:8080`，不提供 TLS，但已全量开启 CORS 支持外部 Agent 跨域请求
4. **无分页**: 流水明细列表无分页，一次性返回所有记录
5. **DELETE 兼容**: 得益于 Go 1.22+ 路由，物理删除显式注册了 POST 和 DELETE 两个方法，以兼容不同客户端
6. **created_at 双轨**: Go 端写入不设 `created_at` (由 SQLite `DEFAULT CURRENT_TIMESTAMP`)，但 Python CSV 导入脚本手动设置 `created_at` 以保留原始时间
7. **SQLite 性能调优**: `InitDB` 时自动注入 `PRAGMA` 开启 WAL、高并发及大缓存机制

---

## 开发指南

### 编译与运行

```bash
# 编译
cd /home/ian/vscode/Go/gobook
go build -o bin/gobook ./cmd/gobook

# 运行
./bin/gobook              # Web 模式 → http://localhost:8080
./bin/gobook 吃饭 25      # CLI 记账模式

# 自动化测试 (纯内存无痕测试)
go test -v ./internal/...
```

### 部署 (Systemd)

```bash
sudo systemctl enable --now gobook
journalctl -u gobook -f
```

### 测试数据

```bash
python3 scripts/seed_2025.py        # 注入 2025 年全年模拟数据
python3 scripts/import_csv.py       # 从钱迹 CSV 导入真实数据
```

### 备份

```bash
# 手动备份
cp data/my_money.db data/my_money_backup_$(date +%Y%m%d).db

# 自动备份 (crontab)
0 23 * * * /home/ian/vscode/Go/gobook/scripts/auto_backup.sh
```

---

## 扩展路线提示

| 方向 | 切入点 |
|------|--------|
| 真实 AI 分类 | 替换 `mockCallAI()` 为 LLM API 调用 |
| 多账本支持 | 激活 `accounts` 表，修改 `account_id` 逻辑 |
| 预算管理 | 新增 `budgets` 表，前端增加预算对比视图 |
| 流水分页 | `QueryTransactions` 增加 `LIMIT/OFFSET` 参数 |
| 软删除 | `transactions` 增加 `deleted_at` 字段 |
