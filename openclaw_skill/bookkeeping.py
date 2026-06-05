import requests
from typing import Dict, Any

def record_financial_transaction(amount: float, category: str, remark: str, raw_text: str) -> str:
    """
    当你需要记录一笔本地的财务流水（无论是日常开销支出，还是工资、红包等资金收入）时，调用此工具。
    
    Args:
        amount (float): 账单的具体金额。必须是正数数字（例如: 35.0, 12000.50）。
        category (str): 账单的分类名称。必须**严格限定**在以下允许的分类列表中选择一个最接近的填入:
                        【支出类】: 三餐, 零食, 衣服, 交通, 旅行, 孩子, 宠物, 话费网费, 烟酒, 学习,
                                   日用品, 住房, 美妆, 医疗, 发红包, 汽车, 娱乐, 请客送礼, 电器数码, 
                                   运动, 水电煤, 办公, 其它
                        【收入类】: 工资, 生活费, 收红包, 股票基金, 其他
        remark (str): 简短的备注、特定物品或商户名称（例如: "端午节发红包", "5月工资", "加油站充值", "麦当劳"）。
        raw_text (str): 用户输入的原始大白话文本，用于留痕溯源（例如: "发工资 8000" 或 "打车花分 25"）。
        
    Returns:
        str: 写入本地高级存储系统的同步回执状态信息。
    """
    url = "http://localhost:8080/api/transactions"
    payload = {
        "amount": amount,
        "category": category,
        "remark": remark,
        "raw_text": raw_text
    }
    
    try:
        response = requests.post(url, json=payload, timeout=5)
        if response.status_code == 200:
            res_data = response.json()
            tx_type = "【收入】" if res_data.get("type") == "income" else "【支出】"
            return f"成功！本地账本已同步录入{tx_type}：[{category}] ¥{amount:.2f}，备注：{remark}。"
        else:
            return f"❌ 本地存储失败，后端返回错误代码: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常：调用系统 Skill 失败: {str(e)}"


def delete_financial_transaction_by_id(transaction_id: int) -> str:
    """
    当用户明确指示要删除、擦除、撤销或作废某笔特定账单记录时，调用此工具。
    
    Args:
        transaction_id (int): 要抹除的账单唯一物理 ID（必须是正整数数字，例如: 12, 105）。
                              通常用户会参考 Web UI 看板明细流水表中的 ID 号。
                              
    Returns:
        str: 后端核心存储系统执行物理擦除的结果回执。
    """
    url = f"http://localhost:8080/api/transactions/delete?id={transaction_id}"
    
    try:
        response = requests.post(url, timeout=5)
        if response.status_code == 200:
            return f"✅ 擦除成功！账单编号 #{transaction_id} 已移入回收站（软删除），Web UI 看板已同步更新。如果属于误删，可随时通过 recover_financial_transaction_by_id 进行恢复。"
        else:
            return f"❌ 擦除失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常：执行删除工具时发生错误: {str(e)}"

def recover_financial_transaction_by_id(transaction_id: int) -> str:
    """
    当用户想要撤销删除、找回或恢复一笔被误删的账单记录时，调用此工具。
    由于系统采用了安全的软删除机制，最近删除的账单可以被完美恢复。
    
    Args:
        transaction_id (int): 要恢复的账单唯一 ID（必须是正整数数字，例如: 12, 105）。
                              
    Returns:
        str: 后端核心存储系统执行恢复的结果回执。
    """
    url = f"http://localhost:8080/api/transactions/recover?id={transaction_id}"
    
    try:
        response = requests.post(url, timeout=5)
        if response.status_code == 200:
            return f"✅ 恢复成功！账单编号 #{transaction_id} 已从回收站中找回，并重新在 Web UI 看板中显示。"
        else:
            return f"❌ 恢复失败，可能该 ID 不存在于回收站中。后端返回: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常：执行恢复工具时发生错误: {str(e)}"

def export_financial_data(period: str = "") -> str:
    """
    当用户要求导出账单、下载 CSV 或全量备份财务数据时，调用此工具。
    
    Args:
        period (str): 可选的时间跨度（如 "2026", "2026-05", "2026-05-31"）。留空 "" 则导出数据库所有的历史记录。
        
    Returns:
        str: 后端核心存储系统执行导出的结果回执，包含本地保存路径。
    """
    import os
    url = f"http://localhost:8080/api/export?period={period}"
    
    try:
        response = requests.get(url, timeout=10)
        if response.status_code == 200:
            filename = f"gobook_export_{period}.csv" if period else "gobook_export_all.csv"
            save_path = os.path.abspath(filename)
            with open(save_path, "wb") as f:
                f.write(response.content)
            return f"✅ 导出成功！{period or '全量'} 的财务明细数据已保存至本地文件：{save_path}"
        else:
            return f"❌ 导出失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常：执行导出工具时发生错误: {str(e)}"

