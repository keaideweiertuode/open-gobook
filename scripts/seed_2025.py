import sqlite3
import random
from datetime import datetime, timedelta

def generate_2025_data():
    db_path = "my_money.db"
    print(f"正在连接数据库 {db_path} 并准备注入 2025 全年测试数据...")
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # 1. 获取分类的名称到 ID 的映射
    cursor.execute("SELECT id, name FROM categories WHERE type = 'expense'")
    category_map = {name: cid for cid, name in cursor.fetchall()}
    
    if not category_map:
        print("❌ 错误：在数据库中未找到任何分类，请先运行一次 Go 服务初始化数据库！")
        return

    # 2. 清理可能存在的历史 2025 数据，防止重复注入
    cursor.execute("DELETE FROM transactions WHERE transaction_date LIKE '2025-%'")
    
    # 3. 配置每种消费类目的生成规则 (权重, 金额范围, 备注示例)
    # frequency: 'daily'(每天都有), 'weekly'(每周几次), 'monthly'(每月一次), 'random'(随机出现)
    rules = {
        "三餐": {"freq": "daily", "count_range": (2, 3), "amount_range": (15, 60), "remarks": ["工作餐", "麦当劳", "面条", "盖浇饭", "路边摊"]},
        "零食": {"freq": "weekly", "count_range": (1, 3), "amount_range": (10, 40), "remarks": ["买饮料", "便利店零食", "水果", "咖啡", "奶茶"]},
        "交通": {"freq": "weekly", "count_range": (2, 5), "amount_range": (3, 30), "remarks": ["地铁充值", "打车", "共享单车", "公交车"]},
        "日用品": {"freq": "weekly", "count_range": (1, 2), "amount_range": (15, 80), "remarks": ["纸巾", "洗发水", "超市采购", "垃圾袋"]},
        "烟酒": {"freq": "weekly", "count_range": (1, 2), "amount_range": (30, 150), "remarks": ["买包烟", "便利店买酒", "聚会带酒"]},
        "娱乐": {"freq": "weekly", "count_range": (1, 2), "amount_range": (30, 200), "remarks": ["看电影", "抓娃娃", "网吧赛高", "密室逃脱"]},
        
        "住房": {"freq": "monthly", "day": 1, "amount_range": (2500, 4000), "remarks": ["房租支出", "房贷扣款"]},
        "水电煤": {"freq": "monthly", "day": 5, "amount_range": (120, 350), "remarks": ["本月电费", "水费煤气费结算"]},
        "话费网费": {"freq": "monthly", "day": 10, "amount_range": (50, 128), "remarks": ["手机话费充值", "宽带续费"]},
        "学习": {"freq": "monthly", "day": 15, "amount_range": (50, 300), "remarks": ["购买技术书籍", "网络教程", "软件订阅费"]},
        "宠物": {"freq": "monthly", "day": 20, "amount_range": (100, 500), "remarks": ["买猫粮狗粮", "宠物零食", "宠物玩具"]},

        "衣服": {"freq": "random", "yearly_count": 15, "amount_range": (100, 600), "remarks": ["网购外套", "买双鞋", "优衣库", "买件T恤"]},
        "医疗": {"freq": "random", "yearly_count": 6, "amount_range": (20, 400), "remarks": ["药店买感冒药", "医院挂号", "常规体检", "牙医"]},
        "发红包": {"freq": "random", "yearly_count": 8, "amount_range": (50, 500), "remarks": ["过年红包", "朋友生日", "群里发红包"]},
        "请客送礼": {"freq": "random", "yearly_count": 12, "amount_range": (200, 800), "remarks": ["请客吃饭", "朋友结婚份子钱", "节日送礼"]},
        "电器数码": {"freq": "random", "yearly_count": 8, "amount_range": (299, 5999), "remarks": ["更换机械键盘", "买个固态硬盘", "换手机", "无线耳机"]},
        "运动": {"freq": "random", "yearly_count": 20, "amount_range": (15, 150), "remarks": ["羽毛球场地费", "游泳", "买件运动服", "功能饮料"]},
        "旅行": {"freq": "random", "yearly_count": 3, "amount_range": (1500, 5000), "remarks": ["五一出游", "十一长假旅行", "周末周边游"]},
        "其它": {"freq": "random", "yearly_count": 24, "amount_range": (5, 100), "remarks": ["随手花", "不知道丢哪了", "小物件"]},
    }

    start_date = datetime(2025, 1, 1)
    end_date = datetime(2025, 12, 31)
    current_date = start_date

    transactions_batch = []
    total_records = 0

    # 4. 开始逐日循环模拟
    while current_date <= end_date:
        date_str = current_date.Format("%Y-%m-%d") if hasattr(current_date, 'Format') else current_date.strftime("%Y-%m-%d")
        
        for name, rule in rules.items():
            if name not in category_map:
                continue
            cid = category_map[name]
            
            # 判断今天是否需要生成该类目的账单
            should_generate = False
            count = 1
            
            if rule["freq"] == "daily":
                should_generate = True
                count = random.randint(*rule["count_range"])
            elif rule["freq"] == "weekly":
                # 每周概率触发
                if random.random() < (random.randint(*rule["count_range"]) / 7.0):
                    should_generate = True
            elif rule["freq"] == "monthly":
                # 每月固定某一天
                if current_date.day == rule["day"]:
                    should_generate = True
            elif rule["freq"] == "random":
                # 全年总数均匀分摊到365天的概率
                if random.random() < (rule["yearly_count"] / 365.0):
                    should_generate = True

            if should_generate:
                for _ in range(count):
                    amount = round(random.uniform(*rule["amount_range"]), 2)
                    remark = random.choice(rule["remarks"])
                    
                    # 随机生成一个记账的小时和分钟 (主要为了测试日视图的时段柱状图)
                    # 绝大多数消费发生在 07 点到 23 点之间
                    hour = random.randint(7, 23)
                    minute = random.randint(0, 59)
                    second = random.randint(0, 59)
                    created_at_str = f"{date_str} {hour:02d}:{minute:02d}:{second:02d}"
                    
                    raw_text = f"模拟数据：{remark} 花了 {amount}"
                    
                    # 组合一条完整的账单数据
                    transactions_batch.append((
                        amount, 
                        'expense', 
                        1, # account_id = 1 (默认账本)
                        cid, 
                        date_str, 
                        raw_text, 
                        remark, 
                        created_at_str
                    ))
                    total_records += 1

        current_date += timedelta(days=1)

    # 5. 批量写入数据库，极大提高写入性能
    cursor.executemany("""
        INSERT INTO transactions (amount, type, account_id, category_id, transaction_date, raw_text, remark, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    """, transactions_batch)
    
    conn.commit()
    conn.close()
    print(f"✅ 成功！已向 2025 年静默注入了 {total_records} 条结构化消费记录。")

if __name__ == "__main__":
    generate_2025_data()