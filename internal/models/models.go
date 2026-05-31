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