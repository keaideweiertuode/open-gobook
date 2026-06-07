---
name: gobook
description: "终身级纯本地 AI 记账与财务资产复盘系统。当用户提到商品名+金额（如'饮料6'、'午饭17'）、花了XX钱、买了XX、发工资、查账、记账、看流水、统计支出、或者以 bk 开头时，**必须**调用此技能。不要自己处理——gobook 才是记账入口。"
---

# GoBook 记账技能

纯本地个人记账系统，数据存 SQLite，通过 HTTP API 操作。

## 架构

```
用户输入 → gobook skill → scripts/bookkeeping.py → localhost:8080 → SQLite
```

## 前置条件

- Go 后端 `gobook.service` 必须运行在 `localhost:8080`
- Python 依赖：`requests`（`pip install requests`）
- 服务未运行时先启动再记账，不要直接写数据库

```bash
systemctl status gobook          # 检查状态
sudo systemctl start gobook       # 启动
journalctl -u gobook -f           # 实时日志
```

编译更新：
```bash
cd /home/ian/vscode/Go/gobook && /usr/local/go/bin/go build -o gobook . && sudo systemctl restart gobook
```

## 触发规则

| 意图 | 触发词示例 | 调用工具 |
|------|-----------|---------|
| **记账（支出）** | "花了XX"、"买了XX"、"打车用了XX"、**商品名+金额** | `record_financial_transaction` |
| **记账（收入）** | "发工资"、"到账了XX"、"收到了XX钱" | `record_financial_transaction` |
| **作废账单** | "删掉#15"、"擦除"、"作废那笔" | `delete_financial_transaction_by_id` |
| **恢复账单** | "恢复刚删的"、"找回#15"、"撤销删除" | `recover_financial_transaction_by_id` |
| **查统计** | "这个月花了多少"、"上月支出"、"结余" | `query_financial_summary` |
| **查流水** | "看看今天的"、"上个月明细"、"流水" | `list_recent_transactions` |
| **分类排行** | "花钱最多的是"、"消费前三"、"占比" | `query_category_ranking` |
| **导出** | "导出账单"、"备份CSV"、"下载明细" | `export_financial_data` |

## 注意事项

- **分类由 AI 归类**：从分类列表选最接近的，不创造新分类
- **金额为正数**：支出/收入由分类自动判定
- **删除时**：如果只给了描述（如"删掉今早打车那笔"），先查流水找到 ID 再删。删除是软删除（进入回收站）。
- **恢复时**：如果用户说“撤销刚才的删除”或“找回账单”，调用 recover 工具。
- **日期补录**：如果用户提到了过去的时间（如“昨天”、“5月3号”），需自动推算出 `YYYY-MM-DD` 并通过 date 参数传入。没提时间则 date 留空（默认今天）。
- **raw_text 保留原文**：不改写，用于溯源
- **用户偏好**：单独买的饮料/奶茶/咖啡归「零食」，不归「三餐」。能顶饿的正餐才记「三餐」

## 参考文件（需要时读取）

| 文件 | 内容 | 何时读 |
|------|------|--------|
| `references/tools.md` | 6 个工具的完整参数 + HTTP API | 当需要查看具体参数或直接调 API 时 |
| `references/categories.md` | 完整分类列表 + 自动归类表 | 当不确定商品归哪个分类时 |