def query_financial_summary(period: str) -> str:
    """
    查询指定时间段的收支汇总（收入、支出、净结余）。
    
    Args:
        period (str): 时间跨度，格式 "YYYY", "YYYY-MM", "YYYY-MM-DD"。
        
    Returns:
        str: 汇总数据结果。
    """
    url = f"http://localhost:8080/api/stats/daily?period={period}"
    try:
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            trend = response.json()
            total_expense = sum(item.get("expense", 0) for item in trend)
            total_income = sum(item.get("income", 0) for item in trend)
            return f"📊 {period} 汇总：收入 ¥{total_income:.2f}，支出 ¥{total_expense:.2f}，净结余 ¥{total_income - total_expense:.2f}"
        return f"❌ 查询失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常: {str(e)}"


def query_category_ranking(period: str, top_n: int = 10) -> str:
    """
    查询指定时间段的支出分类排行。
    
    Args:
        period (str): 时间跨度，格式 "YYYY", "YYYY-MM", "YYYY-MM-DD"。
        top_n (int): 返回前 N 个分类，默认 10。
        
    Returns:
        str: 分类排行结果。
    """
    url = f"http://localhost:8080/api/stats/category?period={period}"
    try:
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            data = response.json()
            if not data:
                return f"📊 {period} 暂无支出分类数据。"
            total = sum(item["total_amount"] for item in data)
            display = data[:top_n]
            medals = ["🥇", "🥈", "🥉"]
            lines = []
            for i, item in enumerate(display):
                rank = medals[i] if i < 3 else f"  #{i+1}"
                pct = item["total_amount"] / total * 100 if total > 0 else 0
                lines.append(f"  {rank} {item['category_name']}: ¥{item['total_amount']:.2f} ({pct:.1f}%)")
            more = f"\n  ... 共 {len(data)} 个分类" if len(data) > top_n else ""
            return f"📊 {period} 支出分类排行（总支出 ¥{total:.2f}）:{'\n'.join(lines)}{more}"
        return f"❌ 查询失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。"
    except Exception as e:
        return f"❌ 运行异常: {str(e)}"

def list_recent_transactions(period: str, category: str = "", page: int = 1, page_size: int = 20) -> str:
    """
    查询指定时间段的流水明细列表（包含每笔账单的 ID、分类、金额、备注）。
    当查询月度或年度等数据量较大的时期时，可配合 page 和 page_size 参数实现翻页。
    
    Args:
        period (str): 时间跨度，格式 "YYYY", "YYYY-MM", "YYYY-MM-DD"。
        category (str): 可选，按分类筛选（如 "三餐"、"交通"）。留空则显示所有分类。
        page (int): 请求的页码，默认为 1。
        page_size (int): 每页包含的条目数，默认为 20。
        
    Returns:
        str: 账单明细列表与分页信息。
    """
    url = f"http://localhost:8080/api/transactions/list?period={period}&page={page}&pageSize={page_size}"
    if category:
        import urllib.parse
        url += f"&category={urllib.parse.quote(category)}"
        
    try:
        response = requests.get(url, timeout=5)
        if response.status_code == 200:
            res_data = response.json()
            # 兼容老版本后端可能返回纯列表的情况
            if isinstance(res_data, list):
                data = res_data
                total = len(data)
            else:
                data = res_data.get("data", [])
                total = res_data.get("total", 0)
                
            if not data:
                cat_str = f" [{category}]" if category else ""
                return f"📋 {period}{cat_str} (第 {page} 页) 暂无流水记录。"
            
            lines = []
            for t in data:
                sign = "+" if t.get("type") == "income" else "-"
                lines.append(f"  #{t['id']} {t['transaction_date']} [{t['category_name']}] {sign}¥{t['amount']:.2f} 备注:{t['remark']}")
            
            total_pages = (total + page_size - 1) // page_size if page_size > 0 else 1
            cat_str = f" [{category}]" if category else ""
            return f"📋 {period}{cat_str} 流水明细（共 {total} 条，第 {page}/{total_pages} 页）:\n" + "\n".join(lines)
        return f"❌ 查询失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。"
    except Exception as e:
        return f"❌ 运行异常: {str(e)}"