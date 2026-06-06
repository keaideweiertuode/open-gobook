package db

import (
	"database/sql"
	"fmt"
	"gobook/internal/models"

	_ "modernc.org/sqlite" 
)

var db *sql.DB

func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA cache_size=-8000;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("设置 PRAGMA 失败: %w", err)
		}
	}

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			balance REAL NOT NULL DEFAULT 0.0,
			is_active INTEGER NOT NULL DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 1
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount REAL NOT NULL,
			type TEXT NOT NULL,
			account_id INTEGER NOT NULL,
			category_id INTEGER,
			transaction_date DATE NOT NULL,
			raw_text TEXT,
			remark TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tx_date ON transactions(transaction_date);`,
		`CREATE INDEX IF NOT EXISTS idx_tx_category ON transactions(category_id);`,
	}

	for _, query := range schemas {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("建表失败: %w", err)
		}
	}

	_, _ = db.Exec("ALTER TABLE transactions ADD COLUMN deleted_at DATETIME")

	initDefaultData()
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func initDefaultData() {
	_, _ = db.Exec("INSERT OR IGNORE INTO accounts (id, name, balance) VALUES (1, '默认账本', 0.0)")

	expenseCategories := []string{
		"三餐", "零食", "衣服", "交通", "旅行", "孩子", "宠物", "话费网费", "烟酒", "学习",
		"日用品", "住房", "美妆", "医疗", "发红包", "汽车", "娱乐", "请客送礼", "电器数码", "运动",
		"其它", "水电煤", "办公", "借出", "还款",
	}
	for _, cat := range expenseCategories {
		_, _ = db.Exec("INSERT OR IGNORE INTO categories (name, type) VALUES (?, 'expense')", cat)
	}

	incomeCategories := []string{"工资", "生活费", "收红包", "股票基金", "借入", "收回", "其他"}
	for _, cat := range incomeCategories {
		_, _ = db.Exec("INSERT OR IGNORE INTO categories (name, type) VALUES (?, 'income')", cat)
	}
}

func SaveTransaction(tx *models.Transaction) error {
	query := `INSERT INTO transactions (amount, type, account_id, category_id, transaction_date, raw_text, remark) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, tx.Amount, tx.Type, tx.AccountID, tx.CategoryID, tx.TransactionDate.Format("2006-01-02"), tx.RawText, tx.Remark)
	return err
}

// 【新增】根据主键 ID 瞬间抹除单条账单记录
func DeleteTransaction(id int) error {
	result, err := db.Exec("UPDATE transactions SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("账单 ID [%d] 不存在", id)
	}
	return nil
}

// 【新增】恢复被软删除的记录
func RecoverTransaction(id int) error {
	result, err := db.Exec("UPDATE transactions SET deleted_at = NULL WHERE id = ? AND deleted_at IS NOT NULL", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("回收站中找不到账单 ID [%d]", id)
	}
	return nil
}



