package models

import "time"

// Category 分类结构
type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // expense / income
	IsActive int    `json:"is_active"`
}

// Transaction 账单明细
type Transaction struct {
	ID              int       `json:"id"`
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"` // expense / income
	AccountID       int       `json:"account_id"`
	CategoryID      int       `json:"category_id"`
	TransactionDate time.Time `json:"transaction_date"`
	RawText         string    `json:"raw_text"` 
	Remark          string    `json:"remark"`   
	CreatedAt       time.Time `json:"created_at"`
}

// TransactionDetail 用于前端表格展示的结构体
type TransactionDetail struct {
	ID              int     `json:"id"`
	Amount          float64 `json:"amount"`
	Type            string  `json:"type"`
	CategoryName    string  `json:"category_name"`
	TransactionDate string  `json:"transaction_date"`
	Remark          string  `json:"remark"`
}

// TrendStat 双轨收支统计结构体
type TrendStat struct {
	Date    string  `json:"date"`    // 时间轴刻度
	Expense float64 `json:"expense"` // 支出金额
	Income  float64 `json:"income"`  // 收入金额
}

// CategoryStat 分类支出排行结构体
type CategoryStat struct {
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
}

// DebtStat 借贷/债权追踪结构体
type DebtStat struct {
	TotalDebt        float64 `json:"total_debt"`        // 我的净负债 (借入 - 还款)
	TotalReceivables float64 `json:"total_receivables"` // 待收净债权 (借出 - 收回)
	BorrowedIn       float64 `json:"borrowed_in"`       // 累计借入
	Repaid           float64 `json:"repaid"`            // 累计还款
	LentOut          float64 `json:"lent_out"`          // 累计借出
	Recovered        float64 `json:"recovered"`         // 累计收回
}