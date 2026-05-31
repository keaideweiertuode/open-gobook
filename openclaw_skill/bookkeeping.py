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
    # 映射最新的物理删除 API 路由
    url = f"http://localhost:8080/api/transactions/delete?id={transaction_id}"
    
    try:
        response = requests.post(url, timeout=5)
        if response.status_code == 200:
            return f"✅ 物理擦除成功！账单编号 #{transaction_id} 已彻底从本地 SQLite 数据库中抹除，Web UI 看板已同步更新。"
        else:
            return f"❌ 擦除失败，后端返回错误详情: {response.text}"
    except requests.exceptions.ConnectionError:
        return "❌ 连接失败：未检测到本地核心财务 API。请确认 Go 服务是否在 localhost:8080 挂起运行。"
    except Exception as e:
        return f"❌ 运行异常：执行删除工具时发生错误: {str(e)}"

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