---
name: gobook
description: 终身级纯本地 AI 记账与财务资产复盘系统。记录收支、查看统计、查询流水、删除账单、管理预算。
---

# GoBook 记账技能

GoBook 是一个面向**长周期运行设计**的个人记账系统，数据存储在本地 SQLite 中。支持通过 CLI、Web UI、HTTP API 三种方式使用。

## 架构概要

```
用户说大白话 → OpenClaw 提取参数 → {baseDir}/bookkeeping.py → HTTP → localhost:8080 → SQLite
```

### 前置条件

GoBook 已注册为 systemd 服务（`gobook.service`），开机自启，监听 `localhost:8080`。

```bash
# 服务管理（需要 sudo）
sudo systemctl start gobook      # 启动
sudo systemctl stop gobook       # 停止
sudo systemctl restart gobook    # 重启（编译新代码后使用）
systemctl status gobook          # 查看状态
journalctl -u gobook -f          # 查看实时日志
```

### 编译（修改代码后）

```bash
cd /home/ian/vscode/Go/gobook
/usr/local/go/bin/go build -o gobook .
sudo systemctl restart gobook
```

如果服务不在运行，先帮用户启动再记账。不要尝试绕过服务直接写数据库。

## 工具

### 1. record_financial_transaction — 记录账单

当用户描述一笔支出或收入时，调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import record_financial_transaction

result = record_financial_transaction(
    amount=<float>,       # 正数金额，如 35.0, 12000.50
    category=<str>,       # 分类名称（见下方分类列表）
    remark=<str>,         # 备注/商户名
    raw_text=<str>        # 用户原始输入文本，用于留痕溯源
)
```

**使用示例（场景 A：智能双轨录入）：**

```
用户说："刚才打车花了 32.5"
→ record_financial_transaction(amount=32.5, category="交通", remark="打车", raw_text="刚才打车花了 32.5")

用户说："股票基金收益到账 1500"
→ record_financial_transaction(amount=1500.0, category="股票基金", remark="收益到账", raw_text="股票基金收益到账 1500")
```

**直接 POST 调用：**

```json
POST /api/transactions
{
  "amount": 38.0,
  "category": "三餐",
  "remark": "麦当劳",
  "raw_text": "中午吃麦当劳花了38块钱"
}
```

### 2. delete_financial_transaction_by_id — 删除账单

当用户明确指示要删除、擦除、撤销或作废某笔特定账单记录时，调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import delete_financial_transaction_by_id

result = delete_financial_transaction_by_id(
    transaction_id=<int>   # 要删除的账单 ID（正整数）
)
```

**使用示例（场景 B：智能流水物理擦除）：**

```
用户说："刚才看网页发现 #15 那笔账记重了，帮我删掉"
→ delete_financial_transaction_by_id(transaction_id=15)
→ 回复："✅ 物理擦除成功！账单编号 #15 已彻底从本地 SQLite 数据库中抹除，Web UI 看板已同步更新。"
```

**直接 POST/DELETE 调用：**

```bash
POST /api/transactions/delete?id=15
# 或
DELETE /api/transactions/delete?id=15
```

### 3. export_financial_data — 导出账单明细

当用户要求备份、下载、导出账单数据时，调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import export_financial_data

result = export_financial_data(
    period=<str>   # 要导出的时间段（"YYYY"、"YYYY-MM" 或 "YYYY-MM-DD"）。留空则导出所有。
)
```

**使用示例（场景 C：数据一键导出）：**

```
用户说："把2026年5月的账单给我导出来"
→ export_financial_data(period="2026-05")
→ 回复："✅ 导出成功！2026-05 的财务明细数据已保存至本地文件：/xxx/gobook_export_2026-05.csv"
```

### 4. query_financial_summary — 查账统计

当用户询问特定时间段花了多少钱、存了多少钱、结余多少时调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import query_financial_summary

result = query_financial_summary(
    period=<str>   # 查询周期（"YYYY"、"YYYY-MM" 或 "YYYY-MM-DD"）
)
```

**使用示例：**
```
用户说："这个月花了多少"
→ query_financial_summary(period="2026-05")
→ 回复："📊 2026-05 汇总：收入 ¥...，支出 ¥...，净结余 ¥..."
```

### 5. list_recent_transactions — 查账流水

当用户要看具体的消费明细、核对账单时调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import list_recent_transactions

result = list_recent_transactions(
    period=<str>,      # 查询周期（"YYYY"、"YYYY-MM" 或 "YYYY-MM-DD"）
    category=<str>,    # 可选，按分类筛选（如 "三餐"），留空显示全部
    page=<int>,        # 可选，请求的页码，默认为 1。配合翻页使用。
    page_size=<int>    # 可选，每页展示数，默认为 20。
)
```

**使用示例：**
```
用户说："看看今天的流水"
→ list_recent_transactions(period="2026-05-31")

用户说："看看上个月交通花了多少"
→ list_recent_transactions(period="2026-04", category="交通")