// 【新增】获取指定时间区间内所有的流水明细列表（支持年月日降序排列，支持分页和分类筛选）
func QueryTransactions(period, category string, page, pageSize int) ([]models.TransactionDetail, int, error) {
	var countQuery string
	var query string
	var args []any
	var countArgs []any

	if len(period) == 10 {
		countQuery = `SELECT count(*) FROM transactions t JOIN categories c ON t.category_id = c.id WHERE t.transaction_date = ? AND t.deleted_at IS NULL`
		query = `SELECT t.id, t.amount, t.type, c.name, t.transaction_date, t.remark 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE t.transaction_date = ? AND t.deleted_at IS NULL`
		args = append(args, period)
		countArgs = append(countArgs, period)
	} else if len(period) == 7 {
		countQuery = `SELECT count(*) FROM transactions t JOIN categories c ON t.category_id = c.id WHERE strftime('%Y-%m', t.transaction_date) = ? AND t.deleted_at IS NULL`
		query = `SELECT t.id, t.amount, t.type, c.name, t.transaction_date, t.remark 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE strftime('%Y-%m', t.transaction_date) = ? AND t.deleted_at IS NULL`
		args = append(args, period)
		countArgs = append(countArgs, period)
	} else if len(period) == 4 {
		countQuery = `SELECT count(*) FROM transactions t JOIN categories c ON t.category_id = c.id WHERE strftime('%Y', t.transaction_date) = ? AND t.deleted_at IS NULL`
		query = `SELECT t.id, t.amount, t.type, c.name, t.transaction_date, t.remark 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE strftime('%Y', t.transaction_date) = ? AND t.deleted_at IS NULL`
		args = append(args, period)
		countArgs = append(countArgs, period)
	} else {
		countQuery = `SELECT count(*) FROM transactions t JOIN categories c ON t.category_id = c.id WHERE t.deleted_at IS NULL`
		query = `SELECT t.id, t.amount, t.type, c.name, t.transaction_date, t.remark 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE t.deleted_at IS NULL`
	}

	if category != "" {
		countQuery += ` AND c.name = ?`
		query += ` AND c.name = ?`
		args = append(args, category)
		countArgs = append(countArgs, category)
	}

	query += ` ORDER BY t.transaction_date DESC, t.id DESC`

	var total int
	if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if pageSize > 0 {
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []models.TransactionDetail
	for rows.Next() {
		var t models.TransactionDetail
		if err := rows.Scan(&t.ID, &t.Amount, &t.Type, &t.CategoryName, &t.TransactionDate, &t.Remark); err != nil {
			return nil, 0, err
		}
		list = append(list, t)
	}
	return list, total, nil
}

func GetCategoryDetailsByName(name string) (int, string, error) {
	var id int
	var catType string
	err := db.QueryRow("SELECT id, type FROM categories WHERE name = ?", name).Scan(&id, &catType)
	if err != nil {
		_ = db.QueryRow("SELECT id, type FROM categories WHERE name = '其它'").Scan(&id, &catType)
		return id, catType, nil
	}
	return id, catType, nil
}

// QueryTrendStats 核心升级：同时计算各时间段的支出与收入总和
func QueryTrendStats(period string) ([]models.TrendStat, error) {
	var query string

	if len(period) == 10 {
		// 按天复盘
		query = `SELECT strftime('%H', created_at, 'localtime') || '点' AS date,
					ROUND(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 2) AS expense,
					ROUND(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 2) AS income
				 FROM transactions WHERE transaction_date = ? AND deleted_at IS NULL
				 GROUP BY strftime('%H', created_at, 'localtime') ORDER BY date ASC`
	} else if len(period) == 7 {
		// 按月复盘
		// 【关键修复点】：强制使用 strftime('%Y-%m-%d') 输出纯文本格式，防止 Go 驱动带上 T00:00:00Z 尾巴
		query = `SELECT strftime('%Y-%m-%d', transaction_date) AS date,
					ROUND(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 2) AS expense,
					ROUND(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 2) AS income
				 FROM transactions WHERE strftime('%Y-%m', transaction_date) = ? AND deleted_at IS NULL
				 GROUP BY strftime('%Y-%m-%d', transaction_date) ORDER BY date ASC`
	} else {
		// 按年复盘
		query = `SELECT strftime('%Y-%m', transaction_date) AS date,
					ROUND(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 2) AS expense,
					ROUND(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 2) AS income
				 FROM transactions WHERE strftime('%Y', transaction_date) = ? AND deleted_at IS NULL
				 GROUP BY strftime('%Y-%m', transaction_date) ORDER BY date ASC`
	}

	rows, err := db.Query(query, period)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.TrendStat
	for rows.Next() {
		var s models.TrendStat
		if err := rows.Scan(&s.Date, &s.Expense, &s.Income); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func QueryCategoryStats(period string) ([]models.CategoryStat, error) {
	var query string
	if len(period) == 10 {
		query = `SELECT c.name, ROUND(SUM(t.amount), 2) as total 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE t.type = 'expense' AND t.transaction_date = ? AND t.deleted_at IS NULL
				 GROUP BY t.category_id ORDER BY total DESC`
	} else if len(period) == 7 {
		query = `SELECT c.name, ROUND(SUM(t.amount), 2) as total 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE t.type = 'expense' AND strftime('%Y-%m', t.transaction_date) = ? AND t.deleted_at IS NULL
				 GROUP BY t.category_id ORDER BY total DESC`
	} else {
		query = `SELECT c.name, ROUND(SUM(t.amount), 2) as total 
				 FROM transactions t JOIN categories c ON t.category_id = c.id 
				 WHERE t.type = 'expense' AND strftime('%Y', t.transaction_date) = ? AND t.deleted_at IS NULL
				 GROUP BY t.category_id ORDER BY total DESC`
	}
	rows, err := db.Query(query, period)
	if err != nil { return nil, err }
	defer rows.Close()
	var stats []models.CategoryStat
	for rows.Next() {
		var s models.CategoryStat
		if err := rows.Scan(&s.CategoryName, &s.TotalAmount); err != nil { return nil, err }
		stats = append(stats, s)
	}
	return stats, nil
}

func QueryAvailableYears() ([]string, error) {
	query := `SELECT DISTINCT strftime('%Y', transaction_date) AS year FROM transactions WHERE deleted_at IS NULL UNION SELECT strftime('%Y', 'now') ORDER BY year ASC`
	rows, err := db.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()
	var years []string
	for rows.Next() {
		var y string
		if err := rows.Scan(&y); err != nil { return nil, err }
		years = append(years, y)
	}
	return years, nil
}