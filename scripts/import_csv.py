import sqlite3
import csv
from datetime import datetime

def parse_time(time_str):
    """
    智能解析时间字符串，兼容各种格式
    """
    time_str = time_str.strip()
    formats = [
        "%m/%d/%y %H:%M",   # 匹配 '5/26/26 19:01'
        "%m/%d/%y %H:%M:%S",
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%d %H:%M",
        "%Y/%m/%d %H:%M:%S"
    ]
    for fmt in formats:
        try:
            return datetime.strptime(time_str, fmt)
        except ValueError:
            continue
            
    print(f"⚠️ 警告: 无法识别的时间格式 '{time_str}'，默认使用当前时间")
    return datetime.now()

def import_csv_to_sqlite(csv_filepath, db_filepath="my_money.db"):
    print(f"🚀 开始准备从 {csv_filepath} 导入数据...")
    
    conn = sqlite3.connect(db_filepath)
    cursor = conn.cursor()

    cursor.execute("SELECT id, name, type FROM categories")
    categories = cursor.fetchall()
    
    if not categories:
        print("❌ 错误：数据库中没有分类数据，请先运行 go run . 初始化数据库！")
        return

    cat_map = {row[1]: (row[0], row[2]) for row in categories}
    fallback_expense_id = cat_map.get("其它", (None, 'expense'))[0]
    fallback_income_id = cat_map.get("其他", (None, 'income'))[0]

    transactions_to_insert = []
    success_count = 0
    fallback_count = 0

    try:
        with open(csv_filepath, mode='r', encoding='utf-8-sig') as f:
            reader = csv.DictReader(f)
            
            for row in reader:
                # 【核心修复】：使用 str(row.get(key) or "") 彻底杜绝 NoneType 报错
                raw_time = str(row.get("时间") or "")
                category_name = str(row.get("分类") or "").strip()
                tx_type_str = str(row.get("类型") or "").strip() 
                amount_str = str(row.get("金额") or "0")
                remark = str(row.get("备注") or "").strip()

                if not raw_time or amount_str == "0":
                    continue

                try:
                    amount = float(amount_str)
                except ValueError:
                    continue

                parsed_date = parse_time(raw_time)
                tx_date = parsed_date.strftime("%Y-%m-%d")               
                created_at = parsed_date.strftime("%Y-%m-%d %H:%M:%S") 

                if category_name in cat_map:
                    cat_id, cat_type = cat_map[category_name]
                else:
                    fallback_count += 1
                    if tx_type_str == "收入":
                        cat_id = fallback_income_id
                        cat_type = "income"
                        remark = f"[{category_name}] {remark}" 
                    else:
                        cat_id = fallback_expense_id
                        cat_type = "expense"
                        remark = f"[{category_name}] {remark}"

                raw_text = f"历史导入数据：{category_name} - {remark}"

                transactions_to_insert.append((
                    amount, cat_type, 1, cat_id, tx_date, raw_text, remark, created_at
                ))
                success_count += 1

    except Exception as e:
        print(f"❌ 读取 CSV 失败: {e}")
        return

    if transactions_to_insert:
        cursor.executemany("""
            INSERT INTO transactions 
            (amount, type, account_id, category_id, transaction_date, raw_text, remark, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """, transactions_to_insert)
        
        conn.commit()
        print(f"✅ 导入成功！共计无损汇入 {success_count} 条记录。")
        if fallback_count > 0:
            print(f"⚠️ 提示: 其中有 {fallback_count} 条记录因找不到对应分类，已安全兜底至「其它/其他」类目，原分类名已保留在备注中。")
    else:
        print("⚠️ 未发现有效数据可导入。")

    conn.close()

if __name__ == "__main__":
    CSV_FILENAME = "Qian_Ji_2026-05-27-21-39.csv" 
    import_csv_to_sqlite(CSV_FILENAME)