用户说："我要看今年流水的第二页"
→ list_recent_transactions(period="2026", page=2)
```

### 6. query_category_ranking — 分类排行

当用户问某种时间段哪类消费最多、排名前三、分类占比时调用此工具。

```python
import sys; sys.path.insert(0, '{baseDir}')
from bookkeeping import query_category_ranking

result = query_category_ranking(
    period=<str>,     # 时间跨度（"YYYY"、"YYYY-MM"、"YYYY-MM-DD"）
    top_n=<int>       # 可选，返回前 N 个分类，默认 10
)
```

**使用示例：**
```
用户说："今年花钱最多的地方是哪儿"
→ query_category_ranking(period="2026")

用户说："上个月消费前三是什么"
→ query_category_ranking(period="2026-04", top_n=3)
```

### 7. 查询统计（HTTP GET） (内部 API 参考)

如果你需要自行调用 HTTP API 获取更复杂的数据分布（例如分类排行）：
| 接口 | 参数 | 说明 |
|------|------|------|
| `GET /api/stats/daily?period=2026-05` | period: YYYY-MM-DD / YYYY-MM / YYYY | 双轨收支统计（支出+收入） |
| `GET /api/stats/category?period=2026-05` | period: YYYY-MM-DD / YYYY-MM / YYYY | 支出分类排行 |
| `GET /api/stats/years` | — | 获取所有可用年份 |
| `GET /api/transactions/list?period=2026-05-27` | period: YYYY-MM-DD / YYYY-MM / YYYY | 流水明细列表（含 ID、金额、分类、备注） |
| `GET /api/export?period=2026-05` | period: YYYY-MM-DD / YYYY-MM / YYYY (可选) | 导出账单明细（CSV格式，带 UTF-8 BOM） |

### 8. 终端 CLI 记账

```bash
cd /home/ian/vscode/Go/gobook && ./gobook <描述词> <金额>
```

示例：

```bash
./gobook 吃饭 25
./gobook 发工资 8000
./gobook 打车去机场 120
```

### 9. Web 看板

浏览器打开 `http://localhost:8080`，查看充满科技感（Cyberpunk 风格）的日/月/年三级联动赛博 AI 收支看板。

## 分类参考

**支出类目（23个）：** 三餐 · 零食 · 衣服 · 交通 · 旅行 · 孩子 · 宠物 · 话费网费 · 烟酒 · 学习 · 日用品 · 住房 · 美妆 · 医疗 · 发红包 · 汽车 · 娱乐 · 请客送礼 · 电器数码 · 运动 · 其它 · 水电煤 · 办公

**收入类目（5个）：** 工资 · 生活费 · 收红包 · 股票基金 · 其他

## 触发规则

当用户表达以下意图时，触发此技能：

| 意图 | 触发词示例 | 调用的工具 |
|------|-----------|-----------|
| **记账（支出）** | "花了XX钱"、"今天XX花了XX"、"打车用了XX"、**"商品名+数字"** | `record_financial_transaction` |
| **记账（收入）** | "发工资了"、"到账了XX"、"收到了XX钱"、"入账" | `record_financial_transaction` |
| **删除/擦除** | "删掉"、"#15那笔删了"、"擦除"、"作废账单" | `delete_financial_transaction_by_id` |
| **恢复/撤销删除** | "恢复刚刚删掉的"、"把#15那笔找回来" | `recover_financial_transaction_by_id` |
| **数据导出** | "导出本月账单"、"把账单备份成CSV"、"帮我下载明细" | `export_financial_data` |
| **查账/统计** | "这个月花了多少"、"上个月支出"、"本月结余" | `query_financial_summary` |
| **查明细** | "看看今天的流水"、"上个月的明细"、"列表" | `list_recent_transactions` |
| **分类排行** | "花钱最多的是哪儿"、"消费前三"、"支出占比" | `query_category_ranking` |

## 注意事项

- **分类由 ub 负责归类**：用户只说商品名+金额（如"可乐5"、"午饭17"），ub 根据商品名从分类列表中选择最接近的，不要自行创造新分类
- **金额必须是正数**：支出/收入方向由分类自动判定（expense/income）
- **用户说"删掉某笔"时**：如果用户只说了描述（如"删掉今早打车那笔"），优先通过 `GET /api/transactions/list?period=今天` 查到对应的 ID，再用 ID 删除
- **raw_text 必须保留原文**：用于日后溯源，不要改写
- **用户偏好（2026-05-28）**：单独购买的饮料/奶茶/咖啡等饮品，统一归 **「零食」** 类，不归「三餐」。只有能顶饿的正餐/主食才记「三餐」。

### 常用商品自动归类参考

| 用户说 | 自动归为 |
|--------|----------|
| 可乐、雪碧、柠檬水、奶茶、咖啡、水、瓜子、薯片、零食 | 零食 |
| 午饭、晚饭、早餐、饭、面、快餐 | 三餐 |
| 打车、地铁、公交 | 交通 |
| 房租、水电、网费 | 住房 / 水电煤 |
